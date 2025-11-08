package snmp

import (
	"baize-monitor/pkg/models"
	"encoding/asn1"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

// ResponseResult Response result
type ResponseResult struct {
	Response   *gosnmp.SnmpPacket
	ShouldSend bool
	Error      error
}

// V1_V2cResponseEngine SNMP v1 response engine
// v1支持以下PDU类型：
// GetRequest
// GetNextRequest
// SetRequest
// Trap
// v1不支持：
// GetBulkRequest（v2c引入）
// InformRequest（v2c引入）
// only accept trap,inform 暂时
type V1_V2cResponseEngine struct {
	enabled       bool
	version       gosnmp.SnmpVersion
	securityModel SecurityModelInterface
	engineID      string // 用于标识（虽然不是v1标准，但有助于跟踪）
	decodeerMap   map[string]*gosnmp.GoSNMP
}

func NewV1_V2cResponseEngine(readCommunity, readWriteCommunity string, enabled bool, v gosnmp.SnmpVersion) *V1_V2cResponseEngine {
	if enabled && (readCommunity == "" || readWriteCommunity == "") {
		panic("Can not enable engine without read and write communities")
	}
	rd := &gosnmp.GoSNMP{
		Community: readCommunity,
		Version:   v,
	}
	rwd := &gosnmp.GoSNMP{
		Community: readWriteCommunity,
		Version:   v,
	}
	return &V1_V2cResponseEngine{
		enabled:       enabled,
		version:       v,
		securityModel: NewCommunitySecurityModel(readCommunity, readWriteCommunity),
		engineID:      generateEngineID(v),
		decodeerMap: map[string]*gosnmp.GoSNMP{
			readCommunity:      rd,
			readWriteCommunity: rwd,
		},
	}
}

func (e *V1_V2cResponseEngine) createInformResponse(request *gosnmp.SnmpPacket) (*gosnmp.SnmpPacket, error) {
	if e.version == gosnmp.Version1 {
		return nil, ErrInvalidInformRequest
	}

	if e.version == gosnmp.Version2c {
		response := &gosnmp.SnmpPacket{
			RequestID:  request.RequestID, // 与请求id要一致
			Version:    e.version,
			Community:  request.Community,
			PDUType:    gosnmp.GetResponse, //GetResponse 是唯一合法且标准的用于响应 InformRequest 的 PDU 类型
			Error:      0,
			ErrorIndex: 0,
			Variables:  []gosnmp.SnmpPDU{}, // 可为空
		}
		return response, nil
	}
	return nil, ErrV1InformNotSupported
}

func (e *V1_V2cResponseEngine) createResponse(request *gosnmp.SnmpPacket) (*gosnmp.SnmpPacket, error) {
	if request == nil {
		return nil, ErrNilRequest
	}

	// 验证版本
	if request.Version != gosnmp.Version1 {
		return nil, ErrVersionMissMatch
	}

	// 验证安全
	if err := e.validateAndProcessSecurity(request); err != nil {
		return nil, err
	}

	switch request.PDUType {
	case gosnmp.Trap:
		return nil, ErrV1TrapNotSupported // Trap不需要响应
	case gosnmp.InformRequest:
		return e.createInformResponse(request)
	default:
		return nil, ErrUnsupportedPDUType
	}
}

func (e *V1_V2cResponseEngine) validateAndProcessSecurity(request *gosnmp.SnmpPacket) error {
	return e.GetSecurityModel().Authenticate(request)
}

func (e *V1_V2cResponseEngine) shouldRespond(request *gosnmp.SnmpPacket) bool {
	if request == nil {
		return false
	}

	// v1 中 Trap 不需要响应
	if request.PDUType == gosnmp.Trap {
		return false
	}

	// 其他请求类型需要响应
	switch request.PDUType {
	case gosnmp.GetRequest, gosnmp.GetNextRequest, gosnmp.SetRequest:
		return true
	default:
		return false
	}
}

func (e *V1_V2cResponseEngine) canProcessRequest(request *gosnmp.SnmpPacket) bool {
	if request == nil || !e.enabled {
		return false
	}

	// 检查版本
	if request.Version == gosnmp.Version1 {
		// 检查支持的PDU类型
		switch request.PDUType {
		case gosnmp.Trap:
			return true
		default:
			return false
		}
	}

	if request.Version == gosnmp.Version2c {
		switch request.PDUType {
		case gosnmp.GetRequest, gosnmp.GetNextRequest, gosnmp.SetRequest:
			return false
		case gosnmp.InformRequest, gosnmp.SNMPv2Trap:
			return true
		}

	}

	return false
}

func (e *V1_V2cResponseEngine) GetVersion() gosnmp.SnmpVersion {
	return e.version
}

func (e *V1_V2cResponseEngine) GetEngineID() string {
	return e.engineID
}

func (e *V1_V2cResponseEngine) GetSecurityModel() SecurityModelInterface {
	return e.securityModel
}

func (e *V1_V2cResponseEngine) GetMaxRepetitions() uint32 {
	return 0
}

func (e *V1_V2cResponseEngine) Enable() bool {
	return e.enabled
}

func (engine *V1_V2cResponseEngine) ProcessRequest(rawPacket *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	requestCommunity, pCErr := parseSNMPv1v2cCommunity(rawPacket.Data)
	if pCErr != nil {
		return nil, pCErr
	}

	decoder := engine.decodeerMap[requestCommunity]
	if decoder == nil {
		return nil, ErrInvalidCommunity
	}

	snmpPacket, decodeErr := decoder.SnmpDecodePacket(rawPacket.Data)
	if decodeErr != nil {
		return nil, ErrDecodeSNMPrequest
	}

	if !engine.canProcessRequest(snmpPacket) {
		return nil, ErrUnsupportedPDUType
	}

	if err := engine.validateAndProcessSecurity(snmpPacket); err != nil {
		return nil, fmt.Errorf("failed to validate and process security: %v", err)
	}

	if !engine.shouldRespond(snmpPacket) {
		return snmpPacket, nil
	}

	response, createRespErr := engine.createResponse(snmpPacket)
	if createRespErr != nil {
		return nil, fmt.Errorf("failed to create response: %v", createRespErr)
	}

	sendSNMPResponse(response, rawPacket)

	return snmpPacket, nil

}

// parseSNMPv1v2cCommunity 解析SNMP v1/v2c报文的community字符串
// 仅适用于v1和v2c版本，v3版本会返回错误
func parseSNMPv1v2cCommunity(data []byte) (string, error) {
	// 解析外层SEQUENCE
	var outerSeq asn1.RawValue
	_, err := asn1.Unmarshal(data, &outerSeq)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal outer sequence: %v", err)
	}

	if outerSeq.Tag != asn1.TagSequence || outerSeq.Class != asn1.ClassUniversal {
		return "", fmt.Errorf("expected SEQUENCE at outer level")
	}

	// 解析第一个字段（版本号，直接跳过）
	versionBytes := outerSeq.Bytes
	var versionField asn1.RawValue
	rest, err := asn1.Unmarshal(versionBytes, &versionField)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal version field: %v", err)
	}

	// 解析第二个字段（community字符串）
	var community asn1.RawValue
	_, err = asn1.Unmarshal(rest, &community)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal community: %v", err)
	}

	if community.Tag != asn1.TagOctetString {
		return "", fmt.Errorf("expected OCTET STRING for community, got tag %d", community.Tag)
	}

	return string(community.Bytes), nil
}

