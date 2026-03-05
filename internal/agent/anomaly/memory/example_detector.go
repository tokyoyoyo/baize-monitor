// 示例：内存温度检测器
// 复制此文件作为模板来创建新的内存检测器
// 文件名应该使用小写和下划线，例如：memory_temperature_detector.go

package memory

import (
	"baize-monitor/pkg/dto/request"
)

// memoryExampleDetector 示例检测器结构体
// 命名规范：小写开头，私有类型
type memoryExampleDetector struct {
	// 配置参数应该在这里定义
	threshold int
}

// 注意：示例检测器不会注册，仅供学习参考
// 如需使用，请复制此文件并取消注释下面的 init 函数
/*
// init 函数会在包加载时自动执行，用于注册检测器
func init() {
	memory.RegisterDetector(&memoryExampleDetector{
		threshold: 85, // 设置默认阈值
	})
}
*/

// CheckItemName 返回检测项名称
// 命名规范：使用小写字母和下划线，例如：memory_temperature
func (d *memoryExampleDetector) CheckItemName() string {
	return "memory_example"
}

// Detect 执行实际检测逻辑
// 返回：AnomalyResult - 如果检测到异常返回有效结果，否则返回空结果
//      error - 如果检测过程出错返回错误
func (d *memoryExampleDetector) Detect() (request.AnomalyResult, error) {
	// ============================================
	// 在这里实现你的检测逻辑
	// ============================================
	
	// 示例 1: 读取内存错误计数器
	// file, err := os.Open("/sys/devices/system/edac/mc/mc0/ece_count")
	// if err != nil {
	// 	return request.AnomalyResult{}, err // 返回错误，由 Registry 记录
	// }
	// defer file.Close()
	
	// 示例 2: 解析数据
	// var count int
	// fmt.Fscanf(file, "%d", &count)
	
	// 示例 3: 检查数据是否异常
	// if count > d.threshold {
	// 	return utils.CreateAnomalyResult(
	// 		models.CheckTypeMemory,
	// 		d.CheckItemName(),
	// 		models.AnomalyLevelWarning, // 级别：info, warning, critical
	// 		"内存错误计数超过阈值",
	// 		map[string]interface{}{
	// 			"error_count": count,
	// 			"threshold":   d.threshold,
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
