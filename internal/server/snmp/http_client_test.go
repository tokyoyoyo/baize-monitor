package snmp

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"baize-monitor/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleHTTPTrapSenderIpmi_SendTrap_Success(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "BaiZe-SNMP/1.0", r.Header.Get("User-Agent"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		trapData := models.TrapMessage{}
		require.NoError(t, json.Unmarshal(body, &trapData))

		// Verify VariableMap
		varMap := trapData.VariableMap
		assert.Equal(t, "critical", varMap["1.3.6.1.4.1.343.6.1.1"])

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create sender
	sender := NewSimpleHTTPTrapSenderIpmi(ts.URL)

	// Construct test trap
	sourceIP := net.ParseIP("192.168.1.100")
	testTrap := &models.TrapMessage{
		ReceivedAt: time.Now(),
		SourceIP:   sourceIP,
		SourcePort: 162,
		VariableMap: map[string]string{
			"1.3.6.1.4.1.343.6.1.1": "critical",
			"1.3.6.1.4.1.343.6.1.2": "temperature",
		},
		RawData: []byte{0x30, 0x01, 0x02}, // Simulate raw SNMP data
	}

	// Send trap
	sender.SendTrap(testTrap)

	// Verify statistics
	stats := sender.stats()
	assert.Equal(t, uint64(1), stats["success_count"])
	assert.Equal(t, uint64(0), stats["error_count"])
	assert.Equal(t, ts.URL, stats["alert_url"])
}

func TestSimpleHTTPTrapSenderIpmi_SendTrap_Failure(t *testing.T) {
	// Create test server that returns 500 error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// Create sender
	sender := NewSimpleHTTPTrapSenderIpmi(ts.URL)

	// Construct test trap
	sourceIP := net.ParseIP("10.0.0.1")
	testTrap := &models.TrapMessage{
		ReceivedAt: time.Now(),
		SourceIP:   sourceIP,
		SourcePort: 162,
		VariableMap: map[string]string{
			"oid": "error_condition",
		},
		RawData: []byte{0x01, 0x02, 0x03},
	}

	// Send trap
	sender.SendTrap(testTrap)

	// Verify statistics
	stats := sender.stats()
	assert.Equal(t, uint64(0), stats["success_count"])
	assert.Equal(t, uint64(1), stats["error_count"])
}

func TestSimpleHTTPTrapSenderIpmi_SendTrap_ConnectionError(t *testing.T) {
	// Create sender with invalid URL (using inactive port)
	invalidURL := "http://localhost:9999/nonexistent"
	sender := NewSimpleHTTPTrapSenderIpmi(invalidURL)

	// Construct test trap
	sourceIP := net.ParseIP("172.16.0.1")
	testTrap := &models.TrapMessage{
		ReceivedAt: time.Now(),
		SourceIP:   sourceIP,
		SourcePort: 162,
		VariableMap: map[string]string{
			"test": "value",
		},
		RawData: []byte{0x04, 0x05, 0x06},
	}

	// Send trap
	sender.SendTrap(testTrap)

	// Verify statistics
	stats := sender.stats()
	assert.Equal(t, uint64(0), stats["success_count"])
	assert.Equal(t, uint64(1), stats["error_count"])
	assert.Equal(t, invalidURL, stats["alert_url"])
}

func TestSimpleHTTPTrapSenderIpmi_MultipleSends(t *testing.T) {
	// Create test server with successful responses
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewSimpleHTTPTrapSenderIpmi(ts.URL)

	// Send multiple traps
	sourceIP1 := net.ParseIP("192.168.1.1")
	sourceIP2 := net.ParseIP("192.168.1.2")
	sourceIP3 := net.ParseIP("192.168.1.3")

	traps := []*models.TrapMessage{
		{
			ReceivedAt:  time.Now(),
			SourceIP:    sourceIP1,
			SourcePort:  162,
			VariableMap: map[string]string{"trap": "one"},
			RawData:     []byte{0x11},
		},
		{
			ReceivedAt:  time.Now(),
			SourceIP:    sourceIP2,
			SourcePort:  162,
			VariableMap: map[string]string{"trap": "two"},
			RawData:     []byte{0x22},
		},
		{
			ReceivedAt:  time.Now(),
			SourceIP:    sourceIP3,
			SourcePort:  162,
			VariableMap: map[string]string{"trap": "three"},
			RawData:     []byte{0x33},
		},
	}

	for _, trap := range traps {
		sender.SendTrap(trap)
	}

	// Verify statistics
	stats := sender.stats()
	assert.Equal(t, uint64(3), stats["success_count"])
	assert.Equal(t, uint64(0), stats["error_count"])
}

func TestSimpleHTTPTrapSenderIpmi_Timeout(t *testing.T) {
	// Create test server with delayed response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Exceeds client's 3-second timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create client with timeout
	sender := &SimpleHTTPTrapSenderIpmi{
		client: &http.Client{
			Timeout: 1 * time.Second, // Shorter than server delay
		},
		alertURL: ts.URL,
	}

	// Construct test trap
	sourceIP := net.ParseIP("127.0.0.1")
	testTrap := &models.TrapMessage{
		ReceivedAt: time.Now(),
		SourceIP:   sourceIP,
		SourcePort: 162,
		VariableMap: map[string]string{
			"timeout": "test",
		},
		RawData: []byte{0xAA, 0xBB, 0xCC},
	}

	// Send trap
	sender.SendTrap(testTrap)

	// Verify statistics
	stats := sender.stats()
	assert.Equal(t, uint64(0), stats["success_count"])
	assert.Equal(t, uint64(1), stats["error_count"])
}
