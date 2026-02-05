package models

import (
	"baize-monitor/pkg/constants"
	"fmt"
	"net"
	"strings"
	"time"
)

type RawPacket struct {
	Data       []byte
	Conn       *net.UDPConn
	RemoteAddr *net.UDPAddr
}

// TrapMessage represents a processed SNMP trap/inform message
type TrapMessage struct {
	SnmpTrapOID string

	SourceType TrapSourceType
	// Basic information
	ReceivedAt time.Time // When the trap was received

	// Source information
	SourceIP   net.IP // Source IP address
	SourcePort int    // Source port

	// Variable bindings - the core data (common to all versions)
	VariableMap map[string]string // OID -> value mapping for easy access

	// Processing metadata
	RawData []byte // Original raw packet data
}

func (tm *TrapMessage) ParseVendorCode() (string, error) {
	if tm == nil || tm.VariableMap == nil {
		return "", fmt.Errorf("invalid TrapMessage: nil or empty VariableMap")
	}
	// 尝试标准trap OID键
	if val, ok := tm.VariableMap[constants.SnmpTrapOID]; ok {
		tm.SnmpTrapOID = val
		if vendorID, ok := extractVendorID(val); ok {
			return vendorID, nil
		}
	}

	// 统计所有OID键的企业ID分布（作为兜底方案）
	vendorCount := make(map[string]int)
	for oidKey := range tm.VariableMap {
		if vendorID, ok := extractVendorID(oidKey); ok {
			vendorCount[vendorID]++
		}
	}

	// 选择高频厂商ID
	var bestVendor string
	maxCount := 0
	for vendor, count := range vendorCount {
		if count > maxCount {
			maxCount = count
			bestVendor = vendor
		}
	}

	if maxCount == 0 {
		return "", fmt.Errorf("no enterprise OID found in %d bindings", len(tm.VariableMap))
	}
	return bestVendor, nil
}

// extractVendorID 从OID字符串提取厂商ID
func extractVendorID(oid string) (string, bool) {
	oid = strings.TrimSpace(oid)
	matches := constants.EnterpriseOIDRegex.FindStringSubmatch(oid)
	if len(matches) >= 2 {
		return matches[1], true // matches[1] = 厂商ID数字部分
	}
	return "", false

}

type TrapSourceType string

const (
	TrapSourceTypeBMC TrapSourceType = "BMC"
)
