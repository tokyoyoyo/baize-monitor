// pkg/models/bmc_trap_parser.go
package models

import (
	"baize-monitor/pkg/dto/response"
	"errors"
	"time"
)

var ErrDuplicateParserName = errors.New("parser name already exists")

// BMCTrapParser BMC trap 解析器配置模型
type BMCTrapParser struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	ParserName string `json:"parser_name" gorm:"index;size:200;not null;column:parser_name"`
	VendorCode string `json:"vendor_code" gorm:"index;size:100;not null;column:vendor_code"`
	VendorName string `json:"vendor_name" gorm:"index;size:200;not null;column:vendor_name"`

	// 核心OID字段
	AlertLevelOID     string `json:"alert_level_oid" gorm:"size:500;not null;column:alert_level_oid"`
	AlertContentOID   string `json:"alert_content_oid" gorm:"size:500;not null;column:alert_content_oid"`
	AlertTimeOID      string `json:"alert_time_oid" gorm:"size:500;not null;column:alert_time_oid"`
	AlertComponentOID string `json:"alert_component_oid" gorm:"size:500;not null;column:alert_component_oid"`

	EnableAutoClose bool   `json:"enable_auto_close" gorm:"default:false;column:enable_auto_close"`
	AlertIndexOID   string `json:"alert_index_oid" gorm:"size:500;column:alert_index_oid"`
	AlertStatusOID  string `json:"alert_status_oid" gorm:"size:500;column:alert_status_oid"`

	EnableContactInterComponentAlerts  bool   `json:"enable_contact_inter_component_alerts" gorm:"default:false;column:enable_contact_inter_component_alerts"`
	ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid" gorm:"size:500;column:contact_inter_component_identifier_oid"`

	TimeFormat string `json:"time_format" gorm:"size:200;not null;column:time_format"`

	// 映射配置 - 使用JSONB
	LevelMappings         map[string]AlertLevel  `json:"level_mappings" gorm:"type:jsonb;not null;default:'{}';serializer:json;column:level_mappings"`
	StatusMappings        map[string]AlertStatus `json:"status_mappings" gorm:"type:jsonb;not null;default:'{}';serializer:json;column:status_mappings"`
	EnableProductNameList []string               `json:"enable_product_name_list" gorm:"type:jsonb;not null;default:'[]';serializer:json;column:enable_product_name_list"`
	EnableHostNameList    []string               `json:"enable_host_name_list" gorm:"type:jsonb;not null;default:'[]';serializer:json;column:enable_host_name_list"`
	ComponentMappings     map[string][]string    `json:"component_mappings" gorm:"type:jsonb;not null;default:'{}';serializer:json;column:component_mappings"`

	Description string `json:"description" gorm:"type:text;column:description"`
	IsActive    bool   `json:"is_active" gorm:"default:true;column:is_active"`
	IsDeleted   bool   `json:"is_deleted" gorm:"default:false;column:is_deleted"`
}

func (BMCTrapParser) TableName() string {
	return "bmc_trap_parsers"
}

func (p *BMCTrapParser) ToResponse() *response.BMCTrapParserResponse {
	var res response.BMCTrapParserResponse
	res.ID = p.ID
	res.CreatedAt = p.CreatedAt
	res.UpdatedAt = p.UpdatedAt

	res.ParserName = p.ParserName
	res.VendorName = p.VendorName
	res.VendorCode = p.VendorCode

	res.AlertLevelOID = p.AlertLevelOID
	res.AlertContentOID = p.AlertContentOID
	res.AlertTimeOID = p.AlertTimeOID
	res.AlertComponentOID = p.AlertComponentOID

	res.EnableAutoClose = p.EnableAutoClose
	res.AlertIndexOID = p.AlertIndexOID
	res.AlertStatusOID = p.AlertStatusOID

	res.EnableContactInterComponentAlerts = p.EnableContactInterComponentAlerts
	res.ContactInterComponentIdentifierOID = p.ContactInterComponentIdentifierOID

	res.TimeFormat = p.TimeFormat

	res.LevelMappings = make(map[string]string)
	for k, v := range p.LevelMappings {
		res.LevelMappings[k] = string(v)
	}
	res.StatusMappings = make(map[string]string)
	for k, v := range p.StatusMappings {
		res.StatusMappings[k] = string(v)
	}

	res.EnableProductNameList = p.EnableProductNameList
	res.EnableHostNameList = p.EnableHostNameList
	res.ComponentMappings = p.ComponentMappings

	res.Description = p.Description
	res.IsActive = p.IsActive

	return &res
}

