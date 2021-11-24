/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50735
 Source Host           : localhost:3306
 Source Schema         : butterfly_monitor

 Target Server Type    : MySQL
 Target Server Version : 50735
 File Encoding         : 65001

 Date: 28/10/2021 19:26:06
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_monitor_dashboard
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_dashboard`;
CREATE TABLE `t_monitor_dashboard`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `slug` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '英文名',
  `uid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '唯一ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '中文名',
  `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '地址',
  `board_id` bigint(20) NOT NULL COMMENT 'dash得id',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_monitor_dashboard_task
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_dashboard_task`;
CREATE TABLE `t_monitor_dashboard_task`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `task_id` bigint(20) NOT NULL,
  `dashboard_id` bigint(20) NOT NULL,
  `sort` int(10) NOT NULL DEFAULT 0 COMMENT '排序，大的靠前',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for t_monitor_database
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_database`;
CREATE TABLE `t_monitor_database`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `database` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '数据库名称',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '数据库中文名',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '账号',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '密码',
  `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '数据库连接地址',
  `type` tinyint(10) NOT NULL COMMENT '数据库类型：0-mysql',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_monitor_task
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task`;
CREATE TABLE `t_monitor_task`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `pre_execute_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上一次执行时间',
  `task_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务key，对应influxdb表',
  `task_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务名称',
  `time_span` int(10) NOT NULL COMMENT '时间间隔，s为单位',
  `command` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指令，可以是url，也可以是sql',
  `task_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务类型，0数据库，1url',
  `exec_params` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '执行参数',
  `task_status` tinyint(2) NOT NULL DEFAULT 1 COMMENT '任务状态，0关闭，1开启',
  `alert_status` tinyint(2) NOT NULL DEFAULT 1 COMMENT '报警状态，0关闭，1开启',
  `err_msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '错误信息',
  `sampled` tinyint(2) NOT NULL DEFAULT 1 COMMENT '是否需要样本，0不需要，1需要',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_sys_menu
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_menu`;
CREATE TABLE `t_sys_menu`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '菜单名称',
  `path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '菜单路径',
  `icon` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '菜单图标',
  `component` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '菜单组件',
  `sort` int(10) NOT NULL DEFAULT 0 COMMENT '菜单排序',
  `option` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '菜单操作',
  `parent` bigint(20) NOT NULL DEFAULT 0 COMMENT '上级目录',
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  `route` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '路由路径',
  `code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '菜单代码',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_menu
-- ----------------------------
INSERT INTO `t_sys_menu` VALUES (1332302770434215918, '2020-11-30 12:38:28', '2021-10-14 04:31:24', '系统管理', '/sys', 'crown', '', 2, '[]', 0, 0, '/1332302770434215918', 'sys');
INSERT INTO `t_sys_menu` VALUES (1332302770434215920, '2021-10-14 22:58:22', '2021-10-14 23:16:02', '菜单管理', '/sys/sysMenu', 'smile', './SysMenu', 1, '[]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215920', 'sysMenu');
INSERT INTO `t_sys_menu` VALUES (1332302770434215922, '2021-10-12 10:11:28', '2021-10-14 23:17:21', '用户管理', '/sys/sysUser', 'smile', './SysUser', 2, '[]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215922', 'sysUser');
INSERT INTO `t_sys_menu` VALUES (1332302770434215924, '2021-10-12 02:11:56', '2021-10-14 23:20:37', '角色管理', '/sys/sysRole', 'smile', './SysRole', 1, '[{\"id\":405,\"name\":\"xxx\",\"value\":\"xxxx\",\"method\":\"POST\",\"path\":\"/test\"}]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215924', 'sysRole');
INSERT INTO `t_sys_menu` VALUES (1332302770434215926, '2021-10-10 02:12:47', '2021-10-17 11:15:28', '监控管理', '/monitor', 'smile', '', 1, '[]', 0, 0, '/1332302770434215926', 'monitor');
INSERT INTO `t_sys_menu` VALUES (1332302770434215928, '2021-10-12 02:13:06', '2021-10-17 22:41:37', '数据源管理', '/monitor/database', 'smile', './MonitorDatabase', 2, '[]', 1332302770434215926, 0, '/1332302770434215926/1332302770434215928', 'monitorDatabase');
INSERT INTO `t_sys_menu` VALUES (1332302770434215930, '2021-10-12 02:13:51', '2021-10-25 18:06:30', '任务管理', '/monitor/task', 'table', './MonitorTask', 1, '[]', 1332302770434215926, 0, '/1332302770434215926/1332302770434215930', 'monitorTask');
INSERT INTO `t_sys_menu` VALUES (1452284009022230528, '2021-10-27 22:41:05', '2021-10-27 21:18:48', '面板管理', '/monitor/dashboard', 'table', './MonitorDashboard', 3, '', 1332302770434215926, 0, '/1332302770434215926/1452284009022230528', 'monitorDashboard');

-- ----------------------------
-- Table structure for t_sys_menu_option
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_menu_option`;
CREATE TABLE `t_sys_menu_option`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '名称',
  `value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '权限串',
  `method` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'URL方法',
  `path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'URL路径',
  `code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '唯一编码',
  `menu_id` bigint(20) NOT NULL COMMENT '菜单id',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `unq_code`(`code`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_menu_option
-- ----------------------------
INSERT INTO `t_sys_menu_option` VALUES (1447759564626726912, '2021-10-12 18:47:33', '2021-10-14 15:16:01', 0, '菜单查看', 'sys:menu:query', 'GET', '/api/sys/menu', 'caa126a343b0e1cef0774b637c246af3', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448238719118086145, '2021-10-12 19:04:17', '2021-10-14 15:16:01', 0, '菜单新增', 'sys:menu:create', 'POST', '/api/sys/menu', '79102b6efd1174afdf1732d9e7e80629', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927012900866, '2021-10-14 23:16:02', '2021-10-14 23:16:02', 0, '菜单修改', 'sys:menu:modify', 'PUT', '/api/sys/menu', '6b3b68431579eaf6b3d4a69a0ec18b08', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095168, '2021-10-14 23:16:02', '2021-10-14 23:16:02', 0, '菜单删除', 'sys:menu:delete', 'DELETE', '/api/sys/menu/:id', 'dca2d4b95c306fbd21a73e38e82ff007', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095169, '2021-10-14 23:16:02', '2021-10-14 23:16:02', 0, '菜单操作', 'sys:menu:option', 'GET', '/api/sys/menu/option/:id', '315a218f5d7b8f961d41604facedab91', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095170, '2021-10-14 23:16:02', '2021-10-14 23:16:02', 0, '菜单获取', 'sys:menu:queryWithOption', 'GET', '/api/sys/menu/withOption', '9ed6bf710b5b9802534783571be07050', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448669259440852992, '2021-10-14 23:17:21', '2021-10-14 23:17:21', 0, '用户查询', 'sys:user:query', 'GET', '/api/sys/user', 'c4b0789ebee8c17c0ff07829eec8670a', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1448669259440852993, '2021-10-14 23:17:21', '2021-10-14 23:17:21', 0, '用户修改', 'sys:user:modify', 'PUT', '/api/sys/user', '2b6a72c2eeb70feb99b4cd0b65e954eb', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1448669259440852994, '2021-10-14 23:17:21', '2021-10-14 23:17:21', 0, '用户创建', 'sys:menu:create', 'POST', '/api/sys/user', 'afbf478129b34727685739d3ca5d606a', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497152, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '角色查询', 'sys:role:query', 'GET', '/api/sys/role', '442aa2720fc50a72a35b53cb5a5695eb', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497153, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '角色创建', 'sys:role:create', 'POST', '/api/sys/role', 'c422e4a0129da9c465ae422280bf8838', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497154, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '角色修改', 'sys:role:modify', 'PUT', '/api/sys/role', '6c200c9d08b929494efea5790f3e5fa0', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497155, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '角色删除', 'sys:role:delete', 'DELETE', '/api/sys/role/:id', 'f5ea9ec30eead23392954aa9d5759183', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497156, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '查询全部角色', 'sys:role:queryAll', 'GET', '/api/sys/role/all', 'd75094205cf8cbb9927e08b0cc24be84', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497157, '2021-10-14 23:20:37', '2021-10-14 23:20:37', 0, '角色权限查询', 'sys:role:queryPermission', 'GET', '/api/sys/role/permission/:roleId', '65ce3dd61cfdb0f3f3b4e7bfafb59a37', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389440, '2021-10-15 07:21:59', '2021-10-17 14:41:37', 0, '数据源查询', 'monitor:database:query', 'GET', '/api/monitor/database', 'e9fdf345326fce9103dffe4b62c648f3', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389441, '2021-10-15 07:21:59', '2021-10-17 14:41:37', 0, '数据源查看', 'monitor:database:create', 'POST', '/api/monitor/database', '534eab021f3a2451281fff1d1767a0cc', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389442, '2021-10-15 07:21:59', '2021-10-17 14:41:37', 0, '数据源更新', 'monitor:database:modify', 'PUT', '/api/monitor/database', 'b4708d30ae76818d5d3f7ea355e51f65', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380992, '2021-10-17 20:50:48', '2021-10-25 18:06:30', 0, '任务查询', 'monitor:task:query', 'GET', '/api/monitor/task', 'b122c53237e751115ce1ecc913ec6865', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380993, '2021-10-17 20:50:48', '2021-10-25 18:06:30', 0, '任务更新', 'monitor:task:modify', 'PUT', '/api/monitor/task', 'a9837ad678785aaf1a5f8d806a0304bb', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380994, '2021-10-17 20:50:48', '2021-10-25 18:06:30', 0, '任务创建', 'monitor:task:create', 'POST', '/api/monitor/task', '1284964e9851ff3d3393c804c76100df', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449747431762694147, '2021-10-17 22:41:37', '2021-10-17 22:41:37', 0, '全部数据源', 'monitor:database:queryAll', 'GET', '/api/monitor/database/all', '5a16e70771500139bf739ac4b237b9c8', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605955, '2021-10-18 04:46:35', '2021-10-18 04:46:35', 0, '任务状态修改', 'monitor:task:modifyTaskStatus', 'PUT', '/api/monitor/task/taskStatus/:id/:status', '540164c1917f24d11d7359973d6d67e0', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605956, '2021-10-18 04:46:35', '2021-10-18 04:46:35', 0, '报警状态修改', 'monitor:task:modifyAlertStatus', 'PUT', '/api/monitor/task/alertStatus/:id/:status', 'f76e02a14e173e00e633f43797e321e5', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605957, '2021-10-25 18:06:30', '2021-10-25 18:06:30', 0, '收集状态修改', 'monitor:task:modifySampled', 'PUT', '/api/monitor/task/sampled/:id/:status', 'f9c1619590dc32e04e60323963bd597e', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999616, '2021-10-25 07:00:37', '2021-10-27 21:18:48', 0, '面板查询', 'monitor:dashboard:query', 'GET', '/api/monitor/dashboard', '510ac77819b00cd71805f27509d7eb6e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999617, '2021-10-25 07:00:37', '2021-10-27 21:18:48', 0, '面板创建', 'monitor:dashboard:create', 'POST', '/api/monitor/dashboard', 'd15f34654e3b6541a097230f41650071', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999618, '2021-10-25 07:00:37', '2021-10-27 21:18:48', 0, '面板更新', 'monitor:dashboard:modify', 'PUT', '/api/monitor/dashboard', 'acd15a7fc886a223f8afb15ee779587e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999619, '2021-10-25 07:00:37', '2021-10-27 21:18:48', 0, '面板全部', 'monitor:dashboard:queryAll', 'GET', '/api/monitor/dashboard/all', '32914b427951e2a2e88d2a35a5c5891f', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453286781452554244, '2021-10-27 19:28:01', '2021-10-27 21:18:48', 0, '全部任务', 'monitor:dashboard:queryAll', 'GET', '/api/monitor/dashboard/task/:id', '0095c870578fdf68dbae73e02c85e95e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453349603091943429, '2021-10-28 01:05:44', '2021-10-27 21:18:48', 0, '任务排序', 'monitor:dashboard:sort', 'PUT', '/api/monitor/dashboard/taskSort', 'ef422a4fd11cdbb1f416c72567da10a6', 1452284009022230528);

-- ----------------------------
-- Table structure for t_sys_permission
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_permission`;
CREATE TABLE `t_sys_permission`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `menu_id` bigint(20) NOT NULL COMMENT '菜单',
  `role_id` bigint(20) NOT NULL COMMENT '角色',
  `option` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '操作',
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  `independent` tinyint(2) NOT NULL DEFAULT 0 COMMENT '是否独立',
  `half` tinyint(2) NOT NULL DEFAULT 0 COMMENT '是否虚拟选中',
  `root` tinyint(2) NOT NULL COMMENT '是否为跟',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_permission
