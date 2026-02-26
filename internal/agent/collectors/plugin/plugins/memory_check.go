package plugins

import (
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCheckPlugin 内存检测插件
type MemoryCheckPlugin struct {
	// 系统内存指标
	memoryTotal       *prometheus.Desc
	memoryAvailable   *prometheus.Desc
	memoryUsed        *prometheus.Desc
	memoryUsedPercent *prometheus.Desc
	memoryFree        *prometheus.Desc
	memoryBuffers     *prometheus.Desc
	memoryCached      *prometheus.Desc
	
	// Swap内存指标
	swapTotal         *prometheus.Desc
	swapUsed          *prometheus.Desc
	swapFree          *prometheus.Desc
	swapUsedPercent   *prometheus.Desc
	
	// Go运行时内存指标
	goAlloc           *prometheus.Desc
	goSys             *prometheus.Desc
	goMallocs         *prometheus.Desc
	goFrees           *prometheus.Desc
	goGoroutines      *prometheus.Desc
	goHeapAlloc       *prometheus.Desc
	goHeapSys         *prometheus.Desc
	goHeapIdle        *prometheus.Desc
	goHeapInuse       *prometheus.Desc
	
	// 告警指标
	memoryPressure    *prometheus.Desc
	swapPressure      *prometheus.Desc
	
	// 配置参数
	warningThreshold  float64 // 内存警告阈值 (%)
	criticalThreshold float64 // 内存危险阈值 (%)
	swapWarningThreshold float64 // Swap警告阈值 (%)
	
	checkInterval     time.Duration
	isInitialized     bool
	lastError         error
}

// NewMemoryCheckPlugin 创建内存检测插件
func NewMemoryCheckPlugin() *MemoryCheckPlugin {
	const subsystem = "memory"
	
	return &MemoryCheckPlugin{
		// 系统内存指标
		memoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "total_bytes"),
			"Total memory in bytes",
			nil, nil,
		),
		memoryAvailable: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "available_bytes"),
			"Available memory in bytes",
			nil, nil,
		),
		memoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "used_bytes"),
			"Used memory in bytes",
			nil, nil,
		),
		memoryUsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "used_percent"),
			"Memory usage percentage",
			nil, nil,
		),
		memoryFree: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "free_bytes"),
			"Free memory in bytes",
			nil, nil,
		),
		memoryBuffers: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "buffers_bytes"),
			"Memory used by kernel buffers",
			nil, nil,
		),
		memoryCached: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "cached_bytes"),
			"Memory used by cache",
			nil, nil,
		),
		
		// Swap指标
		swapTotal: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "swap_total_bytes"),
			"Total swap memory in bytes",
			nil, nil,
		),
		swapUsed: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "swap_used_bytes"),
			"Used swap memory in bytes",
			nil, nil,
		),
		swapFree: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "swap_free_bytes"),
			"Free swap memory in bytes",
			nil, nil,
		),
		swapUsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "swap_used_percent"),
			"Swap usage percentage",
			nil, nil,
		),
		
		// Go运行时指标
		goAlloc: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_alloc_bytes"),
			"Bytes allocated and not yet freed",
			nil, nil,
		),
		goSys: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_sys_bytes"),
			"Bytes obtained from system",
			nil, nil,
		),
		goMallocs: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_mallocs_total"),
			"Total number of mallocs",
			nil, nil,
		),
		goFrees: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_frees_total"),
			"Total number of frees",
			nil, nil,
		),
		goGoroutines: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_goroutines"),
			"Number of goroutines",
			nil, nil,
		),
		goHeapAlloc: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_heap_alloc_bytes"),
			"Heap bytes allocated and not yet freed",
			nil, nil,
		),
		goHeapSys: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_heap_sys_bytes"),
			"Heap bytes obtained from system",
			nil, nil,
		),
		goHeapIdle: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_heap_idle_bytes"),
			"Heap bytes waiting to be used",
			nil, nil,
		),
		goHeapInuse: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_heap_inuse_bytes"),
			"Heap bytes that are in use",
			nil, nil,
		),
		
		// 压力指标
		memoryPressure: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "pressure"),
			"Memory pressure level (0-100)",
			[]string{"level"}, nil,
		),
		swapPressure: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "swap_pressure"),
			"Swap pressure level (0-100)",
			[]string{"level"}, nil,
		),
		
		// 默认配置
		warningThreshold:     80.0,
		criticalThreshold:    90.0,
		swapWarningThreshold: 50.0,
		checkInterval:        30 * time.Second,
	}
}

