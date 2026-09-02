# 🐉 Butterfly Monitor — 云原生时代的智能可观测中枢

> 蜻蜓之眼，洞察系统每一次脉动。告别固定阈值的盲目告警，让监控真正"懂"你的业务。

---

## 🌟 项目简介

**Butterfly Monitor** 是一款面向云原生架构的**企业级统一可观测性平台**。

它不只是收集指标、展示图表，更是一位"会思考"的运维助手——通过历史数据自学习基线，自动适应昼夜峰谷、周末效应，只在真正异常时才触发告警。

**我们始终相信：告警不是为了告警，而是为了可信；监控不是为了看数，而是为了决策。**

---

## 💡 为什么做这个项目？

这是一个关于"用脚投票"的故事。

市面上的开源监控方案多如牛毛——Prometheus、Grafana、Zabbix、Nagios……随便一搜就是一大堆。但当你真正想用它们监控**业务指标**时，却发现一个尴尬的事实：

**几乎全是基础设施监控**：CPU、内存、磁盘、网络。**业务监控领域，竟然没有成熟的开源方案！**

而企业级业务监控的现实需求非常"接地气"：

- 📈 订单量、支付笔数、活跃用户数……这些真正关乎营收的指标，谁来守护？
- 🌙 业务量有昼夜规律、有周末效应、有大促高峰，固定阈值配到怀疑人生
- 🔔 凌晨低谷天天误报，半夜被手机吵醒；大促期间告警刷屏，真故障反被淹没
- 💰 商业方案倒是不少，动辄几十万起步，中小企业只能望而却步

**既然没人做，那就自己来！**

Butterfly Monitor 就是在这样的背景下诞生的。我们不希望任何一个运维兄弟因为凌晨三点被误报吵醒，也不希望任何一家中小企业因为预算有限而无法拥有完善的业务监控。

这是一个**真正懂业务**的开源监控平台，让每一个团队都能用上企业级的业务监控能力。

---

## 😫 你是否也经历过这样的夜晚？

这些场景，相信每一个运维兄弟都不陌生：

| 场景 | 传统做法 | 结果 |
|------|----------|------|
| **凌晨低谷误报** | 配了固定阈值"订单数 > 1000" | 凌晨 3 点订单自然降到 500，手机狂响，爬起来一看——没事 |
| **阈值调到怀疑人生** | 改了又改，改了还报 | 上周改成 800，这周大促又漏报了，老板问"怎么没发现问题" |
| **大促期间告警刷屏** | 索性关掉全站告警 | 图个清静，结果真故障也被淹没了，大促翻车 |
| **十几个监控平台** | MySQL 一个、Redis 一个、应用一个…… | 出问题时要在多个平台反复切换，排查时间翻倍 |
| **Grafana 面板维护** | 手动建 Dashboard、加 Panel | 任务多了，维护面板比写代码还累，想加个指标都要折腾半天 |

**我们太懂这种痛了。Butterfly Monitor 就是来终结这些噩梦的。**

---

## ✨ 核心亮点

| 维度 | 说明 |
|------|------|
| **异构纳管** | 一套平台统一接入 MySQL / PostgreSQL / ClickHouse / MongoDB / Prometheus / VictoriaMetrics / TDengine / OpenSearch / Elasticsearch 等十余种数据源，打破数据孤岛，全域可观测 |
| **时序引擎** | VictoriaMetrics / TDengine / Prometheus remote_write 可插拔后端，自研批量写入与分片策略，毫秒级写入、秒级查询，无缝对接 Grafana 与云端 |
| **自学习基线** | MAD 稳健聚合 + 中位数平滑，无需人工调参，运行数天自动建立"这个时刻该是多少"的参照系 |
| **精准告警** | 规则组灵活编排，相对偏离检测替代死阈值；告警生命周期完整闭环（Pending → Firing → Processing → Complete），多通道触达（企业微信 / 钉钉 / 飞书 / 邮件） |
| **即建即视** | 创建任务即自动同步 Grafana Dashboard / Panel，实时曲线与基线对比线自动叠加，运维零额外成本 |

---

### 🔮 一站式异构数据纳管

一套平台，统一接入 **14+ 数据源**：

```
MySQL / MariaDB / TiDB / OceanBase / Doris / StarRocks
PostgreSQL / ClickHouse / MongoDB
Prometheus / VictoriaMetrics / TDengine
OpenSearch / Elasticsearch
```

