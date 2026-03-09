package types

import (
	"testing"
	"time"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

type MockCollector struct {
	collectCalled bool
}

func (m *MockCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	m.collectCalled = true
	hardwareInfo.CPUs = hardwareRequest.CPURequest{
		Success: true,
		Message: "mock collected",
	}
}

func TestNewCollectorRegistry(t *testing.T) {
	registry := NewCollectorRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}
	if len(registry.collectors) != 0 {
		t.Errorf("Expected empty registry, got %d collectors", len(registry.collectors))
	}
}

func TestCollectorRegistry_Register(t *testing.T) {
	registry := NewCollectorRegistry()
	mockCollector := &MockCollector{}

	registry.Register(mockCollector)

	if len(registry.collectors) != 1 {
		t.Errorf("Expected 1 collector, got %d", len(registry.collectors))
	}
}

func TestCollectorRegistry_CollectAll(t *testing.T) {
	registry := NewCollectorRegistry()
	mockCollector := &MockCollector{}

	registry.Register(mockCollector)

	hardwareInfo := registry.CollectAll()

	if !mockCollector.collectCalled {
		t.Error("Expected collector to be called")
	}

	if !hardwareInfo.CPUs.Success {
		t.Error("Expected successful collection")
	}

	if hardwareInfo.CPUs.Message != "mock collected" {
		t.Errorf("Expected 'mock collected', got '%s'", hardwareInfo.CPUs.Message)
	}
}

func TestCollectorRegistry_CollectAllWithTimestamp(t *testing.T) {
	registry := NewCollectorRegistry()
	mockCollector := &MockCollector{}

	registry.Register(mockCollector)

	before := time.Now()
	hardwareInfo := registry.CollectAllWithTimestamp()
	after := time.Now()

	if hardwareInfo.CollectedAt.Before(before) || hardwareInfo.CollectedAt.After(after) {
		t.Error("Expected CollectedAt to be set to current time")
	}
}

func TestCollectorRegistry_MultipleCollectors(t *testing.T) {
	registry := NewCollectorRegistry()

	mockCollector1 := &MockCollector{}
	mockCollector2 := &MockCollector{}

	registry.Register(mockCollector1)
	registry.Register(mockCollector2)

	registry.CollectAll()

	if !mockCollector1.collectCalled {
		t.Error("Expected first collector to be called")
	}

	if !mockCollector2.collectCalled {
		t.Error("Expected second collector to be called")
	}
}
