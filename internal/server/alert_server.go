package server

import (
	"baize-monitor/internal/server/http/routes"
	"baize-monitor/pkg/config"
	"context"
	"fmt"
	"net/http"
	"time"
)

type AlertServer struct {
	s http.Server
}

func NewAlertServer(cfg *config.ServerConfig, r routes.AlertRouter) (*AlertServer, error) {
	if cfg.AlertServerConfig.Port < 0 || cfg.AlertServerConfig.Port > 65535 {
		return nil, fmt.Errorf("invalid alert server port: %d", cfg.AlertServerConfig.Port)
	}

	addr := fmt.Sprintf(":%d", cfg.AlertServerConfig.Port)

	return &AlertServer{
		s: http.Server{
			Addr:    addr,
			Handler: r.GetAlertRouter(),

			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}, nil
}

func (s *AlertServer) Start() error {
	return s.s.ListenAndServe()
}

func (s *AlertServer) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
