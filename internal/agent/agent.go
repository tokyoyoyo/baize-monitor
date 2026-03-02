package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"baize-monitor/internal/agent/anomaly"
	"baize-monitor/internal/agent/hardware"
	"baize-monitor/internal/agent/machineinfo"
	"baize-monitor/internal/agent/metrics"
	"baize-monitor/pkg/dto/response"
)

// startTime 记录 Agent 启动时间
var startTime time.Time

// AgentConfig Agent 配置结构
type AgentConfig struct {
	ServerURL         string        `yaml:"server_url"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	MetricsPort       int           `yaml:"metrics_port"`
	NodeName          string        `yaml:"node_name"`
	LogLevel          string        `yaml:"log_level"`
}

// Agent Agent 核心结构
type Agent struct {
	config      *AgentConfig
	registry    *prometheus.Registry
	metrics     *metrics.Metrics
	machineInfo *machineinfo.MachineInfo
	hardware    *hardware.Hardware
	anomaly     *anomaly.Anomaly
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mutex       sync.RWMutex
}

// NewAgent 创建 Agent 实例
func NewAgent(config *AgentConfig) *Agent {
	startTime = time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	registry := prometheus.NewRegistry()

	return &Agent{
		config:      config,
		registry:    registry,
		metrics:     metrics.New(),
		machineInfo: machineinfo.New(),
		hardware:    hardware.New(),
		anomaly:     anomaly.New(),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动 Agent
func (a *Agent) Start() error {
	log.Printf("Starting BaiZe Agent on node: %s", a.config.NodeName)

	// 启动所有模块
	if err := a.metrics.Start(); err != nil {
		return fmt.Errorf("failed to start metrics module: %w", err)
	}
	if err := a.machineInfo.Start(); err != nil {
		return fmt.Errorf("failed to start machine info module: %w", err)
	}
	if err := a.hardware.Start(); err != nil {
		return fmt.Errorf("failed to start hardware module: %w", err)
	}
	if err := a.anomaly.Start(); err != nil {
		return fmt.Errorf("failed to start anomaly module: %w", err)
	}

	// 注册 metrics 到 Prometheus
	a.registry.MustRegister(a.metrics)
	a.registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	a.registry.MustRegister(collectors.NewGoCollector())

	// 启动 HTTP 服务器
	if err := a.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	log.Println("Agent started successfully")
	return nil
}

// startHTTPServer 启动 HTTP 服务器
func (a *Agent) startHTTPServer() error {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		mux := http.NewServeMux()

		// Prometheus metrics 端点
		mux.Handle("/metrics", promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))

		// 健康检查
		mux.HandleFunc("/health", a.healthHandler)

		// 信息模块 API 端点
		mux.HandleFunc("/api/v1/machine-info", a.machineInfoHandler)
		mux.HandleFunc("/api/v1/hardware", a.hardwareHandler)
		mux.HandleFunc("/api/v1/anomaly", a.anomalyHandler)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", a.config.MetricsPort),
			Handler: mux,
		}

		fmt.Printf("HTTP server listening on :%d\n", a.config.MetricsPort)

		go func() {
			<-a.ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// healthHandler 健康检查处理器
func (a *Agent) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "healthy",
		Data: map[string]interface{}{
			"timestamp":    time.Now().Format(time.RFC3339),
			"node_name":    a.config.NodeName,
			"uptime":       time.Since(startTime).String(),
			"version":      "1.0.0",
			"metrics_port": a.config.MetricsPort,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// dataProvider 数据提供者接口
type dataProvider interface {
	GetData() (interface{}, error)
}

// writeJSONResponse 写入 JSON 响应
func writeJSONResponse(w http.ResponseWriter, statusCode int, resp interface{}) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// handleDataRequest 处理数据请求
func (a *Agent) handleDataRequest(w http.ResponseWriter, provider dataProvider) {
	w.Header().Set("Content-Type", "application/json")

	data, err := provider.GetData()
	if err != nil {
		resp := response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: err.Error(),
		}
		writeJSONResponse(w, http.StatusInternalServerError, resp)
		return
	}

	resp := response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "success",
		Data:    data,
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// machineInfoHandler 机器信息处理器
func (a *Agent) machineInfoHandler(w http.ResponseWriter, r *http.Request) {
	a.handleDataRequest(w, a.machineInfo)
}

// hardwareHandler 硬件信息处理器
func (a *Agent) hardwareHandler(w http.ResponseWriter, r *http.Request) {
	a.handleDataRequest(w, a.hardware)
}

// anomalyHandler 异常检测处理器
func (a *Agent) anomalyHandler(w http.ResponseWriter, r *http.Request) {
	a.handleDataRequest(w, a.anomaly)
}

// Stop 停止 Agent
func (a *Agent) Stop() {
	log.Println("Shutting down Agent...")

	// 停止所有模块
	if err := a.metrics.Stop(); err != nil {
		log.Printf("Error stopping metrics module: %v", err)
	}
	if err := a.machineInfo.Stop(); err != nil {
		log.Printf("Error stopping machine info module: %v", err)
	}
	if err := a.hardware.Stop(); err != nil {
		log.Printf("Error stopping hardware module: %v", err)
	}
	if err := a.anomaly.Stop(); err != nil {
		log.Printf("Error stopping anomaly module: %v", err)
	}

	// 取消上下文
	a.cancel()

	// 等待所有 goroutine 结束
	a.wg.Wait()

	log.Println("Agent shutdown complete")
}

// Run 运行 Agent 主循环
func (a *Agent) Run() error {
	// 启动 Agent
	if err := a.Start(); err != nil {
		return err
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-a.ctx.Done():
		log.Println("Context cancelled")
	}

	// 停止 Agent
	a.Stop()
	return nil
}
