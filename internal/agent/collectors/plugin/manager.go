package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin 接口定义
type Plugin interface {
	// Name 插件名称
	Name() string
	
	// Description 插件描述
	Description() string
	
	// Interval 执行间隔
	Interval() time.Duration
	
	// Collect 收集指标
	Collect(ch chan<- prometheus.Metric) error
	
	// Setup 初始化插件
	Setup() error
	
	// Teardown 清理插件
	Teardown() error
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name     string        `yaml:"name"`
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
	Config   interface{}   `yaml:"config,omitempty"`
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins    map[string]Plugin
	configs    map[string]*PluginConfig
	registry   *prometheus.Registry
	isRunning  bool
	
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:  make(map[string]Plugin),
		configs:  make(map[string]*PluginConfig),
		registry: prometheus.NewRegistry(),
	}
}

// RegisterPlugin 注册插件
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	
	pm.plugins[name] = plugin
	fmt.Printf("Plugin registered: %s - %s\n", name, plugin.Description())
	return nil
}

// UnregisterPlugin 注销插件
func (pm *PluginManager) UnregisterPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	
	// 清理插件
	if err := plugin.Teardown(); err != nil {
		return fmt.Errorf("failed to teardown plugin %s: %w", name, err)
	}
	
	delete(pm.plugins, name)
	delete(pm.configs, name)
	
	fmt.Printf("Plugin unregistered: %s\n", name)
	return nil
}

// ConfigurePlugin 配置插件
func (pm *PluginManager) ConfigurePlugin(config *PluginConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.configs[config.Name] = config
	fmt.Printf("Plugin configured: %s (enabled: %v, interval: %v)\n", 
		config.Name, config.Enabled, config.Interval)
	return nil
}

// Start 启动插件管理器
func (pm *PluginManager) Start(ctx context.Context) error {
	pm.mutex.Lock()
	if pm.isRunning {
		pm.mutex.Unlock()
		return fmt.Errorf("plugin manager already running")
	}
	pm.isRunning = true
	pm.mutex.Unlock()
	
	pm.ctx, pm.cancel = context.WithCancel(ctx)
	
	// 初始化并启动已启用的插件
	pm.mutex.RLock()
	for name, plugin := range pm.plugins {
		if config, exists := pm.configs[name]; exists && config.Enabled {
			if err := pm.startPlugin(plugin, config); err != nil {
				fmt.Printf("Failed to start plugin %s: %v\n", name, err)
				continue
			}
		}
	}
	pm.mutex.RUnlock()
	
	fmt.Println("Plugin manager started")
	return nil
}

// Stop 停止插件管理器
func (pm *PluginManager) Stop() {
	pm.mutex.Lock()
	if !pm.isRunning {
		pm.mutex.Unlock()
		return
	}
	pm.isRunning = false
	pm.mutex.Unlock()
	
	if pm.cancel != nil {
		pm.cancel()
	}
	
	// 停止所有插件
	pm.mutex.RLock()
	for name, plugin := range pm.plugins {
		fmt.Printf("Stopping plugin: %s\n", name)
		if err := plugin.Teardown(); err != nil {
			fmt.Printf("Error stopping plugin %s: %v\n", name, err)
		}
	}
	pm.mutex.RUnlock()
	
	pm.wg.Wait()
	fmt.Println("Plugin manager stopped")
}

// startPlugin 启动单个插件
func (pm *PluginManager) startPlugin(plugin Plugin, config *PluginConfig) error {
	// 初始化插件
	if err := plugin.Setup(); err != nil {
		return fmt.Errorf("failed to setup plugin %s: %w", plugin.Name(), err)
	}
	
	// 启动插件循环
	pm.wg.Add(1)
	go pm.pluginLoop(plugin, config)
	
	fmt.Printf("Plugin started: %s\n", plugin.Name())
	return nil
}

// pluginLoop 插件执行循环
func (pm *PluginManager) pluginLoop(plugin Plugin, config *PluginConfig) {
	defer pm.wg.Done()
	
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	
	// 立即执行一次
	pm.executePlugin(plugin)
	
	for {
		select {
		case <-ticker.C:
			pm.executePlugin(plugin)
		case <-pm.ctx.Done():
			return
		}
	}
}

// executePlugin 执行插件收集
func (pm *PluginManager) executePlugin(plugin Plugin) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Plugin %s panic: %v\n", plugin.Name(), r)
		}
	}()
	
	// 创建指标通道
	metricCh := make(chan prometheus.Metric, 100)
	
	// 在goroutine中收集指标以避免阻塞
	go func() {
		defer close(metricCh)
		if err := plugin.Collect(metricCh); err != nil {
			fmt.Printf("Plugin %s collect error: %v\n", plugin.Name(), err)
		}
	}()
	
	// 处理收集到的指标
	for metric := range metricCh {
		// 这里可以将指标发送到注册表或其他地方
		// 目前简单记录指标数量
		_ = metric // 避免未使用变量警告
		fmt.Printf("Plugin %s collected metric\n", plugin.Name())
	}
}

// GetPlugins 获取所有插件信息
func (pm *PluginManager) GetPlugins() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	result := make([]map[string]interface{}, 0, len(pm.plugins))
	for name, plugin := range pm.plugins {
		config := pm.configs[name]
		status := "stopped"
		if config != nil && config.Enabled {
			status = "running"
		}
		
		result = append(result, map[string]interface{}{
			"name":        name,
			"description": plugin.Description(),
			"status":      status,
			"interval":    plugin.Interval().String(),
		})
	}
	
	return result
}

// IsRunning 检查管理器是否运行
func (pm *PluginManager) IsRunning() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.isRunning
}

// GetRegistry 获取指标注册表
func (pm *PluginManager) GetRegistry() *prometheus.Registry {
	return pm.registry
}