package anomaly

import (
	"log"

	"baize-monitor/internal/agent/anomaly/types"

	"baize-monitor/internal/agent/anomaly/cpu"
	_ "baize-monitor/internal/agent/anomaly/cpu"
	"baize-monitor/internal/agent/anomaly/disk"
	_ "baize-monitor/internal/agent/anomaly/disk"
	"baize-monitor/internal/agent/anomaly/memory"
	_ "baize-monitor/internal/agent/anomaly/memory"
	"baize-monitor/internal/agent/anomaly/network"
	_ "baize-monitor/internal/agent/anomaly/network"
)

type Anomaly struct {
	registry *types.CheckerRegistry
}

func New(nodeName, targetIP string) *Anomaly {
	registry := types.NewCheckerRegistry()

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