// V3ResponseEngine SNMP v3 response engine
type V3ResponseEngine struct {
	enabled  bool
	version  gosnmp.SnmpVersion
	engineID string // 用于标识
	decoder  *gosnmp.GoSNMP
}

// NewV3ResponseEngine Create v3 response engine instance
func NewV3ResponseEngine(config *models.V3EngineConfig) *V3ResponseEngine {
	msgFlags := gosnmp.NoAuthNoPriv
	switch config.MsgFlags {
	case "NoAuthNoPriv":
		msgFlags = gosnmp.NoAuthNoPriv
	case "AuthNoPriv":
		msgFlags = gosnmp.AuthNoPriv
	case "AuthPriv":
		msgFlags = gosnmp.AuthPriv
	default:
		panic("Invalid msg_flags")
	}

	authProtocol := gosnmp.NoAuth
	switch config.AuthProtocol {
	case "MD5":
		authProtocol = gosnmp.MD5
	case "SHA":
		authProtocol = gosnmp.SHA
	default:
		panic("Invalid auth_protocol")
	}

	privProtocol := gosnmp.NoPriv
	switch config.PrivProtocol {
	case "DES":
		privProtocol = gosnmp.DES
	case "AES":
		privProtocol = gosnmp.AES
	case "AES192":
		privProtocol = gosnmp.AES192
	case "AES256":
		privProtocol = gosnmp.AES256
	default:
		panic("Invalid priv_protocol")
	}

	securityParameters := &gosnmp.UsmSecurityParameters{
		UserName:                 config.UserName,
		AuthenticationProtocol:   authProtocol,
		AuthenticationPassphrase: config.AuthPassphrase,
		PrivacyProtocol:          privProtocol,
		PrivacyPassphrase:        config.PrivPassphrase,
	}

	return &V3ResponseEngine{
		enabled:  config.Enabled,
		version:  gosnmp.Version3,
		engineID: generateEngineID(gosnmp.Version3),
		decoder: &gosnmp.GoSNMP{
			Version:            gosnmp.Version3,
			Timeout:            time.Duration(30) * time.Second,
			Retries:            3,
			ExponentialTimeout: true,
			SecurityModel:      gosnmp.UserSecurityModel,
			SecurityParameters: securityParameters,
			MsgFlags:           msgFlags,
		},
	}
}

