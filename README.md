# BaiZe - Unified Hardware Monitoring Platform

## Project Overview

BaiZe is a unified hardware monitoring platform designed to address the complexity of hardware monitoring in enterprise data centers. By integrating in-band and out-of-band monitoring mechanisms, BaiZe can standardize BMC (Baseboard Management Controller) alerts from different vendors and provide proactive early warning capabilities to help operations teams detect and resolve hardware failures in advance.

## Architecture Design

The BaiZe platform adopts a distributed architecture, mainly consisting of the following components:

### Server Side

- **SNMP-trap Receiver**: Responsible for receiving SNMP-trap alert information from devices such as BMCs and switches
- **OID Mapping Configuration**: Standardizes vendor-specific alert information into a unified format
- **Alert Processing Engine**: Implements alert compression, correlation analysis, and routing functions
- **Configuration Management Module**: Synchronizes configuration information to various components via Redis

### Machine Agent

Lightweight agent deployed on physical machines, responsible for:

- **Heartbeat Reporting**: Regularly report device status to the server
- **System Log Monitoring**: Monitor abnormal information in system logs in real-time
- **Plugin-based Detection Scripts**: Support for extending various hardware detection scripts

```mermaid
graph TB
    %% Define styles
    classDef core fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef input fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef config fill:#fff3e0,stroke:#ff6f00,stroke-width:2px
    classDef output fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef lock fill:#ffebee,stroke:#c62828,stroke-width:2px

    %% Data Input Layer
    A1[BMCOut-of-Band Alerts]
    A2[Network Switch]
    A3[Machine-AgentIn-Band Monitoring]
    A4[Active Detection Plugins]

    %% BaiZe Core Engine
    B1[SNMP Trap/InformReceiverUDP Listener]

    %% Added: Distributed Lock Control Layer
    B1_5[Message Deduplication & Lock ManagementMD5 Fingerprint + Redis Distributed Lock]

    B2[Standardization EngineV1, V2c, V3]
    B3[OID Semantic MappingAlert Time/Component/Content/Level/SN]
    B4[Intelligent Alert ProcessingCompression-Correlation-Routing]

    %% Configuration & Governance
    C[Unified Configuration CenterRedis]

    %% Output & Control
    D1[Unified ConsoleWeb UI]
    D2[Multi-Channel NotificationEmail/DingTalk/WeChat]
    D3[Event & Ticket System]

    %% Connection Relationships
    A1 -->|SNMP Trap| B1
    A2 -->|SNMP Trap| B1
    A3 -->|SNMP InformHeartbeat/Log/Detection| B1
    A3 --> A4

    B1 -->|Raw Message| B1_5
    B1_5 -->|Lock Acquired| B2
    B1_5 -.->|MD5 Fingerprint| C
    B2 --> B3
    B3 --> B4

    C -->|Sync Configuration| B3
    C -->|Deploy Plugins/Rules| A3

    B4 --> D1
    B4 --> D2
    B4 --> D3

    %% Apply styles
    class B1,B2,B3,B4 core
    class A1,A2,A3,A4 input
    class C config
    class D1,D2,D3 output
    class B1_5 lock

    %% Grouping boxes
    subgraph Data Input Layer
        A1
        A2
        A3
        A4
    end

    subgraph BaiZe Core Engine
        B1
        B1_5
        B2
        B3
        B4
    end

    subgraph Configuration & Governance
        C
    end

    subgraph Output & Control
        D1
        D2
        D3
    end
```

## Core Advantages

1. **Unified**: Unified processing of multi-vendor hardware alerts, eliminating information silos
2. **Standardized**: Private alerts are converted into standard format through OID mapping
3. **Proactive**: Not only passively receives alerts but also actively discovers potential issues
4. **Scalable**: Plugin-based architecture supports flexible extension of detection capabilities
5. **Containerized Deployment**: Supports Docker and Kubernetes deployment for easy operations and maintenance

## Problems Solved

- Insufficient hardware failure early warning
- Vendor tools are isolated, lacking unified management
- System log anomalies are difficult to capture
- Inconsistent alert information formats from multi-vendor devices
- Lack of proactive monitoring capabilities

## Deployment Methods

BaiZe supports multiple deployment methods, including traditional physical machine deployment and containerized deployment, to meet different environmental requirements.

## Community and Contributions

We welcome you to join the BaiZe community and help improve this hardware monitoring platform. Whether you are a developer, operations engineer, or user interested in hardware monitoring, you are welcome to contribute code, documentation, or valuable suggestions to the project.

Submit Issues to report problems or suggest new features
Fork the project and submit Pull Requests to contribute code
Participate in discussions and documentation improvement work

[中文版本](README_CN.md)
