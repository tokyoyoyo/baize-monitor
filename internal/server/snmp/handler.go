package snmp

import (
	"baize-monitor/pkg/models"
	pkg_snmp "baize-monitor/pkg/snmp"
	"baize-monitor/pkg/storage"
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
)

// TrapHandler processes SNMP traps with distributed locking and multi-worker support
type TrapHandler struct {
	locker      storage.DistributedLockerInterface
	responseMgr pkg_snmp.ResponseManagerInterface
	workersWg   sync.WaitGroup
	running     bool
	LockTimeout time.Duration
	mu          sync.RWMutex
	inputChan   chan *models.RawPacket
	outputChan  chan *models.TrapMessage
	stopChan    chan struct{}
}

// NewTrapHandler creates a new trap handler instance
func newTrapHandler(
	dl storage.DistributedLockerInterface,
	responseMgr pkg_snmp.ResponseManagerInterface,
	lt time.Duration,
	inputChan chan *models.RawPacket,
	outputChan chan *models.TrapMessage,
) *TrapHandler {

	return &TrapHandler{
		locker:      dl,
		responseMgr: responseMgr,
		LockTimeout: lt,
		stopChan:    make(chan struct{}),
		inputChan:   inputChan,
		outputChan:  outputChan,
	}
}

// start starts the trap handler workers with provided channels
func (h *TrapHandler) start(workerCount int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("trap handler already started")
	}

	h.running = true

	// start worker goroutines
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

	close(h.stopChan)
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
		case <-h.stopChan:
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
	// dont want use context here
	lockCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	acquired, err := h.locker.AcquireLock(lockCtx, lockKey, h.LockTimeout)
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
		"alert_type", trapMessage.AlertType,
		"source_ip", trapMessage.SourceIP)

	h.sendProcessedTrapMessage(trapMessage)
}

// convertToTrapMessage converts gosnmp packet to our TrapMessage model with robust error handling
func (h *TrapHandler) convertToTrapMessage(
	snmpPacket *gosnmp.SnmpPacket,
	rawPacket *models.RawPacket,
) *models.TrapMessage {
	trap := &models.TrapMessage{
		ReceivedAt:    time.Now(),
		SourceIP:      rawPacket.RemoteAddr.IP,
		SourcePort:    rawPacket.RemoteAddr.Port,
		Version:       snmpPacket.Version,
		Community:     snmpPacket.Community,
		PDUType:       snmpPacket.PDUType,
		RequestID:     snmpPacket.RequestID,
		RawData:       rawPacket.Data,
		NeedsResponse: snmpPacket.PDUType == gosnmp.InformRequest,
		VariableMap:   make(map[string]string),
	}

	// Extract version-specific fields with robust error handling
	switch snmpPacket.Version {
	case gosnmp.Version1:
		h.extractV1Fields(trap, snmpPacket)
	case gosnmp.Version2c, gosnmp.Version3:
		h.extractV2cV3Fields(trap, snmpPacket)
	}

	for _, pdu := range snmpPacket.Variables {
		trap.VariableMap[pdu.Name] = fmt.Sprintf("%v", pdu.Value)
	}

	// Extract v3 security information if applicable
	if snmpPacket.Version == gosnmp.Version3 {
		h.extractV3SecurityInfo(trap, snmpPacket)
	}

	return trap
}

// extractV1Fields extracts SNMP v1 specific fields with error tolerance
func (h *TrapHandler) extractV1Fields(trap *models.TrapMessage, snmpPacket *gosnmp.SnmpPacket) {
	if len(snmpPacket.Variables) < 4 {
		snmp_logger.Warn("SNMP v1 packet has insufficient variables, using defaults",
			"variable_count", len(snmpPacket.Variables),
			"remote_addr", trap.SourceIP.String())
		// Set defaults for v1 fields when variables are insufficient
		trap.V1GenericTrap = 0  // Default to coldstart (0)
		trap.V1SpecificTrap = 0 // Default to no specific trap
		return
	}

	// Extract Enterprise OID (Variables[0]) - handle various data types
	if enterprisePDU := snmpPacket.Variables[0]; enterprisePDU.Value != nil {
		switch v := enterprisePDU.Value.(type) {
		case []byte:
			// Try to convert bytes to OID string using gosnmp utility
			trap.V1EnterpriseOID = gosnmp.ToBigInt(v).String()
		case string:
			// Direct string value
			trap.V1EnterpriseOID = v
		case int, int64, uint, uint64:
			// Numeric OID representation
			trap.V1EnterpriseOID = fmt.Sprintf("%v", v)
		default:
			// Unknown type, log and use string representation
			snmp_logger.Warn("Unknown v1 enterprise OID type, using string representation",
				"value_type", fmt.Sprintf("%T", v),
				"value", fmt.Sprintf("%v", v))
			trap.V1EnterpriseOID = fmt.Sprintf("%v", v)
		}
	} else {
		snmp_logger.Warn("v1 enterprise OID is nil")
	}

	// Extract Generic Trap (Variables[2])
	if genericPDU := snmpPacket.Variables[2]; genericPDU.Value != nil {
		switch v := genericPDU.Value.(type) {
		case int:
			trap.V1GenericTrap = v
		case int64:
			trap.V1GenericTrap = int(v)
		case uint32:
			trap.V1GenericTrap = int(v)
		default:
			snmp_logger.Warn("Unknown v1 generic trap type, defaulting to 0",
				"value_type", fmt.Sprintf("%T", v),
				"value", fmt.Sprintf("%v", v))
			trap.V1GenericTrap = 0
		}
	} else {
		trap.V1GenericTrap = 0
	}

	// Extract Specific Trap (Variables[3])
	if specificPDU := snmpPacket.Variables[3]; specificPDU.Value != nil {
		switch v := specificPDU.Value.(type) {
		case int:
			trap.V1SpecificTrap = v
		case int64:
			trap.V1SpecificTrap = int(v)
		case uint32:
			trap.V1SpecificTrap = int(v)
		default:
			snmp_logger.Warn("Unknown v1 specific trap type, defaulting to 0",
				"value_type", fmt.Sprintf("%T", v),
				"value", fmt.Sprintf("%v", v))
			trap.V1SpecificTrap = 0
		}
	} else {
		trap.V1SpecificTrap = 0
	}
}

