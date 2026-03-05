package anomaly

import (
	"log"

	"baize-monitor/internal/agent/anomaly/check_items"

	"baize-monitor/internal/agent/anomaly/check_items/cpu"
	_ "baize-monitor/internal/agent/anomaly/check_items/cpu/detectors"
	"baize-monitor/internal/agent/anomaly/check_items/disk"
	_ "baize-monitor/internal/agent/anomaly/check_items/disk/detectors"
	"baize-monitor/internal/agent/anomaly/check_items/memory"
	_ "baize-monitor/internal/agent/anomaly/check_items/memory/detectors"
	"baize-monitor/internal/agent/anomaly/check_items/network"
	_ "baize-monitor/internal/agent/anomaly/check_items/network/detectors"
)

type Anomaly struct {
	registry *check_items.CheckerRegistry
}

func New(nodeName, targetIP string) *Anomaly {
	registry := check_items.NewCheckerRegistry()

	// 注册所有检测器
	registry.Register(cpu.NewCPUChecker())
	registry.Register(memory.NewMemoryChecker())
	registry.Register(disk.NewDiskChecker())
	registry.Register(network.NewNetworkChecker())

	return &Anomaly{
		registry: registry,
	}
}

func (a *Anomaly) Start() error {
	log.Println("Anomaly module started successfully")
	return nil
}

func (a *Anomaly) Stop() error {
	log.Println("Anomaly module stopped")
	return nil
}

func (a *Anomaly) GetData() (interface{}, error) {
	anomalyReq := a.registry.CheckerALL()
	return anomalyReq, nil
}
