package anomaly

// Anomaly 异常检测模块
type Anomaly struct {
	enabled bool
}

// New 创建异常检测模块
func New() *Anomaly {
	return &Anomaly{
		enabled: true,
	}
}

// GetData 获取异常检测数据（实现 dataProvider 接口）
func (a *Anomaly) GetData() (interface{}, error) {
	if !a.enabled {
		return nil, nil
	}
	
	// TODO: 实现异常检测逻辑
	return map[string]interface{}{
		"status": "not_implemented",
	}, nil
}

// Start 启动异常检测模块
func (a *Anomaly) Start() error {
	return nil
}

// Stop 停止异常检测模块
func (a *Anomaly) Stop() error {
	return nil
}
