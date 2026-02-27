package plugins

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"baize-monitor/internal/agent/plugins"
)

// TestPlugin 测试插件，用于验证自动发现机制
type TestPlugin struct {
	name             string
	description      string
	tool             string
	parameters       []string
	interval         time.Duration
	enabled          bool
	lastExecution    time.Time
	lastStatus       plugins.ExecutionStatus
	mutex            sync.RWMutex
}

// init 函数在包初始化时自动注册插件
func init() {
	plugin := NewTestPlugin()
	plugins.MachineInfoPlugins.Register(plugin)
}

// NewTestPlugin 创建测试插件实例
func NewTestPlugin() *TestPlugin {
	return &TestPlugin{
		name:        "test_plugin",
		description: "Test plugin for auto-discovery verification",
		tool:        "echo",
		parameters:  []string{"test"},
		interval:    30 * time.Second,
		enabled:     true,
		lastStatus:  plugins.StatusPending,
	}
}

// Name 返回插件名称
func (t *TestPlugin) Name() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.name
}

// Description 返回插件描述
func (t *TestPlugin) Description() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.description
}

// Tool 返回执行工具
func (t *TestPlugin) Tool() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tool
}

// Parameters 返回执行参数
func (t *TestPlugin) Parameters() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.parameters
}

// Interval 返回执行间隔
func (t *TestPlugin) Interval() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.interval
}

// Enabled 返回是否启用
func (t *TestPlugin) Enabled() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.enabled
}

// SetEnabled 设置启用状态
func (t *TestPlugin) SetEnabled(enabled bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.enabled = enabled
	if !enabled {
		t.lastStatus = plugins.StatusDisabled
	}
}

// LastExecutionStatus 返回最后执行状态
func (t *TestPlugin) LastExecutionStatus() plugins.ExecutionStatus {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.lastStatus
}

// LastExecutionTime 返回最后执行时间
func (t *TestPlugin) LastExecutionTime() time.Time {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.lastExecution
}

// Execute 执行插件逻辑
func (t *TestPlugin) Execute() (interface{}, error) {
	t.mutex.Lock()
	t.lastExecution = time.Now()
	t.lastStatus = plugins.StatusRunning
	t.mutex.Unlock()

	// 简单的测试逻辑
	result := map[string]interface{}{
		"message":   "Test plugin executed successfully",
		"timestamp": t.lastExecution,
		"tool":      t.tool,
	}

	t.mutex.Lock()
	t.lastStatus = plugins.StatusSuccess
	t.mutex.Unlock()

	fmt.Println("Test plugin executed")
	return result, nil
}

// Collect 实现Prometheus Collector接口
func (t *TestPlugin) Collect(ch chan<- prometheus.Metric) {
	// 测试插件的指标收集
}

// Describe 实现Prometheus Collector接口
func (t *TestPlugin) Describe(ch chan<- *prometheus.Desc) {
	// 测试插件的指标描述
}