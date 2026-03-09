# 采集器架构详解

## 1. 组件职责

### Hardware
模块入口，实现 dataProvider 接口

### Registry
管理采集器，执行采集

### Collectors
具体硬件信息采集

## 2. 数据流

```
GetData()
  ↓
Registry.CollectAll()
  ↓
创建 HardwareInfoRequest{} (空容器)
  ↓
┌─────────────────────────────────────┐
│ CPU Collector    → 填充 CPUs       │
│ Memory Collector → 填充 Memory     │
│ Disk Collector   → 填充 Disks      │
│ Network Collector→ 填充 NetworkInt │
└─────────────────────────────────────┘
  ↓
返回完整的 HardwareInfoRequest
```

## 3. 错误处理

### 错误隔离
每个采集器独立执行，一个失败不影响其他

### 错误记录
采集失败时，记录到对应部件的 `Success` 和 `Message` 字段

## 4. 扩展方法

添加新采集器：
1. 实现 `HardwareCollector` 接口
2. 在 `hardware.New()` 中注册

## 相关文档

- [模块概览](01-overview.md) - 高层介绍
- [采集器总览](03-collectors/README.md) - 接口定义
- [采集策略](03-strategies.md) - 分级采集详情
