package exporters

import "baize-monitor/internal/agent/core"

func init() {
	core.GlobalExporterRegistry.Register(newAnomalyExporter())
	core.GlobalExporterRegistry.Register(newHardwareExporter())
	core.GlobalExporterRegistry.Register(newMachineInfoExporter())
	core.GlobalExporterRegistry.Register(newMetricsExporter())
}
