package plugin

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PluginExecutor 插件执行器
type PluginExecutor struct {
	plugin     Plugin
	interval   time.Duration
	timeout    time.Duration
	maxRetries int
	
	executionCount prometheus.Counter
	errorCount     prometheus.Counter
	executionTime  prometheus.Histogram
	lastExecution  prometheus.Gauge
	
	ctx    context.Context
	cancel context.CancelFunc
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success     bool
	Error       error
	Duration    time.Duration
	Metrics     []prometheus.Metric
	Timestamp   time.Time
	RetryCount  int
}

// NewPluginExecutor 创建插件执行器
func NewPluginExecutor(plugin Plugin, interval time.Duration) *PluginExecutor {
	const subsystem = "plugin"
	
	labels := prometheus.Labels{"plugin": plugin.Name()}
	
	return &PluginExecutor{
		plugin:     plugin,
		interval:   interval,
		timeout:    interval * 2, // 超时时间为间隔的2倍
		maxRetries: 3,
		
		executionCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "baize",
			Subsystem:   subsystem,
			Name:        "executions_total",
			Help:        "Total number of plugin executions",
			ConstLabels: labels,
		}),
		
		errorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "baize",
			Subsystem:   subsystem,
			Name:        "errors_total",
			Help:        "Total number of plugin execution errors",
			ConstLabels: labels,
		}),
		
		executionTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   "baize",
			Subsystem:   subsystem,
			Name:        "execution_duration_seconds",
			Help:        "Plugin execution duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		}),
		
		lastExecution: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "baize",
			Subsystem:   subsystem,
			Name:        "last_execution_timestamp",
			Help:        "Timestamp of last plugin execution",
			ConstLabels: labels,
		}),
	}
}

// Describe 实现prometheus.Collector接口
func (pe *PluginExecutor) Describe(ch chan<- *prometheus.Desc) {
	pe.executionCount.Describe(ch)
	pe.errorCount.Describe(ch)
	pe.executionTime.Describe(ch)
	pe.lastExecution.Describe(ch)
}

// Collect 实现prometheus.Collector接口
func (pe *PluginExecutor) Collect(ch chan<- prometheus.Metric) {
	pe.executionCount.Collect(ch)
	pe.errorCount.Collect(ch)
	pe.executionTime.Collect(ch)
	pe.lastExecution.Collect(ch)
}

// Execute 执行插件
func (pe *PluginExecutor) Execute() *ExecutionResult {
	startTime := time.Now()
	
	result := &ExecutionResult{
		Timestamp: startTime,
	}
	
	defer func() {
		result.Duration = time.Since(startTime)
		pe.executionCount.Inc()
		pe.lastExecution.Set(float64(startTime.Unix()))
		pe.executionTime.Observe(result.Duration.Seconds())
		
		if !result.Success {
			pe.errorCount.Inc()
		}
	}()
	
	// 捕获panic
	defer func() {
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("plugin panicked: %v", r)
			result.Success = false
			
			// 记录堆栈信息
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			fmt.Printf("Plugin %s panic stack:\n%s\n", pe.plugin.Name(), string(buf[:n]))
		}
	}()
	
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), pe.timeout)
	defer cancel()
	
	// 创建指标通道
	metricCh := make(chan prometheus.Metric, 100)
	
	// 在goroutine中执行插件收集
	done := make(chan error, 1)
	go func() {
		defer close(metricCh)
		done <- pe.plugin.Collect(metricCh)
	}()
	
	// 等待执行完成或超时
	select {
	case err := <-done:
		if err != nil {
			result.Error = fmt.Errorf("plugin collect failed: %w", err)
			result.Success = false
			return result
		}
	case <-ctx.Done():
		result.Error = fmt.Errorf("plugin execution timeout after %v", pe.timeout)
		result.Success = false
		return result
	}
	
	// 收集指标
	close(metricCh)
	for metric := range metricCh {
		result.Metrics = append(result.Metrics, metric)
	}
	
	result.Success = true
	return result
}

// ExecuteWithRetry 带重试的执行
func (pe *PluginExecutor) ExecuteWithRetry(maxRetries int) *ExecutionResult {
	var lastResult *ExecutionResult
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result := pe.Execute()
		lastResult = result
		
		if result.Success {
			if attempt > 0 {
				fmt.Printf("Plugin %s succeeded after %d retries\n", pe.plugin.Name(), attempt)
			}
			return result
		}
		
		// 最后一次尝试不需要等待
		if attempt < maxRetries {
			waitTime := time.Duration(attempt+1) * time.Second
			fmt.Printf("Plugin %s failed (attempt %d/%d), retrying in %v: %v\n", 
				pe.plugin.Name(), attempt+1, maxRetries+1, waitTime, result.Error)
			
			select {
			case <-time.After(waitTime):
				// 继续下一次尝试
			case <-pe.ctx.Done():
				result.Error = fmt.Errorf("execution cancelled: %w", pe.ctx.Err())
				return result
			}
		}
	}
	
	fmt.Printf("Plugin %s failed after %d attempts\n", pe.plugin.Name(), maxRetries+1)
	return lastResult
}

// Start 启动执行器
func (pe *PluginExecutor) Start(ctx context.Context) {
	pe.ctx, pe.cancel = context.WithCancel(ctx)
	
	go pe.executionLoop()
	fmt.Printf("Plugin executor started for %s with interval %v\n", pe.plugin.Name(), pe.interval)
}

// Stop 停止执行器
func (pe *PluginExecutor) Stop() {
	if pe.cancel != nil {
		pe.cancel()
	}
	fmt.Printf("Plugin executor stopped for %s\n", pe.plugin.Name())
}

// executionLoop 执行循环
func (pe *PluginExecutor) executionLoop() {
	ticker := time.NewTicker(pe.interval)
	defer ticker.Stop()
	
	// 立即执行一次
	pe.ExecuteWithRetry(pe.maxRetries)
	
	for {
		select {
		case <-ticker.C:
			pe.ExecuteWithRetry(pe.maxRetries)
		case <-pe.ctx.Done():
			return
		}
	}
}

// GetPlugin 获取插件
func (pe *PluginExecutor) GetPlugin() Plugin {
	return pe.plugin
}

// GetInterval 获取执行间隔
func (pe *PluginExecutor) GetInterval() time.Duration {
	return pe.interval
}

// GetTimeout 获取超时时间
func (pe *PluginExecutor) GetTimeout() time.Duration {
	return pe.timeout
}

// SetTimeout 设置超时时间
func (pe *PluginExecutor) SetTimeout(timeout time.Duration) {
	pe.timeout = timeout
}

// GetMaxRetries 获取最大重试次数
func (pe *PluginExecutor) GetMaxRetries() int {
	return pe.maxRetries
}

// SetMaxRetries 设置最大重试次数
func (pe *PluginExecutor) SetMaxRetries(retries int) {
	pe.maxRetries = retries
}