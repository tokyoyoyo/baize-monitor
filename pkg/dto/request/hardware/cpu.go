package request

// CPURequest CPU信息请求
type CPURequest struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	Content []CPUInfo  `json:"content"` // CPU信息
	Summary CPUSummary `json:"summary"` // 摘要信息
}

type CPUInfo struct {
	Model    string `json:"model" binding:"required"`  // CPU型号
	Vendor   string `json:"vendor" binding:"required"` // 制造商
	Family   string `json:"family"`                    // CPU家族
	Stepping string `json:"stepping"`                  // 步进
	Flags    string `json:"flags"`                     // CPU特性标志

	// 硬件规格
	Cores        int    `json:"cores" binding:"required"`      // 物理核心数
	Threads      int    `json:"threads" binding:"required"`    // 逻辑线程数
	BaseSpeed    string `json:"base_speed" binding:"required"` // 基础频率
	CacheSizeL1  string `json:"cache_size_l1"`                 // L1缓存大小
	CacheSizeL2  string `json:"cache_size_l2"`                 // L2缓存大小
	CacheSizeL3  string `json:"cache_size_l3"`                 // L3缓存大小
	Architecture string `json:"architecture"`                  // 架构
}

type CPUSummary struct {
	TotalCores   int `json:"total_cores"`   // 总物理核心数
	TotalThreads int `json:"total_threads"` // 总逻辑线程数
}