func (e *V3ResponseEngine) ProcessRequest(rawPacket *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	snmpPacket, decodeErr := e.decoder.SnmpDecodePacket(rawPacket.Data)
	if decodeErr != nil {
		return nil, ErrDecodeSNMPrequest
	}

	if snmpPacket.PDUType != gosnmp.SNMPv2Trap && snmpPacket.PDUType != gosnmp.InformRequest {
		return nil, ErrUnsupportedPDUType
	}

	if snmpPacket.PDUType == gosnmp.InformRequest {

		response := &gosnmp.SnmpPacket{
			RequestID:  snmpPacket.RequestID, // 与请求id要一致
			Version:    e.version,
			Community:  snmpPacket.Community,
			PDUType:    gosnmp.GetResponse, //GetResponse 是唯一合法且标准的用于响应 InformRequest 的 PDU 类型
			Error:      0,
			ErrorIndex: 0,
			Variables:  []gosnmp.SnmpPDU{}, // 可为空
		}
		sendSNMPResponse(response, rawPacket)
	}
	return snmpPacket, nil
}

func (e *V3ResponseEngine) GetVersion() gosnmp.SnmpVersion {
	return e.version
}

func (e *V3ResponseEngine) GetEngineID() string {
	return e.engineID
}

func (e *V3ResponseEngine) Enable() bool {
	return e.enabled
}

// ResponseEngineFactory Response engine factory
type ResponseEngineFactory struct {
	engines map[gosnmp.SnmpVersion]SNMPResponseEngine
}

// NewResponseEngineFactory Create response engine factory
func NewResponseEngineFactory(config *models.ResponseEngineFactoryConfig) *ResponseEngineFactory {
	factory := &ResponseEngineFactory{
		engines: make(map[gosnmp.SnmpVersion]SNMPResponseEngine),
	}

	// Create engines for each version based on configuration
	if config.V1Config != nil {
		rc := config.V1Config.ReadCommunity
		rwc := config.V1Config.ReadWriteCommunity
		enable := config.V1Config.Enabled
		factory.engines[gosnmp.Version1] = NewV1_V2cResponseEngine(rc, rwc, enable, gosnmp.Version1)
	}

	if config.V2cConfig != nil && config.V2cConfig.Enabled {
		rc := config.V2cConfig.ReadCommunity
		rwc := config.V2cConfig.ReadWriteCommunity
		enable := config.V2cConfig.Enabled
		factory.engines[gosnmp.Version2c] = NewV1_V2cResponseEngine(rc, rwc, enable, gosnmp.Version2c)
	}

	if config.V3Config != nil && config.V3Config.Enabled {
		factory.engines[gosnmp.Version3] = NewV3ResponseEngine(config.V3Config)
	}

	return factory
}

