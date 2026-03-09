package collectors

import (
	"testing"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

func TestNewCPUCollector(t *testing.T) {
	collector := NewCPUCollector()
	if collector == nil {
		t.Fatal("Expected CPUCollector to be created, got nil")
	}
}

func TestCPUCollector_Collect(t *testing.T) {
	collector := NewCPUCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if !hardwareInfo.CPUs.Success {
		t.Error("Expected CPU collection to succeed")
	}

	if len(hardwareInfo.CPUs.Content) == 0 {
		t.Error("Expected CPU content to be populated")
	}
}

func TestCPUCollector_Collect_WithEmptyHardwareInfo(t *testing.T) {
	collector := NewCPUCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if hardwareInfo.CPUs.Message == "" {
		t.Error("Expected message to be set")
	}
}