// AlertLevel 标准告警级别枚举类型
type AlertLevel string

const (
	Critical     AlertLevel = "critical"
	Warning      AlertLevel = "warning"
	Info         AlertLevel = "info"
	Notification AlertLevel = "notification"
)

// AlertStatus 标准告警状态枚举类型
type AlertStatus string

const (
	Asserted   AlertStatus = "asserted"
	Deasserted AlertStatus = "deasserted"
)

type AlertComponent string

type AlertComponentDescription struct {
	Component                AlertComponent              `json:"component"`
	NameCN                   string                      `json:"name"`
	NameEN                   string                      `json:"name_en"`
	DescriptionCN            string                      `json:"description"`
	DescriptionEN            string                      `json:"description_en"`
	SubComponentsDescription []AlertComponentDescription `json:"sub_components"`
}

// 通用 - 未知部件

const AlertComponentUnknown AlertComponent = "UNKNOWN"

const (
	// ==================== BMC相关告警部件 ====================
	// 处理器相关
	BMC_AlertComponentCPU         AlertComponent = "CPU"
	BMC_AlertComponentCPU_Core    AlertComponent = "CPU_CORE"
	BMC_AlertComponentCPU_Thermal AlertComponent = "CPU_THERMAL"
	BMC_AlertComponentCPU_Voltage AlertComponent = "CPU_VOLTAGE"
	BMC_AlertComponentCPU_Power   AlertComponent = "CPU_POWER"

	// 内存相关
	BMC_AlertComponentMemory         AlertComponent = "MEMORY"
	BMC_AlertComponentMemory_DIMM    AlertComponent = "MEMORY_DIMM"
	BMC_AlertComponentMemory_Channel AlertComponent = "MEMORY_CHANNEL"
	BMC_AlertComponentMemory_Voltage AlertComponent = "MEMORY_VOLTAGE"
	BMC_AlertComponentMemory_Thermal AlertComponent = "MEMORY_THERMAL"
	BMC_AlertComponentMemory_ECC     AlertComponent = "MEMORY_ECC"

	// 存储相关
	BMC_AlertComponentStorage           AlertComponent = "STORAGE"
	BMC_AlertComponentStorage_HDD       AlertComponent = "STORAGE_HDD"
	BMC_AlertComponentStorage_SSD       AlertComponent = "STORAGE_SSD"
	BMC_AlertComponentStorage_NVMe      AlertComponent = "STORAGE_NVME"
	BMC_AlertComponentStorage_RAID      AlertComponent = "STORAGE_RAID"
	BMC_AlertComponentStorage_Backplane AlertComponent = "STORAGE_BACKPLANE"

	// 电源相关
	BMC_AlertComponentPower        AlertComponent = "POWER"
	BMC_AlertComponentPower_Supply AlertComponent = "POWER_SUPPLY"
	BMC_AlertComponentPower_Input  AlertComponent = "POWER_INPUT"
	BMC_AlertComponentPower_Output AlertComponent = "POWER_OUTPUT"
	BMC_AlertComponentPower_Domain AlertComponent = "POWER_DOMAIN"

	// 散热相关
	BMC_AlertComponentCooling          AlertComponent = "COOLING"
	BMC_AlertComponentCooling_Fan      AlertComponent = "COOLING_FAN"
	BMC_AlertComponentCooling_FanZone  AlertComponent = "COOLING_FAN_ZONE"
	BMC_AlertComponentCooling_Pump     AlertComponent = "COOLING_PUMP"
	BMC_AlertComponentCooling_Radiator AlertComponent = "COOLING_RADIATOR"

	// 温度传感器
	BMC_AlertComponentTemperature         AlertComponent = "TEMPERATURE"
	BMC_AlertComponentTemperature_Inlet   AlertComponent = "TEMPERATURE_INLET"
	BMC_AlertComponentTemperature_Outlet  AlertComponent = "TEMPERATURE_OUTLET"
	BMC_AlertComponentTemperature_CPU     AlertComponent = "TEMPERATURE_CPU"
	BMC_AlertComponentTemperature_Memory  AlertComponent = "TEMPERATURE_MEMORY"
	BMC_AlertComponentTemperature_Storage AlertComponent = "TEMPERATURE_STORAGE"
	BMC_AlertComponentTemperature_Power   AlertComponent = "TEMPERATURE_POWER"

	// 主板相关
	BMC_AlertComponentMotherboard          AlertComponent = "MOTHERBOARD"
	BMC_AlertComponentMotherboard_VR       AlertComponent = "MOTHERBOARD_VR"
	BMC_AlertComponentMotherboard_BIOS     AlertComponent = "MOTHERBOARD_BIOS"
	BMC_AlertComponentMotherboard_CMOS     AlertComponent = "MOTHERBOARD_CMOS"
	BMC_AlertComponentMotherboard_Firmware AlertComponent = "MOTHERBOARD_FIRMWARE"

	// PCIe相关
	BMC_AlertComponentPCIe        AlertComponent = "PCIE"
	BMC_AlertComponentPCIe_Slot   AlertComponent = "PCIE_SLOT"
	BMC_AlertComponentPCIe_Device AlertComponent = "PCIE_DEVICE"

	// 网络接口
	BMC_AlertComponentNetwork           AlertComponent = "NETWORK"
	BMC_AlertComponentNetwork_NIC       AlertComponent = "NETWORK_NIC"
	BMC_AlertComponentNetwork_LOM       AlertComponent = "NETWORK_LOM"
	BMC_AlertComponentNetwork_Interface AlertComponent = "NETWORK_INTERFACE"

	// BMC自身
	BMC_AlertComponentBMC_Self     AlertComponent = "BMC_SELF"
	BMC_AlertComponentBMC_Firmware AlertComponent = "BMC_FIRMWARE"
	BMC_AlertComponentBMC_Health   AlertComponent = "BMC_HEALTH"
)

