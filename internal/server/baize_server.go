package server

import (
	"baize-monitor/internal/server/snmp"
	"baize-monitor/pkg/config"
	"context"
	"fmt"
	"time"

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
	if err := s.adminServer.Start(); err != nil {
		return fmt.Errorf("failed to start administration server: %v", err)
	}

	if err := s.AlertServer.Start(); err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.adminServer.Shutdown(shutdownCtx)
		return fmt.Errorf("failed to start alert server: %v", err)
	}

	if err := s.snmpServer.Start(); err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.adminServer.Shutdown(shutdownCtx)
		_ = s.AlertServer.Shutdown(shutdownCtx)
		return fmt.Errorf("failed to start snmp server: %v", err)
	}

	return nil
}

func (s *BaiZeServer) Shutdown(ctx context.Context) error {
	err := s.adminServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown administration server: %w", err)
	}

	err = s.AlertServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown alert server: %w", err)
	}

	err = s.snmpServer.Stop()
	if err != nil {
		return fmt.Errorf("failed to shutdown snmp server: %w", err)
	}

	return nil
}
