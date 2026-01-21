// internal/server/snmp/http_client.go
package snmp

import (
	"baize-monitor/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPTrapSender interface for sending SNMP Trap to the alert system
type HTTPTrapSender interface {
	SendTrap(trap *models.TrapMessage)
	stats() map[string]interface{}
	// for testing
}

// SimpleHTTPTrapSenderIpmi simple HTTP client
type SimpleHTTPTrapSenderIpmi struct {
	client       *http.Client
	alertURL     string
	successCount uint64 // success count
	errorCount   uint64 // error count
}

// NewSimpleHTTPTrapSenderIpmi create an HTTP client
func NewSimpleHTTPTrapSenderIpmi(alertURL string) *SimpleHTTPTrapSenderIpmi {
	return &SimpleHTTPTrapSenderIpmi{
		client: &http.Client{
			Timeout: 3 * time.Second, // 3 second timeout
		},
		alertURL: alertURL,
	}
}

// SendTrap sends the trap to the alert interface
func (c *SimpleHTTPTrapSenderIpmi) SendTrap(trap *models.TrapMessage) {
	if err := c.send(trap); err != nil {
		snmp_logger.Error("Failed to send trap to alert API",
			"source", trap.SourceIP,
			"error", err)
		c.errorCount++
	} else {
		c.successCount++
	}
}

// send actually sends the HTTP request
func (c *SimpleHTTPTrapSenderIpmi) send(trap *models.TrapMessage) error {
	jsonData, err := json.Marshal(trap)
	if err != nil {
		return fmt.Errorf("marshal trap failed: %w", err)
	}

	req, err := http.NewRequest("POST", c.alertURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "BaiZe-SNMP/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	snmp_logger.Debug("Trap sent to alert API",
		"source", trap.SourceIP,
		"status", resp.StatusCode)

	return nil
}

// Stats gets statistics information
func (c *SimpleHTTPTrapSenderIpmi) stats() map[string]interface{} {
	return map[string]interface{}{
		"success_count": c.successCount,
		"error_count":   c.errorCount,
		"alert_url":     c.alertURL,
	}
}
