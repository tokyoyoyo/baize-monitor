package snmp

import (
	"baize-monitor/pkg/models"
	"fmt"
	"net"
	"sync"
	"time"
)

// udpReceiver SNMP Trap receiver
type udpReceiver struct {
	listener *net.UDPConn
	locker   sync.RWMutex
	running  bool
	wg       sync.WaitGroup

	outputChan chan *models.RawPacket
}

// newUDPReceiver Create a new UDP receiver
func newUDPReceiver(oc chan *models.RawPacket) *udpReceiver {
	return &udpReceiver{
		running:    false,
		outputChan: oc,
	}
}

// start starts the UDP receiver and begins sending packets to the output channel
func (r *udpReceiver) start(port int) error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if r.running {
		return fmt.Errorf("UDP receiver already started")
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %w", port, err)
	}
	r.listener = listener

	r.running = true
	r.wg.Add(1)
	go r.receiveLoop()

	snmp_logger.Info("UDP receiver started", "port", port)
	return nil
}

// receiveLoop continuously reads UDP packets and sends them to the output channel
func (r *udpReceiver) receiveLoop() {
	defer r.wg.Done()

	buffer := make([]byte, 65507) // Maximum UDP packet size

	for {
		// Check if we should stop
		if !r.isRunning() {
			snmp_logger.Debug("UDP receiver loop stopping")
			return
		}

		// Set read timeout to allow periodic checking of running status
		r.listener.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, remoteAddr, err := r.listener.ReadFromUDP(buffer)
		if err != nil {
			// Check if it's a timeout (expected for periodic status checking)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			// Check if we should stop (listener was closed)
			if !r.isRunning() {
				snmp_logger.Debug("UDP receiver loop stopping due to closed listener")
				return
			}

			// Brief pause before retrying after an error
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Create a copy of the data to avoid buffer reuse issues
		packetData := make([]byte, n)
		copy(packetData, buffer[:n])

		packet := &models.RawPacket{
			Data:       packetData,
			RemoteAddr: remoteAddr,
			Conn:       r.listener,
		}

		if r.running {
			// Send the packet to the output channel
			select {
			case r.outputChan <- packet:
				snmp_logger.Debug("Received bytes", "bytes", n, "from", remoteAddr.String())
			default:
				// Channel is full, log warning and discard packet
				snmp_logger.Warn("Output channel full, discarding packet", "from", remoteAddr.String())
			}
		}
	}
}

// stop stops the UDP receiver
func (r *udpReceiver) stop() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	if !r.running {
		return fmt.Errorf("UDP receiver already stopped")
	}

	r.running = false
	snmp_logger.Info("Stopping UDP receiver...")

	// Close the listener to break out of ReadFromUDP
	if r.listener != nil {
		if err := r.listener.Close(); err != nil {
			snmp_logger.Error("Error closing UDP listener", "error", err)
			return err
		}
	}

	// Wait for receiveLoop to exit with timeout
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		snmp_logger.Debug("UDP receiver loop stopped gracefully")
	case <-time.After(5 * time.Second):
		snmp_logger.Warn("Timeout waiting for UDP receiver loop to stop")
	}

	snmp_logger.Info("UDP receiver stopped")

	return nil
}

// isRunning returns whether the receiver is running
func (r *udpReceiver) isRunning() bool {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.running
}