-- ----------------------------
INSERT INTO `t_sys_permission` VALUES (1453350103837315072, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215922, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315073, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215920, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315074, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215924, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315075, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215928, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315076, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215930, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315077, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215918, 1, '', 0, 1, 0, 1);
INSERT INTO `t_sys_permission` VALUES (1453350103837315078, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1452284009022230528, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315079, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215926, 1, '', 0, 1, 0, 1);
INSERT INTO `t_sys_permission` VALUES (1453350103837315080, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215920, 1, '1447759564626726912,1448238719118086145,1448668927012900866,1448668927017095168,1448668927017095169,1448668927017095170', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315081, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215922, 1, '1448669259440852992,1448669259440852993,1448669259440852994', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315082, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215924, 1, '1448670082216497152,1448670082216497153,1448670082216497154,1448670082216497156,1448670082216497157,1448670082216497155', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315083, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215928, 1, '1449574882915389440,1449574882915389441,1449574882915389442,1449747431762694147', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315084, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1332302770434215930, 1, '1449718480839380992,1449718480839380993,1449718480839380994,1452577297956605956,1452577297956605955,1452577297956605957', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1453350103837315085, '2021-10-27 21:17:21', '2021-10-27 21:17:21', 1452284009022230528, 1, '1453201790357999616,1453201790357999617,1453201790357999618,1453201790357999619,1453286781452554244,1453349603091943429', 0, 0, 0, 0);

