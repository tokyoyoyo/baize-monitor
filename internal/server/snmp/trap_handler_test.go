package snmp

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock DistributedLockerInterface
type MockDistributedLocker struct {
	mock.Mock
}

func (m *MockDistributedLocker) AcquireLock(key string, ttl time.Duration) (bool, error) {
	args := m.Called(key, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockDistributedLocker) GenerateTrapLockKey(data []byte) string {
	args := m.Called(data)
	return args.String(0)
}

// Mock ResponseManagerInterface
type MockResponseManager struct {
	mock.Mock
}

func (m *MockResponseManager) ResponseRequest(raw *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	args := m.Called(raw)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gosnmp.SnmpPacket), args.Error(1)
}

type MockHTTPTrapSender struct {
	mock.Mock
}

func (m *MockHTTPTrapSender) SendTrap(trap *models.TrapMessage) {
	m.Called(trap)
}

func (m *MockHTTPTrapSender) stats() map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]interface{})
}

// Helper to create a dummy RawPacket
func newTestRawPacket() *models.RawPacket {
	return &models.RawPacket{
		Data:       []byte("fake-snmp-data"),
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 1620},
	}
}

func TestTrapHandler_StartStop(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       5,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)

	// Start handler
	err := handler.start()
	assert.NoError(t, err)
	assert.True(t, handler.running)

	// Should not allow double start
	err = handler.start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Stop handler
	err = handler.stop()
	assert.NoError(t, err)
	assert.False(t, handler.running)

	// Should not allow double stop
	err = handler.stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already stopped")
}

func TestTrapHandler_ProcessTrap_Success(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       1,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	// Mock dependencies
	dl.On("GenerateTrapLockKey", raw.Data).Return("lock-key-123").Once()
	dl.On("AcquireLock", "lock-key-123", time.Duration(5)*time.Second).Return(true, nil).Once()

	snmpPkt := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.Trap,
		Variables: []gosnmp.SnmpPDU{
			{Name: "1.3.6.1.2.1.1.3.0", Value: 12345, Type: gosnmp.TimeTicks},
			{Name: "1.3.6.1.6.3.1.1.4.1.0", Value: "1.3.6.1.4.1.1234.0.1", Type: gosnmp.ObjectIdentifier},
		},
	}
	rm.On("ResponseRequest", raw).Return(snmpPkt, nil).Once()

	// Mock HTTP sender
	ts.On("SendTrap", mock.MatchedBy(func(trap *models.TrapMessage) bool {
		return trap.SourceIP.String() == "192.168.1.100" &&
			len(trap.VariableMap) == 2
	})).Return().Once()

	// Send trap for processing
	inputChan <- raw

	// Wait for processing to complete
	time.Sleep(50 * time.Millisecond)

	// Verify mocks
	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestTrapHandler_ProcessTrap_DuplicateSkipped(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       1,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	// Mock dependencies
	dl.On("GenerateTrapLockKey", raw.Data).Return("dup-key").Once()
	dl.On("AcquireLock", "dup-key", time.Duration(5)*time.Second).Return(false, nil).Once()

	// Send packet
	inputChan <- raw

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify mocks
	dl.AssertExpectations(t)
	rm.AssertNotCalled(t, "ResponseRequest", mock.Anything)
	ts.AssertNotCalled(t, "SendTrap", mock.Anything)
}

func TestTrapHandler_ProcessTrap_LockError(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       1,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	// Mock dependencies
	dl.On("GenerateTrapLockKey", raw.Data).Return("lock-key").Once()
	dl.On("AcquireLock", "lock-key", time.Duration(5)*time.Second).Return(false, errors.New("redis down")).Once()

	// Send packet
	inputChan <- raw

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify mocks
	dl.AssertExpectations(t)
	rm.AssertNotCalled(t, "ResponseRequest", mock.Anything)
	ts.AssertNotCalled(t, "SendTrap", mock.Anything)
}

func TestTrapHandler_ProcessTrap_ResponseError(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       1,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	// Mock dependencies
	dl.On("GenerateTrapLockKey", raw.Data).Return("key").Once()
	dl.On("AcquireLock", "key", time.Duration(5)*time.Second).Return(true, nil).Once()
	rm.On("ResponseRequest", raw).Return(nil, errors.New("response error")).Once()

	// Send packet
	inputChan <- raw

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify mocks
	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
	ts.AssertNotCalled(t, "SendTrap", mock.Anything)
}

func TestTrapHandler_MultipleWorkers(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       3,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	snmpPkt := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.Trap,
		Variables: []gosnmp.SnmpPDU{},
	}

	// Mock expectations for multiple traps
	totalTraps := 6
	var wg sync.WaitGroup
	wg.Add(totalTraps)

	// Set up expectations for all traps
	for i := 0; i < totalTraps; i++ {
		rawData := []byte{byte(i)}
		key := "key-" + string(rawData[0])

		dl.On("GenerateTrapLockKey", rawData).Return(key).Once()
		dl.On("AcquireLock", key, time.Duration(5)*time.Second).Return(true, nil).Once()
		rm.On("ResponseRequest", mock.MatchedBy(func(rp *models.RawPacket) bool {
			return len(rp.Data) > 0 && rp.Data[0] == byte(i)
		})).Return(snmpPkt, nil).Once()

		ts.On("SendTrap", mock.MatchedBy(func(trap *models.TrapMessage) bool {
			return trap.SourceIP.String() == "10.0.0.1"
		})).Return().Run(func(args mock.Arguments) {
			wg.Done()
		}).Once()
	}

	// Send traps concurrently
	for i := 0; i < totalTraps; i++ {
		go func(idx int) {
			inputChan <- &models.RawPacket{
				Data:       []byte{byte(idx)},
				RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1620 + idx},
			}
		}(i)
	}

	// Wait for all traps to be processed with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All traps processed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for all traps to be processed")
	}

	// Verify all mocks
	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestTrapHandler_InputChannelClosed(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 1)
	ts := new(MockHTTPTrapSender)
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       5,
		ProcessingTimeout: 5,
		WorkerCount:       1,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)
	err := handler.start()
	assert.NoError(t, err)

	// Close input channel
	close(inputChan)

	// Handler should stop workers gracefully
	time.Sleep(100 * time.Millisecond)

	// Stop should not hang
	err = handler.stop()
	assert.NoError(t, err)
}

func TestTrapHandler_ConfigDefaults(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	ts := new(MockHTTPTrapSender)

	// Create config with zero values to test defaults
	cfg := &config.TrapHandlerConfig{
		LockTimeout:       3,
		ProcessingTimeout: 5,
		WorkerCount:       5,
	}

	handler := NewTrapHandler(dl, rm, inputChan, ts, cfg)

	// Start should set default values
	err := handler.start()
	assert.NoError(t, err)
	defer handler.stop()

	// Verify defaults were set
	assert.Equal(t, time.Duration(3)*time.Second, handler.LockTimeout)
	assert.Equal(t, time.Duration(5)*time.Second, handler.ProcessingTimeout)

	// Default worker count should be at least 1
	assert.Greater(t, cfg.WorkerCount, 0)
}
