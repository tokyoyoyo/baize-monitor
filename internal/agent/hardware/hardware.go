package hardware

import (
	"baize-monitor/internal/agent/hardware/collectors"
	"baize-monitor/internal/agent/hardware/types"
)

type Hardware struct {
	registry *types.CollectorRegistry
	enabled  bool
}

func New() *Hardware {
	registry := types.NewCollectorRegistry()

	registry.Register(collectors.NewCPUCollector())
	registry.Register(collectors.NewMemoryCollector())
	registry.Register(collectors.NewDiskCollector())
	registry.Register(collectors.NewNetworkCollector())

	return &Hardware{
		registry: registry,
		enabled:  true,
	}
}

func (h *Hardware) GetData() (interface{}, error) {
	if !h.enabled {
		return nil, nil
	}

	hardwareInfo := h.registry.CollectAllWithTimestamp()

	return hardwareInfo, nil
}

func (h *Hardware) Start() error {
	return nil
}

func (h *Hardware) Stop() error {
	return nil
}
