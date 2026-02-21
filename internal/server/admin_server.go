package server

import (
	"baize-monitor/internal/server/http/routes"
	"baize-monitor/pkg/config"
	"context"
	"fmt"
	"net/http"
	"time"
)

type AdminServer struct {
	s http.Server
}

func NewAdminServer(cfg *config.ServerConfig, r routes.AdminRouter) *AdminServer {
	addr := fmt.Sprintf(":%d", cfg.AdminServerConfig.Port)
	return &AdminServer{
		s: http.Server{
			Addr:    addr,
			Handler: r.GetAdminRouter(),

			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *AdminServer) Start() error {
	go func() {
		if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Admin server error: %v\n", err)
		}
	}()
	return nil
}

func (s *AdminServer) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
