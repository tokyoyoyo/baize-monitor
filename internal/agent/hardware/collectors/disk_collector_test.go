package collectors

import (
	"testing"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

func TestNewDiskCollector(t *testing.T) {
	collector := NewDiskCollector()
	if collector == nil {
		t.Fatal("Expected DiskCollector to be created, got nil")
	}
}

func TestDiskCollector_Collect(t *testing.T) {
	collector := NewDiskCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if !hardwareInfo.Disks.Success {
		t.Error("Expected disk collection to succeed")
	}
}

func TestDiskCollector_Collect_WithEmptyHardwareInfo(t *testing.T) {
	collector := NewDiskCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if hardwareInfo.Disks.Message == "" {
		t.Error("Expected message to be set")
	}
}
