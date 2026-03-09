# 扩展指南

## 1. 添加采集器

### 步骤

1. **创建采集器文件** - `internal/agent/hardware/collectors/gpu_collector.go`

2. **实现接口** - 实现 `HardwareCollector` 接口

3. **定义数据结构** - 在 `pkg/dto/request/hardware/` 中定义 GPU 相关结构

4. **注册采集器** - 在 `hardware.New()` 中注册

### 示例

```go
// 1. 实现采集器
type GPUCollector struct{}

func (g *GPUCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
    // 实现采集逻辑
}

// 2. 在 hardware.New() 中注册
registry.Register(collectors.NewGPUCollector())
```

## 2. 配置化

### 添加配置项

```yaml
# config/agent.yaml
hardware:
  collectors:
    gpu:
      enabled: true
```

### 实现配置化

根据配置决定是否注册采集器

## 3. 测试

### 单元测试

为采集器编写测试用例

### 集成测试

验证采集器正确集成到模块中

## 相关文档

- [架构详解](02-collector-architecture.md) - 注册流程
- [采集器总览](03-collectors/README.md) - 接口定义
- [最佳实践](05-best-practices.md) - 开发规范
