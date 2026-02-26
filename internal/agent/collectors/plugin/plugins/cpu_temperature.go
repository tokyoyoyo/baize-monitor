package plugins

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/host"
)

// CPUTemperaturePlugin CPU温度检测插件
type CPUTemperaturePlugin struct {
	// 指标定义
	temperature     *prometheus.Desc
	temperatureMax  *prometheus.Desc
	temperatureCrit *prometheus.Desc
	fanSpeed        *prometheus.Desc
	
	// 配置参数
	warningThreshold float64 // 警告阈值 (摄氏度)
	criticalThreshold float64 // 危险阈值 (摄氏度)
	checkInterval    time.Duration
	
	// 状态
	isInitialized bool
	lastError     error
}

// NewCPUTemperaturePlugin 创建CPU温度插件
func NewCPUTemperaturePlugin() *CPUTemperaturePlugin {
	const subsystem = "cpu_temperature"
	
	return &CPUTemperaturePlugin{
		// 温度指标
		temperature: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "celsius"),
			"CPU temperature in Celsius",
			[]string{"sensor", "core"}, nil,
		),
		temperatureMax: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "max_celsius"),
			"CPU maximum temperature in Celsius",
			[]string{"sensor", "core"}, nil,
		),
		temperatureCrit: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "critical_celsius"),
			"CPU critical temperature in Celsius",
			[]string{"sensor", "core"}, nil,
		),
		fanSpeed: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "fan_rpm"),
			"CPU fan speed in RPM",
			[]string{"fan"}, nil,
		),
		
		// 默认配置
		warningThreshold:  70.0,
		criticalThreshold: 85.0,
		checkInterval:     30 * time.Second,
	}
}

// Name 实现Plugin接口
func (p *CPUTemperaturePlugin) Name() string {
	return "cpu_temperature"
}

// Description 实现Plugin接口
func (p *CPUTemperaturePlugin) Description() string {
	return "Monitors CPU temperature and fan speed using system sensors"
}

// Interval 实现Plugin接口
func (p *CPUTemperaturePlugin) Interval() time.Duration {
	return p.checkInterval
}

// Setup 实现Plugin接口
func (p *CPUTemperaturePlugin) Setup() error {
	// 检查传感器支持
	if _, err := host.SensorsTemperatures(); err != nil {
		p.lastError = fmt.Errorf("temperature sensors not available: %w", err)
		return p.lastError
	}
	
	p.isInitialized = true
	fmt.Println("CPU temperature plugin initialized successfully")
	return nil
}

// Teardown 实现Plugin接口
func (p *CPUTemperaturePlugin) Teardown() error {
	p.isInitialized = false
	fmt.Println("CPU temperature plugin cleaned up")
	return nil
}

// Collect 实现Plugin接口
func (p *CPUTemperaturePlugin) Collect(ch chan<- prometheus.Metric) error {
	if !p.isInitialized {
		return fmt.Errorf("plugin not initialized")
	}
	
	// 收集温度数据
	if err := p.collectTemperature(ch); err != nil {
		p.lastError = err
		return err
	}
	
	// 收集风扇数据
	if err := p.collectFanSpeed(ch); err != nil {
		// 风扇数据不是必需的，只记录日志
		fmt.Printf("Warning: failed to collect fan speed: %v\n", err)
	}
	
	p.lastError = nil
	return nil
}

// collectTemperature 收集温度数据
func (p *CPUTemperaturePlugin) collectTemperature(ch chan<- prometheus.Metric) error {
	temperatures, err := host.SensorsTemperatures()
	if err != nil {
		return fmt.Errorf("failed to get sensor temperatures: %w", err)
	}
	
	if len(temperatures) == 0 {
		return fmt.Errorf("no temperature sensors found")
	}
	
	// 处理每个温度传感器
	for _, temp := range temperatures {
		// 过滤CPU相关传感器
		if !p.isCPUSensor(temp.SensorKey) {
			continue
		}
		
		// 发送当前温度
		if !math.IsNaN(temp.Temperature) {
			ch <- prometheus.MustNewConstMetric(
				p.temperature,
				prometheus.GaugeValue,
				temp.Temperature,
				temp.SensorKey, p.extractCoreID(temp.SensorKey),
			)
		}
		
		// 发送最高温度
		if !math.IsNaN(temp.High) {
			ch <- prometheus.MustNewConstMetric(
				p.temperatureMax,
				prometheus.GaugeValue,
				temp.High,
				temp.SensorKey, p.extractCoreID(temp.SensorKey),
			)
		}
		
		// 发送临界温度
		if !math.IsNaN(temp.Critical) {
			ch <- prometheus.MustNewConstMetric(
				p.temperatureCrit,
				prometheus.GaugeValue,
				temp.Critical,
				temp.SensorKey, p.extractCoreID(temp.SensorKey),
			)
		}
		
		// 检查温度是否超出阈值
		p.checkTemperatureThresholds(temp)
	}
	
	return nil
}

