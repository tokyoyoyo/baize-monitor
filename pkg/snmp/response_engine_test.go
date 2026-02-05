package snmp

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	"fmt"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

var (
	nilSNMPPacket              *gosnmp.SnmpPacket
	emptySNMPPacket            = &gosnmp.SnmpPacket{}
	invalidCommunitySNMPPacket = &gosnmp.SnmpPacket{
		Community: "wrongCommunity",
	}
	getRequestSNMPPacket = &gosnmp.SnmpPacket{
		PDUType:   gosnmp.GetRequest,
		Community: "public",
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.3.6.1.2.1.1.1.0",
				Type: gosnmp.NoSuchObject,
			},
		},
	}
	getNextRequestSNMPPacket = &gosnmp.SnmpPacket{
		PDUType:   gosnmp.GetNextRequest,
		Community: "public",
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.3.6.1.2.1.1.1.0",
				Type: gosnmp.NoSuchObject,
			},
		},
	}
	setRequestSNMPPacket = &gosnmp.SnmpPacket{
		Community: "public",
		PDUType:   gosnmp.SetRequest,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.3.6.1.2.1.1.1.0",
				Type: gosnmp.NoSuchObject,
			},
		},
	}
	v1SNMPTrapPacket = &gosnmp.SnmpPacket{
		Version:   gosnmp.Version1,
		Community: "public",
		PDUType:   gosnmp.Trap,
		SnmpTrap: gosnmp.SnmpTrap{ // V1trap必须的一些字段
			Enterprise:   ".1.3.6.1.4.1.9.1.1", // 企业 OID
			AgentAddress: "192.168.1.100",      // 发送方 IP
			GenericTrap:  6,                    // enterpriseSpecific
			SpecificTrap: 1,
			Timestamp:    12345, // TimeTicks (centiseconds)
		},
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  ".1.3.6.1.2.1.1.3.0", // sysUpTime.0
				Type:  gosnmp.TimeTicks,
				Value: uint32(12345), // TimeTicks is uint32
			},
			{
				Name:  ".1.3.6.1.6.3.1.1.4.1.0", // snmpTrapOID.0
				Type:  gosnmp.ObjectIdentifier,
				Value: ".1.3.6.1.4.1.9.1.1.1", // your specific trap OID
			},
			{
				Name:  ".1.3.6.1.2.1.1.1.0",
				Type:  gosnmp.OctetString,
				Value: []byte("hello from baize"),
			},
		},
	}
	v2cSNMPv2TrapPacket = &gosnmp.SnmpPacket{
		Version:   gosnmp.Version2c,
		Community: "public",
		PDUType:   gosnmp.SNMPv2Trap,
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  ".1.3.6.1.2.1.1.3.0", // sysUpTime.0
				Type:  gosnmp.TimeTicks,
				Value: uint32(12345), // TimeTicks is uint32
			},
			{
				Name:  ".1.3.6.1.2.1.1.1.0",
				Type:  gosnmp.OctetString,
				Value: []byte("hello from baize v2c"),
			},
		},
	}
	informRequestSNMPPacket = &gosnmp.SnmpPacket{
		PDUType: gosnmp.InformRequest,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.3.6.1.2.1.1.1.0",
				Type: gosnmp.NoSuchObject,
			},
		},
	}
)

var mockManagerConfig = config.ResponseManagerConfig{
	EngineFactoryConfig: &config.ResponseEngineFactoryConfig{
		V1Config: &config.V1EngineConfig{
			ReadCommunity:      "public",
			ReadWriteCommunity: "private",
			Enabled:            true,
		},
		V2cConfig: &config.V2cEngineConfig{
			ReadCommunity:      "public",
			ReadWriteCommunity: "private",
			Enabled:            true,
		},
		V3Config: &config.V3EngineConfig{
			Enabled:        true,
			UserName:       "baize",
			MsgFlags:       "AuthPriv",
			AuthProtocol:   "MD5",
			PrivProtocol:   "AES",
			PrivPassphrase: "PrivPassphrasetest",
			AuthPassphrase: "AuthPassphrasetest",
		},
	},
}

