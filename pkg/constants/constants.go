package constants

import "regexp"

const ServerConfigPath = "./config/server.yaml"

const (
	APIV1Prefix            = "/api/v1"
	BMCAlertUploadEndpoint = "upload_bmc_trap"
	// 完整路径 - 仅用于客户端构造 (避免硬编码拼接)
	AlertUploadFullPathV1 = APIV1Prefix + "/" + BMCAlertUploadEndpoint // 客户端完整路径

	SnmpTrapOID = ".1.3.6.1.6.3.1.1.4.1.0"
)

var (
	// 只验证OID格式
	OidFormatRegex = regexp.MustCompile(`^\.?\d+(?:\.\d+)+$`)

	// 匹配标准企业OID格式：.1.3.6.1.4.1.xxx.xxx...
	EnterpriseOIDRegex = regexp.MustCompile(`^\.1\.3\.6\.1\.4\.1\.(\d+)(?:\.|$)`)
)