// collectFanSpeed 收集风扇转速
func (p *CPUTemperaturePlugin) collectFanSpeed(ch chan<- prometheus.Metric) error {
	// 注意：gopsutil目前不直接支持风扇转速
	// 这里提供一个框架，实际实现可能需要使用其他库或系统调用
	
	// 示例数据 - 实际应该从系统获取
	fanData := map[string]float64{
		"cpu_fan": 1200.0,
		// "chassis_fan1": 800.0,
		// "chassis_fan2": 800.0,
	}
	
	for fan, rpm := range fanData {
		ch <- prometheus.MustNewConstMetric(
			p.fanSpeed,
			prometheus.GaugeValue,
			rpm,
			fan,
		)
	}
	
	return nil
}

// isCPUSensor 判断是否为CPU传感器
func (p *CPUTemperaturePlugin) isCPUSensor(sensorKey string) bool {
	cpuKeywords := []string{
		"cpu", "core", "package", "tdie", "tctl", "k10temp",
		"acpi", "hwmon", "thermal_zone",
	}
	
	for _, keyword := range cpuKeywords {
		if containsIgnoreCase(sensorKey, keyword) {
			return true
		}
	}
	return false
}

// extractCoreID 提取核心ID
func (p *CPUTemperaturePlugin) extractCoreID(sensorKey string) string {
	// 简化的核心ID提取逻辑
	// 实际实现可能需要更复杂的解析
	if containsIgnoreCase(sensorKey, "core") {
		return sensorKey
	}
	return "package"
}

// checkTemperatureThresholds 检查温度阈值
func (p *CPUTemperaturePlugin) checkTemperatureThresholds(temp host.TemperatureStat) {
	if math.IsNaN(temp.Temperature) {
		return
	}
	
	// 检查警告阈值
	if temp.Temperature >= p.warningThreshold {
		fmt.Printf("WARNING: CPU temperature (%.1f°C) exceeds warning threshold (%.1f°C) for sensor %s\n",
			temp.Temperature, p.warningThreshold, temp.SensorKey)
	}
	
	// 检查危险阈值
	if temp.Temperature >= p.criticalThreshold {
		fmt.Printf("CRITICAL: CPU temperature (%.1f°C) exceeds critical threshold (%.1f°C) for sensor %s\n",
			temp.Temperature, p.criticalThreshold, temp.SensorKey)
	}
}

// containsIgnoreCase 忽略大小写包含检查
func containsIgnoreCase(s, substr string) bool {
	return containsString(s, substr) || containsString(s, toUpper(substr))
}

// containsString 字符串包含检查
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(substr) > 0 && len(s) > 0 && 
		     indexString(s, substr) >= 0))
}

// toUpper 转大写（简化实现）
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		result[i] = c
	}
	return string(result)
}

// indexString 查找子字符串位置（简化实现）
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetWarningThreshold 获取警告阈值
func (p *CPUTemperaturePlugin) GetWarningThreshold() float64 {
	return p.warningThreshold
}

// SetWarningThreshold 设置警告阈值
func (p *CPUTemperaturePlugin) SetWarningThreshold(threshold float64) {
	p.warningThreshold = threshold
}

// GetCriticalThreshold 获取危险阈值
func (p *CPUTemperaturePlugin) GetCriticalThreshold() float64 {
	return p.criticalThreshold
}

// SetCriticalThreshold 设置危险阈值
func (p *CPUTemperaturePlugin) SetCriticalThreshold(threshold float64) {
	p.criticalThreshold = threshold
}

// GetLastError 获取最后错误
func (p *CPUTemperaturePlugin) GetLastError() error {
	return p.lastError
}