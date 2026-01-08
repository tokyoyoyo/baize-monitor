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
	AlterLevelOID     string `json:"alter_level_oid" gorm:"size:500;not null;column:alter_level_oid"`
	AlterContentOID   string `json:"alter_content_oid" gorm:"size:500;not null;column:alter_content_oid"`
	AlterTimeOID      string `json:"alter_time_oid" gorm:"size:500;not null;column:alter_time_oid"`
	AlterComponentOID string `json:"alter_component_oid" gorm:"size:500;not null;column:alter_component_oid"`

	EnableAutoClose bool   `json:"enable_auto_close" gorm:"default:false;column:enable_auto_close"`
	AlterIndexOID   string `json:"alter_index_oid" gorm:"size:500;column:alter_index_oid"`
	AlterStatusOID  string `json:"alter_status_oid" gorm:"size:500;column:alter_status_oid"`

	EnableContactInterComponentAlters  bool   `json:"enable_contact_inter_component_alters" gorm:"default:false;column:enable_contact_inter_component_alters"`
	ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid" gorm:"size:500;column:contact_inter_component_identifier_oid"`

	TimeFormat string `json:"time_format" gorm:"size:200;not null;column:time_format"`

	// 映射配置 - 使用JSONB
	LevelMappings         map[string]AlterLevel  `json:"level_mappings" gorm:"type:jsonb;not null;default:'{}';serializer:json;column:level_mappings"`
	StatusMappings        map[string]AlterStatus `json:"status_mappings" gorm:"type:jsonb;not null;default:'{}';serializer:json;column:status_mappings"`
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

	res.AlterLevelOID = p.AlterLevelOID
	res.AlterContentOID = p.AlterContentOID
	res.AlterTimeOID = p.AlterTimeOID
	res.AlterComponentOID = p.AlterComponentOID

	res.EnableAutoClose = p.EnableAutoClose
	res.AlterIndexOID = p.AlterIndexOID
	res.AlterStatusOID = p.AlterStatusOID

	res.EnableContactInterComponentAlters = p.EnableContactInterComponentAlters
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

// AlterLevel 标准告警级别枚举类型
type AlterLevel string

const (
	Critical     AlterLevel = "critical"
	Warning      AlterLevel = "warning"
	Info         AlterLevel = "info"
	Notification AlterLevel = "notification"
)

// AlterStatus 标准告警状态枚举类型
type AlterStatus string

const (
	Asserted   AlterStatus = "asserted"
	Deasserted AlterStatus = "deasserted"
)

type AlterComponent string

type AlterComponentDescription struct {
	Component                AlterComponent              `json:"component"`
	NameCN                   string                      `json:"name"`
	NameEN                   string                      `json:"name_en"`
	DescriptionCN            string                      `json:"description"`
	DescriptionEN            string                      `json:"description_en"`
	SubComponentsDescription []AlterComponentDescription `json:"sub_components"`
}

// 通用 - 未知部件

const AlterComponentUnknown AlterComponent = "UNKNOWN"