func TestNewV1_V2cResponseEngine(t *testing.T) {

	noPanicCS := []struct {
		Name               string
		ReadCommunity      string
		ReadWriteCommunity string
		Enabled            bool
		Version            gosnmp.SnmpVersion
	}{
		{
			"Valid V1 Enabled",
			"public",
			"private",
			true,
			gosnmp.Version1,
		},
		{
			"Valid V1 Disabled",
			"public",
			"private",
			false,
			gosnmp.Version1,
		},
		{
			"inValid V1 Disabled",
			"",
			"",
			false,
			gosnmp.Version1,
		},
		{
			"Valid V2c Enabled",
			"public",
			"private",
			true,
			gosnmp.Version2c,
		},
		{
			"Valid V2c Disabled",
			"public",
			"private",
			false,
			gosnmp.Version2c,
		},
		{
			"inValid V2c Disabled",
			"",
			"",
			false,
			gosnmp.Version2c,
		},
	}

	t.Run("Test do not panic NewV1_V2cResponseEngine", func(t *testing.T) {
		for _, c := range noPanicCS {
			engine := NewV1_V2cResponseEngine(c.ReadCommunity, c.ReadWriteCommunity, c.Enabled, c.Version)
			assert.NotNil(t, engine)
			assert.Equal(t, c.Version, engine.GetVersion())
			assert.Equal(t, c.Enabled, engine.Enable())
			assert.NotNil(t, engine.GetSecurityModel())
			assert.NotNil(t, engine.GetEngineID())
		}
	})

	panicCS := []struct {
		Name               string
		ReadCommunity      string
		ReadWriteCommunity string
		Enabled            bool
		Version            gosnmp.SnmpVersion
	}{
		{
			"inValid Empty communities V1",
			"",
			"",
			true,
			gosnmp.Version1,
		},
		{
			"inValid Empty communities V2c",
			"",
			"",
			true,
			gosnmp.Version2c,
		},
	}

	t.Run("Test panic NewV1_V2cResponseEngine", func(t *testing.T) {
		for _, c := range panicCS {
			fmt.Println(c)
			t.Run("Invalid V1_V2c config panic "+c.Name, func(t *testing.T) {
				assert.Panics(t, func() {
					NewV1_V2cResponseEngine(c.ReadCommunity, c.ReadWriteCommunity, c.Enabled, c.Version)
				}, "Expected NewV1_V2cResponseEngine to panic with invalid config and enabled config")
			})
		}
	})

}

