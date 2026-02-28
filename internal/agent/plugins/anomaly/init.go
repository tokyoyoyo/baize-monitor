package anomaly

import "baize-monitor/internal/agent/core"

func init() {
	core.AnomalyPlugins.Register(newDiskLifetimeDetector())
}