// Name 实现Plugin接口
func (p *MemoryCheckPlugin) Name() string {
	return "memory_check"
}

// Description 实现Plugin接口
func (p *MemoryCheckPlugin) Description() string {
	return "Monitors system memory usage, swap usage and Go runtime memory statistics"
}

// Interval 实现Plugin接口
func (p *MemoryCheckPlugin) Interval() time.Duration {
	return p.checkInterval
}

// Setup 实现Plugin接口
func (p *MemoryCheckPlugin) Setup() error {
	p.isInitialized = true
	fmt.Println("Memory check plugin initialized successfully")
	return nil
}

// Teardown 实现Plugin接口
func (p *MemoryCheckPlugin) Teardown() error {
	p.isInitialized = false
	fmt.Println("Memory check plugin cleaned up")
	return nil
}

// Collect 实现Plugin接口
func (p *MemoryCheckPlugin) Collect(ch chan<- prometheus.Metric) error {
	if !p.isInitialized {
		return fmt.Errorf("plugin not initialized")
	}
	
	// 收集系统内存信息
	if err := p.collectSystemMemory(ch); err != nil {
		p.lastError = err
		return err
	}
	
	// 收集Swap信息
	if err := p.collectSwapMemory(ch); err != nil {
		fmt.Printf("Warning: failed to collect swap memory: %v\n", err)
	}
	
	// 收集Go运行时信息
	p.collectGoRuntime(ch)
	
	p.lastError = nil
	return nil
}

// collectSystemMemory 收集系统内存信息
func (p *MemoryCheckPlugin) collectSystemMemory(ch chan<- prometheus.Metric) error {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory stats: %w", err)
	}
	
	// 发送各项内存指标
	ch <- prometheus.MustNewConstMetric(p.memoryTotal, prometheus.GaugeValue, float64(memStat.Total))
	ch <- prometheus.MustNewConstMetric(p.memoryAvailable, prometheus.GaugeValue, float64(memStat.Available))
	ch <- prometheus.MustNewConstMetric(p.memoryUsed, prometheus.GaugeValue, float64(memStat.Used))
	ch <- prometheus.MustNewConstMetric(p.memoryUsedPercent, prometheus.GaugeValue, memStat.UsedPercent)
	ch <- prometheus.MustNewConstMetric(p.memoryFree, prometheus.GaugeValue, float64(memStat.Free))
	ch <- prometheus.MustNewConstMetric(p.memoryBuffers, prometheus.GaugeValue, float64(memStat.Buffers))
	ch <- prometheus.MustNewConstMetric(p.memoryCached, prometheus.GaugeValue, float64(memStat.Cached))
	
	// 检查内存压力
	p.checkMemoryPressure(memStat.UsedPercent, ch)
	
	return nil
}

// collectSwapMemory 收集Swap内存信息
func (p *MemoryCheckPlugin) collectSwapMemory(ch chan<- prometheus.Metric) error {
	swapStat, err := mem.SwapMemory()
	if err != nil {
		return fmt.Errorf("failed to get swap memory stats: %w", err)
	}
	
	// 发送Swap指标
	ch <- prometheus.MustNewConstMetric(p.swapTotal, prometheus.GaugeValue, float64(swapStat.Total))
	ch <- prometheus.MustNewConstMetric(p.swapUsed, prometheus.GaugeValue, float64(swapStat.Used))
	ch <- prometheus.MustNewConstMetric(p.swapFree, prometheus.GaugeValue, float64(swapStat.Free))
	ch <- prometheus.MustNewConstMetric(p.swapUsedPercent, prometheus.GaugeValue, swapStat.UsedPercent)
	
	// 检查Swap压力
	p.checkSwapPressure(swapStat.UsedPercent, ch)
	
	return nil
}

