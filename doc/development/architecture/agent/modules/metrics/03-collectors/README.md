# 采集器总览

## 1. 接口定义

### PrometheusMetricsCollector 接口

所有采集器必须实现的接口，继承自 `prometheus.Collector`

### Describe 方法

返回指标描述符，Prometheus 用于了解指标元数据

### Collect 方法

采集指标数据并发送到 Prometheus channel

## 2. 采集器列表

### 已实现

| 采集器 | 采集对象 | 详细文档 |
|--------|---------|---------|
| CPUCollector | CPU 指标 | [CPU 采集器](cpu.md) |
| MemoryCollector | 内存指标 | [内存采集器](memory.md) |

### 待扩展

| 采集器 | 采集对象 | 说明 |
|--------|---------|------|
| DiskCollector | 磁盘指标 | 磁盘 I/O、使用率等 |
| NetworkCollector | 网络指标 | 网络流量、连接数等 |

## 3. 设计原则

### 插件化
- 独立实现
- 集中注册
- 易于扩展

### 标准化
- 遵循 Prometheus 规范
- 统一接口

### 错误隔离
- 独立执行
- 静默失败

## 4. 扩展方法

添加新采集器：
1. 实现 `PrometheusMetricsCollector` 接口
2. 在 metrics.go 的 New() 方法中注册

## 相关文档

- [模块概览](../01-overview.md) - 高层介绍
- [架构设计](../02-architecture.md) - Prometheus 集成
- [CPU 采集器](cpu.md) - CPU 指标采集
- [内存采集器](memory.md) - 内存指标采集
