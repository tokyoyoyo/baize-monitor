// 示例：CPU 温度传感器故障检测器
// 复制此文件作为模板来创建新的 CPU 检测器
// 文件名应该使用小写和下划线，例如：cpu_temperature_detector.go

package detectors

import (
	"baize-monitor/pkg/dto/request"
)

// cpuExampleDetector 示例检测器结构体
// 命名规范：小写开头，私有类型
type cpuExampleDetector struct {
	// 配置参数应该在这里定义
	threshold int
}

// 注意：示例检测器不会注册，仅供学习参考
// 如需使用，请复制此文件并取消注释下面的 init 函数
/*
// init 函数会在包加载时自动执行，用于注册检测器
func init() {
	cpu.RegisterDetector(&cpuExampleDetector{
		threshold: 100, // 设置默认阈值
	})
}
*/

// DetectorType 返回检测器类型标识符
// 命名规范：使用小写字母和下划线，例如：cpu_temperature
func (d *cpuExampleDetector) DetectorType() string {
	return "cpu_example"
}

// Detect 执行实际检测逻辑
// 返回：AnomalyResult - 如果检测到异常返回有效结果，否则返回空结果
//
//	error - 如果检测过程出错返回错误
func (d *cpuExampleDetector) Detect() (request.AnomalyResult, error) {
	// ============================================
	// 在这里实现你的检测逻辑
	// ============================================

	// 示例 1: 读取系统文件获取数据
	// data, err := os.Open("/proc/cpuinfo")
	// if err != nil {
	// 	return request.AnomalyResult{}, err // 返回错误，由 Registry 记录
	// }
	// defer data.Close()

	// 示例 2: 调用系统 API 获取数据
	// cpuInfo, err := gopsutilcpu.Info()
	// if err != nil {
	// 	return request.AnomalyResult{}, err // 返回错误
	// }

	// 示例 3: 检查数据是否异常
	// if someValue > d.threshold {
	// 	return cpu.CreateAnomalyResult(
	// 		d.DetectorType(),
	// 		"warning", // 级别：info, warning, critical
	// 		"检测到异常",
	// 		map[string]interface{}{
	// 			"key": "value", // 额外数据
	// 		},
	// 	), nil
	// }

	// ============================================
	// 重要原则:
	// 1. 零影响：只读内核缓存数据，不主动扫描硬件
	// 2. 持续性：检测累积性问题，非瞬时波动
	// 3. 错误处理：检测失败返回 error，由 Registry 统一处理
	// 4. 轻量级：单次检测 < 100ms，内存 < 10MB
	// ============================================

	return request.AnomalyResult{}, nil
}