**不用再在十几个平台之间反复切换了。** 所有业务指标，一个平台，一屏掌握。打破数据孤岛，让可观测性真正"可观"。

### 🧠 自学习趋势基线 — 告警从此"懂"业务

这是本项目的**杀手锏**。也是我们最引以为傲的功能。

告别死板阈值，让系统自己学会"这个时刻该是多少"——就像一个经验丰富的老运维，一眼就能看出"这个点不对劲"：

```
采集 → 写入实时值 + 未来8天同时刻样本原料点
      ↓
采样任务 MAD 稳健聚合 + 中位数平滑
      ↓
生成基线指标 {taskKey}_sample
      ↓
比对实时值 vs 基线值 → 相对偏离检测
```

**实际效果**：
- 凌晨低谷自然低于基线 → 不告警 ✅ *（终于可以睡个好觉了）*
- 白天高峰突然跌一半 → 立即告警 🚨 *（真正的故障，绝不放过）*
- 大促期间量级翻 3 倍 → 波动日策略自动放宽，不刷屏 ✅ *（大促不再是噩梦）*

### 📊 Grafana 看板零维护

我们深知运维同学的日常：创建一个监控任务，还得去 Grafana 里拖拽面板、配置查询、调整样式……任务多了，光是维护面板就能累死人。

**现在，解放双手的时候到了。**

创建监控任务时**自动创建 Grafana Dashboard/Panel**，修改任务时**自动同步**。开启"样本展示"后，实时曲线与基线对比线自动叠加。

**创建任务，即有图表。运维再也不用手工拖拽面板了！**

### 🔔 精准告警，完整生命周期

告警不是"响一下就完事了"，它应该有始有终：

```
Normal → Pending（首次命中，等待持续确认）
       → Firing（持续超 Duration，创建事件）
              → Processing（人工处理中）
              → Complete（处理完成）
              → Ignore（误报标定）
```

每一次告警都有迹可循，每一个事件都能闭环。**不再有"这个告警后来怎么样了"的疑惑。**

- **多通道触达**：企业微信 / 钉钉 / 飞书 / 邮件
- **外发退避**：首次延迟 + 后续翻倍退避，告别告警风暴
- **灵活规则组**：组内 And/Or，组间固定 Or，想怎么配就怎么配

### 🎉 波动日策略 — 大促不再告警刷屏

618、双11、春节……这些特殊时段，运维同学都是提心吊胆度过的。

我们懂你。所以设计了三层保护机制：

| 机制 | 阶段 | 作用 |
|------|------|------|
| 原料剔除 | 采样阶段 | 大促数据不污染日常基线 |
| 基线冻结 | 采样阶段 | 使用普通日基线，不写大促数据 |
| 阈值放大 | 告警检测 | 敏感任务自动放宽阈值 |

**大促不应该是成功率下降的借口。**

质量类指标（成功率、RT）不开启波动日策略——无论什么时候，真正的故障绝不放过。

### ⚡ 高性能时序引擎

- **可插拔后端**：VictoriaMetrics（默认）/ TDengine / Prometheus remote_write
- **毫秒级写入、秒级查询**
- **无缝对接 Grafana 可视化**

---

## 🛠 技术栈

我们选择了成熟、稳定、社区活跃的技术栈：

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.26+ / Gin / GORM / MySQL，DDD 分层架构 |
| **前端** | React 19 / Umi Max 4 / Ant Design 6 / Pro Components 3 / Tailwind CSS 4 |
| **时序** | VictoriaMetrics / TDengine / Prometheus remote_write（可插拔） |
| **可视化** | Grafana（API 自动管理） |
| **调度** | XXL-JOB 分布式任务调度 |
| **通知** | 企业微信 / 钉钉 / 飞书 / 邮件 |

---

## 📸 界面预览

| 首页概览 | 监控任务 |
|:---:|:---:|
| ![首页](docs/images/首页.png) | ![任务](docs/images/task.png) |

| 告警事件 | 数据源管理 |
|:---:|:---:|
| ![事件](docs/images/event.png) | ![数据源](docs/images/datasource.png) |

| Grafana 自动看板 |
|:---:|
| ![Grafana](docs/images/grafana.png) |

---

## 🚀 快速开始

**只需要三步，你就能拥有自己的业务监控平台。**