// collectGoRuntime 收集Go运行时内存信息
func (p *MemoryCheckPlugin) collectGoRuntime(ch chan<- prometheus.Metric) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	
	// 发送Go运行时指标
	ch <- prometheus.MustNewConstMetric(p.goAlloc, prometheus.GaugeValue, float64(ms.Alloc))
	ch <- prometheus.MustNewConstMetric(p.goSys, prometheus.GaugeValue, float64(ms.Sys))
	ch <- prometheus.MustNewConstMetric(p.goMallocs, prometheus.CounterValue, float64(ms.Mallocs))
	ch <- prometheus.MustNewConstMetric(p.goFrees, prometheus.CounterValue, float64(ms.Frees))
	ch <- prometheus.MustNewConstMetric(p.goGoroutines, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	ch <- prometheus.MustNewConstMetric(p.goHeapAlloc, prometheus.GaugeValue, float64(ms.HeapAlloc))
	ch <- prometheus.MustNewConstMetric(p.goHeapSys, prometheus.GaugeValue, float64(ms.HeapSys))
	ch <- prometheus.MustNewConstMetric(p.goHeapIdle, prometheus.GaugeValue, float64(ms.HeapIdle))
	ch <- prometheus.MustNewConstMetric(p.goHeapInuse, prometheus.GaugeValue, float64(ms.HeapInuse))
}

// checkMemoryPressure 检查内存压力
func (p *MemoryCheckPlugin) checkMemoryPressure(usedPercent float64, ch chan<- prometheus.Metric) {
	var pressureLevel string
	
	switch {
	case usedPercent >= p.criticalThreshold:
		pressureLevel = "critical"
		fmt.Printf("CRITICAL: Memory usage (%.1f%%) exceeds critical threshold (%.1f%%)\n",
			usedPercent, p.criticalThreshold)
	case usedPercent >= p.warningThreshold:
		pressureLevel = "warning"
		fmt.Printf("WARNING: Memory usage (%.1f%%) exceeds warning threshold (%.1f%%)\n",
			usedPercent, p.warningThreshold)
	default:
		pressureLevel = "normal"
	}
	
	// 发送压力指标
	ch <- prometheus.MustNewConstMetric(
		p.memoryPressure,
		prometheus.GaugeValue,
		usedPercent,
		pressureLevel,
	)
}

// checkSwapPressure 检查Swap压力
func (p *MemoryCheckPlugin) checkSwapPressure(usedPercent float64, ch chan<- prometheus.Metric) {
	var pressureLevel string
	
	if usedPercent >= p.swapWarningThreshold {
		pressureLevel = "warning"
		fmt.Printf("WARNING: Swap usage (%.1f%%) exceeds warning threshold (%.1f%%)\n",
			usedPercent, p.swapWarningThreshold)
	} else {
		pressureLevel = "normal"
	}
	
	// 发送Swap压力指标
	ch <- prometheus.MustNewConstMetric(
		p.swapPressure,
		prometheus.GaugeValue,
		usedPercent,
		pressureLevel,
	)
}

// GetWarningThreshold 获取内存警告阈值
func (p *MemoryCheckPlugin) GetWarningThreshold() float64 {
	return p.warningThreshold
}

// SetWarningThreshold 设置内存警告阈值
func (p *MemoryCheckPlugin) SetWarningThreshold(threshold float64) {
	p.warningThreshold = threshold
}

// GetCriticalThreshold 获取内存危险阈值
func (p *MemoryCheckPlugin) GetCriticalThreshold() float64 {
	return p.criticalThreshold
}

// SetCriticalThreshold 设置内存危险阈值
func (p *MemoryCheckPlugin) SetCriticalThreshold(threshold float64) {
	p.criticalThreshold = threshold
}

// GetSwapWarningThreshold 获取Swap警告阈值
func (p *MemoryCheckPlugin) GetSwapWarningThreshold() float64 {
	return p.swapWarningThreshold
}

// SetSwapWarningThreshold 设置Swap警告阈值
func (p *MemoryCheckPlugin) SetSwapWarningThreshold(threshold float64) {
	p.swapWarningThreshold = threshold
}

// GetLastError 获取最后错误
func (p *MemoryCheckPlugin) GetLastError() error {
	return p.lastError
}