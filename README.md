# Dragonfly Monitor

**业务趋势分析平台**。基于 [butterfly-admin](https://github.com/pwh19920920/butterfly-admin) / [butterfly-admin-web](https://github.com/pwh19920920/butterfly-admin-web) 二开，面向业务指标做趋势分析与异常预警：持续采集业务数据，用历史同时刻数据生成自学习基线，比对当前趋势的偏离程度，在真正发生异常时按需告警，并通过多通道触达责任人。

与传统"设个死阈值、超了就报"的告警不同，本平台的告警是**基于历史基线的相对偏离**——业务本身有周期性波动（昼夜峰谷、周末效应），固定阈值要么误报频繁、要么漏报。平台通过样本平滑算法把周期性"消化"进基线，只在偏离正常趋势时才触发，从而显著降低误报、让告警真正可信。

## 解决什么问题 / 达到什么效果

- **看清业务在怎么走**：把分散在各数据源（MySQL / MongoDB、HTTP 接口、外部推送）的业务数据，周期性采集到时序库并自动呈现在 Grafana 面板上，无需手写看板，一眼看到当前值与历史基线的对比趋势。
- **告警跟着业务节奏走**：基线由历史同时刻数据生成，自动适应昼夜 / 周末规律——凌晨低峰不再误报，白天高峰也能识别出"比平时高太多"的真异常。
- **偏离多大才报，可控**：规则支持按百分比、绝对值或样本差值衡量偏离，规则组内 And、组间 Or/And 组合，可设生效时段与事件等级，既能抓突增也能抓突降。
- **告警不刷屏**：事件有完整生命周期，命中后先 Pending、到期才外发；人工介入期间暂停检测；误报可标定忽略；恢复后自动发 resolved 通知，避免告警风暴。
- **通知直达**：到期告警经模板渲染，通过邮件 / 企业微信 / 钉钉外发给指定接收人分组，谁该收到、收什么内容都可控。

## 核心链路

```
指标采集(dataCollect) → 样本生成(dataSampling) → 规则检测(alertCheck) → 事件通知(eventCheck)
```

四个任务由 **XXL-JOB** 调度（cron 在 admin 端配置）：

| 任务 | 职责 |
|---|---|
| `dataCollect` | 按周期从数据源采集指标写入 VictoriaMetrics，同时写未来 8 天同时刻的样本原料点 |
| `dataSampling` | 按窗口生成平滑样本，写入基线指标 `{taskKey}_sample`；生成窗口已扩展到 `now+1天`，可提前产出未来一天的格子；`_sampling` 缺失或查询失败时跳过该格，绝不混入实时值 |
| `alertCheck` | 比对实时均值与平滑基线均值，命中且持续达阈值则生成告警事件 |
| `eventCheck` | 到期事件经模板渲染，通过邮件 / 企业微信 / 钉钉外发 |

### 告警生命周期

```
规则命中(异常持续) → 生成事件(Pending) → 到期外发(firing) → 人工处理 → 完成(恢复通知 resolved)
                         ↑ 失败/误报可标 Ignore
```

状态：`Pending`（待外发）→ `Processing`（处理中，暂停检测）→ `Complete`（完成，若曾告警发恢复通知）/ `Ignore`（误报，不外发）。

## 技术栈

- **后端**：Go 1.26 + Gin（[butterfly](https://github.com/pwh19920920/butterfly)）+ GORM + MySQL，DDD 分层，JWT + 菜单权限
- **前端**：React 19 + Umi Max 4 + Ant Design 6 / Pro Components 3 + Tailwind CSS 4
- **时序**：VictoriaMetrics
- **可视化**：Grafana（自动创建 / 同步 dashboard 与 panel）
- **调度**：XXL-JOB
- **通知**：邮件 / 企业微信 / 钉钉

## 目录结构

```
dragonfly-monitor/
├── backend/     # 后端 API + Job（DDD 分层，详见 backend/README.md）
├── frontend/    # 管理端（Ant Design Pro / Umi Max，详见 frontend/README.md）
└── docs/        # 功能文档（完整接口与业务规则）
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
# 按需修改 configs/config.yml：db / victoriaMetrics / grafana / xxl / business
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

## 功能介绍

- **业务数据接入**：支持三类来源——直接对 MySQL / MongoDB 执行 SQL、调用 HTTP 接口拉取、或由业务方主动推送。数据源密码加密存储，新建时先验连通性再落库。这意味着接入一个新指标只需要写好查询语句、定个周期，不用改代码、不用重新发版。

- **趋势基线自学习**：每个采集周期除了写实时值，还会向未来 8 天的同时刻写入"样本原料点"；采样任务按窗口对这些原料点做稳健聚合（默认点数 ≥5 时排序去最大最小后取平均，<5 时直接算术平均）生成平滑基线指标 `{taskKey}_sample`。采样窗口已扩展到当前时间之后一天，因此未来一天的同时刻格子也会被提前生成。效果是：系统运行几天后自动建立起"这个指标这个点本来该是多少"的参照系，无需人工设阈值。

- **相对偏离检测**：检测时比对实时均值与基线均值，而非比对一个写死的数。偏离多少算异常由规则定义——可按**百分比**（涨了 30%）、**绝对值**（多了 5000）、或**样本差值**衡量；比较方式 5 种（`>` `>=` `<` `<=` `==`），既能抓突增也能抓突降；多条规则组内 And、组间 Or/And 组合，可设生效时段与事件等级。

- **可视化看板**：对接 Grafana，创建任务时自动在大盘上加 panel，修改时同步增删与重排，开启样本展示后自动叠加基线对比线。运营侧无需手动维护 Grafana，任务建好即有图，且图上直接看当前值 vs 历史基线。

- **告警闭环管理**：事件经历 Pending（待外发）→ Processing（处理中，暂停检测）→ Complete（完成，发恢复通知）/ Ignore（误报标定）四态。命中不等于外发——先 Pending 等到期，避免瞬时抖动告警；人工介入期间不再重复检测；恢复后自动发 resolved。这从机制上压住了告警风暴。

- **多通道触达**：到期告警经模板渲染，通过邮件 / 企业微信 / 钉钉外发。接收人按"告警分组"组织，谁收什么、内容怎么拼都可控；通道支持测试发送，配置即用。

- **依赖分组与首页概览**：监控对象按树形分组组织，关联预警链路，便于按业务域聚合查看；首页聚合任务 / 事件 / 面板 / 数据源数量，一眼掌握平台整体运行规模。

## 典型场景

以下三个场景串联起"采集 → 基线 → 偏离检测 → 闭环告警 → 恢复通知"完整链路，帮助理解平台相对固定阈值告警的实际收益。

### 场景一：订单量异常突增（百分比偏离）

电商订单量有明显的昼夜规律——白天午高峰高、凌晨低谷。若用固定阈值"订单数 > 10000 告警"，凌晨或工作日低谷必然误报，大促又可能漏报。

- **接入**：建一条 Database 任务，`TaskKey = order_count_per_min`，SQL 按分钟统计下单量，`TimeSpan = 60`（每分钟采一次）。
- **建基线**：dataCollect 每分钟写实时值，同时往未来 8 天同时刻写样本原料点；dataSampling 按窗口聚合生成基线 `order_count_per_min_sample`（生成窗口覆盖到 `now+1天`，未来一天的格子也会提前产出）。运行几天后，每个时刻都有了"平时这个点该是多少"的参照。
- **配规则**：阈值类型选**样本百分比**（`(实时-基线)*100/基线`）、值 `50`、比较方式选**超出**，含义是"实时均值比基线均值高出 50% 才算异常"。
- **触发与外发**：alertCheck 命中后并不立刻告警——先进 Pending 等待 Duration（持续时长），持续达阈值才置 Firing 并建事件，事件 `NextAlertTime = now + FirstDelay`（默认 60 秒）后到期外发；到期 eventCheck 渲染模板，经企业微信机器人通知值班组。
- **闭环**：值班组在企微收到告警，平台上点"处理"暂停检测；排查确认流量异常后处置，恢复后点"完成"，平台自动发 resolved 通知。若判定为正常大促，标"忽略"即可，不计告警。

**效果**：周期性低谷不再误报，只在"比这个时刻平时高出一半"时才报；命中后还要持续够 Duration 才升级 Firing，瞬时抖动（如一次重试尖峰）不会触发。

### 场景二：接口成功率下降（样本百分比 + 突降检测）

支付成功率平时在 99.5% 以上，跌破 97% 通常意味着下游出问题。这类"本应平稳、突然掉下来"的指标适合用样本百分比 + 低于比较。

- **接入**：HTTP 任务定时探测支付结果接口，`resultFieldPath` 取 `success_rate`。
- **配规则**：阈值类型选**样本百分比**（`(实时-基线)*100/基线`）、值 `-2.5`、比较方式选**低于**，表示"当前值比基线低 2.5% 即异常"；生效时段设为业务高峰时段，避开深夜无人值守打扰。
- **触发**：成功率先小幅下滑，alertCheck 检测到偏离基线 2.5% 且持续过 Duration，置 Firing 建事件；到期邮件通知支付组，事件等级标 Critical。

**效果**：用基线做参照，无需为"99.5% 还是 97%"反复调参；只在工作时间内告警，夜间安静。

### 场景三：数据库慢查询堆积（直接值阈值）

并非所有指标都适合相对偏离——慢查询数"平时接近 0"时基线本身极小，相对偏离会失真，这时用**直接值**比较（不依赖基线，直接拿实时值比阈值）。

- **接入**：Database 任务对慢查询日志表按分钟 `count(*)`，`TaskKey = slow_query_count`。
- **配规则**：阈值类型选**直接值**、值 `20`、比较方式选**超出**，即"实时慢查询数 > 20 即异常"。
- **触发与收敛**：慢查询突增并持续达阈值后告警；若同一根因引发多条任务同时命中，借助监控分组的依赖关系与告警分组，相关告警收敛后只触达对应 DBA 组，避免告警风暴。处置完成发恢复通知。

**效果**：相对基线适合"有起伏"的业务量，直接值适合"本该趋零"的故障型计数；三种阈值类型按指标特性灵活混用，一套平台覆盖两类监控需求。

## 如何在平台上配置一个任务

以场景一（订单量按分钟统计）为例，从前端页面走一遍完整配置。注意上手顺序：**先建数据源和面板，再建任务并关联面板**——任务保存时会按关联的面板自动加 Grafana panel，所以面板必须先行存在。任务一旦建好，采集、基线、Grafana 面板、告警检测会自动运转，无需再碰代码或 Grafana。

### 1. 准备数据源

进入「数据源」页，新建一条被监控的数据库实例：

- 选类型（MySQL / MongoDB），填连接信息并保存。
- 保存前系统会先**测试连通性**；密码用 DES-CBC + 随机盐加密落库，不以明文留存。
- 连接池由后台每分钟增量扫描，**最多 1 分钟后**新数据源对采集任务可见。

> 如果指标来自接口或外部推送，可跳过本步——任务类型选 URL 或外部推送即可。

### 2. 创建监控面板

进入「监控面板」页，新建一个 Grafana 大盘：

- 填面板名称保存，系统调 Grafana `CreateDashboard` 自动创建大盘，并回填 `Url/Slug/Uid/BoardId`。
- 面板页查询时 `Url` 前会自动拼上 Grafana 地址，可直接跳转查看。
- 后续在任务里关联此面板；想按业务域归类，可配合「监控分组」组织依赖关系。

> 面板本身是空大盘，里面的 timeseries panel 由任务保存时自动增删——无需手动在 Grafana 里画图。

### 3. 新建监控任务

进入「监控任务」页，点新建，填关键字段：

| 字段 | 填法（以订单量为例） | 说明 |
|---|---|---|
| TaskKey | `order_count_per_min` | 全局唯一，**同时作为 VictoriaMetrics 指标名**，命名即指标 |
| TaskType | Database | 数据库执行 SQL；接口取值选 URL，由 HTTP 探测取 JSON 路径值 |
| Command | `SELECT count(*) FROM t_order WHERE create_time >= {{beginTime}} AND create_time < {{endTime}}` | SQL/URL 模板，支持 `text/template` + sprig，可引用 `beginTime`/`endTime`/`startTime` 等 |
| ExecParams | `{ "databaseId": 1 }` | 数据库任务填数据源 id；URL 任务填 `resultFieldPath` |
| TimeSpan | `60` | 采集周期（秒），每分钟采一次 |
| StepSpan | 窗口宽度（秒） | 查询区间宽度 |
| Dashboards | 勾选第 2 步建的面板 | 任务与面板多对多；保存时对每个面板自动加一个 timeseries panel |
| MonitorGroup | 选依赖分组 | 树形分组，用于关联预警链路（"A -> B -> C"） |
| TaskStatus / AlertStatus / Sampled | 开 / 开 / 开 | 任务开关、告警开关、样本对比线展示开关 |

保存时系统会**校验 TaskKey 唯一**（重复报"任务key已存在"），对每个关联的 Grafana 面板自动加一个 timeseries panel（以 `panel.Description == TaskKey` 为锚点），并事务性保存 task + dashboardTasks + taskAlert。

### 4. 配置告警规则

任务与告警规则一对一，在任务的 create/modify 里一并配。规则 `Params` 是 JSON 数组，每组含：

- **组间关系 Relation**：`Or`（一组命中即整体命中）/ `And`（全命中才命中）。
- **组内规则 Rules**：每条 = 比较值类型 + 阈值 + 比较方式。
- **生效时段 EffectTimes**：`[startHH:mm:ss, endHH:mm:ss]`，不在时段内跳过该组。
- **事件等级 Level**：不填默认 Critical（可选 Low/Medium/High/Critical）。

三种比较值类型，按指标特性选：

| 类型 | 计算 | 适用 |
|---|---|---|
| 样本百分比 Percent | `(实时-基线)*100/基线` | 有周期起伏的业务量（订单量、访问量） |
| 绝对差值 AbsoluteValue | `实时-基线` | 平稳指标偏离基线的具体差值 |
| 直接值 Value | 不依赖基线，直接比实时值 | 本该趋零的故障型计数（慢查询、错误数） |

五种比较方式（差值模式下值带正负语义：超出 `diff>value`、低于 `diff<-value`）：

| 方式 | 差值模式 | 绝对值/直接值模式 |
|---|---|---|
| 超出 Gt | `diff > value` | `real > value` |
| 低于 Lt | `diff < -value` | `real < value` |
| 等于 / 超出或等于 / 低于或等于 | 同理 | 同理 |

订单量场景的规则配置：类型选**样本百分比**、值 `50`、比较方式选**超出**、生效时段全天、等级 Critical。配多条规则时，组内是 And、组间按 Relation 走 Or/And。

### 5. 配置通知通道与接收人

任务配好后还需让告警能"送达"：

- 「告警通道」：新建邮件 / 企业微信 / 钉钉通道，保存时会**测试发送**验证连通性；企业微信填机器人 webhook。
- 「告警分组」：组织接收人分组与成员，决定谁收。
- 「告警配置」：全局 KV——`firstDelay`（命中转 Firing 后多久首次外发，默认 60 秒）、`alertSpan`（外发后下次再报间隔，默认 600 秒/10 分钟）、`template`（通知模板）、`simplePageSize`/`simpleMaxSecond`（样本生成参数）。`alertMgmtKey` 为告警收敛中心所需 labelKey，空则跳过推送。

### 6. 跑起来后会发生什么

无需手动干预，四个定时任务自动接力：

1. **dataCollect** 每过一个 TimeSpan 采一次，写实时值 `avg({taskKey})` + 未来 8 天同时刻样本原料点。
2. **dataSampling** 按 TimeSpan 逐格补点，生成窗口覆盖到 `now+1天`，可提前产出未来一天的基线；聚合时仅使用 `_sampling` 历史原料（**已移除 realtime 回退**——`_sampling` 缺失或查询失败时跳过该格，绝不拿当前实时值当历史基线），写入 `avg({taskKey}_sample)`。运行几天后基线趋于稳定。
3. **alertCheck** 比对实时均值与基线：命中先进 Pending、持续超 Duration 才 Firing 建事件。
4. **eventCheck** 到 `NextAlertTime` 渲染模板，经通道外发给告警分组成员。

Grafana 面板上：target A 查 `avg({taskKey})` 实时曲线，开启 Sampled 后 target B 叠加 `avg({taskKey}_sample)` 基线对比线，一眼看出当前值 vs 历史趋势。

事件产生后在「告警事件」页处理：点「处理」暂停检测并介入，处置后点「完成」发恢复通知，误报则标「忽略」。

---

## 配置说明

主配置：`backend/configs/config.yml`，按 `engineMode` 命名的环境配置（如 `config-release.yml`）会被 viper 自动 merge。关键配置段：`db` / `victoriaMetrics` / `grafana` / `xxl` / `business`。

完整接口、业务规则与字段说明见 [`docs/功能文档.md`](docs/功能文档.md)，后端 / 前端工程细节分别见 [`backend/README.md`](backend/README.md) 与 [`frontend/README.md`](frontend/README.md)。