// CreateEngine Create corresponding response engine based on version
func (f *ResponseEngineFactory) createEngine(version gosnmp.SnmpVersion) (SNMPResponseEngine, error) {
	engine, exists := f.engines[version]
	if !exists {
		return nil, ErrUnsupportedVersion
	}
	return engine, nil
}

// ResponseManager Response manager
type ResponseManager struct {
	factory *ResponseEngineFactory
}

// NewResponseManager Create response manager (completely based on configuration)
func NewResponseManager(config *models.ResponseManagerConfig) *ResponseManager {
	return &ResponseManager{
		factory: NewResponseEngineFactory(config.EngineFactoryConfig),
	}
}

func (m *ResponseManager) ResponseRequest(rawPacket *models.RawPacket) (*gosnmp.SnmpPacket, error) {
	if rawPacket.Data == nil {
		return nil, ErrNilRequest
	}
	packetVersion, err := m.parseSNMPVersion(rawPacket.Data)
	if err != nil {
		return nil, err
	}

	if packetVersion != gosnmp.Version1 && packetVersion != gosnmp.Version2c && packetVersion != gosnmp.Version3 {
		return nil, ErrVersionMissMatch
	}

	engine, err := m.factory.createEngine(packetVersion)
	if err != nil {
		return nil, err
	}
	if !engine.Enable() {
		return nil, ErrDisableVersion
	}
	return engine.ProcessRequest(rawPacket)
}

// parseSNMPVersion 解析SNMP报文的版本号
// 返回版本号 (0=v1, 1=v2c, 3=v3) 和错误信息
func (m *ResponseManager) parseSNMPVersion(data []byte) (gosnmp.SnmpVersion, error) {
	// 解析外层SEQUENCE
	var outerSeq asn1.RawValue
	_, err := asn1.Unmarshal(data, &outerSeq)
	if err != nil {
		return gosnmp.Version1, ErrParseSNMPVersion
	}

	if outerSeq.Tag != asn1.TagSequence || outerSeq.Class != asn1.ClassUniversal {
		return gosnmp.Version1, ErrParseSNMPVersion
	}

	// 解析版本字段（第一个元素）
	var version asn1.RawValue
	_, err = asn1.Unmarshal(outerSeq.Bytes, &version)
	if err != nil {
		return gosnmp.Version1, ErrParseSNMPVersion
	}

	if version.Tag != asn1.TagInteger {
		return gosnmp.Version1, ErrParseSNMPVersion
	}

	versionInt := 0
	_, err = asn1.Unmarshal(version.FullBytes, &versionInt)
	if err != nil {
		return gosnmp.Version1, ErrParseSNMPVersion
	}

	switch versionInt {
	case 0:
		return gosnmp.Version1, nil
	case 1:
		return gosnmp.Version2c, nil
	case 3:
		return gosnmp.Version3, nil
	}

	return gosnmp.Version1, ErrIligelVersion
}

func sendSNMPResponse(response *gosnmp.SnmpPacket, rawPacket *models.RawPacket) error {
	// 使用 gosnmp 的 Marshal 方法序列化响应包
	data, err := response.MarshalMsg()
	if err != nil {
		return fmt.Errorf("failed to marshal SNMP response: %w", err)
	}

	// 通过 UDP 连接发送响应
	_, err = rawPacket.Conn.WriteToUDP(data, rawPacket.RemoteAddr)
	if err != nil {
		return fmt.Errorf("failed to send SNMP response: %w", err)
	}

	return nil
}
