package snmp

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	pkg_snmp "baize-monitor/pkg/snmp"
	"baize-monitor/pkg/storage"
	"fmt"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
)

// TrapHandler processes SNMP traps with distributed locking and multi-worker support
type TrapHandler struct {
	locker            storage.DistributedLockerInterface
	responseMgr       pkg_snmp.ResponseManagerInterface
	workersWg         sync.WaitGroup
	running           bool
	LockTimeout       time.Duration
	ProcessingTimeout time.Duration
	mu                sync.RWMutex
	inputChan         chan *models.RawPacket
	stopSign          chan struct{}
	httpClient        HTTPTrapSender
	cfg               *config.TrapHandlerConfig
}

// NewTrapHandler creates a new trap handler instance
func NewTrapHandler(
	dl storage.DistributedLockerInterface,
	responseMgr pkg_snmp.ResponseManagerInterface,
	inputChan chan *models.RawPacket,
	httpClient HTTPTrapSender,
	config *config.TrapHandlerConfig,
) *TrapHandler {

	return &TrapHandler{
		locker:      dl,
		responseMgr: responseMgr,
		httpClient:  httpClient,
		stopSign:    make(chan struct{}),
		inputChan:   inputChan,
		cfg:         config,
	}
}

// start starts the trap handler workers with provided channels
func (h *TrapHandler) start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("trap handler already started")
	}

	h.running = true

	if h.cfg.LockTimeout <= 0 {
		h.cfg.LockTimeout = 3
	}

	h.LockTimeout = time.Duration(h.cfg.LockTimeout) * time.Second

	if h.cfg.ProcessingTimeout <= 0 {
		h.cfg.ProcessingTimeout = 5
	}
	h.ProcessingTimeout = time.Duration(h.cfg.ProcessingTimeout) * time.Second

	// start worker goroutines
	workerCount := h.cfg.WorkerCount
	for i := 0; i < workerCount; i++ {
		h.workersWg.Add(1)
		go h.worker(i)
	}

	snmp_logger.Info("Trap handler started", "worker_count", workerCount)
	return nil
}

// stop stops the trap handler and all workers
func (h *TrapHandler) stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return fmt.Errorf("trap handler already stopped")
	}

	close(h.stopSign)
	h.workersWg.Wait()

	h.running = false
	snmp_logger.Info("Trap handler stopped")
	return nil
}

// worker processes traps from the input channel
func (h *TrapHandler) worker(id int) {
	defer h.workersWg.Done()

	snmp_logger.Debug("Trap worker started", "worker_id", id)

	for {
		select {
		case <-h.stopSign:
			snmp_logger.Debug("Trap worker stopping", "worker_id", id)
			return
		case rawPacket, ok := <-h.inputChan:
			if !ok {
				snmp_logger.Debug("Input channel closed, worker exiting", "worker_id", id)
				return
			}
			h.processTrap(rawPacket)
		}
	}
}

// processTrap processes a single trap with distributed locking
func (h *TrapHandler) processTrap(rawPacket *models.RawPacket) {
	startTime := time.Now()
	// Generate lock key for deduplication
	lockKey := h.locker.GenerateTrapLockKey(rawPacket.Data)

	acquired, err := h.locker.AcquireLock(lockKey, h.LockTimeout)
	if err != nil {
		snmp_logger.Error("Failed to acquire lock for trap", "remote_addr", rawPacket.RemoteAddr, "error", err)
		return
	}
	if !acquired {
		snmp_logger.Debug("Duplicate trap detected, skipping", "remote_addr", rawPacket.RemoteAddr)
		return
	}

	// response the trap
	snmpPacket, err := h.responseMgr.ResponseRequest(rawPacket)
	if err != nil {
		snmp_logger.Error("Failed to decode SNMP packet", "rawPacketData", rawPacket.Data, "remote_addr", rawPacket.RemoteAddr, "error", err)
		return
	}

	// Convert to TrapMessage
	trapMessage := h.convertToTrapMessage(snmpPacket, rawPacket)

	snmp_logger.Debug("Trap processed",
		"processing_time", time.Since(startTime),
		"source_ip", trapMessage.SourceIP)

	// 直接发送到告警HTTP接口
	h.httpClient.SendTrap(trapMessage)
}

// convertToTrapMessage converts gosnmp packet to our TrapMessage model with robust error handling
func (h *TrapHandler) convertToTrapMessage(
	snmpPacket *gosnmp.SnmpPacket,
	rawPacket *models.RawPacket,
) *models.TrapMessage {
	trap := &models.TrapMessage{
		SourceType:  models.TrapSourceTypeBMC,
		ReceivedAt:  time.Now(),
		SourceIP:    rawPacket.RemoteAddr.IP,
		SourcePort:  rawPacket.RemoteAddr.Port,
		RawData:     rawPacket.Data,
		VariableMap: make(map[string]string),
	}

	for _, pdu := range snmpPacket.Variables {
		trap.VariableMap[pdu.Name] = fmt.Sprintf("%v", pdu.Value)
	}

	return trap
}
