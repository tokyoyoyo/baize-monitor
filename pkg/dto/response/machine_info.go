package response

import "time"

// MachineInfoResponse 机器信息响应DTO（极简版）
type MachineInfoResponse struct {
	ID          int64               `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	TargetIP    string              `json:"target_ip"`
	Hostname    string              `json:"hostname"`
	IPAddresses map[string][]string `json:"ip_addresses"`
}