package snmp

import (
	"baize-monitor/pkg/config"
	logger "baize-monitor/pkg/logger"
	"baize-monitor/pkg/models"
	pkg_snmp "baize-monitor/pkg/snmp"
	"baize-monitor/pkg/storage"
	"context"
	"fmt"
	"sync"
	"time"
)

var snmp_logger = logger.Snmp_logger

const (
	// PipelineBufferScale 处理管道缓冲区缩放因子
	PipelineBufferScale = 2
)

// SNMPServer main SNMP server that coordinates all components
type SNMPServer struct {
	config            *config.SNMPServerConfig
	running           bool
	receiver          *udpReceiver
	handler           *TrapHandler
	distributedLocker storage.DistributedLockerInterface
	responseMgr       pkg_snmp.ResponseManagerInterface
	midChannel        chan *models.RawPacket
	outChannel        chan *models.TrapMessage
	mu                sync.RWMutex
}

// NewSNMPServer creates a new SNMP server instance
func NewSNMPServer(config *config.SNMPServerConfig,
	dl storage.DistributedLockerInterface,
	rm pkg_snmp.ResponseManagerInterface,
) (*SNMPServer, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if dl == nil {
		return nil, fmt.Errorf("distributed locker cannot be nil")
	}
	if rm == nil {
		return nil, fmt.Errorf("response manager cannot be nil")
	}

	return &SNMPServer{
		config:            config,
		running:           false,
		distributedLocker: dl,
		responseMgr:       rm,
	}, nil
}

// Start starts the SNMP server
func (s *SNMPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SNMP server already started")
	}

	outChannelSize := s.config.MidChannelSize * PipelineBufferScale
	s.midChannel = make(chan *models.RawPacket, s.config.MidChannelSize)
	s.outChannel = make(chan *models.TrapMessage, outChannelSize)

	// Create handler with lock timeout
	lockTimeout := time.Duration(s.config.TrapHandlerConf.LockTimeout)
	s.handler = newTrapHandler(
		s.distributedLocker,
		s.responseMgr,
		lockTimeout,
		s.midChannel,
		s.outChannel,
	)

	// Start trap handler first
	if err := s.handler.start(s.config.TrapHandlerConf.WorkerCount); err != nil {
		return fmt.Errorf("failed to start trap handler: %w", err)
	}

	s.receiver = newUDPReceiver(s.midChannel)

	// Start UDP receiver
	if err := s.receiver.start(int(s.config.ReceiverConf.Port)); err != nil {
		s.handler.stop()
		return fmt.Errorf("failed to start UDP receiver: %w", err)
	}

	s.running = true
	snmp_logger.Info("SNMP server started successfully",
		"port", s.config.ReceiverConf.Port,
		"workers", s.config.TrapHandlerConf.WorkerCount,
		"mid_channel_size", s.config.MidChannelSize,
		"out_channel_size", outChannelSize,
		"lock_timeout", lockTimeout)
	return nil
}

// Stop stops the SNMP server
func (s *SNMPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("SNMP server already stopped")
	}

	// Stop components in reverse order
	if err := s.receiver.stop(); err != nil {
		snmp_logger.Error("Error stopping receiver", "error", err)
	}

	for {
		if len(s.midChannel) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err := s.handler.stop(); err != nil {
		snmp_logger.Error("Error stopping handler", "error", err)
	}

	s.running = false
	snmp_logger.Info("SNMP server stopped")
	return nil
}
