package response

// CPUResponse CPU信息响应
type CPUResponse struct {
	Model    string `json:"model"`    // CPU型号
	Vendor   string `json:"vendor"`   // 制造商
	Family   string `json:"family"`   // CPU家族
	Stepping string `json:"stepping"` // 步进
	Flags    string `json:"flags"`    // CPU特性标志

	// 硬件规格
	Cores        int    `json:"cores"`         // 物理核心数
	Threads      int    `json:"threads"`       // 逻辑线程数
	BaseSpeed    string `json:"base_speed"`    // 基础频率
	CacheSizeL1  string `json:"cache_size_l1"` // L1缓存大小
	CacheSizeL2  string `json:"cache_size_l2"` // L2缓存大小
	CacheSizeL3  string `json:"cache_size_l3"` // L3缓存大小
	Architecture string `json:"architecture"`  // 架构
}
