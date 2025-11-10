package models

import (
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
)

type RawPacket struct {
	Data       []byte
	Conn       *net.UDPConn
	RemoteAddr *net.UDPAddr
}

// TrapMessage represents a processed SNMP trap/inform message
type TrapMessage struct {
	// Basic information
	ReceivedAt  time.Time `json:"received_at"`  // When the trap was received
	ProcessedAt time.Time `json:"processed_at"` // When processing started

	// Source information
	SourceIP   net.IP `json:"source_ip"`   // Source IP address
	SourcePort int    `json:"source_port"` // Source port
	AgentType  string `json:"agent_type"`  // bmc, switch, machine-agent, etc.
	Hostname   string `json:"hostname"`    // Source hostname if available

	// SNMP protocol information
	Version       gosnmp.SnmpVersion `json:"version"`        // SNMP version
	Community     string             `json:"community"`      // Community string (v1/v2c)
	SecurityModel string             `json:"security_model"` // Security model (v3)
	UserName      string             `json:"user_name"`      // Security name (v3)
	PDUType       gosnmp.PDUType     `json:"pdu_type"`       // Trap, Inform, etc.
	RequestID     uint32             `json:"request_id"`     // Request ID for response

	// SNMP v1 specific fields
	V1EnterpriseOID string `json:"v1_enterprise_oid"` // Enterprise OID (v1 only)
	V1GenericTrap   int    `json:"v1_generic_trap"`   // Generic trap type (v1 only)
	V1SpecificTrap  int    `json:"v1_specific_trap"`  // Specific trap type (v1 only)

	// SNMP v2c/v3 specific fields
	V2cV3Timestamp uint32 `json:"v2c_v3_timestamp"` // SysUpTime when trap was generated (v2c/v3 only)

	// Variable bindings - the core data (common to all versions)
	VariableMap map[string]string `json:"variable_map"` // OID -> value mapping for easy access

	// Processing metadata
	RawData       []byte `json:"-"`              // Original raw packet data
	NeedsResponse bool   `json:"needs_response"` // Whether this needs a response (Inform)

	// Standardized fields after processing
	AlertType    string `json:"alert_type"`    // Standardized alert type
	Severity     string `json:"severity"`      // Standardized severity level
	Component    string `json:"component"`     // Affected component
	Message      string `json:"message"`       // Human readable message
	SerialNumber string `json:"serial_number"` // Hardware serial number
}
