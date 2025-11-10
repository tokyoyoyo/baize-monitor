package snmp

import (
	"baize-monitor/pkg/models"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUDPReceiver_NewReceiver tests receiver creation
func TestUDPReceiver_NewReceiver(t *testing.T) {
	t.Run("create receiver", func(t *testing.T) {
		receiver := newUDPReceiver(make(chan *models.RawPacket, 10))
		assert.NotNil(t, receiver)
		assert.NotNil(t, receiver.outputChan)
		assert.False(t, receiver.running)
		assert.Nil(t, receiver.listener)
	})
}

// TestUDPReceiver_StartStop tests basic start and stop functionality
func TestUDPReceiver_StartStop(t *testing.T) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 10))

	port := 8080 // Let system assign port

	t.Run("start receiver", func(t *testing.T) {
		err := receiver.start(port)
		assert.NoError(t, err)
		assert.True(t, receiver.isRunning())
		assert.NotNil(t, receiver.listener)
		assert.NotNil(t, receiver.listener.LocalAddr())

		err = receiver.start(port)
		assert.Error(t, err)

	})

	t.Run("stop receiver", func(t *testing.T) {
		err := receiver.stop()
		assert.NoError(t, err)

		assert.False(t, receiver.isRunning())
		err = receiver.stop()
		assert.Error(t, err)
	})
}

// TestUDPReceiver_PacketReception tests receiving UDP packets
func TestUDPReceiver_PacketReception(t *testing.T) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 10))

	port := 8081

	err := receiver.start(port)
	assert.NoError(t, err)
	defer receiver.stop()

	addr := receiver.listener.LocalAddr().(*net.UDPAddr)

	t.Run("receive single packet", func(t *testing.T) {
		conn, err := net.DialUDP("udp", nil, addr)
		assert.NoError(t, err)
		defer conn.Close()

		testData := []byte("test SNMP trap data")
		_, err = conn.Write(testData)
		assert.NoError(t, err)

		select {
		case packet := <-receiver.outputChan:
			assert.Equal(t, testData, packet.Data)
			assert.NotNil(t, packet.RemoteAddr)
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for packet")
		}
	})

	t.Run("receive multiple packets", func(t *testing.T) {
		conn, err := net.DialUDP("udp", nil, addr)
		assert.NoError(t, err)
		defer conn.Close()

		packets := [][]byte{
			[]byte("packet 1"),
			[]byte("packet 2"),
			[]byte("packet 3"),
		}

		for _, data := range packets {
			_, err = conn.Write(data)
			assert.NoError(t, err)
		}

		// Receive all packets
		for i := 0; i < len(packets); i++ {
			select {
			case packet := <-receiver.outputChan:
				expected := packets[i]
				assert.Equal(t, packet.Data, expected)
			case <-time.After(1 * time.Second):
				t.Errorf("timeout waiting for packet %d", i)
			}
		}
	})
}

// TestUDPReceiver_ChannelFull tests behavior when output channel is full
func TestUDPReceiver_ChannelFull(t *testing.T) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 1))

	port := 8082

	err := receiver.start(port)
	assert.NoError(t, err)
	defer receiver.stop()

	addr := receiver.listener.LocalAddr().(*net.UDPAddr)

	conn, err := net.DialUDP("udp", nil, addr)
	assert.NoError(t, err)
	defer conn.Close()

	// Send multiple packets rapidly to fill the channel
	for i := 0; i < 5; i++ {
		data := []byte(fmt.Sprintf("packet %d", i))
		_, err = conn.Write(data)
		assert.NoError(t, err)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Should only have one packet in the channel (due to small buffer)
	if len(receiver.outputChan) > 1 {
		t.Errorf("expected at most 1 packet in channel due to small buffer, got %d", len(receiver.outputChan))
	}
}

// TestUDPReceiver_MultipleClients tests multiple clients sending simultaneously
func TestUDPReceiver_MultipleClients(t *testing.T) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 100))

	port := 8083

	err := receiver.start(port)
	assert.NoError(t, err)
	defer receiver.stop()

	addr := receiver.listener.LocalAddr().(*net.UDPAddr)

	const numClients = 5
	const packetsPerClient = 10

	var wg sync.WaitGroup
	receivedPackets := make(chan *models.RawPacket, numClients*packetsPerClient)

	// Start packet consumer
	go func() {
		for i := 0; i < numClients*packetsPerClient; i++ {
			select {
			case pkt := <-receiver.outputChan:
				receivedPackets <- pkt
			case <-time.After(2 * time.Second):
				t.Logf("consumer timeout after %d packets", i)
				return
			}
		}
	}()

	// Start multiple clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.DialUDP("udp", nil, addr)
			assert.NoError(t, err)
			defer conn.Close()

			for j := 0; j < packetsPerClient; j++ {
				data := []byte(fmt.Sprintf("client %d packet %d", clientID, j))
				_, err = conn.Write(data)
				assert.NoError(t, err)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	receivedCount := 0
	expectedCount := numClients * packetsPerClient
	timeout := time.After(5 * time.Second)
	for receivedCount < expectedCount {
		select {
		case <-receivedPackets:
			receivedCount++
		case <-timeout:
			t.Fatalf("timeout waiting for packets, got %d/%d", receivedCount, expectedCount)
		}
	}
}

// TestUDPReceiver_StopDuringReceive tests stopping while receiving packets
func TestUDPReceiver_StopDuringReceive(t *testing.T) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 100))

	port := 8084

	err := receiver.start(port)
	assert.NoError(t, err)

	addr := receiver.listener.LocalAddr().(*net.UDPAddr)

	// Start sending packets in background
	go func() {
		conn, err := net.DialUDP("udp", nil, addr)
		assert.NoError(t, err)
		defer conn.Close()

		for i := 0; i < 100; i++ {
			if !receiver.isRunning() {
				break
			}
			data := []byte(fmt.Sprintf("packet %d", i))
			conn.Write(data)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait a bit then stop
	time.Sleep(50 * time.Millisecond)
	err = receiver.stop()
	assert.NoError(t, err)

	// Receiver should not be running
	if receiver.isRunning() {
		t.Error("receiver should not be running after stop")
	}
}

// BenchmarkUDPReceiver_Throughput benchmarks packet receiving throughput
func BenchmarkUDPReceiver_Throughput(b *testing.B) {
	receiver := newUDPReceiver(make(chan *models.RawPacket, 1000))

	port := 8085

	err := receiver.start(port)
	assert.NoError(b, err)
	defer receiver.stop()

	addr := receiver.listener.LocalAddr().(*net.UDPAddr)
	testData := []byte("benchmark packet data")

	// Start consumer to drain the channel
	go func() {
		for range receiver.outputChan {
			// Just consume packets
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn, err := net.DialUDP("udp", nil, addr)
		assert.NoError(b, err)
		defer conn.Close()

		for pb.Next() {
			_, err = conn.Write(testData)
			if err != nil {
				b.Fatalf("failed to send data: %v", err)
			}
		}
	})
}
