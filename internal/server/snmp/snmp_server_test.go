package snmp

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"

	"context"
	"net"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDistributedLocker is a mock implementation of distributed lock
type mockDistributedLocker struct {
}

func (m *mockDistributedLocker) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return true, nil
}

func (m *mockDistributedLocker) ReleaseLock(ctx context.Context, key string) error {
	return nil
}

func (m *mockDistributedLocker) GenerateTrapLockKey(trapData []byte) string {
	return "mock_key"
}

// Add Close method to implement DistributedLockerInterface interface
func (m *mockDistributedLocker) Close() error {
	return nil
}

// mockResponseManager is a mock implementation of response manager
type mockResponseManager struct {
}

func (m *mockResponseManager) ResponseRequest(rawPacket *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	return &gosnmp.SnmpPacket{}, nil
}

func TestNewSNMPServer(t *testing.T) {
	dl := &mockDistributedLocker{}
	rm := &mockResponseManager{}

	t.Run("valid config", func(t *testing.T) {
		cfg := &config.SNMPServerConfig{
			ReceiverConf: &config.ReceiverConfig{
				Port: 0,
			},
			TrapHandlerConf: &config.TrapHandlerConfig{
				WorkerCount: 2,
				LockTimeout: 30,
			},
			MidChannelSize: 100,
		}

		server, err := NewSNMPServer(cfg, dl, rm)
		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.False(t, server.running)
	})

	t.Run("nil config", func(t *testing.T) {
		server, err := NewSNMPServer(nil, dl, rm)
		assert.Error(t, err)
		assert.Nil(t, server)
	})

	t.Run("nil distributed locker", func(t *testing.T) {
		cfg := &config.SNMPServerConfig{
			ReceiverConf: &config.ReceiverConfig{
				Port: 0,
			},
			TrapHandlerConf: &config.TrapHandlerConfig{
				WorkerCount: 2,
				LockTimeout: 30,
			},
			MidChannelSize: 100,
		}

		server, err := NewSNMPServer(cfg, nil, rm)
		assert.Error(t, err)
		assert.Nil(t, server)
	})

	t.Run("nil response manager", func(t *testing.T) {
		cfg := &config.SNMPServerConfig{
			ReceiverConf: &config.ReceiverConfig{
				Port: 0,
			},
			TrapHandlerConf: &config.TrapHandlerConfig{
				WorkerCount: 2,
				LockTimeout: 30,
			},
			MidChannelSize: 100,
		}

		server, err := NewSNMPServer(cfg, dl, nil)
		assert.Error(t, err)
		assert.Nil(t, server)
	})
}

func TestSNMPServer_Start(t *testing.T) {
	dl := &mockDistributedLocker{}
	rm := &mockResponseManager{}

	cfg := &config.SNMPServerConfig{
		ReceiverConf: &config.ReceiverConfig{
			Port: 0,
		},
		TrapHandlerConf: &config.TrapHandlerConfig{
			WorkerCount: 2,
			LockTimeout: 30,
		},
		MidChannelSize: 100,
	}

	server, err := NewSNMPServer(cfg, dl, rm)
	require.NoError(t, err)
	require.NotNil(t, server)

	t.Run("start server", func(t *testing.T) {
		ctx := context.Background()
		err := server.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, server.running)
		assert.NotNil(t, server.midChannel)
		assert.NotNil(t, server.outChannel)
	})

	t.Run("start already started server", func(t *testing.T) {
		ctx := context.Background()
		err := server.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already started")
	})
}

func TestSNMPServer_Stop(t *testing.T) {
	dl := &mockDistributedLocker{}
	rm := &mockResponseManager{}

	cfg := &config.SNMPServerConfig{
		ReceiverConf: &config.ReceiverConfig{
			Port: 0,
		},
		TrapHandlerConf: &config.TrapHandlerConfig{
			WorkerCount: 2,
			LockTimeout: 30,
		},
		MidChannelSize: 100,
	}

	server, err := NewSNMPServer(cfg, dl, rm)
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx := context.Background()
	err = server.Start(ctx)
	require.NoError(t, err)
	require.True(t, server.running)

	t.Run("stop server", func(t *testing.T) {
		err := server.Stop()
		assert.NoError(t, err)
		assert.False(t, server.running)
	})

	t.Run("stop already stopped server", func(t *testing.T) {
		err := server.Stop()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already stopped")
	})

	t.Run("test receiver will not continue receive after close", func(t *testing.T) {

		// Create a UDP connection to test if data can still be sent
		addr := server.receiver.listener.LocalAddr().(*net.UDPAddr)
		conn, err := net.DialUDP("udp", nil, addr)
		require.NoError(t, err)
		defer conn.Close()

		// Try to send data
		testData := []byte("test data after close")
		_, err = conn.Write(testData)
		assert.NoError(t, err)

		// Wait for a short time to ensure processing is complete
		time.Sleep(100 * time.Millisecond)

		// Check if there is data in the channel (should be none)
		select {
		case <-server.midChannel:
			t.Error("Expected no data after receiver closed")
		default:
			// Correct case, no data received
		}
	})

	t.Run("test handler will process all data in channel after stop", func(t *testing.T) {
		// Create a new server instance for this test
		newServer, err := NewSNMPServer(cfg, dl, rm)
		require.NoError(t, err)
		require.NotNil(t, newServer)

		ctx := context.Background()
		err = newServer.Start(ctx)
		require.NoError(t, err)
		require.True(t, newServer.running)

		// Create some test data and send it to midChannel
		testData1 := &models.RawPacket{
			Data:       []byte("test trap 1"),
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		}

		testData2 := &models.RawPacket{
			Data:       []byte("test trap 2"),
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		}

		// Send test data to midChannel
		newServer.midChannel <- testData1
		newServer.midChannel <- testData2

		// Stop the server
		err = newServer.Stop()
		assert.NoError(t, err)
		assert.False(t, newServer.running)

		// Wait for some time to ensure processing is complete
		time.Sleep(100 * time.Millisecond)

		// Check if there is processed data in the output channel
		// Should be able to get the two pieces of data previously sent to midChannel
		count := 0
		timeout := time.After(1 * time.Second)
		for count < 2 {
			select {
			case trapMsg := <-newServer.outChannel:
				assert.NotNil(t, trapMsg)
				count++
			case <-timeout:
				// Timeout, check if the expected amount of data was received
				t.Fatalf("Expected 2 trap messages, got %d", count)
			}
		}
	})
}
