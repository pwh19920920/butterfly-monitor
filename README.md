# Dragonfly Monitor

**业务趋势分析与异常预警平台**。基于 [butterfly-admin](https://github.com/pwh19920920/butterfly-admin) 二开，面向业务指标做趋势分析与异常预警：持续采集业务数据，用历史同时刻数据生成自学习基线，比对当前趋势的偏离程度，在真正发生异常时按需告警，并通过多通道触达责任人。

核心思路：**不是设一个死阈值、超了就报**，而是基于历史基线做相对偏离检测——业务本身有周期性波动（昼夜峰谷、周末效应），固定阈值要么误报频繁、要么漏报。平台通过样本平滑算法把周期性"消化"进基线，只在偏离正常趋势时才触发，从而显著降低误报、让告警真正可信。

## 解决什么问题

| 问题 | 传统方式 | 本平台 |
|------|----------|--------|
| 周期性波动导致误报 | 固定阈值，凌晨低谷必误报 | 历史同时刻基线，自动适应昼夜/周末节奏 |
| 阈值难调 | 反复改数，改了又误报 | 基线自学习，运行几天自动建立"这个点该是多少" |
| 大促/holiday 误报刷屏 | 大促期间全站关告警 | 波动日策略：敏感任务自动放宽阈值，非敏感任务不受影响 |
| 告警风暴 | 一次故障反复报 | 完整生命周期：Pending → 持续确认 → 外发 → 人工闭环 |
| Grafana 面板维护 | 手动建面板、手工加 panel | 创建任务即自动同步 Grafana dashboard/panel |

## 核心链路

```
指标采集(dataCollect) → 样本生成(dataSampling) → 规则检测(alertCheck) → 事件通知(eventCheck)
```

四个任务由 **XXL-JOB** 调度（cron 在 admin 端配置）：

| 任务 | 职责 |
|------|------|
| `dataCollect` | 按周期从数据源采集指标写入时序库，同时写未来 8 天同时刻的样本原料点（day tag 1~8） |
| `dataSampling` | 按窗口对历史原料做稳健聚合（MAD 剔除离群 + 中位数），写入平滑基线指标 `{taskKey}_sample` |
| `alertCheck` | 比对实时均值与基线均值，命中规则且持续达 Duration 则生成告警事件；敏感任务命中波动日时阈值自动放大 |
| `eventCheck` | 到期事件经模板渲染，通过邮件/企业微信外发给告警分组 |

### 告警生命周期

规则与事件是两套状态，不要混读：

```
规则 (t_monitor_task_alert)
  Normal → Pending（首次命中，未达 Duration）→ Firing（持续超 Duration，建事件）
                ↑                                      │
                └──── 恢复正常 / 事件闭环后 ─────────────┘

事件 (t_monitor_task_event，Firing 时创建)
  Pending（待外发，等 NextAlertTime）
      ├─→ Processing（人工处理；规则 DealStatus=Processing，暂停检测）
      │       └─→ Complete（处理完成，规则恢复可检）
      └─→ Ignore（误报标定 / 自动恢复时的 pending 收口，不外发）
```

- **规则 Pending**：第一次命中，进入等待。未达 Duration 不升级，避免瞬时抖动误报。
- **规则 Firing**：持续命中超 Duration，创建事件，等待 `NextAlertTime` 到期后由 `eventCheck` 外发。
- **事件 Processing**：人工介入；对应规则 `DealStatus=Processing`，暂停对该规则的检测。
- **事件 Complete**：处理完成，规则恢复可检。
- **事件 Ignore**：标定为误报，或规则恢复 Normal 时对残留 Pending 事件的收口；**不**向外部平台推送 resolved。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.26 + Gin（[butterfly](https://github.com/pwh19920920/butterfly)）+ GORM + MySQL，DDD 分层 |
| 前端 | React 19 + Umi Max 4 + Ant Design 6 / Pro Components 3 + Tailwind CSS 4 |
| 时序 | 可插拔：VictoriaMetrics（默认）/ TDengine |
| 可视化 | Grafana（自动创建/同步 dashboard 与 panel） |
| 调度 | XXL-JOB |
| 通知 | 邮件 / 企业微信 |
| 鉴权 | JWT + 菜单权限 |
| 被监控数据源 | MongoDB；PostgreSQL；ClickHouse；Prometheus / VictoriaMetrics / TDengine / OpenSearch / Elasticsearch（只读）；MySQL 协议族（Mysql / MariaDB / TiDB / OceanBase MySQL 模式 / Doris / StarRocks） |

## 目录结构

```
dragonfly-monitor/
├── backend/     # 后端 API + Job（DDD 分层）
│   ├── cmd/                     # 入口
│   ├── configs/                 # 配置文件（config.yml）
│   ├── migrations/              # 数据库迁移 SQL
│   ├── internal/
│   │   ├── application/         # 应用服务层（业务编排）
│   │   ├── common/              # 公共工具（BaseEntity、LocalTime 等）
│   │   ├── config/              # 配置加载
│   │   ├── domain/
│   │   │   ├── entity/          # 实体定义
│   │   │   ├── handler/         # 领域接口（时序库、指标查询等）
│   │   │   └── repository/      # 仓储接口
│   │   ├── infrastructure/
│   │   │   ├── handler/         # 领域接口实现（平台时序 VM/TDengine、被监控 ClickHouse/Prometheus/VM/TDengine/OpenSearch/ES 只读、MySQL 协议族、MongoDB、PostgreSQL 等）
│   │   │   └── persistence/     # 仓储实现（GORM）
│   │   ├── interfaces/          # HTTP handler + 路由注册
│   │   ├── job/                 # XXL-JOB 定时任务
│   │   ├── starter/             # 启动装配
│   │   └── types/               # 请求/响应 DTO
│   └── logs/                    # 日志输出
├── frontend/    # 管理端（Ant Design Pro / Umi Max）
└── docs/        # 设计文档
```

后端分层依赖方向严格单向无环：

```
cmd → starter → {interfaces, application, infrastructure, job, config}
                     ↓
                domain ← {types, common}
```

## 快速开始

### 1. 数据库

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS dragonfly_monitor DEFAULT CHARSET utf8mb4;"
mysql -uroot -p dragonfly_monitor < backend/migrations/dragonfly_monitor.sql
```

默认账号：`admin` / `123456`

### 2. 后端

> 配置文件 `configs/config.yml` 与日志目录 `logs/` 均按**进程工作目录**读取，因此必须**在 `backend/` 目录**执行。

```bash
cd backend
# 按需修改 configs/config.yml：db / timeseries / victoriaMetrics 或 tdEngine / grafana / xxl
go mod tidy
go run ./cmd        # 或 make run
```

默认端口 `:8088`。指定配置文件：`go run ./cmd --configFilePath=path/to/config.yml`。

### 3. 前端

```bash
cd frontend
npm install
npm run dev         # 或 npm start
```

浏览器访问 http://localhost:8000 ，开发环境将 `/api/` 代理到 `http://localhost:8088`（`config/proxy.ts`）。

## 功能模块

### 数据接入

支持三类任务数据源，接入新指标只需写好查询语句、定个周期，不用改代码、不用重新发版：

| 任务类型 | 说明 |
|------|------|
| Database | 对关系型 / 文档型库执行 SQL 或聚合查询，取回标量值 |
| URL (HTTP) | 调 HTTP 接口，按 JSONPath 提取字段值 |
| Push（外部推送） | 业务方主动推数据到平台 |

**Database 支持的库类型**（`DataSourceType`）：

| 值 | 类型 | 实现 | 备注 |
|----|------|------|------|
| 1 | Mongo | 独立 handler | Extended JSON 聚合管道 |
| 2 | Mysql | 独立 handler | SQL |
| 3 | PostgreSQL | 独立 handler | SQL，`postgres://` DSN |
| 4 | MariaDB | 复用 Mysql | MySQL 协议兼容 |
| 5 | TiDB | 复用 Mysql | MySQL 协议兼容 |
| 6 | OceanBase | 复用 Mysql | **仅 MySQL 模式** |
| 7 | Doris | 复用 Mysql | 连 FE MySQL 端口 |
| 8 | StarRocks | 复用 Mysql | 连 FE MySQL 端口 |
| 9 | ClickHouse | 独立 handler | SQL，原生协议默认 9000 |
| 10 | Prometheus | 独立 handler | **只读** PromQL，默认 9090 |
| 11 | OpenSearch | 独立 handler | **只读** `_search`/`_count`/SQL，默认 9200 |
| 12 | Elasticsearch | 复用 OpenSearch | **只读**，查询 API 兼容，共用实现 |
| 13 | VictoriaMetrics | 复用 Prometheus | **只读** MetricsQL/PromQL，默认 8428 |
| 14 | TDengine | 独立 handler | **只读** REST SQL，默认 6041 |

- 数据源密码 DES-CBC + 随机盐加密存储，不以明文留存
- 新建数据源时先验连通性再落库
- 连接池每分钟增量扫描，最多 1 分钟后对新任务可见
- 各类型连接字段、`Command` 示例与注意项见 **[docs/数据源添加说明.md](docs/数据源添加说明.md)**

### 趋势基线自学习

每个采集周期除了写实时值，还会向未来 8 天的同时刻写入"样本原料点"（day tag 1~8）；采样任务按窗口对这些原料点做**稳健聚合**，生成平滑基线 `{taskKey}_sample`：

```
原料点集合 (1~8天同时刻) → 剔除 NaN/Inf
                           → 保留 0/负值（由中位数 + MAD 处理）
                           → 1点：原样返回
                           → 2~4点：中位数
                           → ≥5点：MAD (k=3) 剔除离群 → 中位数
                           → 写入 {taskKey}_sample
```

关键设计：
- **少样本也写**：有原料点就写基线，保证 Grafana 样本线连续，不因点数不足拒绝出数
- **MAD 替代掐头去尾**：对单日 0 或尖刺更稳健，两日同时异常大概率双剔
- **绝不混入实时值**：`_sampling` 缺失或查询失败时跳过该格，不拿当前实时值当历史基线

### 异常检测规则

比对实时均值与基线均值的**相对偏离**，而非比对一个写死的数。规则结构：

```
规则组 1 (OR)                      规则组 2 (OR)
├── 规则 A（样本百分比 > 50%）      ├── 规则 C（实时值 > 10000）
└── 规则 B（样本差值 > 5000）       └── 规则 D（实时值 < 100）
     ↓                                  ↓
  任一组命中 → 整体命中（组间固定 OR）
```

| 配置项 | 说明 |
|--------|------|
| 比较值类型 | **样本百分比** `(实时-基线)*100/基线`、**样本差值** `实时-基线`、**实时值**（不依赖基线） |
| 比较方式 | `>` `>=` `<` `<=` `==`，既能抓突增也能抓突降 |
| 组内关系 | And（全部命中）/ Or（任一命中） |
| 组间关系 | 固定 Or（首组命中即打断，不再匹配后续组） |
| 生效时段 | `HH:mm:ss ~ HH:mm:ss`，不在时段内整组跳过 |
| 事件等级 | Low / Medium / High / Critical |
| 持续时长 | Duration（秒），命中持续达此时长才升级 Firing |

### 波动日策略

**适用场景**：大促（618/双11）、节假日（春节/国庆）等业务量级会大幅偏离日常的特殊时段。策略解决两个问题——日常基线不被大促历史污染，大促期间敏感指标不刷屏。

#### 两层开关

```
波动日日历（全局，回答 when）
    ×
promoSensitive（任务级，回答 whether）
    = 策略生效
```

- **波动日日历**：在管理端「波动日」页维护（表 `t_monitor_volatility_day`，API `/api/monitor/volatilityDay`），类型分高峰 `peak=1`（如 618）和低谷 `trough=2`（如春节），按时间区间配置；**不是**告警配置 KV
- **promoSensitive**：每个任务独立标记，默认关闭（=1 不敏感），仅明确需要走策略的任务开启（=2 敏感）

#### 三大机制

| 机制 | 阶段 | 做什么 |
|------|------|--------|
| **原料剔除** | dataSampling | 聚合历史原料时，按 day tag 还原来源日，落在波动日的原料点不参与基线计算 |
| **基线冻结** | dataSampling | 当前时间格落在波动日时，不写大促原料聚合值，改为写入「最近一个普通日同时刻」的历史基线值 |
| **阈值放大** | alertCheck | 敏感任务命中波动日时，对本轮判定副本按配置倍数放大阈值（不改库、不改用户配置） |

#### 放大方向

| 波动日类型 | 放大方向 | 配置项 |
|-----------|----------|--------|
| **高峰 (peak)** | 上偏规则（`>` `>=`）放大 | `promoPeakRatio`，默认 1（不放大），建议 2~5 |
| **低谷 (trough)** | 样本差类下偏规则（`<` `<=`）放大 | `promoTroughRatio`，默认 1（不放大） |
| 低谷 - 实时值下偏 | **不动** | 绝对值门槛表达不了"相对低谷的跌幅"，放大反而掩盖真故障 |

#### 适用/不适用指标

| ✅ 应该开启（量级/计数类） | ❌ 严禁开启（质量/比率类） |
|---------------------------|---------------------------|
| 入口 QPS、下单量、支付笔数、UV/PV | 成功率、错误率 |
| 明确会跟大促量级走的计数指标 | 核心 RT、P99 延迟 |
| 活跃用户数、消息推送量 | CPU/内存使用率、队列积压 |

> **大促不应该是成功率下降的借口。** 成功率/错误率/RT 这类质量指标无论什么时候都该稳定；如果大促期间降了，恰恰是真故障，阈值放宽反而会把问题掩盖。

#### 配置概览

| 配置项 | 存储位置 | 默认值 | 说明 |
|--------|----------|--------|------|
| 波动日日历 | 「波动日」页 / `t_monitor_volatility_day` | 空（无波动日） | 每项含 name / startTime / endTime / type(1高峰 2低谷) |
| `promoPeakRatio` | 告警配置 | `1` | 高峰上偏阈值放大倍数 |
| `promoTroughRatio` | 告警配置 | `1` | 低谷样本差下偏放大倍数 |
| `freezeSampleLookBackDays` | 告警配置 | `3` | 冻结基线向前回溯普通日的最大天数（从「前一日」起算；`<=0` 时采样兜底按 14） |
| `promoSensitive` | 任务字段 | `1`（否） | 任务级开关：1=不敏感 2=敏感 |

### Grafana 看板自动同步

- 创建任务时自动在关联面板上加 timeseries panel（以 `panel.Description == TaskKey` 为锚点）
- 修改任务时同步增删、重排 panel
- 开启"样本展示"后自动叠加 `{taskKey}_sample` 基线对比线
- 运营侧无需手动维护 Grafana，任务建好即有图

### 告警通知

- **告警通道**：邮件 / 企业微信 Webhook，保存时测试发送
- **告警分组**：组织接收人分组与成员，决定谁收
- **通知模板**：全局可配，支持变量渲染
- **外发退避**：首次外发延迟 `firstDelay`（默认 60s），后续按告警次数翻倍退避（`alertSpan` 基数 300s，上限 1800s），避免告警风暴

### 监控分组

监控对象按树形分组组织，关联预警链路（"A → B → C"），便于按业务域聚合查看依赖关系。

### 聚合查询与系统下钻

- **聚合任务**（`dataType=2`）：分组多行结果写入时序库，仅收集不做采样/告警，用于维度拆分场景
- **系统下钻**（`taskType=4`）：从聚合任务结果中按标签过滤取数，实现从大盘到单维度的逐级下钻

### 首页概览

聚合展示任务/事件/面板/数据源数量，一眼掌握平台整体运行规模。

## 典型场景

### 场景一：订单量异常突增（样本百分比 + 波动日策略）

电商订单量有明显的昼夜规律——白天午高峰高、凌晨低谷。若用固定阈值"订单数 > 10000"，凌晨必误报，大促又可能漏报。

- **接入**：Database 任务，`TaskKey = order_count_per_min`，SQL 按分钟统计下单量，`TimeSpan = 60`
- **建基线**：dataSampling 按窗口 MAD + 中位数聚合原料，生成基线 `order_count_per_min_sample`。运行几天后每个时刻都有了"平时该是多少"的参照
- **配规则**：值类型选**样本百分比**、值 `50`、比较**超出**——"比基线高出 50% 才告警"
- **配波动日**：标记 `promoSensitive = 是`；在「波动日」页录入 618 区间（高峰），告警配置里设 `promoPeakRatio = 3`。大促期间订单量涨 3 倍不告警，但涨 5 倍仍报
- **闭环**：规则先 Pending、持续超 Duration 才 Firing 建事件 → 到期外发企业微信 → 值班处理（Processing）→ 完成/忽略

### 场景二：接口成功率下降（样本百分比 + 突降检测，不开启波动日）

支付成功率平时 99.5%+，跌破 97% 通常意味着下游出问题。

- **接入**：URL 任务，`resultFieldPath` 取 `success_rate`
- **配规则**：值类型选**样本百分比**、值 `-2.5`、比较**低于**——"比基线低 2.5% 即异常"。生效时段设为业务高峰时段
- **波动日**：**不开启**。成功率是质量指标，大促期间降了就是真故障
- **外发**：命中后到期邮件通知支付组，事件等级 Critical

### 场景三：数据库慢查询堆积（直接值，不依赖基线）

慢查询数"平时接近 0"，基线本身极小，相对偏离会失真。用**直接值**比较。

- **接入**：Database 任务对慢查询日志表按分钟 `count(*)`，`TaskKey = slow_query_count`
- **配规则**：值类型选**实时值**、值 `20`、比较**超出**——"实时慢查询数 > 20 即异常"
- **波动日**：不开启（稳定性指标）

三种值类型的适用场景总结：

| 值类型 | 计算方式 | 适用场景 |
|--------|----------|----------|
| 样本百分比 | `(实时-基线)*100/基线` | 有周期起伏的业务量（订单、访问量） |
| 样本差值 | `实时-基线` | 平稳指标偏离基线的具体差值 |
| 实时值 | 不依赖基线，直接比 | 本该趋零的故障计数（慢查询、错误数） |

## 如何在平台上配置一个任务

注意上手顺序：**先建数据源和面板，再建任务并关联面板**——任务保存时按关联面板自动加 Grafana panel，面板必须先行存在。

### 1. 准备数据源

进入「数据源」页，新建被监控的数据库实例：选类型（见上表 14 种），填 `url=host:port`、库名、账号密码与可选附加参数，保存。保存前系统会先测试连通性；密码 DES-CBC + 随机盐加密落库。

- Mysql 协议族（含 MariaDB/TiDB/OceanBase MySQL 模式/Doris/StarRocks）：附加参数同 MySQL DSN 查询串
- PostgreSQL：附加参数如 `sslmode=disable`
- ClickHouse：端口默认 **9000**（原生协议），附加参数如 `dial_timeout=10s&readonly=1`
- Prometheus：**只读**；端口默认 **9090**；库名可空；`Command` 写 PromQL；附加参数如 `scheme=https path_prefix=/prometheus timeout=30s`
- VictoriaMetrics：**只读**，复用 Prometheus handler；端口默认 **8428**；`Command` 写 PromQL/MetricsQL
- TDengine：**只读** REST SQL；端口默认 **6041**（taosAdapter）；库名=业务库；`Command` 写 `SELECT`/`SHOW`
- OpenSearch / Elasticsearch：**只读**，共用实现；端口默认 **9200**；库名=默认 index；`params.api=search|count|sql`；`Command` 为 `_search` JSON / 空 / SQL
- Mongo：附加参数如 `connectTimeoutMS=5000`

如果是 URL 或 Push 类型的任务，可跳过本步。

### 2. 创建监控面板

进入「监控面板」页，新建 Grafana 大盘：填面板名称保存，系统自动调 Grafana API 创建 dashboard 并回填 Uid/Url。后续任务关联此面板即可自动加 panel。

### 3. 新建监控任务

进入「监控任务」页，点新建，关键字段：

| 字段 | 说明 |
|------|------|
| TaskKey | 全局唯一，同时作为时序库指标名 |
| TaskType | Database / URL / Push / 系统下钻 |
| Command | SQL/URL 模板，支持 `{{.beginTime}}` `{{.endTime}}` `{{.startTime}}` 等占位符 |
| TimeSpan | 采集周期（秒），30s 的倍数 |
| StepSpan | 查询区间宽度（秒） |
| dataType | 正常查询(单值) / 聚合查询(分组多行) |
| Dashboards | 关联的 Grafana 面板（多选） |
| MonitorGroup | 监控分组（树形依赖） |
| 大促敏感 | 量级指标开启，质量指标关闭 |
| TaskStatus / AlertStatus / Sampled | 任务/告警/样本展示开关 |

### 4. 配置告警规则

在任务的创建/编辑表单中一并配置：

- **报警检查配置**：报警通道、报警分组、检查间隔(秒)、持续时间(秒)
- **异常检测规则**：规则条件组（组内 And/Or、组间固定 Or），每组合生效时段 + 告警等级 + 多条规则（比较类型 + 比较值 + 值类型）

### 5. 配置通知通道与接收人

- 「告警通道」：新建邮件/企业微信通道，保存时测试发送
- 「告警分组」：组织接收人分组
- 「告警配置」：全局 KV 参数（`firstDelay`、`alertSpan`、模板、采样/采集并发与超时、波动日倍率等）
- 「波动日」：录入大促/节假日时间区间（高峰/低谷），与告警配置分离

### 6. 跑起来之后

四个定时任务自动接力，无需手动干预：

1. **dataCollect**：每过一个 TimeSpan 采一次，写实时值 + 未来 8 天样本原料点
2. **dataSampling**：按 TimeSpan 逐格补点，原料 MAD + 中位数 → 写入 `{taskKey}_sample`
3. **alertCheck**：比对实时均值与基线，命中 → Pending → 持续超 Duration → Firing 建事件
4. **eventCheck**：到期事件渲染模板，经通道外发

Grafana 面板上实时曲线 + 样本基线对比线自动叠加，一眼看出当前值 vs 历史趋势。

## 配置说明

主配置：`backend/configs/config.yml`，按 `engineMode` 命名的环境配置（如 `config-release.yml`）会被 viper 自动 merge。生产密钥（DB 密码、Grafana Token 等）勿提交仓库。

关键配置段：

| 配置段 | 说明 |
|--------|------|
| `db` | MySQL 连接（GORM） |
| `timeseries` | 时序后端选择：`victoriaMetrics`（默认）/ `tdengine`（配置值以代码常量为准，推荐小写 `tdengine`） |
| `victoriaMetrics` | VM 连接（`backend=victoriaMetrics` 时使用） |
| `tdEngine` | TDengine 连接（YAML 段名 `tdEngine`；`timeseries.backend=tdengine` 时使用） |
| `grafana` | Grafana API 地址 + Token，用于自动管理 dashboard/panel |
| `xxl` | XXL-JOB admin 地址 + 执行器配置 |

告警配置（全局 KV，在管理端「告警配置」页维护；代码默认值见 `backend/internal/types/alert_conf_type.go`）：

| Key | 默认值 | 说明 |
|-----|--------|------|
| `firstDelay` | `60` | 首次外发延迟（秒） |
| `alertSpan` | `300` | 外发间隔基数（秒），按次数翻倍退避，上限 1800s（`maxAlertShift` 控制最大位移） |
| `defaultTemplate` | - | 全局默认通知模板（通道未配模板时兜底） |
| `defaultChannelEmailHandlerTemplate` | - | 邮件通道默认模板 |
| `defaultChannelWechatHandlerTemplate` | - | 企业微信通道默认模板 |
| `simplePageSize` | `50` | 采样任务单批捞取任务数 |
| `simpleMaxSecond` | `600` | 采样单轮最大补点秒数（安全阀） |
| `collectMaxSecond` | `25` | 单任务采集超时（秒） |
| `alertCheckConcurrency` | `100` | 告警检查并发数 |
| `samplingConcurrency` | `80` | 采样并发数 |
| `sampleRawDays` | `8` | 采集时向未来投射的样本原料天数 |
| `batchWriteChunkSize` | `3000` | 时序写入单批最大点数 |
| `maxAlertShift` | `5` | 告警间隔翻倍最大位移 `2^N` |
| `promoPeakRatio` | `1` | 高峰上偏阈值放大倍数 |
| `promoTroughRatio` | `1` | 低谷样本差下偏放大倍数 |
| `freezeSampleLookBackDays` | `3` | 冻结基线回溯普通日最大天数 |

波动日日历**不在**告警配置 KV 中，请到管理端「波动日」页维护（表 `t_monitor_volatility_day`）。

## 新增业务模块步骤

遵循 DDD 分层，严格单向依赖：

1. `internal/domain/entity/` — 实体定义
2. `internal/domain/repository/` — 仓储接口
3. `internal/infrastructure/persistence/` — 仓储实现
4. `internal/application/` — 应用服务
5. `internal/types/` — 请求/响应 DTO
6. `internal/interfaces/` — HTTP handler + starter 注册路由

命名约定：Handler `*_handler.go` / App `*_app.go` / Repo `*_repository.go`，实体表名 `t_*`，BaseEntity 软删除，中文注释。

## 相关文档

- [`docs/功能文档.md`](docs/功能文档.md) — 完整接口清单与业务规则
- [`docs/数据源添加说明.md`](docs/数据源添加说明.md) — 各数据源连接字段、Command 示例与注意项
- [`backend/CLAUDE.md`](backend/CLAUDE.md) — 后端工程约定（分层、命名、运行方式）
- [`backend/README.md`](backend/README.md) / [`frontend/README.md`](frontend/README.md) — 源自 butterfly-admin 模板说明（界面预览截图等），工程细节以本文与 `docs/` 为准
