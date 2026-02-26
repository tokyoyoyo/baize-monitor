package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"baize-monitor/internal/agent/plugins"
)

// PluginManager 插件管理器
type PluginManager struct {
	plugins   map[string]plugins.Plugin
	isRunning bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]plugins.Plugin),
	}
}

// RegisterPlugin 注册插件
func (pm *PluginManager) RegisterPlugin(plugin plugins.Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	pm.plugins[name] = plugin
	log.Printf("Plugin registered: %s", name)
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

	// 停止插件执行
	plugin.SetEnabled(false)

	delete(pm.plugins, name)
	log.Printf("Plugin unregistered: %s", name)
	return nil
}

// ListPlugins 列出所有插件信息
func (pm *PluginManager) ListPlugins() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make([]map[string]interface{}, 0, len(pm.plugins))
	for name, plugin := range pm.plugins {
		result = append(result, map[string]interface{}{
			"name":        name,
			"description": plugin.Description(),
			"tool":        plugin.Tool(),
			"enabled":     plugin.Enabled(),
			"interval":    plugin.Interval().String(),
			"last_status": string(plugin.LastExecutionStatus()),
			"last_time":   plugin.LastExecutionTime().Format(time.RFC3339),
		})
	}

	return result
}

// EnablePlugin 启用插件
func (pm *PluginManager) EnablePlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.SetEnabled(true)
	log.Printf("Plugin enabled: %s", name)
	return nil
}

// DisablePlugin 禁用插件
func (pm *PluginManager) DisablePlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.SetEnabled(false)
	log.Printf("Plugin disabled: %s", name)
	return nil
}

// GetPlugin 获取插件
func (pm *PluginManager) GetPlugin(name string) (plugins.Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
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

	// 启动所有启用的插件
	pm.mutex.RLock()
	for name, plugin := range pm.plugins {
		if plugin.Enabled() {
			pm.startPlugin(plugin)
		} else {
			log.Printf("Plugin %s is disabled, skipping startup", name)
		}
	}
	pm.mutex.RUnlock()

	log.Println("Plugin manager started")
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

	// 禁用所有插件
	pm.mutex.RLock()
	for name, plugin := range pm.plugins {
		log.Printf("Disabling plugin: %s", name)
		plugin.SetEnabled(false)
	}
	pm.mutex.RUnlock()

	pm.wg.Wait()
	log.Println("Plugin manager stopped")
}

// startPlugin 启动单个插件
func (pm *PluginManager) startPlugin(plugin plugins.Plugin) {
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.runPlugin(plugin)
	}()

	log.Printf("Plugin started: %s", plugin.Name())
}

// runPlugin 运行插件循环
func (pm *PluginManager) runPlugin(plugin plugins.Plugin) {
	ticker := time.NewTicker(plugin.Interval())
	defer ticker.Stop()

	// 立即执行一次
	pm.executePlugin(plugin)

	for {
		select {
		case <-ticker.C:
			if plugin.Enabled() {
				pm.executePlugin(plugin)
			}
		case <-pm.ctx.Done():
			return
		}
	}
}

// executePlugin 执行插件
func (pm *PluginManager) executePlugin(plugin plugins.Plugin) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Plugin %s panic: %v", plugin.Name(), r)
		}
	}()

	log.Printf("Executing plugin: %s", plugin.Name())

	// 执行插件
	result, err := plugin.Execute()
	if err != nil {
		log.Printf("Plugin %s execution failed: %v", plugin.Name(), err)
	} else {
		log.Printf("Plugin %s executed successfully, result: %v", plugin.Name(), result)
	}
}
