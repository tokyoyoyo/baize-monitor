package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// 导入各个Exporter
	anomalyExporter "baize-monitor/internal/agent/exporters/anomaly"
	hardwareExporter "baize-monitor/internal/agent/exporters/hardware"
	machineExporter "baize-monitor/internal/agent/exporters/machine_info"
	metricsExporter "baize-monitor/internal/agent/exporters/metrics"

	// 导入插件接口
	"baize-monitor/internal/agent/plugins"
)

// AgentConfig Agent配置结构
type AgentConfig struct {
	ServerURL         string        `yaml:"server_url"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	MetricsPort       int           `yaml:"metrics_port"`
	NodeName          string        `yaml:"node_name"`
	LogLevel          string        `yaml:"log_level"`
}

// Agent Agent核心结构
type Agent struct {
	config    *AgentConfig
	registry  *prometheus.Registry
	exporters map[string]Exporter
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mutex     sync.RWMutex
}

// NewAgent 创建Agent实例
func NewAgent(config *AgentConfig) *Agent {
	// 记录启动时间
	startTime = time.Now()

	ctx, cancel := context.WithCancel(context.Background())

	registry := prometheus.NewRegistry()

	return &Agent{
		config:    config,
		registry:  registry,
		exporters: make(map[string]Exporter),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// RegisterExporter 注册Exporter
func (a *Agent) RegisterExporter(name string, exporter Exporter) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.exporters[name]; exists {
		return fmt.Errorf("exporter %s already registered", name)
	}

	if err := a.registry.Register(exporter); err != nil {
		return fmt.Errorf("failed to register %s exporter: %w", name, err)
	}

	a.exporters[name] = exporter
	log.Printf("Exporter registered: %s (%s)", name, exporter.Name())
	return nil
}

// UnregisterExporter 注销Exporter
func (a *Agent) UnregisterExporter(name string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	exporter, exists := a.exporters[name]
	if !exists {
		return fmt.Errorf("exporter %s not found", name)
	}

	a.registry.Unregister(exporter)
	delete(a.exporters, name)

	log.Printf("Exporter unregistered: %s", name)
	return nil
}

// GetExporter 获取Exporter
func (a *Agent) GetExporter(name string) (Exporter, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	exporter, exists := a.exporters[name]
	if !exists {
		return nil, fmt.Errorf("exporter %s not found", name)
	}

	return exporter, nil
}

// InitializeExporters 初始化所有Exporter并使用已注册的插件
func (a *Agent) InitializeExporters() error {
	log.Println("Initializing exporters with registered plugins...")

	// 定义Exporter配置
	exporterConfigs := []struct {
		name    string
		newFunc func() Exporter
	}{
		{
			name:    "machine_info",
			newFunc: func() Exporter { return machineExporter.NewExporter() },
		},
		{
			name:    "hardware",
			newFunc: func() Exporter { return hardwareExporter.NewExporter() },
		},
		{
			name:    "metrics",
			newFunc: func() Exporter { return metricsExporter.NewExporter() },
		},
		{
			name:    "anomaly",
			newFunc: func() Exporter { return anomalyExporter.NewExporter() },
		},
	}

	// 依次初始化每个Exporter并加载其已注册的插件
	for _, config := range exporterConfigs {
		if err := a.initializeExporterWithRegisteredPlugins(config.name, config.newFunc); err != nil {
			return fmt.Errorf("failed to initialize %s exporter: %w", config.name, err)
		}
	}

	log.Printf("Initialized %d exporters with registered plugins", len(a.exporters))
	return nil
}

// getRegisteredPluginsByExporter 根据Exporter类型获取已注册的插件
func (a *Agent) getRegisteredPluginsByExporter(exporterName string) []plugins.Plugin {
	switch exporterName {
	case "machine_info":
		return plugins.MachineInfoPlugins.GetAllPlugins()

	case "hardware":
		return plugins.HardwarePlugins.GetAllPlugins()

	case "metrics":
		return plugins.MetricsPlugins.GetAllPlugins()

	case "anomaly":
		return plugins.AnomalyPlugins.GetAllPlugins()

	default:
		log.Printf("Unknown exporter type: %s", exporterName)
		return []plugins.Plugin{}
	}
}

// initializeExporterWithRegisteredPlugins 初始化Exporter并加载已注册的插件
func (a *Agent) initializeExporterWithRegisteredPlugins(name string, newFunc func() Exporter) error {
	// 创建Exporter
	exporter := newFunc()

	// 注册Exporter
	if err := a.RegisterExporter(name, exporter); err != nil {
		return fmt.Errorf("failed to register exporter: %w", err)
	}

	// 获取对应类型的已注册插件
	registeredPlugins := a.getRegisteredPluginsByExporter(name)

	// 注册发现的插件
	for _, plugin := range registeredPlugins {
		if err := exporter.RegisterPlugin(plugin); err != nil {
			log.Printf("Warning: failed to register plugin %s: %v", plugin.Name(), err)
		} else {
			log.Printf("Plugin registered: %s (%s)", plugin.Name(), plugin.Description())
		}
	}

	log.Printf("Exporter %s initialized with %d registered plugins", name, len(registeredPlugins))
	return nil
}

// Start 启动Agent
func (a *Agent) Start() error {
	log.Printf("Starting BaiZe Agent on node: %s", a.config.NodeName)

	// 初始化Exporter
	if err := a.InitializeExporters(); err != nil {
		return fmt.Errorf("failed to initialize exporters: %w", err)
	}

	// 启动所有Exporter
	a.mutex.RLock()
	for name, exporter := range a.exporters {
		if err := exporter.Start(); err != nil {
			a.mutex.RUnlock()
			return fmt.Errorf("failed to start %s exporter: %w", name, err)
		}
		log.Printf("Exporter started: %s", name)
	}
	a.mutex.RUnlock()

	// 启动指标服务器
	if err := a.startMetricsServer(); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	log.Println("Agent started successfully")
	return nil
}

// startMetricsServer 启动指标服务器
func (a *Agent) startMetricsServer() error {
	// 注册默认收集器
	a.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	a.registry.MustRegister(prometheus.NewGoCollector())

	// 创建HTTP处理器
	handler := promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	// 启动HTTP服务器（goroutine中运行）
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		mux := http.NewServeMux()
		mux.Handle("/metrics", handler)
		mux.HandleFunc("/health", a.healthHandler)
		mux.HandleFunc("/api/v1/plugins", a.handlePluginsAPI)
		mux.HandleFunc("/api/v1/plugin/control", a.handlePluginControlAPI)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", a.config.MetricsPort),
			Handler: mux,
		}

		fmt.Printf("Metrics server listening on :%d\n", a.config.MetricsPort)

		go func() {
			<-a.ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		}()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	return nil
}

// healthHandler 健康检查处理器
func (a *Agent) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 计算运行时间
	uptime := time.Since(startTime).String()

	// 构造健康检查响应
	response := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"node_name":    a.config.NodeName,
		"status":       "healthy",
		"uptime":       uptime,
		"version":      "1.0.0",
		"metrics_port": a.config.MetricsPort,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 添加全局变量记录启动时间
var startTime time.Time

// Stop 停止Agent
func (a *Agent) Stop() {
	log.Println("Shutting down Agent...")

	// 停止所有Exporter
	a.mutex.RLock()
	for name, exporter := range a.exporters {
		if err := exporter.Stop(); err != nil {
			log.Printf("Error stopping %s exporter: %v", name, err)
		} else {
			log.Printf("Exporter stopped: %s", name)
		}
	}
	a.mutex.RUnlock()

	// 取消上下文
	a.cancel()

	// 等待所有goroutine结束
	a.wg.Wait()

	log.Println("Agent shutdown complete")
}

// Run 运行Agent主循环
func (a *Agent) Run() error {
	// 启动Agent
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

	// 停止Agent
	a.Stop()
	return nil
}

// handlePluginsAPI 处理插件列表API
func (a *Agent) handlePluginsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 收集所有插件信息
	var allPlugins []map[string]interface{}

	a.mutex.RLock()
	for _, exporter := range a.exporters {
		for _, plugin := range exporter.GetPlugins() {
			allPlugins = append(allPlugins, map[string]interface{}{
				"name":        plugin.Name(),
				"description": plugin.Description(),
				"tool":        plugin.Tool(),
				"enabled":     plugin.Enabled(),
				"interval":    plugin.Interval().String(),
				"last_status": string(plugin.LastExecutionStatus()),
				"last_time":   plugin.LastExecutionTime().Format(time.RFC3339),
			})
		}
	}
	a.mutex.RUnlock()

	response := map[string]interface{}{
		"plugins": allPlugins,
		"count":   len(allPlugins),
	}

	json.NewEncoder(w).Encode(response)
}

// handlePluginControlAPI 处理插件控制API
func (a *Agent) handlePluginControlAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action string `json:"action"` // enable, disable, list
		Name   string `json:"name"`   // plugin name
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "enable":
		// 启用插件逻辑
		a.enablePlugin(req.Name)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "enabled", "plugin": req.Name})

	case "disable":
		// 禁用插件逻辑
		a.disablePlugin(req.Name)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "plugin": req.Name})

	case "list":
		// 列出插件逻辑已在handlePluginsAPI中实现
		http.Redirect(w, r, "/api/v1/plugins", http.StatusSeeOther)

	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// enablePlugin 启用插件
func (a *Agent) enablePlugin(pluginName string) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, exporter := range a.exporters {
		for _, plugin := range exporter.GetPlugins() {
			if plugin.Name() == pluginName {
				plugin.SetEnabled(true)
				log.Printf("Plugin enabled via API: %s", pluginName)
				return
			}
		}
	}
}

// disablePlugin 禁用插件
func (a *Agent) disablePlugin(pluginName string) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, exporter := range a.exporters {
		for _, plugin := range exporter.GetPlugins() {
			if plugin.Name() == pluginName {
				plugin.SetEnabled(false)
				log.Printf("Plugin disabled via API: %s", pluginName)
				return
			}
		}
	}
}

// GetRegistry 获取指标注册表
func (a *Agent) GetRegistry() *prometheus.Registry {
	return a.registry
}

// GetConfig 获取配置
func (a *Agent) GetConfig() *AgentConfig {
	return a.config
}
