// 示例：网络丢包率检测器
// 复制此文件作为模板来创建新的网络检测器
// 文件名应该使用小划线，例如：network_packet_loss_detector.go

package network

import (
	"baize-monitor/pkg/dto/request"
)

// networkExampleDetector 示例检测器结构体
// 命名规范：小写开头，私有类型
type networkExampleDetector struct {
	// 配置参数应该在这里定义
	threshold int
}

// 注意：示例检测器不会注册，仅供学习参考
// 如需使用，请复制此文件并取消注释下面的 init 函数
/*
// init 函数会在包加载时自动执行，用于注册检测器
func init() {
	network.RegisterDetector(&networkExampleDetector{
		threshold: 1000, // 设置默认阈值
	})
}
*/

// CheckItemName 返回检测项名称
// 命名规范：使用小写字母和下划线，例如：network_packet_loss
func (d *networkExampleDetector) CheckItemName() string {
	return "network_example"
}

// Detect 执行实际检测逻辑
// 返回：AnomalyResult - 如果检测到异常返回有效结果，否则返回空结果
//      error - 如果检测过程出错返回错误
func (d *networkExampleDetector) Detect() (request.AnomalyResult, error) {
	// ============================================
	// 在这里实现你的检测逻辑
	// ============================================
	
	// 示例 1: 读取网络接口统计
	// file, err := os.Open("/proc/net/dev")
	// if err != nil {
	// 	return request.AnomalyResult{}, nil // 静默失败
	// }
	// defer file.Close()
	
	// 示例 2: 解析网络数据
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if strings.Contains(line, "eth0") {
	// 		// 解析丢包计数
	// 	}
	// }
	
	// 示例 3: 检查数据是否异常
	// if packetLoss > d.threshold {
	// 	return utils.CreateAnomalyResult(
	// 		models.CheckTypeNetwork,
	// 		d.CheckItemName(),
	// 		models.AnomalyLevelWarning, // 级别：info, warning, critical
	// 		"网络接口丢包严重",
	// 		map[string]interface{}{
	// 			"interface":   "eth0",
	// 			"packet_loss": packetLoss,
	// 			"threshold":   d.threshold,
	// 		},
	// 	), nil
	// }
	
	// ============================================
	// 重要原则:
	// 1. 零影响：只读内核缓存数据，不主动发送探测包
	// 2. 持续性：检测累积性问题，非瞬时波动
	// 3. 错误处理：检测失败返回 error，由 Registry 统一处理
	// 4. 轻量级：单次检测 < 100ms，内存 < 10MB
	// 5. 避免使用 ping 等主动探测命令
	// ============================================
	
	return request.AnomalyResult{}, nil
}
