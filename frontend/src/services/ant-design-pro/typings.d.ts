// @ts-nocheck
/* eslint-disable */

declare namespace API {
  type Resp<T> = {
    status: number;
    data: T;
    message: string;

    pageSize?: number;
    current?: number;
    total?: number;
  };

  type SysUser = {
    id: string;
    name?: string;
    avatar?: string;
    username?: string;
    mobile?: string;
    email?: string;
    roles?: string;
    roleList: string[];
    menus?: SysMenu[];
    codes?: string[];
    permissions?: string[];
  };

  type PageParams = {
    current?: number;
    pageSize?: number;
  };

  type LoginParams = {
    username?: string;
    password?: string;
  };

  type ErrorResponse = {
    errorCode: string;
    errorMessage?: string;
    success?: boolean;
  };

  type SysMenuOption = {
    id: string;
    name: string;
    value: string;
    method: string;
    path: string;
  };

  type SysMenu = {
    id: string;
    code: string;
    name: string;
    path: string;
    icon?: string;
    component?: string;
    sort?: number;
    options: SysMenuOption[];
    parent?: string;
    children?: SysMenu[];
    routes: SysMenu[];
  };

  type SysRole = {
    id: string;
    name: string;
    permissions: SysPermission[];
  };

  interface SysPermission {
    roleId?: string;
    menuId: string;
    option: string;
    independent: boolean;
    half: boolean;
    root: boolean;
  }

  interface SysRolePermission extends SysPermission {
    options: string[];
  }

  type MonitorDatabase = {
    id?: string;
    name?: string;
    database?: string;
    username?: string;
    password?: string;
    url?: string;
    type?: number;
    params?: string;
    /** 0未知 1正常 2异常 */
    healthStatus?: number;
    lastCheckTime?: string;
    lastError?: string;
    consecutiveFail?: number;
  };

  type MonitorTask = {
    id?: string;
    taskKey?: string;
    taskName?: string;
    timeSpan?: number;
    stepSpan?: number;
    command?: string;
    taskType?: number;
    /** 数据类型：1 正常查询(单值) 2 聚合查询(分组多行) */
    dataType?: number;
    execParams?: string;
    taskStatus?: number;
    /** 告警开关：0关闭 1开启 */
    alertStatus?: number;
    /**
     * 规则运行态（来自 taskAlert.alertStatus）：
     * 1正常 2异常 3告警；无告警配置时后端固定返回 1
     */
    taskAlertStatus?: number;
    /** 首次出现异常的时间（taskAlert.firstFlagTime）；从未异常不返回 */
    firstFlagTime?: string;
    sampled?: number;
    /**
     * 大促敏感：1 否 2 是。
     * 仅敏感任务在特殊日做原料剔除 / 冻结基线 / 告警比例放大。
     */
    promoSensitive?: number;
    monitorGroup?: string;
    labels?: string;
    /** 关联任务 ID（逗号分隔），叠加实时/样本曲线到本任务面板 */
    relatedTaskIds?: string;
    collectErrMsg?: string;
    sampleErrMsg?: string;
    preExecuteTime?: string;
  };

  type MonitorTaskCreate = MonitorTask & {
    taskExecParams?: Record<string, any>;
    dashboards?: string[];
    taskAlert?: Record<string, any>;
  };

  // 聚合预览响应：结果列名列表
  type MonitorTaskPreviewResponse = {
    columns?: string[];
  };

  // 告警规则比较项
  type MonitorAlertCheckParamsItem = {
    valueType?: number;
    value?: number;
    compareType?: number;
  };

  // 告警规则组
  type MonitorAlertCheckParams = {
    relation?: number;
    effectTimes?: string[];
    rules?: MonitorAlertCheckParamsItem[];
    level?: number;
  };

  // 任务告警配置（创建/详情）
  type MonitorTaskAlertConfig = {
    alertChannels?: string[];
    alertGroups?: string[];
    timeSpan?: number;
    duration?: number;
    checkParams?: MonitorAlertCheckParams[];
  };

  // 任务详情响应（含 taskAlert 嵌套，供编辑态回显）
  type MonitorTaskQueryResponse = MonitorTask & {
    taskExecParams?: Record<string, any>;
    dashboards?: string[];
    taskAlert?: MonitorTaskAlertConfig;
  };

  type MonitorDashboard = {
    id?: string;
    name?: string;
    slug?: string;
    url?: string;
    uid?: string;
    boardId?: string;
  };

  type MonitorDashboardTask = {
    id?: string;
    taskId?: string;
    dashboardId?: string;
    sort?: number;
    taskName?: string;
  };

  type MonitorGroup = {
    id?: string;
    name?: string;
    route?: string;
    parent?: string;
  };

  type MonitorTaskEvent = {
    id?: string;
    alertId?: string;
    taskId?: string;
    alertMsg?: string;
    dealStatus?: number;
    dealUser?: string;
    dealUserName?: string;
    taskName?: string;
    content?: string;
    eventLevel?: number;
    alertCount?: number;
    nextAlertTime?: string;
    preAlertTime?: string;
    createdAt?: string;
  };

  type AlertConf = {
    id?: string;
    confKey?: string;
    confVal?: string;
    confDesc?: string;
    confType?: number;
  };

  type AlertGroup = {
    id?: string;
    name?: string;
  };

  type AlertChannel = {
    id?: string;
    name?: string;
    type?: number;
    params?: string;
    handler?: string;
    failRoute?: number;
    /** 通道告警模板；空则回落到 alertConf 中对应 handler 的默认模板 */
    template?: string;
  };

  // 通道类型与可用处理器的绑定关系
  type AlertChannelHandler = {
    channelType: number;
    handlers: string[];
  };

  // 测试发送参数（不入库，仅本次请求触发测试发送）
  // 消息内容由通道 template / handler 默认模板 + 假参数渲染
  type AlertChannelTestParams = {
    email?: string;
  };

  // 创建/修改通道请求体，携带临时测试参数
  type AlertChannelSaveRequest = AlertChannel & {
    testParams?: AlertChannelTestParams;
  };

  type MonitorHomeCount = {
    taskCount?: number;
    eventCount?: number;
    dashboardCount?: number;
    databaseCount?: number;
    alertChannelCount?: number;
    alertGroupCount?: number;
    pendingEvents?: number;
    processingEvents?: number;
    completeEvents?: number;
    ignoreEvents?: number;
    taskTypeDistribution?: Record<string, number>;
    alertLevelDistribution?: Record<string, number>;
    recentEvents?: Array<{
      id?: number | string;
      taskName?: string;
      alertMsg?: string;
      dealStatus?: number;
      eventLevel?: number;
      createTime?: string;
    }>;
  };

  // 波动日管理
  type MonitorVolatilityDay = {
    id?: string;
    name?: string;
    startTime?: string;
    endTime?: string;
    /** 1=高峰 2=低谷 */
    type?: number;
  };

  type MonitorVolatilityDayBatchCreateRequest = {
    name?: string;
    /** 1=高峰 2=低谷 */
    type?: number;
    items?: MonitorVolatilityDay[];
  };
}