### 1. 准备数据库

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS butterfly_monitor DEFAULT CHARSET utf8mb4;"
mysql -uroot -p butterfly_monitor < backend/migrations/butterfly_monitor.sql
```

### 2. 启动后端

```bash
cd backend
# 按需修改 configs/config.yml
go mod tidy
go run ./cmd
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

访问 http://localhost:8000，默认账号 `admin` / `123456`

**就是这么简单。** 不需要复杂的安装脚本，不需要漫长的配置过程。几分钟，你就能跑起来一个企业级的业务监控平台。

---

## 🐳 Docker Compose 一键部署（推荐）

**一行命令，全部搞定。** 无需手动安装 Go、Node.js、MySQL、VictoriaMetrics、Grafana、XXL-JOB——所有依赖组件自动编排启动。

### 快速启动

```bash
# 克隆项目
git clone https://github.com/pwh19920920/butterfly-monitor.git
cd butterfly-monitor

# 一键启动所有服务
docker compose up -d
```

首次启动会自动完成以下初始化工作（约 2-3 分钟）：

| 步骤 | 说明 |
|------|------|
| MySQL 初始化 | 自动创建 `butterfly_monitor` 和 `xxl_job` 数据库，执行建表迁移 |
| XXL-JOB 初始化 | 自动创建执行器 + 四个定时任务（dataCollect / dataSampling / alertCheck / eventCheck） |
| Grafana 配置 | 自动创建 VictoriaMetrics 数据源 + Service Account + API Token |
| 服务启动 | 后端 + 前端依次启动 |

### 访问地址

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| 前端控制台 | http://localhost:8000 | `admin` / `123456` |
| Grafana | http://localhost:3000 | `admin` / `admin123` |
| XXL-JOB 管理台 | http://localhost:8080/xxl-job-admin | `admin` / `123456` |
| VictoriaMetrics | http://localhost:8428 | — |

### 服务架构

```
docker compose up -d
        │
        ├── mysql:3306            # 业务数据存储
        ├── victoria-metrics:8428  # 时序数据存储
        ├── grafana:3000           # 可视化看板
        ├── xxl-job-admin:8080     # 分布式任务调度
        ├── backend:8088           # Butterfly Monitor 后端 API
        └── frontend:8000          # Butterfly Monitor 前端（Nginx）
```

### 常用命令

```bash
# 查看服务状态
docker compose ps

# 查看后端日志
docker compose logs -f backend

# 重启所有服务
docker compose restart

# 停止所有服务
docker compose down

# 停止并清除所有数据（重新开始）
docker compose down -v
```

### 持久化存储

所有数据通过 Docker 命名卷持久化，重启不会丢失：

| 卷名 | 用途 |
|------|------|
| `butterfly-mysql-data` | MySQL 业务数据 |
| `butterfly-vm-data` | VictoriaMetrics 时序数据 |
| `butterfly-grafana-data` | Grafana 配置与面板 |
| `butterfly-backend-logs` | 后端日志 |

### 自定义配置

如需修改配置，编辑 `backend/configs/config-docker.yml` 后重启：

```bash
docker compose restart backend
```

> **提示**：Docker Compose 部署适合快速体验和小规模生产使用。如需集群部署或自定义外部依赖，请参考下方手动部署章节。

---

## 🔧 手动部署

部署前需自行准备以下依赖组件：

| 组件 | 版本要求 | 说明 |
|------|----------|------|
| Go | ≥ 1.26 | 后端编译运行 |
| Node.js | ≥ 18 | 前端构建 |
| MySQL | ≥ 5.7 | 业务数据存储 |
| VictoriaMetrics / TDengine / 远端 remote_write 接收端 | 三选一 | 时序存储 |
| Grafana | ≥ 9.0 | 可视化看板，需添加时序数据源并创建 API Token |
| XXL-JOB | ≥ 2.3 | 分布式任务调度，需创建执行器及四个定时任务（dataCollect / dataSampling / alertCheck / eventCheck） |

### ⏱ XXL-JOB 任务配置

在 XXL-JOB 管理台中创建以下四个定时任务。前三个任务配置「分片广播」以利用多节点并行加速；事件通知任务不需要分片，选择「轮询」或「第一个」即可：

