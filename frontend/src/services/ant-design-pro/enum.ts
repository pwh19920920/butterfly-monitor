// 通道类型：1邮件 2Webhook 3短信
export const AlertChannelTypeEnum: Record<number, string> = {
  1: '邮件',
  2: 'Webhook',
  3: '短信',
};

// 失败路由：1否 2是
export const AlertChannelFailRouteEnum: Record<number, string> = {
  1: '否',
  2: '是',
};

// 数据源类型：1 Mongo 2 Mysql 3 PostgreSQL 4 MariaDB 5 TiDB 6 OceanBase 7 Doris 8 StarRocks 9 ClickHouse 10 Prometheus 11 OpenSearch 12 Elasticsearch 13 VictoriaMetrics 14 TDengine
export const DatabaseTypeEnum: Record<number, string> = {
  1: 'Mongo',
  2: 'Mysql',
  3: 'PostgreSQL',
  4: 'MariaDB',
  5: 'TiDB',
  6: 'OceanBase',
  7: 'Doris',
  8: 'StarRocks',
  9: 'ClickHouse',
  10: 'Prometheus',
  11: 'OpenSearch',
  12: 'Elasticsearch',
  13: 'VictoriaMetrics',
  14: 'TDengine',
};

// 数据源探活状态：0未知 1正常 2异常
export const DatabaseHealthStatusEnum: Record<number, string> = {
  0: '未知',
  1: '正常',
  2: '异常',
};

/** 数据源「附加参数」表单提示（按 type 维护，避免页面 if/else 散落） */
export type DatabaseParamsHint = {
  placeholder: string;
  tooltip: string;
};

const mysqlProtocolParamsHint: DatabaseParamsHint = {
  placeholder:
    'MySQL 协议族 DSN 查询串，例如 charset=utf8mb4&parseTime=True&loc=Local',
  tooltip:
    '适用于 Mysql/MariaDB/TiDB/OceanBase(MySQL模式)/Doris/StarRocks；留空默认 charset=utf8mb4&parseTime=True&loc=Local',
};

const openSearchParamsHint: DatabaseParamsHint = {
  placeholder: 'opensearch/es 参数，例如 scheme=https api=search timeout=30s',
  tooltip:
    '只读；url 填 host:port（默认 9200）；库名=默认 index；api=search|count|sql（默认 search）；Command 为 _search JSON / 空(_count) / SQL；OpenSearch 与 Elasticsearch 共用实现',
};

/** 各数据源类型的 params 提示；未单独配置的 MySQL 协议族走默认 */
export const DatabaseParamsHintEnum: Record<number, DatabaseParamsHint> = {
  1: {
    placeholder:
      'mongo URI 查询串，例如 collection=log&connectTimeoutMS=5000',
    tooltip: '拼接为 mongo URI 的 ? 后查询参数；可含 collection=xxx',
  },
  2: mysqlProtocolParamsHint,
  3: {
    placeholder: 'postgres 连接参数，例如 sslmode=disable TimeZone=Local',
    tooltip:
      '拼接为 postgres URL 的 ? 后查询参数；留空时默认 sslmode=disable',
  },
  4: mysqlProtocolParamsHint,
  5: mysqlProtocolParamsHint,
  6: mysqlProtocolParamsHint,
  7: mysqlProtocolParamsHint,
  8: mysqlProtocolParamsHint,
  9: {
    placeholder:
      'clickhouse 连接参数，例如 dial_timeout=10s&readonly=1',
    tooltip:
      '拼接为 clickhouse URL 的 ? 后查询参数；留空默认 dial_timeout=10s&readonly=1；原生协议端口常见 9000',
  },
  10: {
    placeholder:
      'prometheus 参数，例如 scheme=http path_prefix= timeout=30s',
    tooltip:
      '只读 PromQL 数据源；url 填 host:port（默认 9090）；可选 scheme=https、path_prefix=/prometheus、timeout=30s；库名可留空',
  },
  11: openSearchParamsHint,
  12: openSearchParamsHint,
  13: {
    placeholder:
      'victoriaMetrics 参数，例如 scheme=http path_prefix= timeout=30s',
    tooltip:
      '只读 MetricsQL/PromQL；复用 Prometheus 查询 API；url 填 host:port（默认 8428）；可选 scheme=https、path_prefix、timeout；库名可留空；不写对方',
  },
  14: {
    placeholder: 'tdengine 参数，例如 scheme=http timeout=30s',
    tooltip:
      '只读 REST SQL（/rest/sql）；url 填 host:port（默认 6041，taosAdapter）；库名=业务库；账号默认 root；Command 写 SELECT/SHOW；不建表不写数',
  },
};

