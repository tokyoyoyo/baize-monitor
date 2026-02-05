package server

import (
	"baize-monitor/internal/server/snmp"
	"baize-monitor/pkg/config"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

type BaiZeServer struct {
	adminServer *AdminServer
	AlertServer *AlertServer
	snmpServer  *snmp.SNMPServer
}

func NewBaiZeServer(cfg *config.ServerConfig,
	adminS *AdminServer,
	alertS *AlertServer,
	snmpS *snmp.SNMPServer,
) *BaiZeServer {
	// 设置Gin模式
	if cfg.GinDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &BaiZeServer{
		adminServer: adminS,
		AlertServer: alertS,
		snmpServer:  snmpS,
	}

	return s
}

func (s *BaiZeServer) Start() error {
	err := s.adminServer.Start()
	if err != nil {
		return fmt.Errorf("启动 administration server fail:%v", err.Error())
	}

	err = s.AlertServer.Start()
	if err != nil {
		return fmt.Errorf("启动 alert server fail:%v", err.Error())
	}

	err = s.snmpServer.Start()
	if err != nil {
		return fmt.Errorf("启动 snmp server fail:%v", err.Error())
	}

	return nil
}

func (s *BaiZeServer) Shutdown(ctx context.Context) error {
	// TODO 应该记日志
	err := s.adminServer.Shutdown(ctx)
	if err != nil {
		panic(fmt.Sprint("关闭 administration server fail:%v", err.Error()))
	}

	err = s.snmpServer.Stop()
	if err != nil {
		return fmt.Errorf("关闭 snmp server fail:%v", err.Error())
	}

	err = s.AlertServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("关闭 alert server fail:%v", err.Error())
	}

	return nil
}
