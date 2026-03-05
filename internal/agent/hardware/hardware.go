package hardware

import (
	"time"

	"baize-monitor/internal/agent/hardware/collectors"
)

// Hardware 硬件信息收集器
type Hardware struct {
	registry *collectors.Registry
	enabled  bool
}

// New 创建硬件信息收集器
func New() *Hardware {
	registry := collectors.NewRegistry()
	registry.AutoDiscover()
	return &Hardware{
		registry: registry,
		enabled:  true,
	}
}

// GetData 获取硬件信息数据（实现 dataProvider 接口）
func (h *Hardware) GetData() (interface{}, error) {
	if !h.enabled {
		return nil, nil
	}

	hardwareInfo := h.registry.CollectAll()

	// 设置收集时间戳
	hardwareInfo.CollectedAt = time.Now()

	return hardwareInfo, nil
}

// Start 启动（空实现，保持接口一致）
func (h *Hardware) Start() error {
	return nil
}

// Stop 停止（空实现，保持接口一致）
func (h *Hardware) Stop() error {
	return nil
}
