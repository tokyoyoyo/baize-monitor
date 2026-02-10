package constants

import "fmt"

// 功能模块权限常量
const (
	// 告警管理模块
	PermissionAlertManagement = "alert_management"
	PermissionAlertRead       = "alert_read"
	PermissionAlertPass       = "alert_pass"

	// BMC trap parser配置模块
	PermissionBMCTrapParserManagement = "bmc_trap_parser_management"

	// 用户管理模块
	PermissionUserManagement = "user_management"

	// 用户权限管理模块
	PermissionUserAuthPass = "auth"

	PermissionHealthCheckPass = "helth_check"
)

// ModuleRoutes 功能模块对应的路由映射
var ModuleRoutes = map[string][]string{
	PermissionAlertManagement: {
		fmt.Sprintf("%s/%s/*", APIV1Prefix, PermissionAlertManagement),
	},
	PermissionAlertRead: {
		fmt.Sprintf("%s/%s/*", APIV1Prefix, PermissionAlertRead),
	},

	PermissionBMCTrapParserManagement: {
		fmt.Sprintf("%s/%s/*", APIV1Prefix, PermissionBMCTrapParserManagement),
	},
}

var DefaultPermissions = map[string]bool{
	PermissionAlertManagement:         false,
	PermissionAlertRead:               true,
	PermissionBMCTrapParserManagement: false,
}
