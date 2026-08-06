/**
 * @name 路由配置
 * access: routeAccess 依赖 currentUser.codes 与 route.name 匹配
 * 注意：exportStatic 要求每条路由都有 path，404 用 /* 通配
 */
export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        name: 'login',
        path: '/user/login',
        component: './user/login',
      },
      {
        path: '/user',
        redirect: '/user/login',
      },
      {
        path: '/user/*',
        component: './404',
      },
    ],
  },
  {
    path: '/welcome',
    name: 'welcome',
    icon: 'smile',
    component: './Welcome',
  },
  {
    path: '/sys',
    name: 'sys',
    icon: 'crown',
    access: 'routeAccess',
    routes: [
      {
        path: '/sys',
        redirect: '/sys/sysUser',
      },
      {
        name: 'sysMenu',
        path: '/sys/sysMenu',
        access: 'routeAccess',
        component: './SysMenu',
      },
      {
        name: 'sysRole',
        path: '/sys/sysRole',
        access: 'routeAccess',
        component: './SysRole',
      },
      {
        name: 'sysUser',
        path: '/sys/sysUser',
        access: 'routeAccess',
        component: './SysUser',
      },
    ],
  },
  {
    path: '/monitor',
    name: 'monitor',
    icon: 'dashboard',
    access: 'routeAccess',
    routes: [
      {
        path: '/monitor',
        redirect: '/monitor/task',
      },
      {
        name: 'monitorTask',
        path: '/monitor/task',
        access: 'routeAccess',
        component: './MonitorTask',
      },
      {
        name: 'monitorTaskEvent',
        path: '/monitor/taskEvent',
        access: 'routeAccess',
        component: './MonitorTaskEvent',
      },
      {
        name: 'monitorGroup',
        path: '/monitor/group',
        access: 'routeAccess',
        component: './MonitorGroup',
      },
    ],
  },
  {
    path: '/alert',
    name: 'alert',
    icon: 'bell',
    access: 'routeAccess',
    routes: [
      {
        path: '/alert',
        redirect: '/alert/alertConf',
      },
      {
        name: 'alertConf',
        path: '/alert/alertConf',
        access: 'routeAccess',
        component: './AlertConf',
      },
      {
        name: 'alertGroup',
        path: '/alert/alertGroup',
        access: 'routeAccess',
        component: './AlertGroup',
      },
      {
        name: 'alertChannel',
        path: '/alert/alertChannel',
        access: 'routeAccess',
        component: './AlertChannel',
      },
      {
        name: 'monitorVolatilityDay',
        path: '/alert/volatilityDay',
        access: 'routeAccess',
        component: './MonitorVolatilityDay',
      },
      {
        name: 'monitorDatabase',
        path: '/alert/database',
        access: 'routeAccess',
        component: './MonitorDatabase',
      },
      {
        name: 'monitorDashboard',
        path: '/alert/dashboard',
        access: 'routeAccess',
        component: './MonitorDashboard',
      },
    ],
  },
  {
    path: '/',
    redirect: '/welcome',
  },
  {
    path: '*',
    layout: false,
    component: './404',
  },
];
