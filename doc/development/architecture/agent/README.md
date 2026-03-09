# BaiZe Agent 架构文档

BaiZe Agent 是白泽监控系统的轻量级客户端，部署在被监控的服务器上，负责采集硬件信息、机器信息、异常检测，并暴露 Prometheus 监控指标。

## 📚 文档导航

### 核心文档
- **[agent.md](agent.md)** - 快速开始

### 模块文档
- **[硬件采集](modules/hardware/README.md)** - CPU、内存、磁盘、网络
- **[异常检测](modules/anomaly/README.md)** - 基于规则的异常检测
- **[机器信息](modules/machine_info/README.md)** - 机器基本信息
- **[指标采集](modules/metrics/README.md)** - Prometheus 指标

---

## 🔗 相关文档

- **[Server 架构](../server/README.md)** - Server 端架构
- **[数据库文档](../../database.md)** - 数据库表结构
- **[项目背景](../../background.md)** - 项目背景和技术栈