-- ----------------------------
-- Table structure for t_sys_role
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_role`;
CREATE TABLE `t_sys_role`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色名称',
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_role
-- ----------------------------
INSERT INTO `t_sys_role` VALUES (1, '2021-10-29 20:51:40', '2021-10-27 21:17:21', 'super_admin', 0);
INSERT INTO `t_sys_role` VALUES (1447459031953182720, '2021-10-11 15:08:20', '2021-10-11 15:26:45', 'test', 1);
INSERT INTO `t_sys_role` VALUES (1447459092053364736, '2021-10-15 07:08:35', '2021-10-11 15:26:42', 'xx', 1);

-- ----------------------------
-- Table structure for t_sys_token
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_token`;
CREATE TABLE `t_sys_token`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `secret` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `subject` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 119 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for t_sys_user
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_user`;
CREATE TABLE `t_sys_user`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码',
  `salt` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码盐',
  `deleted` tinyint(1) NOT NULL DEFAULT 0,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '名称',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '头像',
  `roles` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '角色',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `unq_username`(`username`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_user
-- ----------------------------
INSERT INTO `t_sys_user` VALUES (1, '2020-11-24 07:49:07', '2021-10-12 22:55:22', 'admin', '593d4632a8c70251d0e9be4b1799bcc1', '54099a65-a235-158c-d610-74d2ff4c789b', 0, '王小二', 'https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png', '1');

SET FOREIGN_KEY_CHECKS = 1;