/** 按数据源 type 取 params 提示，未知类型回落 MySQL 协议族文案 */
export function getDatabaseParamsHint(type?: number): DatabaseParamsHint {
  if (type != null && DatabaseParamsHintEnum[type]) {
    return DatabaseParamsHintEnum[type];
  }
  return mysqlProtocolParamsHint;
}

// 任务类型：4 Drilldown(系统下钻，特殊) 1 Database 2 URL 3 Push（外部推送）
// 下钻不用 0，避免与 Go 零值冲突导致 GORM Updates 跳过
export const TaskTypeEnum: Record<number, string> = {
  4: '系统下钻',
  1: 'Database',
  2: 'URL',
  3: 'Push',
};

// 数据类型：1 正常查询(单值) 2 聚合查询(分组多行)
export const DataTypeEnum: Record<number, string> = {
  1: '正常查询',
  2: '聚合查询',
};

// 任务状态
export const TaskStatusEnum: Record<number, string> = {
  0: '关闭',
  1: '开启',
};

// 告警状态开关
export const TaskAlertStatusEnum: Record<number, string> = {
  0: '关闭',
  1: '开启',
};

// 样本展示开关
export const TaskSampledEnum: Record<number, string> = {
  0: '关闭',
  1: '开启',
};

// 任务告警状态
export const MonitorTaskAlertStatusEnum: Record<number, string> = {
  1: '正常',
  2: '异常',
  3: '告警',
};

// 规则条件关系：1 或 2 且
export const CheckParamRelationEnum: Record<number, string> = {
  1: '或者-or',
  2: '并且-and',
};

// 比较类型
export const CheckParamCompareTypeEnum: Record<number, string> = {
  1: '高于',
  2: '低于',
  3: '等于',
  4: '大于等于',
  5: '小于等于',
};

// 比较值类型
// 1: (实时-样本)/样本 * 100，百分比波动
// 2: 直接用实时值比较（唯一不需要样本）
// 3: 实时-样本，差额比较
export const CheckParamValueTypeEnum: Record<number, string> = {
  1: '样本差阈值百分比',
  2: '实时数值比较',
  3: '样本差阈值比较',
};

// 告警等级
export const CheckParamsLevelTypeEnum: Record<number, string> = {
  0: '严重Critical',
  1: '高级High',
  2: '中级Medium',
  3: '低级Low',
};

// 事件处理状态：1待处理 2处理中 3已完成 4已忽略
export const MonitorTaskEventDealStatusEnum: Record<number, string> = {
  1: '待处理',
  2: '处理中',
  3: '已完成',
  4: '已忽略',
};

// 事件等级：-1正常，其余继承 CheckParamsLevelTypeEnum
export const MonitorTaskEventLevelEnum: Record<number | string, string> = {
  '-1': '正常',
  ...CheckParamsLevelTypeEnum,
};

// 报警配置类型：1数字 2字符串
export const AlertConfTypeEnum: Record<number, string> = {
  1: '数字',
  2: '字符串',
};

// 波动日类型：1高峰 2低谷
export const VolatilityDayTypeEnum: Record<number, string> = {
  1: '高峰',
  2: '低谷',
};

// HTTP 请求方法
export const HttpMethodEnum: Record<string, string> = {
  POST: 'POST',
  GET: 'GET',
  PUT: 'PUT',
  DELETE: 'DELETE',
};
