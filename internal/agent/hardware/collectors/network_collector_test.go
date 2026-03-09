package collectors

import (
	"testing"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

func TestNewNetworkCollector(t *testing.T) {
	collector := NewNetworkCollector()
	if collector == nil {
		t.Fatal("Expected NetworkCollector to be created, got nil")
	}
}

func TestNetworkCollector_Collect(t *testing.T) {
	collector := NewNetworkCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if !hardwareInfo.NetworkInterfaces.Success {
		t.Error("Expected network collection to succeed")
	}

	if len(hardwareInfo.NetworkInterfaces.Content) == 0 {
		t.Error("Expected network interfaces to be populated")
	}
}

func TestNetworkCollector_Collect_WithEmptyHardwareInfo(t *testing.T) {
	collector := NewNetworkCollector()
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	collector.Collect(hardwareInfo)

	if hardwareInfo.NetworkInterfaces.Message == "" {
		t.Error("Expected message to be set")
	}
}
