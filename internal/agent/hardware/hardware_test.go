package hardware

import (
	"testing"
)

func TestNew(t *testing.T) {
	hardware := New()
	if hardware == nil {
		t.Fatal("Expected hardware to be created, got nil")
	}
	if !hardware.enabled {
		t.Error("Expected hardware to be enabled by default")
	}
	if hardware.registry == nil {
		t.Error("Expected registry to be initialized")
	}
}

func TestHardware_Start(t *testing.T) {
	hardware := New()
	err := hardware.Start()
	if err != nil {
		t.Errorf("Expected Start() to return nil, got %v", err)
	}
}

func TestHardware_Stop(t *testing.T) {
	hardware := New()
	err := hardware.Stop()
	if err != nil {
		t.Errorf("Expected Stop() to return nil, got %v", err)
	}
}

func TestHardware_GetData(t *testing.T) {
	hardware := New()
	data, err := hardware.GetData()
	if err != nil {
		t.Errorf("Expected GetData() to return nil error, got %v", err)
	}
	if data == nil {
		t.Error("Expected GetData() to return data, got nil")
	}
}

func TestHardware_GetData_WhenDisabled(t *testing.T) {
	hardware := New()
	hardware.enabled = false
	data, err := hardware.GetData()
	if err != nil {
		t.Errorf("Expected GetData() to return nil error, got %v", err)
	}
	if data != nil {
		t.Error("Expected GetData() to return nil when disabled, got data")
	}
}
