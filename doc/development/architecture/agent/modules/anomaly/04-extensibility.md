# 扩展指南

## 1. 添加检测器

### 步骤

1. **创建检测器文件** - 在对应子包创建文件
2. **实现 Detector 接口** - 实现 `CheckItemName()` 和 `Detect()` 方法
3. **自动注册** - 在 `init()` 中调用 `RegisterDetector()`

### 示例

```go
// 1. 创建检测器
type cpuTemperatureDetector struct {
    threshold int
}

// 2. 自动注册
func init() {
    RegisterDetector(&cpuTemperatureDetector{threshold: 85})
}

// 3. 实现接口
func (d *cpuTemperatureDetector) CheckItemName() string {
    return "cpu_temperature"
}

func (d *cpuTemperatureDetector) Detect() (request.AnomalyResult, error) {
    // 实现检测逻辑
    // 返回异常结果或空结果
}
```

## 2. 配置化

### 添加配置项

```yaml
# config/agent.yaml
anomaly:
  detectors:
    cpu_temperature:
      enabled: true
      threshold: 85
```

### 实现配置化

根据配置决定是否注册检测器

## 3. 测试

### 单元测试

为检测器编写测试用例

### 集成测试

验证检测器正确集成到模块中

## 相关文档

- [插件架构](02-plugin-architecture.md) - 自动注册机制
- [检测器总览](03-detectors/README.md) - 接口定义
- [最佳实践](05-best-practices.md) - 开发规范
