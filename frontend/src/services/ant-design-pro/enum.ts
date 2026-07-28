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

// 数据源类型：1 Mongo 2 Mysql
export const DatabaseTypeEnum: Record<number, string> = {
  1: 'Mongo',
  2: 'Mysql',
};

// 任务类型：1 Database 2 URL 3 Push（外部推送）
export const TaskTypeEnum: Record<number, string> = {
  1: 'Database',
  2: 'URL',
  3: 'Push',
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
// 2: 实时-样本，差额比较
// 3: 直接用实时值比较
export const CheckParamValueTypeEnum: Record<number, string> = {
  1: '样本差阈值百分比',
  2: '样本差阈值比较',
  3: '实时数值比较',
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

// HTTP 请求方法
export const HttpMethodEnum: Record<string, string> = {
  POST: 'POST',
  GET: 'GET',
  PUT: 'PUT',
  DELETE: 'DELETE',
};
