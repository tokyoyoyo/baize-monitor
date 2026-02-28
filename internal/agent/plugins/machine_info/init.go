package machine_info

import "baize-monitor/internal/agent/core"

func init() {
	core.MachineInfoPlugins.Register(newHostnameCollector())
}
