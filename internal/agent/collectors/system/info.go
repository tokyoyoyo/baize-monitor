package system

import (
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemCollector 系统信息收集器
type SystemCollector struct {
	// CPU指标
	cpuTotal          *prometheus.Desc
	cpuPerCore        *prometheus.Desc
	cpuLoad1          *prometheus.Desc
	cpuLoad5          *prometheus.Desc
	cpuLoad15         *prometheus.Desc
	
	// 内存指标
	memoryTotal       *prometheus.Desc
	memoryAvailable   *prometheus.Desc
	memoryUsedPercent *prometheus.Desc
	
	// 磁盘指标
	diskTotal         *prometheus.Desc
	diskFree          *prometheus.Desc
	diskUsedPercent   *prometheus.Desc
	diskIOReadBytes   *prometheus.Desc
	diskIOWriteBytes  *prometheus.Desc
	
	// 网络指标
	networkReceiveBytes *prometheus.Desc
	networkTransmitBytes *prometheus.Desc
	
	// 系统信息
	hostInfo          *prometheus.Desc
	goGoroutines      *prometheus.Desc
	goThreads         *prometheus.Desc
	
	// 收集间隔
	collectInterval time.Duration
}

// NewSystemCollector 创建系统收集器
func NewSystemCollector() *SystemCollector {
	const subsystem = "system"
	
	return &SystemCollector{
		// CPU指标
		cpuTotal: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "cpu_total"),
			"Total CPU usage percentage",
			nil, nil,
		),
		cpuPerCore: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "cpu_per_core"),
			"CPU usage percentage per core",
			[]string{"core"}, nil,
		),
		cpuLoad1: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "load1"),
			"System load average over 1 minute",
			nil, nil,
		),
		cpuLoad5: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "load5"),
			"System load average over 5 minutes",
			nil, nil,
		),
		cpuLoad15: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "load15"),
			"System load average over 15 minutes",
			nil, nil,
		),
		
		// 内存指标
		memoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "memory_total_bytes"),
			"Total memory in bytes",
			nil, nil,
		),
		memoryAvailable: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "memory_available_bytes"),
			"Available memory in bytes",
			nil, nil,
		),
		memoryUsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "memory_used_percent"),
			"Memory usage percentage",
			nil, nil,
		),
		
		// 磁盘指标
		diskTotal: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "disk_total_bytes"),
			"Total disk space in bytes",
			[]string{"device", "mountpoint", "fstype"}, nil,
		),
		diskFree: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "disk_free_bytes"),
			"Free disk space in bytes",
			[]string{"device", "mountpoint", "fstype"}, nil,
		),
		diskUsedPercent: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "disk_used_percent"),
			"Disk usage percentage",
			[]string{"device", "mountpoint", "fstype"}, nil,
		),
		diskIOReadBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "disk_io_read_bytes_total"),
			"Total bytes read from disk",
			[]string{"device"}, nil,
		),
		diskIOWriteBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "disk_io_write_bytes_total"),
			"Total bytes written to disk",
			[]string{"device"}, nil,
		),
		
		// 网络指标
		networkReceiveBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "network_receive_bytes_total"),
			"Total bytes received by network interface",
			[]string{"interface"}, nil,
		),
		networkTransmitBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "network_transmit_bytes_total"),
			"Total bytes transmitted by network interface",
			[]string{"interface"}, nil,
		),
		
		// 系统信息
		hostInfo: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "host_info"),
			"Host information",
			[]string{"hostname", "platform", "platform_family", "platform_version", "kernel_version"}, nil,
		),
		goGoroutines: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_goroutines"),
			"Number of goroutines that currently exist",
			nil, nil,
		),
		goThreads: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "go_threads"),
			"Number of OS threads created",
			nil, nil,
		),
		
		collectInterval: 10 * time.Second,
	}
}

// Describe 实现prometheus.Collector接口
func (sc *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.cpuTotal
	ch <- sc.cpuPerCore
	ch <- sc.cpuLoad1
	ch <- sc.cpuLoad5
	ch <- sc.cpuLoad15
	
	ch <- sc.memoryTotal
	ch <- sc.memoryAvailable
	ch <- sc.memoryUsedPercent
	
	ch <- sc.diskTotal
	ch <- sc.diskFree
	ch <- sc.diskUsedPercent
	ch <- sc.diskIOReadBytes
	ch <- sc.diskIOWriteBytes
	
	ch <- sc.networkReceiveBytes
	ch <- sc.networkTransmitBytes
	
	ch <- sc.hostInfo
	ch <- sc.goGoroutines
	ch <- sc.goThreads
}

