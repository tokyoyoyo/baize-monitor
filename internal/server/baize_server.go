package server

import (
	http_server "baize-monitor/internal/server/http/server"
	"baize-monitor/internal/server/snmp"
	"baize-monitor/pkg/config"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaiZeServer struct {
	adminServer    *http_server.AdminServer
	bmcAlterServer *http.Server
	snmpServer     *snmp.SNMPServer
}

func NewServer(cfg *config.ServerConfig,
	adminS *http_server.AdminServer,
	snmpS *snmp.SNMPServer,
) *BaiZeServer {
	// 设置Gin模式
	if cfg.GinEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &BaiZeServer{
		adminServer: adminS,
		snmpServer:  snmpS,
	}

	return s
}

func (s *BaiZeServer) Start() error {
	// TODO 应该记日志
	err := s.adminServer.Start()
	if err != nil {
		panic(fmt.Sprint("启动 administration server fail:%v", err.Error()))
	}

	return nil
}

func (s *BaiZeServer) Shutdown(ctx context.Context) error {
	// TODO 应该记日志
	err := s.adminServer.Shutdown(ctx)
	if err != nil {
		panic(fmt.Sprint("关闭 administration server fail:%v", err.Error()))
	}

	return nil
}

// func (r *RouterImpl) readyCheck(c *gin.Context) {
// 	// 检查数据库连接
// 	if err := r.db.HealthCheck(); err != nil {
// 		c.JSON(503, gin.H{
// 			"status":  "not ready",
// 			"error":   err.Error(),
// 			"service": "database",
// 		})
// 		return
// 	}

// 	// 可以添加其他依赖检查
// 	c.JSON(200, gin.H{
// 		"status": "ready",
// 		"checks": gin.H{
// 			"database": "ok",
// 		},
// 	})
// }