// BMC告警部件详细描述映射
var BMCAlertComponentDescriptions = []AlertComponentDescription{
	// 处理器相关
	{Component: BMC_AlertComponentCPU,
		NameCN:        "CPU",
		NameEN:        "CPU",
		DescriptionCN: "中央处理器，负责执行计算机指令",
		DescriptionEN: "Central Processing Unit, responsible for executing computer instructions",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentCPU_Core,
				NameCN:        "CPU核心",
				NameEN:        "CPU Core",
				DescriptionCN: "CPU中的独立处理单元",
				DescriptionEN: "Independent processing unit in CPU, modern CPUs typically contain multiple cores",
			},
			{
				Component:     BMC_AlertComponentCPU_Thermal,
				NameCN:        "CPU温度",
				NameEN:        "CPU Temperature",
				DescriptionCN: "CPU工作时产生的热量和温度监控",
				DescriptionEN: "Heat and temperature monitoring generated by CPU during operation",
			},
			{
				Component:     BMC_AlertComponentCPU_Voltage,
				NameCN:        "CPU电压",
				NameEN:        "CPU Voltage",
				DescriptionCN: "CPU供电电压监控",
				DescriptionEN: "CPU power supply voltage monitoring",
			},
			{
				Component:     BMC_AlertComponentCPU_Power,
				NameCN:        "CPU功耗",
				NameEN:        "CPU Power Consumption",
				DescriptionCN: "CPU消耗的电能功率",
				DescriptionEN: "Electrical power consumed by CPU",
			},
		},
	},
	{Component: BMC_AlertComponentMemory,
		NameCN:        "内存",
		NameEN:        "Memory",
		DescriptionCN: "计算机随机存取存储器",
		DescriptionEN: "Computer Random Access Memory",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentMemory_DIMM,
				NameCN:        "内存条",
				NameEN:        "Memory DIMM",
				DescriptionCN: "双列直插式内存模块",
				DescriptionEN: "Dual In-line Memory Module",
			},
			{
				Component:     BMC_AlertComponentMemory_Channel,
				NameCN:        "内存通道",
				NameEN:        "Memory Channel",
				DescriptionCN: "内存控制器与内存模块之间的数据传输通道",
				DescriptionEN: "Data transmission channel between memory controller and memory modules",
			},
			{
				Component:     BMC_AlertComponentMemory_Voltage,
				NameCN:        "内存电压",
				NameEN:        "Memory Voltage",
				DescriptionCN: "内存模块供电电压监控",
				DescriptionEN: "Memory module power supply voltage monitoring",
			},
			{
				Component:     BMC_AlertComponentMemory_Thermal,
				NameCN:        "内存温度",
				NameEN:        "Memory Temperature",
				DescriptionCN: "内存模块工作温度监控",
				DescriptionEN: "Memory module operating temperature monitoring",
			},
			{
				Component:     BMC_AlertComponentMemory_ECC,
				NameCN:        "内存ECC错误",
				NameEN:        "Memory ECC Error",
				DescriptionCN: "内存错误检查和纠正功能报告的错误",
				DescriptionEN: "Errors reported by memory Error Checking and Correction function",
			},
		},
	},
	{Component: BMC_AlertComponentStorage,
		NameCN:        "存储设备",
		NameEN:        "Storage Device",
		DescriptionCN: "数据存储设备总称",
		DescriptionEN: "Data storage device general term",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentStorage_HDD,
				NameCN:        "机械硬盘",
				NameEN:        "Hard Disk Drive",
				DescriptionCN: "传统旋转磁盘存储设备",
				DescriptionEN: "Traditional rotating disk storage device",
			},
			{
				Component:     BMC_AlertComponentStorage_SSD,
				NameCN:        "固态硬盘",
				NameEN:        "Solid State Drive",
				DescriptionCN: "基于闪存的存储设备",
				DescriptionEN: "Flash-based storage device",
			},
			{
				Component:     BMC_AlertComponentStorage_NVMe,
				NameCN:        "NVMe硬盘",
				NameEN:        "NVMe Drive",
				DescriptionCN: "通过PCIe接口连接的高性能固态硬盘",
				DescriptionEN: "High-performance solid state drive connected via PCIe interface",
			},
			{
				Component:     BMC_AlertComponentStorage_RAID,
				NameCN:        "RAID阵列",
				NameEN:        "RAID Array",
				DescriptionCN: "磁盘阵列存储技术",
				DescriptionEN: "Redundant Array of Independent Disks storage technology",
			},
			{
				Component:     BMC_AlertComponentStorage_Backplane,
				NameCN:        "存储背板",
				NameEN:        "Storage Backplane",
				DescriptionCN: "连接存储设备的主板背板",
				DescriptionEN: "Motherboard backplane connecting storage devices",
			},
		},
	},
	{Component: BMC_AlertComponentPower,
		NameCN:        "电源模块",
		NameEN:        "Power Module",
		DescriptionCN: "电源管理模块总称",
		DescriptionEN: "Power management module general term",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentPower_Supply,
				NameCN:        "电源供应器",
				NameEN:        "Power Supply",
				DescriptionCN: "将交流电转换为直流电的设备",
				DescriptionEN: "Device that converts AC to DC",
			},
			{
				Component:     BMC_AlertComponentPower_Input,
				NameCN:        "电源输入",
				NameEN:        "Power Input",
				DescriptionCN: "电源模块的输入电路",
				DescriptionEN: "Input circuit of power module",
			},
			{
				Component:     BMC_AlertComponentPower_Output,
				NameCN:        "电源输出",
				NameEN:        "Power Output",
				DescriptionCN: "电源模块的输出电路",
				DescriptionEN: "Output circuit of power module",
			},
			{
				Component:     BMC_AlertComponentPower_Domain,
				NameCN:        "电源域",
				NameEN:        "Power Domain",
				DescriptionCN: "电源管理的不同电压域",
				DescriptionEN: "Different voltage domains of power management",
			},
		},
	},
	{Component: BMC_AlertComponentCooling,
		NameCN:        "散热系统",
		NameEN:        "Cooling System",
		DescriptionCN: "设备散热系统总称",
		DescriptionEN: "Device cooling system general term",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentCooling_Fan,
				NameCN:        "风扇",
				NameEN:        "Fan",
				DescriptionCN: "用于空气对流散热的风扇设备",
				DescriptionEN: "Fan device used for air convection cooling",
			},
			{
				Component:     BMC_AlertComponentCooling_FanZone,
				NameCN:        "风扇区域",
				NameEN:        "Fan Zone",
				DescriptionCN: "一组风扇组成的散热区域",
				DescriptionEN: "Cooling zone composed of a group of fans",
			},
			{
				Component:     BMC_AlertComponentCooling_Pump,
				NameCN:        "水泵",
				NameEN:        "Water Pump",
				DescriptionCN: "液冷系统中的循环水泵",
				DescriptionEN: "Circulating water pump in liquid cooling system",
			},
			{
				Component:     BMC_AlertComponentCooling_Radiator,
				NameCN:        "散热器",
				NameEN:        "Radiator",
				DescriptionCN: "用于散发热量的散热装置",
				DescriptionEN: "Heat dissipation device used to dissipate heat",
			},
		},
	},
	{Component: BMC_AlertComponentTemperature,
		NameCN:        "温度传感器",
		NameEN:        "Temperature Sensor",
		DescriptionCN: "检测设备各部位温度的传感器",
		DescriptionEN: "Sensor detecting temperatures at various parts of the device",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentTemperature_Inlet,
				NameCN:        "进气口温度",
				NameEN:        "Inlet Temperature",
				DescriptionCN: "设备进风口处的环境温度",
				DescriptionEN: "Ambient temperature at device air intake",
			},
			{
				Component:     BMC_AlertComponentTemperature_Outlet,
				NameCN:        "出气口温度",
				NameEN:        "Outlet Temperature",
				DescriptionCN: "设备出风口处的热空气温度",
				DescriptionEN: "Hot air temperature at device air outlet",
			},
			{
				Component:     BMC_AlertComponentTemperature_CPU,
				NameCN:        "CPU温度",
				NameEN:        "CPU Temperature",
				DescriptionCN: "中央处理器工作温度",
				DescriptionEN: "Central Processing Unit operating temperature",
			},
			{
				Component:     BMC_AlertComponentTemperature_Memory,
				NameCN:        "内存温度",
				NameEN:        "Memory Temperature",
				DescriptionCN: "内存模块工作温度",
				DescriptionEN: "Memory module operating temperature",
			},
			{
				Component:     BMC_AlertComponentTemperature_Storage,
				NameCN:        "存储设备温度",
				NameEN:        "Storage Device Temperature",
				DescriptionCN: "存储设备工作温度",
				DescriptionEN: "Storage device operating temperature",
			},
			{
				Component:     BMC_AlertComponentTemperature_Power,
				NameCN:        "电源温度",
				NameEN:        "Power Supply Temperature",
				DescriptionCN: "电源模块工作温度",
				DescriptionEN: "Power module operating temperature",
			},
		},
	},
	{Component: BMC_AlertComponentMotherboard,
		NameCN:        "主板",
		NameEN:        "Motherboard",
		DescriptionCN: "计算机主机板",
		DescriptionEN: "Computer main board",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentMotherboard_VR,
				NameCN:        "主板电压调节器",
				NameEN:        "Motherboard Voltage Regulator",
				DescriptionCN: "为主板各组件提供稳定电压的调节器",
				DescriptionEN: "Regulator providing stable voltage to motherboard components",
			},
			{
				Component:     BMC_AlertComponentMotherboard_BIOS,
				NameCN:        "BIOS固件",
				NameEN:        "BIOS Firmware",
				DescriptionCN: "基本输入输出系统固件",
				DescriptionEN: "Basic Input/Output System firmware",
			},
			{
				Component:     BMC_AlertComponentMotherboard_CMOS,
				NameCN:        "CMOS电池",
				NameEN:        "CMOS Battery",
				DescriptionCN: "为主板CMOS芯片供电的电池",
				DescriptionEN: "Battery powering the motherboard CMOS chip",
			},
			{
				Component:     BMC_AlertComponentMotherboard_Firmware,
				NameCN:        "主板固件",
				NameEN:        "Motherboard Firmware",
				DescriptionCN: "主板上的嵌入式软件系统",
				DescriptionEN: "Embedded software system on motherboard",
			},
		},
	},
	{Component: BMC_AlertComponentPCIe,
		NameCN:        "PCIe总线",
		NameEN:        "PCIe Bus",
		DescriptionCN: "高速串行计算机扩展总线标准",
		DescriptionEN: "High-speed serial computer expansion bus standard",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentPCIe_Slot,
				NameCN:        "PCIe插槽",
				NameEN:        "PCIe Slot",
				DescriptionCN: "主板上PCIe扩展插槽",
				DescriptionEN: "PCIe expansion slot on motherboard",
			},
			{
				Component:     BMC_AlertComponentPCIe_Device,
				NameCN:        "PCIe设备",
				NameEN:        "PCIe Device",
				DescriptionCN: "插入PCIe插槽的扩展设备",
				DescriptionEN: "Expansion device inserted into PCIe slot",
			},
		},
	},
	{Component: BMC_AlertComponentNetwork,
		NameCN:        "网络设备",
		NameEN:        "Network Device",
		DescriptionCN: "网络通信设备总称",
		DescriptionEN: "Network communication device general term",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentNetwork_NIC,
				NameCN:        "网络接口控制器",
				NameEN:        "Network Interface Controller",
				DescriptionCN: "计算机与网络连接的硬件接口",
				DescriptionEN: "Hardware interface for computer to network connection",
			},
			{
				Component:     BMC_AlertComponentNetwork_LOM,
				NameCN:        "板载网络",
				NameEN:        "LOM (LAN on Motherboard)",
				DescriptionCN: "集成在主板上的网络接口",
				DescriptionEN: "Network interface integrated on motherboard",
			},
			{
				Component:     BMC_AlertComponentNetwork_Interface,
				NameCN:        "网络接口",
				NameEN:        "Network Interface",
				DescriptionCN: "网络连接的逻辑或物理接口",
				DescriptionEN: "Logical or physical interface for network connection",
			},
		},
	},
	{Component: BMC_AlertComponentBMC_Self,
		NameCN:        "BMC自身状态",
		NameEN:        "BMC Self Status",
		DescriptionCN: "BMC控制器自身的运行状态",
		DescriptionEN: "Operating status of BMC controller itself",
		SubComponentsDescription: []AlertComponentDescription{
			{
				Component:     BMC_AlertComponentBMC_Firmware,
				NameCN:        "BMC固件",
				NameEN:        "BMC Firmware",
				DescriptionCN: "BMC控制器的固件系统",
				DescriptionEN: "Firmware system of BMC controller",
			},
			{
				Component:     BMC_AlertComponentBMC_Health,
				NameCN:        "BMC健康状态",
				NameEN:        "BMC Health Status",
				DescriptionCN: "BMC控制器整体健康状况",
				DescriptionEN: "Overall health status of BMC controller",
			},
		},
	},
}
