package snmp

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/constants"
	logger "baize-monitor/pkg/logger"
	"baize-monitor/pkg/models"
	pkg_snmp "baize-monitor/pkg/snmp"
	"baize-monitor/pkg/storage"
	"fmt"
	"time"
)

// TODO 调整
var snmp_logger = logger.Snmp_logger

// SNMPServer main SNMP server that coordinates all components
type SNMPServer struct {
	running    bool
	receiver   *UdpReceiver
	handler    *TrapHandler
	midChannel chan *models.RawPacket
}

// NewSNMPServer creates a new SNMP server instance
func NewSNMPServer(config *config.ServerConfig, locker storage.DistributedLockerInterface, responseMgr pkg_snmp.ResponseManagerInterface) (*SNMPServer, error) {
	alertConfig := config.AlertServerConfig
	snmpConfig := config.SNMPServerConfig

	if alertConfig.Port <= 0 || alertConfig.Port > 65535 {
		return nil, fmt.Errorf("invalid alert server port: %d (must be 1-65535)", alertConfig.Port)
	}

	alertHost := "127.0.0.1"
	alertURL := fmt.Sprintf("http://%s:%d%s",
		alertHost,
		alertConfig.Port,
		constants.AlertUploadFullPathV1,
	)
	httpClient := NewSimpleHTTPTrapSenderIpmi(alertURL)

	if snmpConfig.MidChannelSize <= 0 {
		snmpConfig.MidChannelSize = 10000
	}
	midChannel := make(chan *models.RawPacket, snmpConfig.MidChannelSize)

	udpReceiver := NewUDPReceiver(*snmpConfig.ReceiverConf, midChannel)

	trapTrapHandler := NewTrapHandler(
		locker,
		responseMgr,
		midChannel,
		httpClient,
		snmpConfig.TrapHandlerConf,
	)

	return &SNMPServer{
		running:    false,
		receiver:   udpReceiver,
		handler:    trapTrapHandler,
		midChannel: midChannel,
	}, nil
}

// Start starts the SNMP server
func (s *SNMPServer) Start() error {
	if s.running {
		return fmt.Errorf("SNMP server already started")
	}

	// Start trap handler first
	if err := s.handler.start(); err != nil {
		return fmt.Errorf("failed to start trap handler: %w", err)
	}

	snmp_logger.Info("SNMP server trap handler started",
		"worker_count", s.handler.cfg.WorkerCount)

	// Start UDP receiver
	if err := s.receiver.start(); err != nil {
		s.handler.stop()
		return fmt.Errorf("failed to start UDP receiver: %w", err)
	}
	snmp_logger.Info("SNMP server UDP receiver started",
		"port", s.receiver.Port)

	s.running = true
	snmp_logger.Info("SNMP server started successfully")
	return nil
}

// Stop stops the SNMP server
func (s *SNMPServer) Stop() error {
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