func TestNewV3ResponseEngine(t *testing.T) {
	noPanicCS := []struct {
		Name string
		Conf *config.V3EngineConfig
	}{
		{
			"Valid V3 enabled",
			&config.V3EngineConfig{
				Enabled:        true,
				UserName:       "baize",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "MD5",
				PrivProtocol:   "AES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
		{
			"Valid V3 disabled",
			&config.V3EngineConfig{
				Enabled:        false,
				UserName:       "baize",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "MD5",
				PrivProtocol:   "AES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
		{
			"inValid V3 disabled",
			&config.V3EngineConfig{
				Enabled:        true,
				UserName:       "baize",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "MD5",
				PrivProtocol:   "AES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
	}

	t.Run("Test do not panic NewV3ResponseEngine test", func(t *testing.T) {
		for _, c := range noPanicCS {
			engine := NewV3ResponseEngine(c.Conf)
			assert.NotNil(t, engine)
			assert.Equal(t, c.Conf.Enabled, engine.Enable())
			assert.Equal(t, gosnmp.Version3, engine.GetVersion())
			assert.NotNil(t, engine.GetEngineID())
			assert.NotNil(t, engine.decoder)
		}
	})

	panicCS := []struct {
		Name string
		Conf *config.V3EngineConfig
	}{
		{
			"inValid V3 enabled",
			&config.V3EngineConfig{
				Enabled:        true,
				UserName:       "baize",
				MsgFlags:       "iliggleAuthPriv",
				AuthProtocol:   "MD5",
				PrivProtocol:   "AES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
		{
			"inValid V3 enabled",
			&config.V3EngineConfig{
				Enabled:        true,
				UserName:       "baize",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "iligelMD5",
				PrivProtocol:   "AES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
		{
			"inValid V3 enabled",
			&config.V3EngineConfig{
				Enabled:        true,
				UserName:       "baize",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "MD5",
				PrivProtocol:   "iligelAES",
				AuthPassphrase: "AuthPassphrasetest",
				PrivPassphrase: "PrivPassphrasetest",
			},
		},
	}

	t.Run("Test panic NewV3ResponseEngine", func(t *testing.T) {
		for _, c := range panicCS {
			fmt.Println(c)
			t.Run("Invalid V3 config panic "+c.Name, func(t *testing.T) {
				assert.Panics(t, func() {
					NewV3ResponseEngine(c.Conf)
				}, "Expected NewV3ResponseEngine to panic with invalid config and enabled config")
			})
		}
	})
}

func TestNewResponseManager(t *testing.T) {
	manager := NewResponseManager(&mockManagerConfig)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.factory)

	e, err := manager.factory.createEngine(gosnmp.Version1)
	assert.NoError(t, err)
	assert.NotNil(t, e)

	e, err = manager.factory.createEngine(gosnmp.Version2c)
	assert.NoError(t, err)
	assert.NotNil(t, e)

	e, err = manager.factory.createEngine(gosnmp.Version3)
	assert.NoError(t, err)
	assert.NotNil(t, e)

}

func TestV1_V2cResponseEngine_GetVersion(t *testing.T) {
	engine := NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version1)
	assert.Equal(t, gosnmp.Version1, engine.GetVersion())
	NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version2c)
	assert.Equal(t, gosnmp.Version2c, NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version2c).GetVersion())
}

func TestV1_V2cResponseEngine_GetEngineID(t *testing.T) {
	engine := NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version1)
	assert.NotNil(t, engine.GetEngineID())
	assert.NotEmpty(t, engine.GetEngineID())
	engine = NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version2c)
	assert.NotNil(t, engine.GetEngineID())
	assert.NotEmpty(t, engine.GetEngineID())
}

func TestV1_V2cResponseEngine_GetSecurityModel(t *testing.T) {
	rc := "public"
	rwc := "private"

	engine := NewV1_V2cResponseEngine(rc, rwc, true, gosnmp.Version1)
	securityModel := engine.GetSecurityModel()
	assert.NotNil(t, securityModel)
	assert.IsType(t, &CommunitySecurityModel{}, securityModel)
	assert.Equal(t, rc, securityModel.(*CommunitySecurityModel).readCommunity)
	assert.Equal(t, rwc, securityModel.(*CommunitySecurityModel).readWriteCommunity)

	engine = NewV1_V2cResponseEngine(rc, rwc, true, gosnmp.Version2c)
	securityModel = engine.GetSecurityModel()
	assert.NotNil(t, securityModel)
	assert.IsType(t, &CommunitySecurityModel{}, securityModel)
	assert.Equal(t, rc, securityModel.(*CommunitySecurityModel).readCommunity)
	assert.Equal(t, rwc, securityModel.(*CommunitySecurityModel).readWriteCommunity)
}

func TestV1_V2cResponseEngine_Enable(t *testing.T) {
	engine := NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version1)
	assert.True(t, engine.Enable())

	disabledEngine := NewV1_V2cResponseEngine("public", "private", false, gosnmp.Version1)
	assert.False(t, disabledEngine.Enable())

	engine = NewV1_V2cResponseEngine("public", "private", true, gosnmp.Version2c)
	assert.True(t, engine.Enable())
	disabledEngine = NewV1_V2cResponseEngine("public", "private", false, gosnmp.Version2c)
	assert.False(t, disabledEngine.Enable())
}

func TestV3ResponseEngine_GetVersion(t *testing.T) {
	v3Engine := NewV3ResponseEngine(mockManagerConfig.EngineFactoryConfig.V3Config)
	assert.Equal(t, gosnmp.Version3, v3Engine.GetVersion())
}

func TestV3ResponseEngine_GetEngineID(t *testing.T) {
	v3Engine := NewV3ResponseEngine(mockManagerConfig.EngineFactoryConfig.V3Config)
	assert.NotNil(t, v3Engine.GetEngineID())
	assert.NotEmpty(t, v3Engine.GetEngineID())
}

func TestV3ResponseEngine_Enable(t *testing.T) {
	enabledConfig := *mockManagerConfig.EngineFactoryConfig.V3Config
	enabledConfig.Enabled = true
	v3Engine := NewV3ResponseEngine(&enabledConfig)
	assert.True(t, v3Engine.Enable())

	disabledConfig := *mockManagerConfig.EngineFactoryConfig.V3Config
	disabledConfig.Enabled = false
	disabledV3Engine := NewV3ResponseEngine(&disabledConfig)
	assert.False(t, disabledV3Engine.Enable())
}

func TestResponseEngineFactory_CreateEngine(t *testing.T) {
	factory := NewResponseEngineFactory(mockManagerConfig.EngineFactoryConfig)

	v1Engine, err := factory.createEngine(gosnmp.Version1)
	assert.NoError(t, err)
	assert.NotNil(t, v1Engine)
	assert.IsType(t, &V1_V2cResponseEngine{}, v1Engine)

	v2cEngine, err := factory.createEngine(gosnmp.Version2c)
	assert.NoError(t, err)
	assert.NotNil(t, v2cEngine)
	assert.IsType(t, &V1_V2cResponseEngine{}, v2cEngine)

	v3Engine, err := factory.createEngine(gosnmp.Version3)
	assert.NoError(t, err)
	assert.NotNil(t, v3Engine)
	assert.IsType(t, &V3ResponseEngine{}, v3Engine)

	unsupportedEngine, err := factory.createEngine(99)
	assert.Error(t, err)
	assert.Nil(t, unsupportedEngine)
	assert.Equal(t, ErrUnsupportedVersion, err)
}

func TestResponseManager_ResponseRequest(t *testing.T) {
	manager := NewResponseManager(&mockManagerConfig)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.factory)

	e, err := manager.factory.createEngine(gosnmp.Version1)
	assert.NoError(t, err)
	assert.NotNil(t, e)
	e, err = manager.factory.createEngine(gosnmp.Version2c)
	assert.NoError(t, err)
	assert.NotNil(t, e)
	e, err = manager.factory.createEngine(gosnmp.Version3)
	assert.NoError(t, err)
	assert.NotNil(t, e)

	t.Run("Test unsport request V1 engine", func(t *testing.T) {
		unSportV1RequestList := []struct {
			Name    string
			Request *gosnmp.SnmpPacket
		}{
			{
				"nil request",
				nilSNMPPacket,
			},
			{
				"empty request",
				emptySNMPPacket,
			},
			{
				"invalid community",
				invalidCommunitySNMPPacket,
			},
			{
				"invalid PDU type GetRequest",
				getRequestSNMPPacket,
			},
			{
				"invalid PDU type GetNextRequest",
				getNextRequestSNMPPacket,
			},
			{
				"invalid PDU type SetRequest",
				setRequestSNMPPacket,
			},
			{
				"invalid PDU type SNMP v2cTrap",
				v2cSNMPv2TrapPacket,
			},
			{
				"invalid PDU type InformRequest",
				informRequestSNMPPacket,
			},
		}

		v1validCommunitySNMP := gosnmp.GoSNMP{
			Community: "public",
			Version:   gosnmp.Version1,
			Port:      162,
			Transport: "udp",
		}
		v1invalidCommunitySNMP := gosnmp.GoSNMP{
			Community: "invalid community",
			Version:   gosnmp.Version1,
			Port:      162,
			Transport: "udp",
		}

		for _, r := range unSportV1RequestList {
			rawPacket := models.RawPacket{}
			if r.Name == "nil request" {
				rawPacket.Data = nil
			} else {
				if r.Name == "invalid community" {

					request, err := v1invalidCommunitySNMP.SnmpEncodePacket(gosnmp.GetRequest, r.Request.Variables, 1, 1)
					assert.NoError(t, err)
					rawPacket.Data = request
				} else {
					request, err := v1validCommunitySNMP.SnmpEncodePacket(r.Request.PDUType, r.Request.Variables, 1, 1)
					assert.NoError(t, err)
					rawPacket.Data = request
				}

			}
			requestPacker, err := manager.ResponseRequest(&rawPacket)
			assert.Error(t, err)
			assert.Nil(t, requestPacker)
			fmt.Printf("%s: %v\n", r.Name, err)
		}
	})

	t.Run("Test sport request V1 engine", func(t *testing.T) {
		sportV1RequestList := []struct {
			Name    string
			Request *gosnmp.SnmpPacket
		}{
			{
				"valid SNMPv1 Trap",
				v1SNMPTrapPacket,
			},
		}

		for _, r := range sportV1RequestList {
			rawPacket := models.RawPacket{}
			request, err := r.Request.MarshalMsg()
			assert.NoError(t, err)
			rawPacket.Data = request
			requestPacket, err := manager.ResponseRequest(&rawPacket)
			assert.NoError(t, err)
			assert.NotNil(t, requestPacket)
			assert.Equal(t, r.Request.PDUType, requestPacket.PDUType)
			assert.Equal(t, r.Request.Variables, requestPacket.Variables)
		}
	})

	t.Run("Test unsport request V2c engine", func(t *testing.T) {
		unSportV2cRequestList := []struct {
			Name    string
			Request *gosnmp.SnmpPacket
		}{
			{
				"nil request",
				nilSNMPPacket,
			},
			{
				"empty request",
				emptySNMPPacket,
			},
			{
				"invalid community",
				invalidCommunitySNMPPacket,
			},
			{
				"invalid PDU type GetRequest",
				getRequestSNMPPacket,
			},
			{
				"invalid PDU type GetNextRequest",
				getNextRequestSNMPPacket,
			},
			{
				"invalid PDU type SetRequest",
				setRequestSNMPPacket,
			},
		}

		v2cValidCommunitySNMP := gosnmp.GoSNMP{
			Community: "public",
			Version:   gosnmp.Version2c,
			Port:      162,
			Transport: "udp",
		}
		v2cInvalidCommunitySNMP := gosnmp.GoSNMP{
			Community: "invalid community",
			Version:   gosnmp.Version2c,
			Port:      162,
			Transport: "udp",
		}

		for _, r := range unSportV2cRequestList {
			rawPacket := models.RawPacket{}

			if r.Name == "nil request" {
				rawPacket.Data = nil
			} else {
				if r.Name == "invalid community" {
					request, err := v2cInvalidCommunitySNMP.SnmpEncodePacket(gosnmp.GetRequest, r.Request.Variables, 1, 1)
					assert.NoError(t, err)
					rawPacket.Data = request
				} else {
					request, err := v2cValidCommunitySNMP.SnmpEncodePacket(r.Request.PDUType, r.Request.Variables, 1, 1)
					assert.NoError(t, err)
					rawPacket.Data = request
				}
			}
			requestPacker, err := manager.ResponseRequest(&rawPacket)
			assert.Error(t, err)
			assert.Nil(t, requestPacker)
			fmt.Printf("%s: %v\n", r.Name, err)

		}
	})

	t.Run("Test sport request V2c engine", func(t *testing.T) {
		sportV2cRequestList := []struct {
			Name    string
			Request *gosnmp.SnmpPacket
		}{
			{
				Name:    "valid SNMPv2c Trap",
				Request: v2cSNMPv2TrapPacket,
			},
		}

		for _, r := range sportV2cRequestList {
			rawPacket := models.RawPacket{}
			request, err := r.Request.MarshalMsg()
			assert.NoError(t, err)
			rawPacket.Data = request
			requestPacket, err := manager.ResponseRequest(&rawPacket)
			assert.NoError(t, err)
			assert.NotNil(t, requestPacket)
			assert.Equal(t, r.Request.PDUType, requestPacket.PDUType)
			assert.Equal(t, r.Request.Variables, requestPacket.Variables)
		}
	})
}