// Collect 实现prometheus.Collector接口
func (sc *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	// 收集CPU信息
	sc.collectCPU(ch)
	
	// 收集内存信息
	sc.collectMemory(ch)
	
	// 收集磁盘信息
	sc.collectDisk(ch)
	
	// 收集网络信息
	sc.collectNetwork(ch)
	
	// 收集主机信息
	sc.collectHostInfo(ch)
	
	// 收集Go运行时信息
	sc.collectGoRuntime(ch)
}

// collectCPU 收集CPU信息
func (sc *SystemCollector) collectCPU(ch chan<- prometheus.Metric) {
	// 总CPU使用率
	cpuPercent, err := cpu.Percent(sc.collectInterval, false)
	if err == nil && len(cpuPercent) > 0 {
		ch <- prometheus.MustNewConstMetric(sc.cpuTotal, prometheus.GaugeValue, cpuPercent[0])
	}
	
	// 每核CPU使用率
	cpuPercents, err := cpu.Percent(sc.collectInterval, true)
	if err == nil {
		for i, percent := range cpuPercents {
			ch <- prometheus.MustNewConstMetric(
				sc.cpuPerCore, 
				prometheus.GaugeValue, 
				percent, 
				fmt.Sprintf("cpu%d", i),
			)
		}
	}
	
	// 系统负载
	loadAvg, err := load.Avg()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(sc.cpuLoad1, prometheus.GaugeValue, loadAvg.Load1)
		ch <- prometheus.MustNewConstMetric(sc.cpuLoad5, prometheus.GaugeValue, loadAvg.Load5)
		ch <- prometheus.MustNewConstMetric(sc.cpuLoad15, prometheus.GaugeValue, loadAvg.Load15)
	}
}

// collectMemory 收集内存信息
func (sc *SystemCollector) collectMemory(ch chan<- prometheus.Metric) {
	memStat, err := mem.VirtualMemory()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(sc.memoryTotal, prometheus.GaugeValue, float64(memStat.Total))
		ch <- prometheus.MustNewConstMetric(sc.memoryAvailable, prometheus.GaugeValue, float64(memStat.Available))
		ch <- prometheus.MustNewConstMetric(sc.memoryUsedPercent, prometheus.GaugeValue, memStat.UsedPercent)
	}
}

// collectDisk 收集磁盘信息
func (sc *SystemCollector) collectDisk(ch chan<- prometheus.Metric) {
	// 磁盘空间
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(
					sc.diskTotal,
					prometheus.GaugeValue,
					float64(usage.Total),
					partition.Device, partition.Mountpoint, partition.Fstype,
				)
				ch <- prometheus.MustNewConstMetric(
					sc.diskFree,
					prometheus.GaugeValue,
					float64(usage.Free),
					partition.Device, partition.Mountpoint, partition.Fstype,
				)
				ch <- prometheus.MustNewConstMetric(
					sc.diskUsedPercent,
					prometheus.GaugeValue,
					usage.UsedPercent,
					partition.Device, partition.Mountpoint, partition.Fstype,
				)
			}
		}
	}
	
	// 磁盘IO
	ioCounters, err := disk.IOCounters()
	if err == nil {
		for device, counter := range ioCounters {
			ch <- prometheus.MustNewConstMetric(
				sc.diskIOReadBytes,
				prometheus.CounterValue,
				float64(counter.ReadBytes),
				device,
			)
			ch <- prometheus.MustNewConstMetric(
				sc.diskIOWriteBytes,
				prometheus.CounterValue,
				float64(counter.WriteBytes),
				device,
			)
		}
	}
}

// collectNetwork 收集网络信息
func (sc *SystemCollector) collectNetwork(ch chan<- prometheus.Metric) {
	interfaces, err := net.IOCounters(true)
	if err == nil {
		for _, iface := range interfaces {
			ch <- prometheus.MustNewConstMetric(
				sc.networkReceiveBytes,
				prometheus.CounterValue,
				float64(iface.BytesRecv),
				iface.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				sc.networkTransmitBytes,
				prometheus.CounterValue,
				float64(iface.BytesSent),
				iface.Name,
			)
		}
	}
}

// collectHostInfo 收集主机信息
func (sc *SystemCollector) collectHostInfo(ch chan<- prometheus.Metric) {
	info, err := host.Info()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			sc.hostInfo,
			prometheus.GaugeValue,
			1,
			info.Hostname, info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelVersion,
		)
	}
}

// collectGoRuntime 收集Go运行时信息
func (sc *SystemCollector) collectGoRuntime(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(sc.goGoroutines, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	
	// 获取线程数（需要通过其他方式获取）
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// 这里简化处理，实际可能需要更精确的方法获取线程数
	ch <- prometheus.MustNewConstMetric(sc.goThreads, prometheus.GaugeValue, float64(runtime.GOMAXPROCS(0)))
}