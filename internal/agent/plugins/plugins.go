// Package plugins 包含所有监控插件的自动注册机制
// 通过导入此包，会自动注册所有子目录中的插件
package plugins

import (
	// 导入所有插件子包以触发它们的init函数
	_ "baize-monitor/internal/agent/plugins/anomaly"
	_ "baize-monitor/internal/agent/plugins/hardware"
	_ "baize-monitor/internal/agent/plugins/machine_info"
	_ "baize-monitor/internal/agent/plugins/metrics"
)
