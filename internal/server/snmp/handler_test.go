package snmp

import (
	"baize-monitor/pkg/models"
	"context"
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

func (m *MockDistributedLocker) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockDistributedLocker) ReleaseLock(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockDistributedLocker) GenerateTrapLockKey(data []byte) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *MockDistributedLocker) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock ResponseManagerInterface
type MockResponseManager struct {
	mock.Mock
}

func (m *MockResponseManager) ResponseRequest(raw *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	args := m.Called(raw)
	return args.Get(0).(*gosnmp.SnmpPacket), args.Error(1)
}

// Helper to create a dummy RawPacket
func newTestRawPacket() *models.RawPacket {
	return &models.RawPacket{
		Data:       []byte("fake-snmp-data"),
		RemoteAddr: &net.UDPAddr{IP: net.ParseIP("192.168.1.100"), Port: 1620},
	}
}

func TestTrapHandler_start_stop(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)

	// start
	err := handler.start(2)
	assert.NoError(t, err)

	// Should not allow double start
	err = handler.start(1)
	assert.Error(t, err)

	// stop
	err = handler.stop()
	assert.NoError(t, err)

	// Should not allow double stop
	err = handler.stop()
	assert.Error(t, err)
}

func TestTrapHandler_ProcessTrap_Success(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	// Mock lock key generation - changed to use OnceOrMoreTimes() to allow multiple calls
	dl.On("GenerateTrapLockKey", raw.Data).Return("lock-key-123")
	// Mock lock acquisition
	dl.On("AcquireLock", mock.Anything, "lock-key-123", 5*time.Second).Return(true, nil).Once()
	// Mock successful response
	snmpPkt := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.Trap,
		Variables: []gosnmp.SnmpPDU{},
	}
	rm.On("ResponseRequest", raw).Return(snmpPkt, nil).Once()

	// Send trap
	inputChan <- raw

	// Wait for output
	select {
	case msg := <-outputChan:
		assert.Equal(t, "192.168.1.100", msg.SourceIP.String())
	case <-time.After(2000 * time.Millisecond):
		t.Fatal("Expected trap message not received")
	}

	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
}

func TestTrapHandler_ProcessTrap_DuplicateSkipped(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	dl.On("GenerateTrapLockKey", raw.Data).Return("dup-key").Once()
	dl.On("AcquireLock", mock.Anything, "dup-key", mock.Anything).Return(false, nil).Once()
	// responseMgr should NOT be called

	inputChan <- raw

	// Ensure nothing is sent to output
	select {
	case <-outputChan:
		t.Fatal("Should not output duplicate trap")
	case <-time.After(100 * time.Millisecond):
		// OK
	}

	dl.AssertExpectations(t)
	rm.AssertNotCalled(t, "ResponseRequest", mock.Anything)
}

func TestTrapHandler_ProcessTrap_LockError(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	dl.On("GenerateTrapLockKey", raw.Data).Return("lock-key").Once()
	dl.On("AcquireLock", mock.Anything, "lock-key", mock.Anything).Return(false, errors.New("redis down")).Once()

	inputChan <- raw

	// Should not output anything
	select {
	case <-outputChan:
		t.Fatal("Should not output on lock error")
	case <-time.After(100 * time.Millisecond):
		// OK
	}

	dl.AssertExpectations(t)
	rm.AssertNotCalled(t, "ResponseRequest", mock.Anything)
}

func TestTrapHandler_ProcessTrap_ResponseError(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	dl.On("GenerateTrapLockKey", raw.Data).Return("key").Once()
	dl.On("AcquireLock", mock.Anything, "key", mock.Anything).Return(true, nil).Once()
	rm.On("ResponseRequest", raw).Return((*gosnmp.SnmpPacket)(nil), errors.New("decode failed")).Once()

	inputChan <- raw

	// Should not output
	select {
	case <-outputChan:
		t.Fatal("Should not output on decode error")
	case <-time.After(100 * time.Millisecond):
		// OK
	}

	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
}

func TestTrapHandler_OutputChannelBlocked(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage) // unbuffered → will block

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)
	defer handler.stop()

	raw := newTestRawPacket()

	dl.On("GenerateTrapLockKey", raw.Data).Return("key").Once()
	dl.On("AcquireLock", mock.Anything, "key", mock.Anything).Return(true, nil).Once()
	snmpPkt := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.Trap,
		Variables: []gosnmp.SnmpPDU{},
	}
	rm.On("ResponseRequest", raw).Return(snmpPkt, nil).Once()

	// Send trap → output channel is blocked, should timeout and warn
	inputChan <- raw

	// Wait a bit to let it process and hit the timeout
	time.Sleep(150 * time.Millisecond)

	// No assertion on log, but test that it doesn't hang
	// If `sendProcessedTrapMessage` didn't have timeout, this test would deadlock
	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
}

func TestTrapHandler_MultipleWorkers(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 10)
	outputChan := make(chan *models.TrapMessage, 10)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(3)
	assert.NoError(t, err)
	defer handler.stop()

	snmpPkt := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.Trap,
		Variables: []gosnmp.SnmpPDU{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			raw := &models.RawPacket{
				Data:       []byte{byte(idx)},
				RemoteAddr: &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1620 + idx},
			}
			key := "key-" + string(raw.Data[0])
			dl.On("GenerateTrapLockKey", raw.Data).Return(key).Once()
			dl.On("AcquireLock", mock.Anything, key, mock.Anything).Return(true, nil).Once()
			rm.On("ResponseRequest", raw).Return(snmpPkt, nil).Once()

			inputChan <- raw
		}(i)
	}

	wg.Wait()

	// Collect all outputs
	received := make([]*models.TrapMessage, 0, 6)
	timeout := time.After(500 * time.Millisecond)
	for len(received) < 6 {
		select {
		case msg := <-outputChan:
			received = append(received, msg)
		case <-timeout:
			t.Fatalf("Only received %d messages, expected 6", len(received))
		}
	}

	assert.Len(t, received, 6)
	dl.AssertExpectations(t)
	rm.AssertExpectations(t)
}

func TestTrapHandler_InputChannelClosed(t *testing.T) {
	dl := new(MockDistributedLocker)
	rm := new(MockResponseManager)
	inputChan := make(chan *models.RawPacket, 1)
	outputChan := make(chan *models.TrapMessage, 1)

	handler := newTrapHandler(dl, rm, 5*time.Second, inputChan, outputChan)
	err := handler.start(1)
	assert.NoError(t, err)

	// Close input channel
	close(inputChan)

	// Worker should exit gracefully
	time.Sleep(50 * time.Millisecond)

	// stop should not hang
	err = handler.stop()
	assert.NoError(t, err)
}
