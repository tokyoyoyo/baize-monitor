# CPU 采集器

## 1. 概述

CPU 采集器负责采集 CPU 使用率指标。

**采集对象**：CPU 使用率  
**采集器**：CPUCollector  
**指标类型**：Gauge（实时值）

## 2. 采集指标

### CPU 使用率

指标名：`baize_metrics_cpu_usage_percent`  
类型：Gauge  
说明：CPU 使用率百分比（所有核心的平均值）

## 3. 采集方法

### 数据源

使用 gopsutil 的 `cpu.Percent()` API

### 采集逻辑

1. 调用 `cpu.Percent(0, false)` 获取 CPU 使用率
2. 参数说明：
   - 第一个参数 0：不指定时间间隔，直接返回
   - 第二个参数 false：返回所有核心的平均值
3. 创建 Prometheus 指标
4. 发送到 Prometheus channel

## 4. 指标说明

### 命名规范

遵循 Prometheus 指标命名规范：
- 小写字母
- 下划线分隔
- 以 `_percent` 结尾

### 标签

当前无额外标签，可扩展：
- `instance`: 实例标识
- `cpu`: CPU 核心编号

## 5. 错误处理

### 采集失败

- 记录日志
- 返回空指标
- 不影响其他采集器

### 数据异常

- 验证数据范围（0-100）
- 异常值丢弃
- 记录告警日志

## 6. 性能特点

### 采集耗时

- 单次采集：< 10ms
- 系统调用：1 次
- 内存分配：可忽略

### 采集频率建议

- Prometheus 默认：15 秒
- 可根据需要调整
- CPU 信息变化较快，建议 15-30 秒采集一次

## 7. 扩展示例

### 添加分核心指标

- 采集每个 CPU 核心的使用率
- 添加 `cpu` 标签区分核心

### 添加 CPU 状态指标

- User: 用户空间使用率
- System: 内核空间使用率
- Idle: 空闲率
- IOWait: I/O 等待率

## 相关文档

- [采集器总览](README.md) - 接口定义
- [架构设计](../02-architecture.md) - Prometheus 集成
- [模块概览](../01-overview.md) - 模块整体
- [内存采集器](memory.md) - 内存指标采集
