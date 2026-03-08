# 数据库与 DTO 文档

本文档详细描述 BaiZe 监控平台的数据库表结构设计和 DTO（Data Transfer Object）数据结构。

## 目录

### 📊 数据库表结构

#### 核心业务表
- [告警表 (alerts)](#告警表-alerts) - 存储 BMC Trap 告警记录
- [BMC Trap 解析器表 (bmc_trap_parsers)](#bmc-trap-解析器表-bmc_trap_parsers) - 存储 Trap 解析器配置

#### 硬件信息采集表
- [hardware_info](#hardware_info-表) - 硬件信息主表（完整快照）
- [hardware_cpu](#hardware_cpu-表) - CPU 摘要信息
- [hardware_memory](#hardware_memory-表) - 内存摘要信息
- [hardware_disk](#hardware_disk-表) - 磁盘摘要信息
- [hardware_network_interface](#hardware_network_interface-表) - 网络接口摘要信息

#### 基础信息表
- [机器信息表 (machine_infos)](#机器信息表-machine_infos) - 机器基本信息
- [异常检测结果表 (anomaly_results)](#异常检测结果表-anomaly_results) - 异常检测结果
- [用户表 (users)](#用户表-users) - 用户信息和权限

### 📦 DTO 数据结构

#### 告警相关 DTO
- [AlertFilter](#alertfilter-请求) - 告警查询过滤条件
- [AlertUpdate](#alertupdate-请求) - 告警状态更新
- [AlertListResult](#alertlistresult-响应) - 告警列表查询结果

#### BMC Trap 解析器 DTO
- [BMCTrapParserCreate](#bmctrapparsercreate-请求) - 创建解析器
- [BMCTrapParserUpdate](#bmctrapparserupdate-请求) - 更新解析器
- [BMCTrapParserFilter](#bmctrapparserfilter-请求) - 解析器查询
- [BMCTrapParserResponse](#bmctrapparserresponse-响应) - 解析器详情
- [BMCTrapParserListResult](#bmctrapparserlistresult-响应) - 解析器列表

#### 用户相关 DTO
- [LoginRequest](#loginrequest-请求) - 登录请求
- [CreateUserRequest](#createuserrequest-请求) - 创建用户
- [LoginResponse](#loginresponse-响应) - 登录响应
- [UserResponse](#userresponse-响应) - 用户信息
- [UsersListResponse](#userslistresponse-响应) - 用户列表

#### 硬件信息 DTO
- [HardwareInfoRequest](#hardwareinforequest-请求) - 硬件信息采集请求
- [HardwareInfoResponse](#hardwareinforesponse-响应) - 硬件信息响应
- [CPURequest](#cpurequest-请求) / [CPUResponse](#cpuresponse-响应) - CPU 信息
- [MemoryRequest](#memoryrequest-请求) / [MemoryResponse](#memoryresponse-响应) - 内存信息
- [DiskRequest](#diskrequest-请求) / [DiskResponse](#diskresponse-响应) - 磁盘信息
- [NetworkInterfaceRequest](#networkinterfacerequest-请求) / [NetworkInterfaceResponse](#networkinterfaceresponse-响应) - 网络接口信息
- [HardwareSummaryResponse](#hardwaresummaryresponse-响应) - 硬件摘要

#### 通用响应 DTO
- [SuccessResponse](#successresponse) - 成功响应
- [ErrorResponse](#errorresponse) - 错误响应

### 📚 设计原则与最佳实践

- [数据模型关系](#数据模型关系) - 表关系图
- [设计原则](#设计原则) - 数据库、DTO、索引设计原则
- [最佳实践](#最佳实践) - 查询优化、数据一致性、扩展性考虑

---

[🔝 返回顶部](#目录)

## 数据库表结构

### 告警表 (alerts)

存储 BMC Trap 告警记录，支持告警的完整生命周期管理。

#### 表结构

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 告警 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `parser_id` | `BIGINT` | `INDEX` | 解析器 ID（关联 bmc_trap_parsers.id） |
| `alert_status` | `VARCHAR(100)` | `INDEX` | 告警状态：ACTIVE/CLEARED |
| `trap_oid` | `VARCHAR(200)` | `INDEX` | Trap OID（告警类型标识） |
| `source_ip` | `VARCHAR(45)` | `INDEX` | 告警源 IP 地址 |
| `source_type` | `VARCHAR(100)` | `INDEX` | 告警源类型：BMC |
| `vendor_code` | `VARCHAR(50)` | `INDEX` | 厂商标识代码（企业 OID 数字部分） |
| `vendor_name` | `VARCHAR(100)` | | 厂商名称 |
| `alert_level` | `VARCHAR(100)` | `INDEX` | 告警级别：critical/warning/info/notification |
| `alert_time` | `TIMESTAMP` | `INDEX` | 告警发生时间 |
| `component` | `VARCHAR(100)` | `INDEX` | 告警组件：CPU/MEMORY/STORAGE 等 |
| `content` | `TEXT` | | 告警详细内容 |
| `raw_data` | `TEXT` | | 原始 Trap 数据（JSON 格式） |
| `trap_raw_time` | `VARCHAR(100)` | | 原始告警时间字符串 |
| `variable_map` | `JSONB` | `NOT NULL` | OID 到值的映射关系 |
| `enable_auto_close` | `BOOLEAN` | | 是否启用自动关闭 |
| `trap_index` | `VARCHAR(200)` | `INDEX` | Trap 索引 OID（用于告警关联） |
| `trap_status` | `VARCHAR(200)` | `INDEX` | Trap 状态 OID：Asserted/Deasserted |
| `enable_contact_inter_component_alerts` | `BOOLEAN` | | 是否启用组件间关联告警 |
| `identifier_of_the_same_component` | `VARCHAR(200)` | `INDEX` | 同一组件标识符 OID |

#### 表说明

**关联关系**
- `parser_id`: 关联 `bmc_trap_parsers.id`，使用哪个解析器解析的 Trap
- `variable_map`: JSONB 格式存储所有 OID 到值的映射，便于扩展

**枚举值说明**

**alert_status**
- `ACTIVE`: 活动告警
- `CLEARED`: 已清除告警

**alert_level**
- `critical`: 严重告警
- `warning`: 警告告警
- `info`: 信息告警
- `notification`: 通知告警

**trap_status**
- `Asserted`: 告警触发
- `Deasserted`: 告警恢复

**source_type**
- `BMC`: Baseboard Management Controller

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构)

---

### BMC Trap 解析器表 (bmc_trap_parsers)

存储 BMC Trap 解析器配置，支持多厂商、多型号服务器的 Trap 解析配置。

#### 表结构

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 解析器 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `parser_name` | `VARCHAR(200)` | `INDEX, NOT NULL` | 解析器名称（唯一） |
| `vendor_code` | `VARCHAR(100)` | `INDEX, NOT NULL` | 厂商代码（企业 OID 数字部分） |
| `vendor_name` | `VARCHAR(200)` | `INDEX, NOT NULL` | 厂商名称 |
| `alert_level_oid` | `VARCHAR(500)` | `NOT NULL` | 告警级别 OID |
| `alert_content_oid` | `VARCHAR(500)` | `NOT NULL` | 告警内容 OID |
| `alert_time_oid` | `VARCHAR(500)` | `NOT NULL` | 告警时间 OID |
| `alert_component_oid` | `VARCHAR(500)` | `NOT NULL` | 告警组件 OID |
| `enable_auto_close` | `BOOLEAN` | `DEFAULT FALSE` | 是否启用自动关闭 |
| `alert_index_oid` | `VARCHAR(500)` | | 告警索引 OID（自动关闭用） |
| `alert_status_oid` | `VARCHAR(500)` | | 告警状态 OID（自动关闭用） |
| `enable_contact_inter_component_alerts` | `BOOLEAN` | `DEFAULT FALSE` | 是否启用组件间关联 |
| `contact_inter_component_identifier_oid` | `VARCHAR(500)` | | 组件标识符 OID |
| `level_mappings` | `JSONB` | `NOT NULL` | 告警级别映射配置 |
| `status_mappings` | `JSONB` | `NOT NULL` | 告警状态映射配置 |
| `enable_product_name_list` | `JSONB` | `NOT NULL` | 启用的产品名称列表 |
| `enable_host_name_list` | `JSONB` | `NOT NULL` | 启用的主机名列表 |
| `component_mappings` | `JSONB` | `NOT NULL` | 组件映射配置（关键词->组件） |
| `description` | `TEXT` | | 解析器描述 |
| `is_active` | `BOOLEAN` | `DEFAULT TRUE` | 是否激活 |
| `is_deleted` | `BOOLEAN` | `DEFAULT FALSE` | 是否删除 |

#### 表说明

**JSONB 字段结构**

**level_mappings**
```json
{
  "1.3.6.1.4.1.2011.2.12.3.8.1.1.4.1.3": "critical",
  "1.3.6.1.4.1.2011.2.12.3.8.1.1.4.1.4": "warning",
  "1.3.6.1.4.1.2011.2.12.3.8.1.1.4.1.5": "info"
}
```

**status_mappings**
```json
{
  "99": "Asserted",
  "100": "Deasserted"
}
```

**component_mappings**
```json
{
  "CPU": ["processor", "cpu", "core"],
  "MEMORY": ["memory", "dimm", "ram"],
  "STORAGE": ["disk", "hdd", "ssd", "nvme"]
}
```

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构)

---

### 硬件信息表

存储 Agent 采集的详细硬件信息，采用主表 + 摘要表的设计，支持硬件信息的完整快照和快速摘要查询。

#### hardware_info 表

硬件信息主表，存储完整的硬件配置快照。

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `INET` | `INDEX, NOT NULL` | 目标 IP 地址（PostgreSQL inet 类型） |
| `collected_at` | `TIMESTAMP` | `NOT NULL` | 硬件信息收集时间 |
| `hardware_info_snapshot` | `JSONB` | `NOT NULL` | 硬件信息完整快照（包含 CPU、内存、磁盘、网络） |
| `version` | `VARCHAR(50)` | `NOT NULL` | 数据格式版本 |
| `checksum` | `VARCHAR(128)` | `NOT NULL` | 数据校验和

#### 表说明

**JSONB 字段结构 (hardware_info_snapshot)**

```json
{
  "collected_at": "2024-01-01T10:00:00Z",
  "cpu": {
    "collected_at": "2024-01-01T10:00:00Z",
    "success": true,
    "message": "collected successfully",
    "cpus": [
      {
        "id": 0,
        "model_name": "Intel(R) Xeon(R) Platinum 8358 CPU",
        "vendor_id": "GenuineIntel",
        "family": "6",
        "model": "106",
        "stepping": "6",
        "physical_id": "0",
        "core_id": "0",
        "cores": 32,
        "threads": 64,
        "mhz": 2600.0
      }
    ]
  },
  "memory": {
    "collected_at": "2024-01-01T10:00:00Z",
    "success": true,
    "message": "collected successfully",
    "summary": {
      "total_size": 274877906944
    },
    "content": [
      {
        "locator": "DIMM_A1",
        "size": 34359738368,
        "type": "DDR4",
        "speed": "3200 MHz",
        "manufacturer": "Samsung"
      }
    ]
  },
  "disks": {
    "collected_at": "2024-01-01T10:00:00Z",
    "success": true,
    "message": "collected successfully",
    "disks": [
      {
        "name": "sda",
        "size": 960170500096,
        "type": "ssd",
        "serial": "S61CNB0KB00000"
      }
    ],
    "partitions": [
      {
        "device": "/dev/sda1",
        "mountpoint": "/",
        "size": 960170500096
      }
    ]
  },
  "network_interfaces": {
    "collected_at": "2024-01-01T10:00:00Z",
    "success": true,
    "message": "collected successfully",
    "interfaces": [
      {
        "name": "eth0",
        "mac": "00:11:22:33:44:55",
        "ips": ["192.168.1.100"],
        "mtu": 1500,
        "speed": 10000
      }
    ]
  }
}
```

---

#### hardware_cpu 表

CPU 摘要信息表，存储 CPU 核心数等摘要信息，便于快速查询。

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `INET` | `INDEX, NOT NULL` | 目标 IP 地址 |
| `success` | `BOOLEAN` | `NOT NULL` | 采集是否成功 |
| `message` | `TEXT` | | 采集消息（失败时记录错误） |
| `total_cores` | `INTEGER` | `NOT NULL` | 总物理核心数 |
| `total_threads` | `INTEGER` | `NOT NULL` | 总逻辑线程数 |

#### 表说明

GORM 索引定义：`target_ip` 字段使用 `gorm:"index"` tag 自动创建索引

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构) | [⬆️ 返回硬件信息表](#硬件信息表)

---

#### hardware_memory 表

内存摘要信息表，存储内存容量、类型等摘要信息。

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `INET` | `INDEX, NOT NULL` | 目标 IP 地址 |
| `success` | `BOOLEAN` | `NOT NULL` | 采集是否成功 |
| `message` | `TEXT` | | 采集消息（失败时记录错误） |
| `total_size` | `BIGINT` | | 总容量（字节） |
| `type` | `VARCHAR(100)` | | 内存类型（DDR4、DDR5 等） |
| `speed` | `VARCHAR(100)` | | 内存频率 |
| `total_slots` | `INTEGER` | | 插槽总数 |
| `used_slots` | `INTEGER` | | 已用插槽数 |

#### 表说明

GORM 索引定义：`target_ip` 和 `type` 字段使用 `gorm:"index"` tag 自动创建索引

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构) | [⬆️ 返回硬件信息表](#硬件信息表)

---

#### hardware_disk 表

磁盘摘要信息表，存储磁盘数量、容量等摘要信息。

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `INET` | `INDEX, NOT NULL` | 目标 IP 地址 |
| `success` | `BOOLEAN` | `NOT NULL` | 采集是否成功 |
| `message` | `TEXT` | | 采集消息（失败时记录错误） |
| `total_count` | `INTEGER` | `NOT NULL` | 磁盘总数 |
| `total_size` | `BIGINT` | `NOT NULL` | 总容量（字节） |
| `ssd_count` | `INTEGER` | `NOT NULL` | SSD 数量 |
| `hdd_count` | `INTEGER` | `NOT NULL` | HDD 数量 |
| `total_partitions` | `INTEGER` | `NOT NULL` | 分区总数 |

#### 表说明

GORM 索引定义：`target_ip` 字段使用 `gorm:"index"` tag 自动创建索引

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构) | [⬆️ 返回硬件信息表](#硬件信息表)

---

#### hardware_network_interface 表

网络接口摘要信息表，存储网络接口数量等摘要信息。

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `INET` | `INDEX, NOT NULL` | 目标 IP 地址 |
| `success` | `BOOLEAN` | `NOT NULL` | 采集是否成功 |
| `message` | `TEXT` | | 采集消息（失败时记录错误） |
| `total_count` | `INTEGER` | | 接口总数 |
| `physical_count` | `INTEGER` | | 物理接口数 |
| `virtual_count` | `INTEGER` | | 虚拟接口数 |
| `active_count` | `INTEGER` | | 活跃接口数

#### 表说明

GORM 索引定义：`target_ip` 字段使用 `gorm:"index"` tag 自动创建索引

[🔝 返回顶部](#目录) | [📊 返回数据库表结构](#数据库表结构)

---

### 机器信息表 (machine_infos)

存储 Agent 采集的机器基本信息，用于设备管理和识别。

#### 表结构

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 记录 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `collected_at` | `TIMESTAMP` | `NOT NULL` | 采集时间 |
| `success` | `BOOLEAN` | `NOT NULL` | 采集是否成功 |
| `message` | `TEXT` | | 采集消息（失败时记录错误） |
| `content` | `JSONB` | `NOT NULL` | 机器信息内容 |

#### 表说明

**JSONB 字段结构 (content)**

```json
{
  "hostname": "server-001",
  "ip_addresses": {
    "eth0": ["192.168.1.100", "fe80::1"],
    "eth1": ["192.168.2.100"]
  }
}
```

---

### 异常检测结果表 (anomaly_results)

存储 Agent 端异常检测结果，支持基于规则的异常检测。

#### 表结构

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 检测结果 ID |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `target_ip` | `VARCHAR(45)` | `INDEX, NOT NULL` | 目标 IP 地址 |
| `check_type` | `VARCHAR(50)` | `INDEX, NOT NULL` | 检测类型：cpu/memory/disk/network |
| `check_item` | `VARCHAR(100)` | `INDEX, NOT NULL` | 检测项目 |
| `success` | `BOOLEAN` | `INDEX, NOT NULL` | 检测是否成功 |
| `level` | `VARCHAR(50)` | `INDEX` | 异常级别：info/warning/critical |
| `message` | `TEXT` | | 异常消息 |
| `extra_data` | `JSONB` | `NOT NULL` | 额外数据 |
| `checked_at` | `TIMESTAMP` | `INDEX` | 检测时间 |

#### 表说明

**枚举值说明**

**check_type**
- `cpu`: CPU 相关检测
- `memory`: 内存相关检测
- `disk`: 磁盘相关检测
- `network`: 网络相关检测
- `process`: 进程相关检测

**level**
- `info`: 信息级别
- `warning`: 警告级别
- `critical`: 严重级别

---

### 用户表 (users)

存储系统用户信息和权限配置。

#### 表结构

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` | `PRIMARY KEY` | 用户 ID |
| `username` | `VARCHAR(255)` | `UNIQUE, NOT NULL` | 用户名 |
| `password_hash` | `VARCHAR(255)` | `NOT NULL` | 密码哈希 |
| `is_admin` | `BOOLEAN` | `DEFAULT FALSE` | 是否管理员 |
| `is_active` | `BOOLEAN` | `DEFAULT TRUE` | 是否激活 |
| `is_deleted` | `BOOLEAN` | `DEFAULT FALSE` | 是否删除 |
| `created_at` | `TIMESTAMP` | `NOT NULL` | 创建时间 |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | 更新时间 |
| `permissions` | `JSONB` | `NOT NULL` | 权限配置

#### 表说明

**JSONB 字段结构 (permissions)**

```json
{
  "alert:view": true,
  "alert:manage": true,
  "parser:create": true,
  "parser:edit": true,
  "parser:delete": false,
  "user:manage": false
}
```

---

## DTO 数据结构

[📋 返回目录](#目录)

### 告警相关 DTO

#### AlertFilter (请求)

告警列表查询过滤条件。

```go
type AlertFilter struct {
    Page     int `json:"page" binding:"min=1" default:"1"`
    PageSize int `json:"page_size" binding:"min=1,max=100" default:"10"`
    
    // 查询条件（指针类型，nil 表示不筛选）
    ParserID            *int64     `json:"parser_id,omitempty"`
    Status              *string    `json:"alert_status,omitempty"`
    TrapOID             *string    `json:"trap_oid,omitempty"`
    TrapStatus          *string    `json:"trap_status,omitempty"`
    TrapIndex           *string    `json:"trap_index,omitempty"`
    SourceIP            *string    `json:"source_ip,omitempty"`
    VendorCode          *string    `json:"vendor_code,omitempty"`
    VendorName          *string    `json:"vendor_name,omitempty"`
    AlertLevel          *string    `json:"alert_level,omitempty"`
    AlertTimeRangeStart *time.Time `json:"alert_time_range_start,omitempty"`
    AlertTimeRangeEnd   *time.Time `json:"alert_time_range_end,omitempty"`
    Component           *string    `json:"component,omitempty"`
    Content             *string    `json:"content,omitempty"`
}
```

#### AlertUpdate (请求)

更新告警状态。

```go
type AlertUpdate struct {
    ID     int64              `json:"id"`
    Status models.AlertStatus `json:"alert_status"`
}
```

#### AlertListResult (响应)

告警列表查询结果。

```go
type AlertListResult struct {
    List     []*models.Alert `json:"list"`
    Total    int64           `json:"total"`
    Page     int             `json:"page"`
    PageSize int             `json:"page_size"`
}
```

[🔝 返回顶部](#目录) | [📦 返回 DTO 数据结构](#dto-数据结构)

---

### BMC Trap 解析器 DTO

#### BMCTrapParserCreate (请求)

创建解析器配置。

```go
type BMCTrapParserCreate struct {
    ParserName string `json:"parser_name" binding:"required,min=1,max=200"`
    VendorName string `json:"vendor_name" binding:"required,min=1,max=200"`
    
    // 核心 OID 字段
    AlertLevelOID     string `json:"alert_level_oid" binding:"required,min=5,max=500"`
    AlertContentOID   string `json:"alert_content_oid" binding:"required,min=5,max=500"`
    AlertTimeOID      string `json:"alert_time_oid" binding:"required,min=5,max=500"`
    AlertComponentOID string `json:"alert_component_oid" binding:"required,min=5,max=500"`
    
    // 自动关闭配置
    EnableAutoClose bool   `json:"enable_auto_close"`
    AlertIndexOID   string `json:"alert_index_oid" binding:"required_if=EnableAutoClose true,max=500"`
    AlertStatusOID  string `json:"alert_status_oid" binding:"required_if=EnableAutoClose true,max=500"`
    
    // 组件间关联配置
    EnableContactInterComponentAlerts  bool   `json:"enable_contact_inter_component_alerts"`
    ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid" binding:"required_if=EnableContactInterComponentAlerts true,max=500"`
    
    TimeFormat string `json:"time_format" binding:"required,max=200"`
    
    // 映射配置
    LevelMappings         map[string]string   `json:"level_mappings"`
    StatusMappings        map[string]string   `json:"status_mappings"`
    EnableProductNameList []string            `json:"enable_product_name_list"`
    EnableHostNameList    []string            `json:"enable_host_name_list"`
    ComponentMappings     map[string][]string `json:"component_mappings"`
    
    Description string `json:"description" binding:"max=1000"`
}
```

#### BMCTrapParserUpdate (请求)

更新解析器配置（所有字段可选）。

```go
type BMCTrapParserUpdate struct {
    ID int64 `json:"id" binding:"required"`
    
    ParserName *string `json:"parser_name,omitempty" binding:"omitempty,max=200"`
    VendorName *string `json:"vendor_name,omitempty" binding:"omitempty,max=200"`
    
    // 核心 OID 字段
    AlertLevelOID     *string `json:"alert_level_oid,omitempty" binding:"omitempty,max=500"`
    AlertContentOID   *string `json:"alert_content_oid,omitempty" binding:"omitempty,max=500"`
    AlertTimeOID      *string `json:"alert_time_oid,omitempty" binding:"omitempty,max=500"`
    AlertComponentOID *string `json:"alert_component_oid,omitempty" binding:"omitempty,max=500"`
    
    EnableAutoClose *bool   `json:"enable_auto_close,omitempty"`
    AlertIndexOID   *string `json:"alert_index_oid,omitempty" binding:"omitempty,max=500"`
    AlertStatusOID  *string `json:"alert_status_oid,omitempty" binding:"omitempty,max=500"`
    
    EnableContactInterComponentAlerts  *bool   `json:"enable_contact_inter_component_alerts,omitempty"`
    ContactInterComponentIdentifierOID *string `json:"contact_inter_component_identifier_oid,omitempty" binding:"omitempty,max=500"`
    
    TimeFormat *string `json:"time_format,omitempty" binding:"omitempty,max=200"`
    
    // 映射配置
    LevelMappings         *map[string]string   `json:"level_mappings,omitempty"`
    StatusMappings        *map[string]string   `json:"status_mappings,omitempty"`
    EnableProductNameList *[]string            `json:"enable_product_name_list,omitempty"`
    EnableHostNameList    *[]string            `json:"enable_host_name_list,omitempty"`
    ComponentMappings     *map[string][]string `json:"component_mappings,omitempty"`
    
    Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}
```

#### BMCTrapParserFilter (请求)

解析器列表查询过滤条件。

```go
type BMCTrapParserFilter struct {
    Page     int `json:"page" binding:"min=1" default:"1"`
    PageSize int `json:"page_size" binding:"min=1,max=100" default:"10"`
    
    ParserName  *string `json:"parser_name,omitempty"` // 模糊匹配
    VendorCode  *string `json:"vendor_code,omitempty"` // 精确匹配
    VendorName  *string `json:"vendor_name,omitempty"` // 模糊匹配
    IsActive    *bool   `json:"is_active,omitempty"`   // 精确匹配
    Description *string `json:"description,omitempty"` // 模糊匹配
}
```

#### BMCTrapParserResponse (响应)

解析器详情响应。

```go
type BMCTrapParserResponse struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    ParserName string `json:"parser_name"`
    VendorCode string `json:"vendor_code"`
    VendorName string `json:"vendor_name"`
    
    // 核心 OID 字段
    AlertLevelOID     string `json:"alert_level_oid"`
    AlertContentOID   string `json:"alert_content_oid"`
    AlertTimeOID      string `json:"alert_time_oid"`
    AlertComponentOID string `json:"alert_component_oid"`
    
    EnableAutoClose bool   `json:"enable_auto_close"`
    AlertIndexOID   string `json:"alert_index_oid,omitempty"`
    AlertStatusOID  string `json:"alert_status_oid,omitempty"`
    
    EnableContactInterComponentAlerts  bool   `json:"enable_contact_inter_component_alerts"`
    ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid,omitempty"`
    
    // 映射配置
    LevelMappings         map[string]string   `json:"level_mappings"`
    StatusMappings        map[string]string   `json:"status_mappings"`
    EnableProductNameList []string            `json:"enable_product_name_list"`
    EnableHostNameList    []string            `json:"enable_host_name_list"`
    ComponentMappings     map[string][]string `json:"component_mappings"`
    
    Description string `json:"description"`
    IsActive    bool   `json:"is_active"`
}
```

#### BMCTrapParserListResult (响应)

解析器列表查询结果。

```go
type BMCTrapParserListResult struct {
    List     []*BMCTrapParserResponse `json:"list"`
    Total    int64                    `json:"total"`
    Page     int                      `json:"page"`
    PageSize int                      `json:"page_size"`
}
```

[🔝 返回顶部](#目录) | [📦 返回 DTO 数据结构](#dto-数据结构)

---

### 用户相关 DTO

#### LoginRequest (请求)

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

#### CreateUserRequest (请求)

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

#### LoginResponse (响应)

```go
type LoginResponse struct {
    AccessToken  string          `json:"access_token"`
    RefreshToken string          `json:"refresh_token"`
    ID           int64           `json:"id"`
    Username     string          `json:"username"`
    IsAdmin      bool            `json:"is_admin"`
    Permissions  map[string]bool `json:"permissions"`
}
```

#### UserResponse (响应)

```go
type UserResponse struct {
    ID          int64           `json:"id"`
    Username    string          `json:"username"`
    IsAdmin     bool            `json:"is_admin"`
    IsActive    bool            `json:"is_active"`
    IsDeleted   bool            `json:"is_deleted"`
    CreatedAt   string          `json:"created_at"`
    UpdatedAt   string          `json:"updated_at"`
    Permissions map[string]bool `json:"permissions"`
}
```

#### UsersListResponse (响应)

```go
type UsersListResponse struct {
    Users      []UserResponse `json:"users"`
    Page       int            `json:"page"`
    PageSize   int            `json:"page_size"`
    Total      int64          `json:"total"`
    TotalPages int            `json:"total_pages"`
}
```

---

### 硬件信息 DTO

#### 硬件信息请求 DTO

**HardwareInfoRequest** - 硬件信息请求（总结构）

```go
type HardwareInfoRequest struct {
    CollectedAt       time.Time                 `json:"collected_at" binding:"required"`
    CPUs              CPURequest                `json:"cpu" binding:"required"`
    Memory            MemoryRequest             `json:"memory" binding:"required"`
    Disks             DiskRequest               `json:"disks" binding:"required"`
    NetworkInterfaces NetworkInterfaceRequest   `json:"network_interfaces" binding:"required"`
}
```

**CPURequest** - CPU 信息采集请求

```go
type CPURequest struct {
    Success bool      `json:"success"`
    Message string    `json:"message"`
    Content []CPUInfo `json:"content"`
    Summary CPUSummary `json:"summary"`
}

type CPUInfo struct {
    Model        string `json:"model" binding:"required"`
    Vendor       string `json:"vendor" binding:"required"`
    Family       string `json:"family"`
    Stepping     string `json:"stepping"`
    Flags        string `json:"flags"`
    Cores        int    `json:"cores" binding:"required"`
    Threads      int    `json:"threads" binding:"required"`
    BaseSpeed    string `json:"base_speed" binding:"required"`
    CacheSizeL1  string `json:"cache_size_l1"`
    CacheSizeL2  string `json:"cache_size_l2"`
    CacheSizeL3  string `json:"cache_size_l3"`
    Architecture string `json:"architecture"`
}

type CPUSummary struct {
    TotalCores   int `json:"total_cores"`
    TotalThreads int `json:"total_threads"`
}
```

**MemoryRequest** - 内存信息采集请求

```go
type MemoryRequest struct {
    Success bool             `json:"success"`
    Message string           `json:"message"`
    Content []MemoryModuleInfo `json:"content"`
    Summary MemorySummary    `json:"summary"`
}

type MemoryModuleInfo struct {
    Success    bool   `json:"success"`
    Message    string `json:"message"`
    Slot       string `json:"slot" binding:"required"`
    Size       int64  `json:"size" binding:"required"`
    Type       string `json:"type" binding:"required"`
    Speed      string `json:"speed" binding:"required"`
    Manufacturer string `json:"manufacturer" binding:"required"`
    SerialNumber string `json:"serial_number"`
    PartNumber   string `json:"part_number"`
}

type MemorySummary struct {
    TotalSize  int64  `json:"total_size"`
    Type       string `json:"type"`
    Speed      string `json:"speed"`
    TotalSlots int    `json:"total_slots"`
    UsedSlots  int    `json:"used_slots"`
}
```

**DiskRequest** - 磁盘信息采集请求

```go
type DiskRequest struct {
    Success bool      `json:"success"`
    Message string    `json:"message"`
    Content []DiskInfo `json:"content"`
    Summary DiskSummary `json:"summary"`
}

type DiskInfo struct {
    Success      bool   `json:"success"`
    Message      string `json:"message"`
    Device       string `json:"device" binding:"required"`
    Model        string `json:"model" binding:"required"`
    Type         string `json:"type" binding:"required"`
    Size         int64  `json:"size" binding:"required"`
    SerialNumber string `json:"serial_number"`
    Firmware     string `json:"firmware"`
    RPM          int    `json:"rpm"`
    Manufacturer string `json:"manufacturer"`
    Product      string `json:"product"`
    Vendor       string `json:"vendor"`
    Partitions   []DiskPartitionInfo `json:"partitions"`
}

type DiskPartitionInfo struct {
    Success    bool   `json:"success"`
    Message    string `json:"message"`
    Device     string `json:"device" binding:"required"`
    DiskDevice string `json:"disk_device" binding:"required"`
    Size       int64  `json:"size" binding:"required"`
    Type       string `json:"type"`
    Filesystem string `json:"filesystem"`
    Label      string `json:"label"`
    IsBootable bool   `json:"is_bootable"`
}

type DiskSummary struct {
    TotalCount      int   `json:"total_count"`
    TotalSize       int64 `json:"total_size"`
    SSDCount        int   `json:"ssd_count"`
    HDDCount        int   `json:"hdd_count"`
    TotalPartitions int   `json:"total_partitions"`
}
```

**NetworkInterfaceRequest** - 网络接口信息采集请求

```go
type NetworkInterfaceRequest struct {
    Success bool                   `json:"success"`
    Message string                 `json:"message"`
    Content []NetworkInterfaceInfo `json:"content"`
    Summary NetworkInterfaceSummary `json:"summary"`
}

type NetworkInterfaceInfo struct {
    Success      bool   `json:"success"`
    Message      string `json:"message"`
    Name         string `json:"name" binding:"required"`
    MACAddress   string `json:"mac_address" binding:"required"`
    IPAddress    string `json:"ip_address"`
    SubnetMask   string `json:"subnet_mask"`
    Gateway      string `json:"gateway"`
    Speed        string `json:"speed"`
    Duplex       string `json:"duplex"`
    MTU          int    `json:"mtu"`
    Type         string `json:"type"`
    IsVirtual    bool   `json:"is_virtual"`
    Manufacturer string `json:"manufacturer"`
    Driver       string `json:"driver"`
    PCIAddress   string `json:"pci_address"`
}

type NetworkInterfaceSummary struct {
    TotalCount    int `json:"total_count"`
    PhysicalCount int `json:"physical_count"`
    VirtualCount  int `json:"virtual_count"`
    ActiveCount   int `json:"active_count"`
}
```

---

#### 硬件信息响应 DTO

**HardwareInfoResponse** - 硬件信息响应

```go
type HardwareInfoResponse struct {
    CPU               *CPUResponse               `json:"cpu"`
    Memory            *MemoryResponse            `json:"memory"`
    MemoryModules     []MemoryModuleResponse     `json:"memory_modules"`
    Disks             []DiskResponse             `json:"disks"`
    Partitions        []DiskPartitionResponse    `json:"partitions"`
    NetworkInterfaces []NetworkInterfaceResponse `json:"network_interfaces"`
    Summary           *HardwareSummaryResponse   `json:"summary,omitempty"`
    Version           string                     `json:"version"`
    Checksum          string                     `json:"checksum"`
}
```

**CPUResponse** - CPU 信息响应

```go
type CPUResponse struct {
    Model        string `json:"model"`
    Vendor       string `json:"vendor"`
    Family       string `json:"family"`
    Stepping     string `json:"stepping"`
    Flags        string `json:"flags"`
    Cores        int    `json:"cores"`
    Threads      int    `json:"threads"`
    BaseSpeed    string `json:"base_speed"`
    CacheSizeL1  string `json:"cache_size_l1"`
    CacheSizeL2  string `json:"cache_size_l2"`
    CacheSizeL3  string `json:"cache_size_l3"`
    Architecture string `json:"architecture"`
}
```

**MemoryResponse** - 内存信息响应

```go
type MemoryResponse struct {
    TotalSize int64  `json:"total_size"`
    Type      string `json:"type"`
    Speed     string `json:"speed"`
    Slots     int    `json:"slots"`
}

type MemoryModuleResponse struct {
    Slot         string `json:"slot"`
    Size         int64  `json:"size"`
    Type         string `json:"type"`
    Speed        string `json:"speed"`
    Manufacturer string `json:"manufacturer"`
    SerialNumber string `json:"serial_number"`
    PartNumber   string `json:"part_number"`
}
```

**DiskResponse** - 磁盘信息响应

```go
type DiskResponse struct {
    Device       string `json:"device"`
    Model        string `json:"model"`
    Type         string `json:"type"`
    Size         int64  `json:"size"`
    SerialNumber string `json:"serial_number"`
    Firmware     string `json:"firmware"`
    RPM          int    `json:"rpm"`
    Manufacturer string `json:"manufacturer"`
    Product      string `json:"product"`
    Vendor       string `json:"vendor"`
}

type DiskPartitionResponse struct {
    Device     string `json:"device"`
    DiskDevice string `json:"disk_device"`
    Size       int64  `json:"size"`
    Type       string `json:"type"`
    Filesystem string `json:"filesystem"`
    Label      string `json:"label"`
    IsBootable bool   `json:"is_bootable"`
}
```

**NetworkInterfaceResponse** - 网络接口信息响应

```go
type NetworkInterfaceResponse struct {
    Name         string `json:"name"`
    MACAddress   string `json:"mac_address"`
    IPAddress    string `json:"ip_address"`
    SubnetMask   string `json:"subnet_mask"`
    Gateway      string `json:"gateway"`
    Speed        string `json:"speed"`
    Duplex       string `json:"duplex"`
    MTU          int    `json:"mtu"`
    Type         string `json:"type"`
    IsVirtual    bool   `json:"is_virtual"`
    Manufacturer string `json:"manufacturer"`
    Driver       string `json:"driver"`
    PCIAddress   string `json:"pci_address"`
}
```

**HardwareSummaryResponse** - 硬件摘要信息响应

```go
type HardwareSummaryResponse struct {
    CPUModel    string `json:"cpu_model"`
    CPUCores    int    `json:"cpu_cores"`
    TotalMemory int64  `json:"total_memory"`
    MemorySlots int    `json:"memory_slots"`
    DiskCount   int    `json:"disk_count"`
    TotalDisk   int64  `json:"total_disk"`
    NetworkCount int   `json:"network_count"`
}
```

[🔝 返回顶部](#目录) | [📦 返回 DTO 数据结构](#dto-数据结构)

---

### 通用响应 DTO

#### SuccessResponse

```go
type SuccessResponse struct {
    Code    int         `json:"code"`
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

#### ErrorResponse

```go
type ErrorResponse struct {
    Code    int    `json:"code"`
    Success bool   `json:"success"`
    Message string `json:"message"`
}
```

[🔝 返回顶部](#目录) | [📦 返回 DTO 数据结构](#dto-数据结构)

---

## 数据模型关系

### 表关系图

```
┌─────────────────────┐         ┌──────────────────────┐
│ bmc_trap_parsers    │  1:N    │ alerts               │
│---------------------│─────────│----------------------│
│ id (PK)             │         │ id (PK)              │
│ parser_name         │         │ parser_id (FK)       │
│ vendor_code         │         │ alert_status         │
│ vendor_name         │         │ trap_oid             │
│ alert_level_oid     │         │ source_ip            │
│ ...                 │         │ component            │
└─────────────────────┘         │ ...                  │
                                └──────────────────────┘

┌─────────────────────┐
│ machine_infos       │
│---------------------│
│ id (PK)             │
│ collected_at        │
│ content (JSONB)     │
│ - hostname          │
│ - ip_addresses      │
└─────────────────────┘
         │
         │ 1:1
         ▼
┌─────────────────────┐
│ hardware_info       │
│---------------------│
│ id (PK)             │
│ target_ip (FK)      │
│ collected_at        │
│ hardware_info_snapshot (JSONB)
│ - cpu               │
│ - memory            │
│ - disks             │
│ - network_interfaces│
│ version             │
│ checksum            │
└─────────────────────┘
         │
         ├──────────────┬──────────────┬──────────────────┐
         │ 1:1          │ 1:1          │ 1:1              │ 1:1
         ▼              ▼              ▼                  ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐
│ hardware_cpu│ │hardware_memo│ │hardware_disk│ │hardware_network_    │
│-------------│ │ry-----------│ │-------------│ │interface            │
│ id (PK)     │ │id (PK)      │ │id (PK)      │ │---------------------│
│ target_ip   │ │target_ip    │ │target_ip    │ │id (PK)              │
│ total_cores │ │total_size   │ │total_count  │ │target_ip            │
│ total_threads││type         │ │total_size   │ │total_count          │
│             │ │speed        │ │ssd_count    │ │physical_count       │
│             │ │total_slots  │ │hdd_count    │ │virtual_count        │
│             │ │used_slots   │ │total_partitions                    │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────┘

┌─────────────────────┐
│ anomaly_results     │
│---------------------│
│ id (PK)             │
│ target_ip           │
│ check_type          │
│ check_item          │
│ success             │
│ level               │
│ extra_data (JSONB)  │
└─────────────────────┘

┌─────────────────────┐
│ users               │
│---------------------│
│ id (PK)             │
│ username            │
│ password_hash       │
│ is_admin            │
│ permissions (JSONB) │
└─────────────────────┘
```

---

## 设计原则

### 数据库设计原则

1. **使用 GORM 管理数据库**
   - 通过 GORM struct tag 定义字段和索引
   - 使用 `gorm:"index"` 自动创建索引
   - 使用 `gorm:"uniqueIndex"` 创建唯一索引
   - 无需手动执行 SQL 创建索引

2. **JSONB 灵活存储**
   - 使用 JSONB 存储映射配置、变量映射等灵活结构，便于扩展
   - 支持复杂数据结构的完整保存
   - 新增字段无需修改表结构

3. **软删除支持**
   - 使用 `is_deleted` 字段支持数据软删除
   - 保留历史数据，便于审计和恢复

4. **时间戳追踪**
   - 使用 `created_at`、`updated_at` 追踪数据变更
   - GORM 自动维护这些字段

5. **枚举值标准化**
   - 使用字符串枚举而非数字，提高可读性
   - 便于新增枚举值

### DTO 设计原则

1. **请求验证**
   - 使用 `binding` 标签进行参数验证
   - 必填字段和可选字段明确区分

2. **指针可选**
   - 使用指针类型表示可选字段
   - nil 表示不使用该过滤条件

3. **分页统一**
   - 所有列表查询使用统一的分页参数结构

4. **响应标准化**
   - 使用统一的响应结构，包含分页信息

5. **版本兼容**
   - DTO 结构独立于数据库模型
   - 支持 API 版本演进

### 索引设计原则

使用 GORM 的索引定义方式：

```go
// 基本索引
TargetIP net.IP `gorm:"column:target_ip;type:inet;index"`

// 唯一索引  
Username string `gorm:"uniqueIndex;not null"`

// 复合索引（在 model 中定义）
func (User) TableName() string {
    return "users"
}

// 或者使用 migrate 创建复合索引
db.Exec("CREATE INDEX idx_users_name_email ON users(username, email)")
```

---

[🔝 返回顶部](#目录)

## 最佳实践

### 查询优化

1. 使用复合索引覆盖常用查询组合
2. 对 JSONB 字段使用 GIN 索引支持复杂查询
3. 时间范围查询使用时区一致的 UTC 时间
4. 大数据量表考虑分区策略

### 数据一致性

1. 使用事务保证关联数据的一致性
2. 外键关系在应用层维护（Go 语言特性）
3. 定期清理过期数据（根据业务需求配置保留策略）

### 扩展性考虑

1. JSONB 字段支持新字段无需修改表结构
2. DTO 与模型分离，支持 API 演进
3. 枚举值使用字符串，便于新增枚举值

---

[🔝 返回顶部](#目录)