| 任务名称 | JobHandler | 功能说明 | 建议 Cron | 调度策略 |
|----------|------------|----------|-----------|----------|
| 数据采集 | `dataCollect` | 从各数据源执行查询，采集监控指标值，写入实时序列点与未来 N 天样本原料点。按任务 ID 分片，多节点并行执行 | `0/1 * * * * ?` | 分片广播 |
| 样本平滑 | `dataSampling` | 对原料数据进行 MAD 稳健聚合 + 中位数平滑，生成基线指标 `_sample`。大促敏感任务自动剔除特殊日原料，回退冻结基线 | `0 */5 * * * * ?` | 分片广播 |
| 告警检查 | `alertCheck` | 执行告警规则判定，对比实时值与样本基线，检测相对偏离。管理告警生命周期状态机（Normal → Pending → Firing） | `0/1 * * * * ?` | 分片广播 |
| 事件通知 | `eventCheck` | 将待发送的告警事件按用户组和通知通道分桶聚合，通过企业微信 / 钉钉 / 飞书 / 邮件外发。支持翻倍退避策略，避免告警风暴 | `0/30 * * * * ?` | 轮询 / 第一个 |

> **注意**：四个任务均需绑定同一个执行器（AppName），执行器在 `backend/configs/config.yml` 的 `xxlJob` 配置块中注册。Cron 表达式可根据实际业务需求调整，以上为推荐值。

---

## 📖 典型场景

### 场景一：电商订单量监控

**场景**：电商订单量有明显的昼夜规律，白天高峰、凌晨低谷。固定阈值凌晨必误报，大促必漏报。

**过去**：运维同学每天凌晨被误报吵醒，阈值改来改去，大促期间还是漏报了，背锅。

**现在**：
- 接入 Database 任务，按分钟统计下单量
- 规则选择"样本百分比 > 50%"——比基线高出 50% 才告警
- 开启波动日策略，618 期间阈值自动放大 3 倍

**效果**：凌晨终于可以睡个安稳觉了，大促监控也能精准到位。

### 场景二：支付成功率监控

**场景**：支付成功率平时 99.5%+，跌破 97% 意味着下游出问题。这是核心业务指标，容不得半点马虎。

**过去**：固定阈值"成功率 < 97%"，但日常波动也会触发，次数多了就开始麻木，真的跌了反而反应慢了。

**现在**：
- 规则选择"样本百分比 < -2.5%"——比基线低 2.5% 即异常
- **不开启波动日**：成功率是质量指标，大促降了就是真故障

**效果**：真正的异常秒发现，绝不因大促放过任何一个问题。支付团队，终于放心了。

---

## 📁 目录结构

清晰的项目结构，让参与贡献变得更简单：

```
butterfly-monitor/
├── backend/          # 后端 API + Job（DDD 分层）
│   ├── cmd/          # 入口
│   ├── configs/      # 配置文件
│   ├── migrations/   # 数据库迁移
│   └── internal/     # 业务代码
│       ├── application/    # 应用服务层
│       ├── domain/         # 领域层
│       ├── infrastructure/ # 基础设施层
│       └── interfaces/     # 接口层
├── frontend/         # 管理端（Ant Design Pro）
└── docs/             # 设计文档
```

---

## 🤝 开源协议与贡献

本项目基于 **MIT 协议**开源。

**开源的意义，在于让更多人受益。**

如果你也曾经因为凌晨三点的误报而痛苦，如果你也曾经因为大促监控而焦虑，如果你也认同"业务监控不该是奢侈品"——欢迎 Star、Fork、PR！

你的每一个贡献，都在帮助更多的运维兄弟睡个好觉。

---

## 🔗 相关链接

- **GitHub**：https://github.com/pwh19920920/butterfly-monitor
- **基础框架**：https://github.com/pwh19920920/butterfly-admin
- **MIT 协议**开源
- **详细文档**：[docs/功能文档.md](docs/功能文档.md) ，[docs/数据源添加说明.md](docs/数据源添加说明.md) — 完整部署说明、配置参考、数据源接入指南

---

## 💬 交流反馈

有任何问题、建议，或者只是想聊聊监控那些事儿——欢迎提交 Issue 或 PR！

我们会认真对待每一条反馈。因为我们知道，每一个问题背后，都是一个真实的使用场景，都是一份对更好工具的期待。

---

> **Butterfly Monitor** — 让监控从"看数"进化到"决策"，从"被动告警"进化到"智能感知"。
>
> **愿每一个运维兄弟，都能睡个好觉。** 😴✨