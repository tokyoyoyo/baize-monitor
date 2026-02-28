package plugins

import (
	"baize-monitor/internal/agent/core"
	"baize-monitor/pkg/constants"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HostnameCollector 主机名收集插件
type HostnameCollector struct {
	name           string
	description    string
	tools          []string
	parameters     []string
	interval       time.Duration
	enabled        bool
	lastExecution  time.Time
	lastStatus     constants.ExecutionStatus
	mutex          sync.RWMutex
	executionCount int64
	successCount   int64
	failureCount   int64
}

// init 函数在包初始化时自动注册插件
func init() {
	core.MachineInfoPlugins.Register(newHostnameCollector())
}

// NewHostnameCollector 创建主机名收集插件实例
func newHostnameCollector() *HostnameCollector {
	collector := &HostnameCollector{
		name:        "hostname_collector",
		description: "Collect machine hostname information",
		tools:       []string{"hostname"},
		parameters:  []string{},
		interval:    1 * time.Minute,
		enabled:     true,
		lastStatus:  constants.StatusPending,
	}
	return collector
}

// Name 返回插件名称
func (h *HostnameCollector) Name() string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.name
}

// Description 返回插件描述
func (h *HostnameCollector) Description() string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.description
}

// Tool 返回执行工具
func (h *HostnameCollector) Tools() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.tools
}

// Parameters 返回执行参数
func (h *HostnameCollector) Parameters() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.parameters
}

// Interval 返回执行间隔
func (h *HostnameCollector) Interval() time.Duration {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.interval
}

// Execute 执行插件逻辑
func (h *HostnameCollector) Execute() (interface{}, error) {
	h.mutex.Lock()
	h.lastExecution = time.Now()
	h.lastStatus = constants.StatusRunning
	h.executionCount++
	h.mutex.Unlock()

	// 执行hostname命令
	cmd := exec.Command("hostname", h.parameters...)
	output, err := cmd.Output()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err != nil {
		h.lastStatus = constants.StatusFailed
		h.failureCount++
		return nil, fmt.Errorf("failed to execute hostname command: %w", err)
	}

	hostname := strings.TrimSpace(string(output))
	h.lastStatus = constants.StatusSuccess
	h.successCount++

	result := map[string]interface{}{
		"hostname":  hostname,
		"tool":      h.tools[0],
		"timestamp": h.lastExecution,
	}

	return result, nil
}

// LastExecutionStatus 返回最后执行状态
func (h *HostnameCollector) LastExecutionStatus() constants.ExecutionStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.lastStatus
}

// LastExecutionTime 返回最后执行时间
func (h *HostnameCollector) LastExecutionTime() time.Time {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.lastExecution
}

// Enabled 返回是否启用
func (h *HostnameCollector) Enabled() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.enabled
}

// SetEnabled 设置启用状态
func (h *HostnameCollector) SetEnabled(enabled bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.enabled = enabled
	if !enabled {
		h.lastStatus = constants.StatusDisabled
	}
}

// Describe 实现Prometheus Collector接口
func (h *HostnameCollector) Describe(ch chan<- *prometheus.Desc) {
	// 描述插件相关的指标
}

// Collect 实现Prometheus Collector接口
func (h *HostnameCollector) Collect(ch chan<- prometheus.Metric) {
	// 这里可以收集插件相关的指标
	// 例如执行次数、成功率等
}
