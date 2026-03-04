package request

// MachineInfoRequest 机器信息请求DTO（极简版）
// 用于Agent向服务端上报机器信息，关联服务端上传的管理IP
type MachineInfoRequest struct {
	Hostname    string              `json:"hostname" binding:"required,min=1,max=255"`
	IPAddresses map[string][]string `json:"ip_addresses"`
}
