package collectors

import (
	"testing"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

func TestNewMemoryCollector(t *testing.T) {
	collector := NewMemoryCollector()
	if collector == nil {
		t.Fatal("Expected MemoryCollector to be created, got nil")
	}
}

func TestMemoryCollector_Collect(t *testing.T) {
	collector := NewMemoryCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if !hardwareInfo.Memory.Success {
		t.Error("Expected memory collection to succeed")
	}
}

func TestMemoryCollector_Collect_WithEmptyHardwareInfo(t *testing.T) {
	collector := NewMemoryCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if hardwareInfo.Memory.Message == "" {
		t.Error("Expected message to be set")
	}
}
