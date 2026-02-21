package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"baize-monitor/pkg/constants"
	"baize-monitor/pkg/utils"
)

// Auth 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 检查是否跳过认证的路径
		if shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 2. 获取Token
		token := extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未提供认证Token",
			})
			c.Abort()
			return
		}

		// 3. 验证Token
		claims, err := utils.JWTManagerInstance.ValidateAccessToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "无效的认证Token",
			})
			c.Abort()
			return
		}

		// 4. 检查用户状态
		if !claims.IsActive || claims.IsDelete {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "用户账户不可用",
			})
			c.Abort()
			return
		}
		// 5. 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("permissions", claims.Permissions)

		if claims.IsAdmin {
			c.Next()
			return
		}

		// 6. 权限检查
		if !hasPermission(claims.Permissions, c.Request.URL.Path) {
			c.JSON(403, gin.H{
				"code":    403,
				"message": "没有访问该资源的权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission 检查用户是否有访问指定路径的权限
func hasPermission(userPermissions map[string]bool, path string) bool {
	// 检查每个模块的路由
	for module, routes := range constants.ModuleRoutes {
		if userPermissions[module] {
			for _, routePattern := range routes {
				if matchRoute(routePattern, path) {
					return true
				}
			}
		}
	}

	return false
}

// matchRoute 匹配路由模式
func matchRoute(pattern, pathStr string) bool {
	if strings.Contains(pattern, "*") {
		// 假设通配符模式都是以"/*"结尾，去掉这个部分，然后做前缀匹配
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(pathStr, prefix) {
			return true
		}
	} else {
		// 不含通配符，精确匹配
		if pathStr == pattern {
			return true
		}
	}

	return false
}

// shouldSkipAuth 检查是否应该跳过认证
func shouldSkipAuth(pathStr string) bool {
	skipPatterns := []string{
		fmt.Sprintf("/%s", constants.PermissionHealthCheckPass),
		fmt.Sprintf("%s/%s/*", constants.APIV1Prefix, constants.PermissionAlertPass),
		fmt.Sprintf("%s/%s/*", constants.APIV1Prefix, constants.PermissionUserAuthPass),
	}

	for _, pattern := range skipPatterns {
		// 检查是否包含通配符
		if matchRoute(pattern, pathStr) {
			return true
		}
	}

	return false
}

// extractToken 从请求中提取Token
func extractToken(c *gin.Context) string {
	// 1. 从 Authorization Header 获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 检查 Bearer Token（大小写不敏感）
		const bearerPrefix = "Bearer "
		if len(authHeader) > len(bearerPrefix) &&
			strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
			token := strings.TrimSpace(authHeader[len(bearerPrefix):])
			if token != "" {
				return token
			}
		}
	}

	// 2. 从 Query 参数获取
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		return token
	}

	// 3. 从 Cookie 获取
	if token, err := c.Cookie("token"); err == nil && token != "" {
		return token
	}

	return ""
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成一个简单的请求ID
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	// 使用UUID生成请求ID
	return strings.ToUpper(uuid.New().String())
}
