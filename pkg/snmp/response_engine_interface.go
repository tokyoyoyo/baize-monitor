package snmp

import (
	"baize-monitor/pkg/models"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

// Error constants definition
var (
	ErrNilRequest         = errors.New("request is nil")
	ErrUnsupportedVersion = errors.New("unsupported SNMP version")
	ErrParseSNMPVersion   = errors.New("unable to parse SNMP version")
	ErrIligelVersion      = errors.New("ililegal version")
	ErrDecodeSNMPrequest  = errors.New("unable to decode SNMP request")

	ErrVersionMissMatch = errors.New("version mismatch")
	ErrDisableVersion   = errors.New("disable SNMP version")

	ErrV1InformNotSupported = errors.New("SNMP v1 does not support Inform requests")
	ErrV1TrapNotSupported   = errors.New("SNMP v1 Trap does not require a response")

	ErrInvalidInformRequest = errors.New("not an Inform request")

	ErrEmptyCommunity = errors.New("community is empty")

	ErrInvalidCommunity   = errors.New("invalid community")
	ErrUnsupportedPDUType = errors.New("unsupported PDU type")
)

type SNMPResponseEngine interface {
	//
	ProcessRequest(rp *models.RawPacket) (*gosnmp.SnmpPacket, error)

	// 版本和配置信息
	GetVersion() gosnmp.SnmpVersion
	GetEngineID() string
	Enable() bool
}

// ResponseManagerInterface Response manager interface
type ResponseManagerInterface interface {
	ResponseRequest(rawPacket *models.RawPacket) (*gosnmp.SnmpPacket, error)
}

func generateV3EngineID() string {
	// 简单版本：时间戳 + 随机数
	timestamp := time.Now().UnixNano()
	randomPart := make([]byte, 8)
	rand.Read(randomPart)

	data := fmt.Sprintf("v3-engine-%d-%x", timestamp, randomPart)

	// 确保长度在 RFC 3411 要求范围内 (5-32字节)
	if len(data) > 32 {
		data = data[:32]
	}

	// 转换为十六进制字符串
	return hex.EncodeToString([]byte(data))
}

func generateV1EngineID() string {
	// 简单版本：时间戳 + 随机数
	timestamp := time.Now().UnixNano()
	randomPart := make([]byte, 4)
	rand.Read(randomPart)

	data := fmt.Sprintf("v1-engine-%d-%x", timestamp, randomPart)

	// 编码为十六进制字符串
	// return hex.EncodeToString([]byte(data))
	return data
}

func generateV2cEngineID() string {
	// 简单版本：时间戳 + 随机数
	timestamp := time.Now().UnixNano()
	randomPart := make([]byte, 4)
	rand.Read(randomPart)

	data := fmt.Sprintf("v2c-engine-%d-%x", timestamp, randomPart)

	// 编码为十六进制字符串
	// return hex.EncodeToString([]byte(data))
	return data
}

// generateEngineID 生成 SNMP 引擎标识
// 对于 SNMPv1/v2c，这仅用于内部标识
// 对于 SNMPv3，这应该符合 RFC 3411 标准
func generateEngineID(version gosnmp.SnmpVersion) string {
	switch version {
	case gosnmp.Version3:
		return generateV3EngineID()
	case gosnmp.Version1:
		// v1/v2c
		return generateV1EngineID()
	case gosnmp.Version2c:
		// v1/v2c 使用简化版本
		return generateV2cEngineID()
	default:
		return "baize-snmp-default-engine"
	}
}

// 安全模型抽象
type SecurityModelInterface interface {
	Authenticate(packet *gosnmp.SnmpPacket) error
	Encrypt(packet *gosnmp.SnmpPacket) error
	Decrypt(packet *gosnmp.SnmpPacket) error
	CheckAccess(oid string, operation gosnmp.PDUType) (bool, error)
}

// 社区字符串安全模型 (v1/v2c)
type CommunitySecurityModel struct {
	readCommunity      string
	readWriteCommunity string
}

func NewCommunitySecurityModel(rc, rwc string) *CommunitySecurityModel {
	return &CommunitySecurityModel{
		readCommunity:      rc,
		readWriteCommunity: rwc,
	}
}

func (m *CommunitySecurityModel) Authenticate(packet *gosnmp.SnmpPacket) error {
	if packet == nil {
		return ErrNilRequest
	}

	if packet.Community == "" {
		return ErrEmptyCommunity
	}

	switch packet.PDUType {
	case gosnmp.GetRequest, gosnmp.GetNextRequest, gosnmp.GetBulkRequest:
		if packet.Community != m.readCommunity && packet.Community != m.readWriteCommunity {
			return ErrInvalidCommunity
		}
	case gosnmp.SetRequest:
		if packet.Community != m.readWriteCommunity {
			return ErrInvalidCommunity
		}
	case gosnmp.InformRequest:
		return nil
	case gosnmp.Trap:
		if packet.Version != gosnmp.Version1 {
			return ErrVersionMissMatch
		}
		return nil
	case gosnmp.SNMPv2Trap:
		if packet.Version != gosnmp.Version2c {
			return ErrVersionMissMatch
		}
		return nil
	default:
		return ErrUnsupportedPDUType
	}

	return nil
}

func (m *CommunitySecurityModel) Encrypt(packet *gosnmp.SnmpPacket) error {
	// v1/v2c 不加密
	return nil
}

func (m *CommunitySecurityModel) Decrypt(packet *gosnmp.SnmpPacket) error {
	// v1/v2c 不加密
	return nil
}

func (m *CommunitySecurityModel) CheckAccess(version string, operation gosnmp.PDUType) (bool, error) {
	if operation == gosnmp.Trap {
		return true, nil
	}
	if operation == gosnmp.InformRequest && version == gosnmp.Version2c.String() {
		return true, nil
	}

	// 暂时禁止trap 和 inform 外所有操作
	return false, nil
}