const (
	// ==================== BMC相关告警部件 ====================
	// 处理器相关
	BMC_AlterComponentCPU         AlterComponent = "CPU"
	BMC_AlterComponentCPU_Core    AlterComponent = "CPU_CORE"
	BMC_AlterComponentCPU_Thermal AlterComponent = "CPU_THERMAL"
	BMC_AlterComponentCPU_Voltage AlterComponent = "CPU_VOLTAGE"
	BMC_AlterComponentCPU_Power   AlterComponent = "CPU_POWER"

	// 内存相关
	BMC_AlterComponentMemory         AlterComponent = "MEMORY"
	BMC_AlterComponentMemory_DIMM    AlterComponent = "MEMORY_DIMM"
	BMC_AlterComponentMemory_Channel AlterComponent = "MEMORY_CHANNEL"
	BMC_AlterComponentMemory_Voltage AlterComponent = "MEMORY_VOLTAGE"
	BMC_AlterComponentMemory_Thermal AlterComponent = "MEMORY_THERMAL"
	BMC_AlterComponentMemory_ECC     AlterComponent = "MEMORY_ECC"

	// 存储相关
	BMC_AlterComponentStorage           AlterComponent = "STORAGE"
	BMC_AlterComponentStorage_HDD       AlterComponent = "STORAGE_HDD"
	BMC_AlterComponentStorage_SSD       AlterComponent = "STORAGE_SSD"
	BMC_AlterComponentStorage_NVMe      AlterComponent = "STORAGE_NVME"
	BMC_AlterComponentStorage_RAID      AlterComponent = "STORAGE_RAID"
	BMC_AlterComponentStorage_Backplane AlterComponent = "STORAGE_BACKPLANE"

	// 电源相关
	BMC_AlterComponentPower        AlterComponent = "POWER"
	BMC_AlterComponentPower_Supply AlterComponent = "POWER_SUPPLY"
	BMC_AlterComponentPower_Input  AlterComponent = "POWER_INPUT"
	BMC_AlterComponentPower_Output AlterComponent = "POWER_OUTPUT"
	BMC_AlterComponentPower_Domain AlterComponent = "POWER_DOMAIN"

	// 散热相关
	BMC_AlterComponentCooling          AlterComponent = "COOLING"
	BMC_AlterComponentCooling_Fan      AlterComponent = "COOLING_FAN"
	BMC_AlterComponentCooling_FanZone  AlterComponent = "COOLING_FAN_ZONE"
	BMC_AlterComponentCooling_Pump     AlterComponent = "COOLING_PUMP"
	BMC_AlterComponentCooling_Radiator AlterComponent = "COOLING_RADIATOR"

	// 温度传感器
	BMC_AlterComponentTemperature         AlterComponent = "TEMPERATURE"
	BMC_AlterComponentTemperature_Inlet   AlterComponent = "TEMPERATURE_INLET"
	BMC_AlterComponentTemperature_Outlet  AlterComponent = "TEMPERATURE_OUTLET"
	BMC_AlterComponentTemperature_CPU     AlterComponent = "TEMPERATURE_CPU"
	BMC_AlterComponentTemperature_Memory  AlterComponent = "TEMPERATURE_MEMORY"
	BMC_AlterComponentTemperature_Storage AlterComponent = "TEMPERATURE_STORAGE"
	BMC_AlterComponentTemperature_Power   AlterComponent = "TEMPERATURE_POWER"

	// 主板相关
	BMC_AlterComponentMotherboard          AlterComponent = "MOTHERBOARD"
	BMC_AlterComponentMotherboard_VR       AlterComponent = "MOTHERBOARD_VR"
	BMC_AlterComponentMotherboard_BIOS     AlterComponent = "MOTHERBOARD_BIOS"
	BMC_AlterComponentMotherboard_CMOS     AlterComponent = "MOTHERBOARD_CMOS"
	BMC_AlterComponentMotherboard_Firmware AlterComponent = "MOTHERBOARD_FIRMWARE"

	// PCIe相关
	BMC_AlterComponentPCIe        AlterComponent = "PCIE"
	BMC_AlterComponentPCIe_Slot   AlterComponent = "PCIE_SLOT"
	BMC_AlterComponentPCIe_Device AlterComponent = "PCIE_DEVICE"

	// 网络接口
	BMC_AlterComponentNetwork           AlterComponent = "NETWORK"
	BMC_AlterComponentNetwork_NIC       AlterComponent = "NETWORK_NIC"
	BMC_AlterComponentNetwork_LOM       AlterComponent = "NETWORK_LOM"
	BMC_AlterComponentNetwork_Interface AlterComponent = "NETWORK_INTERFACE"

	// BMC自身
	BMC_AlterComponentBMC_Self     AlterComponent = "BMC_SELF"
	BMC_AlterComponentBMC_Firmware AlterComponent = "BMC_FIRMWARE"
	BMC_AlterComponentBMC_Health   AlterComponent = "BMC_HEALTH"
)

