package middleware

// TODO : add auth middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		if !isValidToken(token) {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "无效的认证Token",
			})
			c.Abort()
			return
		}

		// 4. 设置用户信息到上下文
		// 这里可以解析JWT或从数据库查询用户信息
		c.Set("user_id", "admin") // 示例，实际应该从Token解析
		c.Set("user_role", "admin")

		c.Next()
	}
}

// extractToken 从请求中提取Token
func extractToken(c *gin.Context) string {
	// 1. 从Header中获取
	bearerToken := c.GetHeader("Authorization")
	if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
		return bearerToken[7:] // 去掉"Bearer "前缀
	}

	// 2. 从Query参数中获取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 3. 从Cookie中获取
	token, _ = c.Cookie("token")
	return token
}

// isValidToken 验证Token是否有效
func isValidToken(token string) bool {
	// TODO
	return false
}

// shouldSkipAuth 检查是否应该跳过认证
func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/ready",
		"/metrics",
		"/api/v1/alerts/trap", // SNMP陷阱接收接口通常不需要认证
		"/static",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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
	// TODO
	return time.Now().Format("20060102150405") + "-" + strings.ToUpper(uuid.New().String())
}
