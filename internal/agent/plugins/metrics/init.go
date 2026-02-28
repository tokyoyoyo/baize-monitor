package metrics

import "baize-monitor/internal/agent/core"

func init() {
	core.MetricsPlugins.Register(newLoadMonitor())
}
