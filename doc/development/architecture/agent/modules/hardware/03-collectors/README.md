# 采集器总览

## 1. 接口定义

所有采集器必须实现 `HardwareCollector` 接口

## 2. 采集器

| 采集器 | 采集对象 | 详细文档 |
|--------|---------|---------|
| CPUCollector | CPU | [CPU 采集器](implementations/cpu.md) |
| MemoryCollector | 内存 | [内存采集器](implementations/memory.md) |
| DiskCollector | 磁盘 | [磁盘采集器](implementations/disk.md) |
| NetworkCollector | 网络接口 | [网络采集器](implementations/network.md) |

## 3. 数据结构

每个硬件类型包含：
- `Success`: 采集是否成功
- `Message`: 错误信息
- `Content`: 详细数据列表
- `Summary`: 摘要统计

## 4. 设计原则

1. **无状态** - 采集器不需要初始化
2. **独立** - 采集器之间互不依赖
3. **错误隔离** - 一个失败不影响其他
