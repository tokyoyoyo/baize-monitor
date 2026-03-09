# 网络检测器

## 1. 概述

网络检测器负责检测网络相关的异常情况，包括链路状态、网络连接等问题。

**检测对象**：网络接口和连接状态
**检测器数量**：2 个（1 个功能检测器 + 1 个示例检测器）
**特殊考虑**：排除虚拟接口，使用 ip 命令

## 2. 检测器列表

| 检测器 | 检测项名称 | 异常级别 | 说明 |
|--------|-----------|---------|------|
| network_link_down | network_link_down | critical | 网络接口链路断开 |
| network_example | network_example | - | 示例模板 |

## 3. 检测器详解

### 3.1 network_link_down

**功能**：检测网络接口链路状态，发现 DOWN 状态的接口

**实现文件**：[network_link_down_detector.go](https://github.com/zhaojunjie/baize-monitor/blob/main/internal/agent/anomaly/network/network_link_down_detector.go)

#### 检测逻辑流程

```mermaid
flowchart TD
    A[开始检测] --> B[执行 ip -o link 命令]
    B --> C{命令成功？}
    C -->|否 | D[静默失败]
    C -->|是 | E[逐行解析输出]
    E --> F{包含 state DOWN?}
    F -->|否 | G[还有更多行？]
    F -->|是 | H{是 loopback 接口？}
    H -->|是 | G
    H -->|否 | I[提取接口名称]
    I --> J[报告链路 DOWN]
    J --> K[返回异常结果]
    G -->|是 | E
    G -->|否 | L[正常，所有接口 UP]
    L --> M[返回空结果]
    D --> N[返回空结果]
```

#### `ip -o link` 输出格式

**示例输出**：
```
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000\    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP mode DEFAULT group default qlen 1000\    link/ether 00:0c:29:xx:xx:xx brd ff:ff:ff:ff:ff:ff
3: eth1: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN mode DEFAULT group default qlen 1000\    link/ether 00:0c:29:xx:xx:xx brd ff:ff:ff:ff:ff:ff
```

**字段说明**：
- `1:`: 接口索引
- `lo:`: 接口名称（以冒号结尾）
- `<...>`: 接口标志
- `state UP/DOWN/UNKNOWN`: 链路状态
- `link/ether`: MAC 地址类型

#### 检测场景

**场景 1：物理网卡 DOWN**
- **现象**：网络接口状态为 DOWN
- **可能原因**：
  - 网线未连接
  - 交换机端口故障
  - 网卡硬件故障
  - 驱动程序问题

**场景 2：配置错误**
- **现象**：接口被管理员禁用
- **可能原因**：
  - 配置错误导致接口 DOWN
  - 网络配置变更
  - 安全策略禁用

**场景 3：驱动问题**
- **现象**：网卡驱动加载失败
- **可能原因**：
  - 网卡驱动加载失败
  - 驱动崩溃
  - 内核模块问题

#### 返回数据示例

```json
{
  "check_type": "network",
  "check_item": "network_link_down",
  "success": false,
  "level": "critical",
  "message": "Network interface link is down",
  "extra_data": {
    "interface": "eth1",
    "status": "DOWN",
    "issue": "link_down"
  },
  "checked_at": "2024-01-01T10:00:00Z"
}
```

#### 异常级别说明

**使用 critical 级别的原因**：
- 网络接口 DOWN 是严重问题
- 可能导致服务不可用
- 需要立即处理

### 3.2 network_example

**功能**：示例检测器模板

**实现文件**：[example_detector.go](https://github.com/zhaojunjie/baize-monitor/blob/main/internal/agent/anomaly/network/example_detector.go)

**用途**：
- 作为开发新网络检测器的参考
- 展示检测器的标准结构
- 提供最佳实践示例

## 4. 数据源

### 4.1 /proc/net/dev

**网络接口统计**：

```go
file, err := os.Open("/proc/net/dev")
if err != nil {
    return request.AnomalyResult{}, err
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // 解析网络接口统计
    // 格式：interface: rx_bytes rx_packets rx_errs rx_drop ... 
    //                tx_bytes tx_packets tx_errs tx_drop ...
}
```

**字段说明**：
- `rx_bytes`: 接收字节数
- `rx_packets`: 接收包数
- `rx_errs`: 接收错误数
- `rx_drop`: 接收丢包数
- `tx_bytes`: 发送字节数
- `tx_packets`: 发送包数
- `tx_errs`: 发送错误数
- `tx_drop`: 发送丢包数

### 4.2 /proc/net/snmp

**网络协议统计**：
```go
// 读取网络协议统计
snmp, _ := os.ReadFile("/proc/net/snmp")
// 包含 IP、ICMP、TCP、UDP 等协议统计
```

### 4.3 系统命令

**主要命令**：
```go
// ip 命令
cmd := exec.Command("ip", "-o", "link")
output, _ := cmd.Output()
```

**调试工具**：
```bash
# 查看所有接口状态
ip -o link

# 查看简要状态
ip link show

# 查看接口统计
cat /proc/net/dev

# 查看接口详细信息
ip -s link

# 查看网卡驱动信息
ethtool eth0
```

## 5. 扩展示例

### 5.1 网络丢包率检测器

**功能**：监控网络接口的丢包率

**实现思路**：
1. 读取 `/proc/net/dev`
2. 解析丢包计数（rx_drop 和 tx_drop）
3. 计算丢包率
4. 检查是否超过阈值

**伪代码**：
```go
func (d *networkPacketLossDetector) Detect() (request.AnomalyResult, error) {
    // 1. 读取 /proc/net/dev
    // 2. 解析丢包数和总包数
    lossRate := (totalDrops / totalPackets) * 100.0
    
    // 3. 检查阈值
    if lossRate > d.threshold {
        return utils.CreateAnomalyResult(
            models.CheckTypeNetwork,
            "network_packet_loss",
            models.AnomalyLevelWarning,
            "Network packet loss rate is too high",
            map[string]interface{}{
                "loss_rate": lossRate,
                "threshold": d.threshold,
            },
        ), nil
    }
    return request.AnomalyResult{}, nil
}
```

### 5.2 网络错误检测器

**功能**：监控接收/发送错误数

**实现思路**：
1. 读取 `/proc/net/dev` 的错误计数
2. 检查错误数是否超过阈值
3. 返回异常结果

### 5.3 TCP 连接数异常检测器

**功能**：监控 TCP 连接状态

**实现思路**：
1. 读取 `/proc/net/tcp` 和 `/proc/net/tcp6`
2. 统计连接数
3. 检查连接数是否异常

## 6. 调试方法

### 6.1 查看网络接口状态

```bash
# 查看所有接口状态
ip -o link

# 查看简要状态
ip link show

# 查看接口统计
cat /proc/net/dev

# 查看接口详细信息
ip -s link

# 查看网卡驱动信息
ethtool eth0
```

### 6.2 测试检测器

```go
package network

import (
    "testing"
    "log"
)

func TestNetworkLinkDownDetector(t *testing.T) {
    detector := &networkLinkStatusDetector{}
    
    result, err := detector.Detect()
    if err != nil {
        t.Errorf("Detect() error = %v", err)
    }
    
    log.Printf("Result: %+v", result)
}
```

### 6.3 手动测试

```go
detector := &networkLinkStatusDetector{}

result, err := detector.Detect()
if err != nil {
    log.Printf("Error: %v", err)
}

if result.Success {
    log.Println("No anomaly detected")
} else {
    log.Printf("Anomaly: %s - %s", result.Level, result.Message)
    log.Printf("Extra: %+v", result.ExtraData)
}
```

## 7. 最佳实践

### 7.1 命令执行

**✅ 处理命令执行失败**：
```go
cmd := exec.Command("ip", "-o", "link")
output, err := cmd.Output()
if err != nil {
    return request.AnomalyResult{}, nil  // 静默失败
}
```

### 7.2 接口过滤

**✅ 排除 loopback 接口**：
```go
if strings.Contains(line, "lo:") {
    continue  // 跳过 loopback
}

// ✅ 根据需求过滤虚拟接口
if strings.HasPrefix(iface, "veth") ||
   strings.HasPrefix(iface, "docker") ||
   strings.HasPrefix(iface, "br-") {
    // 根据需求决定是否跳过
}
```

### 7.3 解析健壮性

**✅ 健壮的接口名称提取**：
```go
parts := strings.Fields(line)
iface := ""
for i, part := range parts {
    if strings.HasSuffix(part, ":") {
        iface = strings.TrimSuffix(part, ":")
        break
    }
    if i > 2 {  // 防止无限循环
        break
    }
}

if iface == "" {
    continue  // 无法解析则跳过
}
```

## 8. 故障排查

### 8.1 ip 命令不可用

**问题**：系统没有 ip 命令

**解决方案**：
```go
// 回退到 /sys/class/net/*/operstate
files, _ := filepath.Glob("/sys/class/net/*/operstate")
for _, file := range files {
    state, _ := os.ReadFile(file)
    if strings.TrimSpace(string(state)) == "down" {
        // 提取接口名
        iface := filepath.Base(filepath.Dir(file))
        if iface != "lo" {
            // 报告异常
        }
    }
}
```

### 8.2 解析错误

**问题**：输出格式不符合预期

**调试**：
```go
log.Printf("Raw output: %s", string(output))

lines := strings.Split(string(output), "\n")
for i, line := range lines {
    log.Printf("Line %d: %s", i, line)
    // 添加详细解析日志
}
```

### 8.3 误报问题

**问题**：虚拟接口 DOWN 导致误报

**解决方案**：
```go
// 过滤虚拟接口
if strings.HasPrefix(iface, "veth") ||
   strings.HasPrefix(iface, "docker") ||
   strings.HasPrefix(iface, "br-") ||
   strings.HasPrefix(iface, "virbr") {
    continue  // 跳过虚拟接口
}
```

## 相关文档

- [检测器总览](../README.md)
- [接口规范](../interface.md)
- [类型和分级](../types.md)
- [CPU 检测器](cpu.md)
- [内存检测器](memory.md)
- [磁盘检测器](disk.md)
