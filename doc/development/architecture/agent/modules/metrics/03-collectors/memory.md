# 内存采集器

## 1. 概述

内存采集器负责采集内存使用指标，包括总内存、已用内存、空闲内存和使用率。

**采集对象**：系统内存使用情况  
**采集器**：MemoryCollector  
**指标类型**：Gauge（实时值）

## 2. 采集指标

### 内存总量

指标名：`node_memory_total_bytes`  
类型：Gauge  
说明：系统总内存容量（字节）

### 已用内存

指标名：`node_memory_used_bytes`  
类型：Gauge  
说明：已使用的内存容量（字节）

### 空闲内存

指标名：`node_memory_free_bytes`  
类型：Gauge  
说明：空闲内存容量（字节）

### 内存使用率

指标名：`node_memory_used_percent`  
类型：Gauge  
说明：内存使用率（0-100）

## 3. 采集方法

### 数据源

使用 gopsutil 的 `mem.VirtualMemory()` API

### 采集逻辑

1. 调用 `mem.VirtualMemory()` 获取内存信息
2. 提取 Total、Used、Free、UsedPercent 字段
3. 创建对应的 Prometheus 指标
4. 发送到 Prometheus channel

## 4. 指标说明

### 命名规范

遵循 Prometheus 指标命名规范：
- 小写字母
- 下划线分隔
- 以 `_bytes` 或 `_percent` 结尾

### 标签

当前无额外标签，可扩展：
- `instance`: 实例标识
- `region`: 区域标识

## 5. 错误处理

### 采集失败

- 记录日志
- 返回空指标
- 不影响其他采集器

### 数据异常

- 验证数据范围
- 负值或异常值丢弃
- 记录告警日志

## 6. 性能特点

### 采集耗时

- 单次采集：< 10ms
- 系统调用：1 次
- 内存分配：可忽略

### 采集频率建议

- Prometheus 默认：15 秒
- 可根据需要调整
- 内存信息变化较慢，无需高频采集

## 7. 扩展示例

### 添加更多内存指标

- Available: 可用内存
- Buffers: 缓冲区内存
- Cached: 缓存内存
- Swap: 交换内存

### 添加容器内存

- 采集容器内存使用
- 添加 container_id 标签

## 相关文档

- [采集器总览](README.md) - 接口定义
- [架构设计](../02-architecture.md) - Prometheus 集成
- [模块概览](../01-overview.md) - 模块整体
