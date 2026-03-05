// 示例：磁盘温度检测器
// 复制此文件作为模板来创建新的磁盘检测器
// 文件名应该使用小写和下划线，例如：disk_temperature_detector.go

package detectors

import (
	"baize-monitor/pkg/dto/request"
	
	// 如需使用，请取消注释下面的导入
	// "baize-monitor/internal/agent/anomaly/check_items/disk"
)

// diskExampleDetector 示例检测器结构体
// 命名规范：小写开头，私有类型
type diskExampleDetector struct {
	// 配置参数应该在这里定义
	threshold int
}

// 注意：示例检测器不会注册，仅供学习参考
// 如需使用，请复制此文件并取消注释下面的 init 函数
/*
// init 函数会在包加载时自动执行，用于注册检测器
func init() {
	disk.RegisterDetector(&diskExampleDetector{
		threshold: 50, // 设置默认阈值
	})
}
*/

// DetectorType 返回检测器类型标识符
// 命名规范：使用小写字母和下划线，例如：disk_temperature
func (d *diskExampleDetector) DetectorType() string {
	return "disk_example"
}

// Detect 执行实际检测逻辑
// 返回：AnomalyResult - 如果检测到异常返回有效结果，否则返回空结果
//      error - 如果检测过程出错返回错误
func (d *diskExampleDetector) Detect() (request.AnomalyResult, error) {
	// ============================================
	// 在这里实现你的检测逻辑
	// ============================================
	
	// 示例 1: 读取磁盘统计信息
	// file, err := os.Open("/proc/diskstats")
	// if err != nil {
	// 	return request.AnomalyResult{}, nil // 静默失败
	// }
	// defer file.Close()
	
	// 示例 2: 解析磁盘数据
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	fields := strings.Fields(scanner.Text())
	// 	if len(fields) < 14 {
	// 		continue
	// 	}
	// 	device := fields[2]
	// 	ioErrors := fields[13] // IO 错误计数
	// }
	
	// 示例 3: 检查数据是否异常
	// if someValue > d.threshold {
	// 	return disk.CreateAnomalyResult(  // 需要导入 disk 包
	// 		d.DetectorType(),
	// 		"warning", // 级别：info, warning, critical
	// 		"磁盘异常检测",
	// 		map[string]interface{}{
	// 			"device":    "/dev/sda",
	// 			"value":     someValue,
	// 			"threshold": d.threshold,
	// 		},
	// 	), nil
	// }
	
	// ============================================
	// 重要原则:
	// 1. 零影响：只读内核缓存数据，不主动扫描硬件
	// 2. 持续性：检测累积性问题，非瞬时波动
	// 3. 错误处理：检测失败返回 error，由 Registry 统一处理
	// 4. 轻量级：单次检测 < 100ms，内存 < 10MB
	// 5. 避免使用 smartctl 等可能影响 IO 的外部命令
	// ============================================
	
	return request.AnomalyResult{}, nil
}
