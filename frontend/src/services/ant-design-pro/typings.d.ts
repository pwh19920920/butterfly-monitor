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
    id: string | number;
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
    id: number;
    name: string;
    value: string;
    method: string;
    path: string;
  };

  type SysMenu = {
    id: number;
    code: string;
    name: string;
    path: string;
    icon?: string;
    component?: string;
    sort?: number;
    options: SysMenuOption[];
    parent?: number;
    children?: SysMenu[];
    routes: SysMenu[];
  };

  type SysRole = {
    id: number | string;
    name: string;
    permissions: SysPermission[];
  };

  interface SysPermission {
    roleId?: string | number;
    menuId: string | number;
    option: string;
    independent: boolean;
    half: boolean;
    root: boolean;
  }

  interface SysRolePermission extends SysPermission {
    options: string[];
  }

  type MonitorDatabase = {
    id?: string | number;
    name?: string;
    database?: string;
    username?: string;
    password?: string;
    url?: string;
    type?: number;
    params?: string;
  };

  type MonitorTask = {
    id?: string | number;
    taskKey?: string;
    taskName?: string;
    timeSpan?: number;
    stepSpan?: number;
    command?: string;
    taskType?: number;
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
    monitorGroup?: string;
    labels?: string;
    collectErrMsg?: string;
    sampleErrMsg?: string;
    preExecuteTime?: string;
  };

  type MonitorTaskCreate = MonitorTask & {
    taskExecParams?: Record<string, any>;
    dashboards?: string[];
    taskAlert?: Record<string, any>;
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
    id?: string | number;
    name?: string;
    slug?: string;
    url?: string;
    uid?: string;
    boardId?: string | number;
  };

  type MonitorDashboardTask = {
    id?: string | number;
    taskId?: string | number;
    dashboardId?: string | number;
    sort?: number;
    taskName?: string;
  };

  type MonitorGroup = {
    id?: string | number;
    name?: string;
    route?: string;
    parent?: string | number;
  };

  type MonitorTaskEvent = {
    id?: string | number;
    alertId?: string | number;
    taskId?: string | number;
    alertMsg?: string;
    dealStatus?: number;
    dealUser?: string | number;
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
    id?: string | number;
    confKey?: string;
    confVal?: string;
    confDesc?: string;
    confType?: number;
  };

  type AlertGroup = {
    id?: string | number;
    name?: string;
  };

  type AlertChannel = {
    id?: string | number;
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
}
