package hardware

import "baize-monitor/internal/agent/core"

func init() {
	core.HardwarePlugins.Register(newDiskInfoCollector())
}
