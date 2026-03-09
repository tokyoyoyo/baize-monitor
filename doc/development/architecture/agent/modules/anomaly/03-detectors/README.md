# 检测器总览

## 1. 接口定义

### Detector 接口
所有检测器必须实现的基础接口

### AnomalyChecker 接口
检测器分组管理器接口

### CheckerRegistry
注册表管理所有检测器组，执行全局检测

## 2. 检测器分类

### 按检测对象分类

| 检测类型 | 检测器组 | 说明 |
|---------|---------|------|
| CPU | CPUChecker | CPU 核心数、频率、温度等 |
| 内存 | MemoryChecker | ECC 错误、使用率等 |
| 磁盘 | DiskChecker | I/O 错误、SMART、空间等 |
| 网络 | NetworkChecker | 链路状态、丢包、错误等 |

## 3. 异常级别

- **info** - 信息级，不影响业务
- **warning** - 警告级，可能影响业务
- **critical** - 严重级，已影响业务

### 现有检测器级别

| 检测器 | 级别 | 原因 |
|--------|------|------|
| cpu_core_count_mismatch | warning | CPU 配置异常 |
| memory_ecc_error | warning | 内存错误累积 |
| disk_io_error | warning | 磁盘 I/O 错误 |
| network_link_down | critical | 网络中断 |

## 4. 检测结果

AnomalyResult 包含：
- CheckType: 检测类型
- CheckItem: 检测项名称
- Success: false 表示检测到异常
- Level: 异常级别
- Message: 异常描述
- ExtraData: 额外数据
- CheckedAt: 检测时间

### 返回规则

- **检测到异常** - 返回 AnomalyResult（Success=false）
- **正常情况** - 返回空的 AnomalyResult{}
- **检测失败** - 返回 error

## 5. 命名规范

### 检测项名称
格式：`<type>_<detector_name>`（小写，下划线分隔）

### 文件名
格式：`<type>_<detector_name>_detector.go`

## 6. 错误处理

### 静默失败
非关键错误静默处理

### 返回错误
关键错误返回 error，由 Registry 统一处理

## 7. 设计原则

### 零影响原则
- 只读内核缓存数据
- 不影响系统性能

### 持续性原则
- 检测累积性问题
- 非瞬时波动

## 相关文档

- [模块概览](../01-overview.md) - 高层介绍
- [插件架构](../02-plugin-architecture.md) - 自动注册机制
- [CPU 检测器](implementations/cpu.md)
- [内存检测器](implementations/memory.md)
- [磁盘检测器](implementations/disk.md)
- [网络检测器](implementations/network.md)