// BMC告警部件详细描述映射
var BMCAlterComponentDescriptions = []AlterComponentDescription{
	// 处理器相关
	{Component: BMC_AlterComponentCPU,
		NameCN:        "CPU",
		NameEN:        "CPU",
		DescriptionCN: "中央处理器，负责执行计算机指令",
		DescriptionEN: "Central Processing Unit, responsible for executing computer instructions",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentCPU_Core,
				NameCN:        "CPU核心",
				NameEN:        "CPU Core",
				DescriptionCN: "CPU中的独立处理单元",
				DescriptionEN: "Independent processing unit in CPU, modern CPUs typically contain multiple cores",
			},
			{
				Component:     BMC_AlterComponentCPU_Thermal,
				NameCN:        "CPU温度",
				NameEN:        "CPU Temperature",
				DescriptionCN: "CPU工作时产生的热量和温度监控",
				DescriptionEN: "Heat and temperature monitoring generated by CPU during operation",
			},
			{
				Component:     BMC_AlterComponentCPU_Voltage,
				NameCN:        "CPU电压",
				NameEN:        "CPU Voltage",
				DescriptionCN: "CPU供电电压监控",
				DescriptionEN: "CPU power supply voltage monitoring",
			},
			{
				Component:     BMC_AlterComponentCPU_Power,
				NameCN:        "CPU功耗",
				NameEN:        "CPU Power Consumption",
				DescriptionCN: "CPU消耗的电能功率",
				DescriptionEN: "Electrical power consumed by CPU",
			},
		},
	},
	{Component: BMC_AlterComponentMemory,
		NameCN:        "内存",
		NameEN:        "Memory",
		DescriptionCN: "计算机随机存取存储器",
		DescriptionEN: "Computer Random Access Memory",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentMemory_DIMM,
				NameCN:        "内存条",
				NameEN:        "Memory DIMM",
				DescriptionCN: "双列直插式内存模块",
				DescriptionEN: "Dual In-line Memory Module",
			},
			{
				Component:     BMC_AlterComponentMemory_Channel,
				NameCN:        "内存通道",
				NameEN:        "Memory Channel",
				DescriptionCN: "内存控制器与内存模块之间的数据传输通道",
				DescriptionEN: "Data transmission channel between memory controller and memory modules",
			},
			{
				Component:     BMC_AlterComponentMemory_Voltage,
				NameCN:        "内存电压",
				NameEN:        "Memory Voltage",
				DescriptionCN: "内存模块供电电压监控",
				DescriptionEN: "Memory module power supply voltage monitoring",
			},
			{
				Component:     BMC_AlterComponentMemory_Thermal,
				NameCN:        "内存温度",
				NameEN:        "Memory Temperature",
				DescriptionCN: "内存模块工作温度监控",
				DescriptionEN: "Memory module operating temperature monitoring",
			},
			{
				Component:     BMC_AlterComponentMemory_ECC,
				NameCN:        "内存ECC错误",
				NameEN:        "Memory ECC Error",
				DescriptionCN: "内存错误检查和纠正功能报告的错误",
				DescriptionEN: "Errors reported by memory Error Checking and Correction function",
			},
		},
	},
	{Component: BMC_AlterComponentStorage,
		NameCN:        "存储设备",
		NameEN:        "Storage Device",
		DescriptionCN: "数据存储设备总称",
		DescriptionEN: "Data storage device general term",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentStorage_HDD,
				NameCN:        "机械硬盘",
				NameEN:        "Hard Disk Drive",
				DescriptionCN: "传统旋转磁盘存储设备",
				DescriptionEN: "Traditional rotating disk storage device",
			},
			{
				Component:     BMC_AlterComponentStorage_SSD,
				NameCN:        "固态硬盘",
				NameEN:        "Solid State Drive",
				DescriptionCN: "基于闪存的存储设备",
				DescriptionEN: "Flash-based storage device",
			},
			{
				Component:     BMC_AlterComponentStorage_NVMe,
				NameCN:        "NVMe硬盘",
				NameEN:        "NVMe Drive",
				DescriptionCN: "通过PCIe接口连接的高性能固态硬盘",
				DescriptionEN: "High-performance solid state drive connected via PCIe interface",
			},
			{
				Component:     BMC_AlterComponentStorage_RAID,
				NameCN:        "RAID阵列",
				NameEN:        "RAID Array",
				DescriptionCN: "磁盘阵列存储技术",
				DescriptionEN: "Redundant Array of Independent Disks storage technology",
			},
			{
				Component:     BMC_AlterComponentStorage_Backplane,
				NameCN:        "存储背板",
				NameEN:        "Storage Backplane",
				DescriptionCN: "连接存储设备的主板背板",
				DescriptionEN: "Motherboard backplane connecting storage devices",
			},
		},
	},
	{Component: BMC_AlterComponentPower,
		NameCN:        "电源模块",
		NameEN:        "Power Module",
		DescriptionCN: "电源管理模块总称",
		DescriptionEN: "Power management module general term",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentPower_Supply,
				NameCN:        "电源供应器",
				NameEN:        "Power Supply",
				DescriptionCN: "将交流电转换为直流电的设备",
				DescriptionEN: "Device that converts AC to DC",
			},
			{
				Component:     BMC_AlterComponentPower_Input,
				NameCN:        "电源输入",
				NameEN:        "Power Input",
				DescriptionCN: "电源模块的输入电路",
				DescriptionEN: "Input circuit of power module",
			},
			{
				Component:     BMC_AlterComponentPower_Output,
				NameCN:        "电源输出",
				NameEN:        "Power Output",
				DescriptionCN: "电源模块的输出电路",
				DescriptionEN: "Output circuit of power module",
			},
			{
				Component:     BMC_AlterComponentPower_Domain,
				NameCN:        "电源域",
				NameEN:        "Power Domain",
				DescriptionCN: "电源管理的不同电压域",
				DescriptionEN: "Different voltage domains of power management",
			},
		},
	},
	{Component: BMC_AlterComponentCooling,
		NameCN:        "散热系统",
		NameEN:        "Cooling System",
		DescriptionCN: "设备散热系统总称",
		DescriptionEN: "Device cooling system general term",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentCooling_Fan,
				NameCN:        "风扇",
				NameEN:        "Fan",
				DescriptionCN: "用于空气对流散热的风扇设备",
				DescriptionEN: "Fan device used for air convection cooling",
			},
			{
				Component:     BMC_AlterComponentCooling_FanZone,
				NameCN:        "风扇区域",
				NameEN:        "Fan Zone",
				DescriptionCN: "一组风扇组成的散热区域",
				DescriptionEN: "Cooling zone composed of a group of fans",
			},
			{
				Component:     BMC_AlterComponentCooling_Pump,
				NameCN:        "水泵",
				NameEN:        "Water Pump",
				DescriptionCN: "液冷系统中的循环水泵",
				DescriptionEN: "Circulating water pump in liquid cooling system",
			},
			{
				Component:     BMC_AlterComponentCooling_Radiator,
				NameCN:        "散热器",
				NameEN:        "Radiator",
				DescriptionCN: "用于散发热量的散热装置",
				DescriptionEN: "Heat dissipation device used to dissipate heat",
			},
		},
	},
	{Component: BMC_AlterComponentTemperature,
		NameCN:        "温度传感器",
		NameEN:        "Temperature Sensor",
		DescriptionCN: "检测设备各部位温度的传感器",
		DescriptionEN: "Sensor detecting temperatures at various parts of the device",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentTemperature_Inlet,
				NameCN:        "进气口温度",
				NameEN:        "Inlet Temperature",
				DescriptionCN: "设备进风口处的环境温度",
				DescriptionEN: "Ambient temperature at device air intake",
			},
			{
				Component:     BMC_AlterComponentTemperature_Outlet,
				NameCN:        "出气口温度",
				NameEN:        "Outlet Temperature",
				DescriptionCN: "设备出风口处的热空气温度",
				DescriptionEN: "Hot air temperature at device air outlet",
			},
			{
				Component:     BMC_AlterComponentTemperature_CPU,
				NameCN:        "CPU温度",
				NameEN:        "CPU Temperature",
				DescriptionCN: "中央处理器工作温度",
				DescriptionEN: "Central Processing Unit operating temperature",
			},
			{
				Component:     BMC_AlterComponentTemperature_Memory,
				NameCN:        "内存温度",
				NameEN:        "Memory Temperature",
				DescriptionCN: "内存模块工作温度",
				DescriptionEN: "Memory module operating temperature",
			},
			{
				Component:     BMC_AlterComponentTemperature_Storage,
				NameCN:        "存储设备温度",
				NameEN:        "Storage Device Temperature",
				DescriptionCN: "存储设备工作温度",
				DescriptionEN: "Storage device operating temperature",
			},
			{
				Component:     BMC_AlterComponentTemperature_Power,
				NameCN:        "电源温度",
				NameEN:        "Power Supply Temperature",
				DescriptionCN: "电源模块工作温度",
				DescriptionEN: "Power module operating temperature",
			},
		},
	},
	{Component: BMC_AlterComponentMotherboard,
		NameCN:        "主板",
		NameEN:        "Motherboard",
		DescriptionCN: "计算机主机板",
		DescriptionEN: "Computer main board",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentMotherboard_VR,
				NameCN:        "主板电压调节器",
				NameEN:        "Motherboard Voltage Regulator",
				DescriptionCN: "为主板各组件提供稳定电压的调节器",
				DescriptionEN: "Regulator providing stable voltage to motherboard components",
			},
			{
				Component:     BMC_AlterComponentMotherboard_BIOS,
				NameCN:        "BIOS固件",
				NameEN:        "BIOS Firmware",
				DescriptionCN: "基本输入输出系统固件",
				DescriptionEN: "Basic Input/Output System firmware",
			},
			{
				Component:     BMC_AlterComponentMotherboard_CMOS,
				NameCN:        "CMOS电池",
				NameEN:        "CMOS Battery",
				DescriptionCN: "为主板CMOS芯片供电的电池",
				DescriptionEN: "Battery powering the motherboard CMOS chip",
			},
			{
				Component:     BMC_AlterComponentMotherboard_Firmware,
				NameCN:        "主板固件",
				NameEN:        "Motherboard Firmware",
				DescriptionCN: "主板上的嵌入式软件系统",
				DescriptionEN: "Embedded software system on motherboard",
			},
		},
	},
	{Component: BMC_AlterComponentPCIe,
		NameCN:        "PCIe总线",
		NameEN:        "PCIe Bus",
		DescriptionCN: "高速串行计算机扩展总线标准",
		DescriptionEN: "High-speed serial computer expansion bus standard",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentPCIe_Slot,
				NameCN:        "PCIe插槽",
				NameEN:        "PCIe Slot",
				DescriptionCN: "主板上PCIe扩展插槽",
				DescriptionEN: "PCIe expansion slot on motherboard",
			},
			{
				Component:     BMC_AlterComponentPCIe_Device,
				NameCN:        "PCIe设备",
				NameEN:        "PCIe Device",
				DescriptionCN: "插入PCIe插槽的扩展设备",
				DescriptionEN: "Expansion device inserted into PCIe slot",
			},
		},
	},
	{Component: BMC_AlterComponentNetwork,
		NameCN:        "网络设备",
		NameEN:        "Network Device",
		DescriptionCN: "网络通信设备总称",
		DescriptionEN: "Network communication device general term",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentNetwork_NIC,
				NameCN:        "网络接口控制器",
				NameEN:        "Network Interface Controller",
				DescriptionCN: "计算机与网络连接的硬件接口",
				DescriptionEN: "Hardware interface for computer to network connection",
			},
			{
				Component:     BMC_AlterComponentNetwork_LOM,
				NameCN:        "板载网络",
				NameEN:        "LOM (LAN on Motherboard)",
				DescriptionCN: "集成在主板上的网络接口",
				DescriptionEN: "Network interface integrated on motherboard",
			},
			{
				Component:     BMC_AlterComponentNetwork_Interface,
				NameCN:        "网络接口",
				NameEN:        "Network Interface",
				DescriptionCN: "网络连接的逻辑或物理接口",
				DescriptionEN: "Logical or physical interface for network connection",
			},
		},
	},
	{Component: BMC_AlterComponentBMC_Self,
		NameCN:        "BMC自身状态",
		NameEN:        "BMC Self Status",
		DescriptionCN: "BMC控制器自身的运行状态",
		DescriptionEN: "Operating status of BMC controller itself",
		SubComponentsDescription: []AlterComponentDescription{
			{
				Component:     BMC_AlterComponentBMC_Firmware,
				NameCN:        "BMC固件",
				NameEN:        "BMC Firmware",
				DescriptionCN: "BMC控制器的固件系统",
				DescriptionEN: "Firmware system of BMC controller",
			},
			{
				Component:     BMC_AlterComponentBMC_Health,
				NameCN:        "BMC健康状态",
				NameEN:        "BMC Health Status",
				DescriptionCN: "BMC控制器整体健康状况",
				DescriptionEN: "Overall health status of BMC controller",
			},
		},
	},
}
