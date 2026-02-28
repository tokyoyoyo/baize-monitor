package hardware

import (
	"baize-monitor/pkg/constants"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DiskInfoCollector 磁盘信息收集插件
type DiskInfoCollector struct {
	name                string
	description         string
	tools               []string
	parameters          []string
	interval            time.Duration
	enabled             bool
	lastExecutionTime   time.Time
	lastExecutionStatus constants.ExecutionStatus
}

// newDiskInfoCollector 创建磁盘信息收集插件
func newDiskInfoCollector() *DiskInfoCollector {
	return &DiskInfoCollector{
		name:                "disk_info_collector",
		description:         "Collect disk hardware information",
		tools:               []string{"lsblk"},
		parameters:          []string{"-J"}, // JSON格式输出
		interval:            120 * time.Second,
		enabled:             true,
		lastExecutionStatus: constants.StatusPending,
	}
}

// Name 获取插件名称
func (dic *DiskInfoCollector) Name() string {
	return dic.name
}

// Description 获取插件描述
func (dic *DiskInfoCollector) Description() string {
	return dic.description
}

// Tool 获取工具名称
func (dic *DiskInfoCollector) Tools() []string {
	return dic.tools
}

// Parameters 获取参数
func (dic *DiskInfoCollector) Parameters() []string {
	return dic.parameters
}

// Interval 获取执行间隔
func (dic *DiskInfoCollector) Interval() time.Duration {
	return dic.interval
}

// Execute 执行插件
func (dic *DiskInfoCollector) Execute() (interface{}, error) {
	dic.lastExecutionStatus = constants.StatusRunning
	dic.lastExecutionTime = time.Now()

	// 简化实现，实际应该执行命令获取磁盘信息
	result := map[string]interface{}{
		"disks": []map[string]interface{}{
			{
				"name":  "sda",
				"size":  "1000G",
				"type":  "SSD",
				"model": "Samsung SSD 860",
			},
		},
		"collected_at": time.Now().Unix(),
	}

	dic.lastExecutionStatus = constants.StatusSuccess
	return result, nil
}

// LastExecutionStatus 获取最后执行状态
func (dic *DiskInfoCollector) LastExecutionStatus() constants.ExecutionStatus {
	return dic.lastExecutionStatus
}

// LastExecutionTime 获取最后执行时间
func (dic *DiskInfoCollector) LastExecutionTime() time.Time {
	return dic.lastExecutionTime
}

// Enabled 获取启用状态
func (dic *DiskInfoCollector) Enabled() bool {
	return dic.enabled
}

// SetEnabled 设置启用状态
func (dic *DiskInfoCollector) SetEnabled(enabled bool) {
	dic.enabled = enabled
}

// Describe 描述指标
func (dic *DiskInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	// 插件特定指标描述
}

// Collect 收集指标
func (dic *DiskInfoCollector) Collect(ch chan<- prometheus.Metric) {
	// 插件特定指标收集
}