// extractV2cV3Fields extracts SNMP v2c/v3 specific fields with error tolerance
func (h *TrapHandler) extractV2cV3Fields(trap *models.TrapMessage, snmpPacket *gosnmp.SnmpPacket) {
	if len(snmpPacket.Variables) == 0 {
		snmp_logger.Warn("SNMP v2c/v3 packet has no variables, timestamp will be 0",
			"remote_addr", trap.SourceIP.String(),
			"version", snmpPacket.Version.String())
		return
	}

	// The first variable in v2c/v3 traps is typically sysUpTime.0 (1.3.6.1.2.1.1.3.0)
	// But some vendors might not follow this convention
	firstVar := snmpPacket.Variables[0]
	if firstVar.Value != nil {
		switch v := firstVar.Value.(type) {
		case uint32:
			trap.V2cV3Timestamp = v
		case uint64:
			// Check for overflow
			if v > math.MaxUint32 {
				snmp_logger.Warn("v2c/v3 sysUpTime exceeds uint32 range, truncating",
					"original_value", v)
				trap.V2cV3Timestamp = math.MaxUint32
			} else {
				trap.V2cV3Timestamp = uint32(v)
			}
		case int:
			if v < 0 {
				snmp_logger.Warn("v2c/v3 sysUpTime is negative, using 0",
					"original_value", v)
				trap.V2cV3Timestamp = 0
			} else {
				trap.V2cV3Timestamp = uint32(v)
			}
		case int64:
			if v < 0 {
				snmp_logger.Warn("v2c/v3 sysUpTime is negative, using 0",
					"original_value", v)
				trap.V2cV3Timestamp = 0
			} else if v > math.MaxUint32 {
				snmp_logger.Warn("v2c/v3 sysUpTime exceeds uint32 range, truncating",
					"original_value", v)
				trap.V2cV3Timestamp = math.MaxUint32
			} else {
				trap.V2cV3Timestamp = uint32(v)
			}
		default:
			snmp_logger.Warn("Unknown v2c/v3 sysUpTime type, defaulting to 0",
				"value_type", fmt.Sprintf("%T", v),
				"value", fmt.Sprintf("%v", v))
			trap.V2cV3Timestamp = 0
		}
	} else {
		snmp_logger.Warn("v2c/v3 sysUpTime value is nil, using 0")
		trap.V2cV3Timestamp = 0
	}
}

// extractV3SecurityInfo extracts v3 security information with error tolerance
func (h *TrapHandler) extractV3SecurityInfo(trap *models.TrapMessage, snmpPacket *gosnmp.SnmpPacket) {
	// Extract security level from MsgFlags
	switch gosnmp.SnmpV3MsgFlags(snmpPacket.MsgFlags) & 0x03 {
	case gosnmp.NoAuthNoPriv:
		trap.SecurityModel = "noAuthNoPriv"
	case gosnmp.AuthNoPriv:
		trap.SecurityModel = "authNoPriv"
	case gosnmp.AuthPriv:
		trap.SecurityModel = "authPriv"
	default:
		trap.SecurityModel = fmt.Sprintf("Unknown(%d)", snmpPacket.MsgFlags&0x03)
		snmp_logger.Warn("Unknown v3 security level",
			"msg_flags", snmpPacket.MsgFlags&0x03)
	}

	// Extract user name from security parameters if available
	if secParams := snmpPacket.SecurityParameters; secParams != nil {
		switch params := secParams.(type) {
		case *gosnmp.UsmSecurityParameters:
			trap.UserName = params.UserName
		default:
			snmp_logger.Warn("Unknown v3 security parameters type",
				"type", fmt.Sprintf("%T", params))
		}
	} else {
		snmp_logger.Debug("v3 security parameters is nil")
	}
}

func (h *TrapHandler) sendProcessedTrapMessage(trapMessage *models.TrapMessage) {
	select {
	case h.outputChan <- trapMessage:
		// Successfully sent
	case <-time.After(100 * time.Millisecond):
		snmp_logger.Warn("Output channel blocked, discarding trap result")
	}
}
