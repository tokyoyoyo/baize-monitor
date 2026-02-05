package snmp

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunitySecurityModel_Authenticate(t *testing.T) {
	securityModel := NewCommunitySecurityModel("public", "private")

	tests := []struct {
		name        string
		packet      *gosnmp.SnmpPacket
		expectError error
	}{
		{
			name:        "Nil packet",
			packet:      nil,
			expectError: ErrNilRequest,
		},
		{
			name: "Empty community",
			packet: &gosnmp.SnmpPacket{
				Community: "",
				PDUType:   gosnmp.GetRequest,
			},
			expectError: ErrEmptyCommunity,
		},
		{
			name: "Valid read community for GetRequest",
			packet: &gosnmp.SnmpPacket{
				Community: "public",
				PDUType:   gosnmp.GetRequest,
			},
			expectError: nil,
		},
		{
			name: "Invalid community for GetRequest",
			packet: &gosnmp.SnmpPacket{
				Community: "invalid",
				PDUType:   gosnmp.GetRequest,
			},
			expectError: ErrInvalidCommunity,
		},
		{
			name: "Valid read-write community for SetRequest",
			packet: &gosnmp.SnmpPacket{
				Community: "private",
				PDUType:   gosnmp.SetRequest,
			},
			expectError: nil,
		},
		{
			name: "Read community for SetRequest should fail",
			packet: &gosnmp.SnmpPacket{
				Community: "public",
				PDUType:   gosnmp.SetRequest,
			},
			expectError: ErrInvalidCommunity,
		},
		{
			name: "Trap with v1 version",
			packet: &gosnmp.SnmpPacket{
				Community: "any",
				PDUType:   gosnmp.Trap,
				Version:   gosnmp.Version1,
			},
			expectError: nil,
		},
		{
			name: "Unsupported PDU type",
			packet: &gosnmp.SnmpPacket{
				Community: "public",
				PDUType:   gosnmp.Report, // Unsupported type
			},
			expectError: ErrUnsupportedPDUType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := securityModel.Authenticate(tt.packet)
			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommunitySecurityModel_EncryptDecrypt(t *testing.T) {
	securityModel := NewCommunitySecurityModel("public", "private")
	packet := &gosnmp.SnmpPacket{
		Community: "public",
		PDUType:   gosnmp.GetRequest,
	}

	// v1/v2c should not encrypt/decrypt
	err := securityModel.Encrypt(packet)
	assert.NoError(t, err)

	err = securityModel.Decrypt(packet)
	assert.NoError(t, err)
}

func TestCommunitySecurityModel_CheckAccess(t *testing.T) {
	securityModel := NewCommunitySecurityModel("public", "private")

	tests := []struct {
		name      string
		version   string
		operation gosnmp.PDUType
		expected  bool
	}{
		{
			name:      "Trap should be allowed",
			version:   gosnmp.Version1.String(),
			operation: gosnmp.Trap,
			expected:  true,
		},
		{
			name:      "Inform with v2c should be allowed",
			version:   gosnmp.Version2c.String(),
			operation: gosnmp.InformRequest,
			expected:  true,
		},
		{
			name:      "GetRequest should not be allowed",
			version:   gosnmp.Version1.String(),
			operation: gosnmp.GetRequest,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := securityModel.CheckAccess(tt.version, tt.operation)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

func TestParseSNMPv1v2cCommunity(t *testing.T) {
	// Create a simple SNMP v1 packet for testing
	packet := &gosnmp.SnmpPacket{
		Version:   gosnmp.Version1,
		Community: "test-community",
		PDUType:   gosnmp.GetRequest,
		RequestID: 123,
	}

	data, err := packet.MarshalMsg()
	require.NoError(t, err)

	t.Run("Valid SNMP v1 packet", func(t *testing.T) {
		community, err := parseSNMPv1v2cCommunity(data)
		assert.NoError(t, err)
		assert.Equal(t, "test-community", community)
	})

	t.Run("Invalid data", func(t *testing.T) {
		community, err := parseSNMPv1v2cCommunity([]byte{0x00, 0x01, 0x02})
		assert.Error(t, err)
		assert.Empty(t, community)
	})

	t.Run("Empty data", func(t *testing.T) {
		community, err := parseSNMPv1v2cCommunity(nil)
		assert.Error(t, err)
		assert.Empty(t, community)
	})
}
