/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50736
 Source Host           : localhost:3306
 Source Schema         : butterfly_monitor

 Target Server Type    : MySQL
 Target Server Version : 50736
 File Encoding         : 65001

 Date: 13/05/2022 14:47:27
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_alert_channel
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_channel`;
CREATE TABLE `t_alert_channel`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '渠道名称',
  `type` int(10) NOT NULL DEFAULT 1 COMMENT '渠道类型，1邮件，2webhook，3短信',
  `params` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '渠道参数',
  `handler` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '渠道key',
  `fail_route` int(255) NOT NULL DEFAULT 1 COMMENT '失败路由，1否，2是',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警通道' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_alert_channel
-- ----------------------------
INSERT INTO `t_alert_channel` VALUES (1465598810553061376, '2021-12-01 00:29:21', '2022-05-13 14:42:19', 0, '企业微信群提醒', 2, '{\"addr\":\"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=d1d5354d-be44-4539-b6b0-d7534bde1e33\"}', 'ChannelWechatHandler', 1);
INSERT INTO `t_alert_channel` VALUES (1465618975940415488, '2021-12-02 01:49:28', '2022-05-13 14:42:19', 0, '邮箱', 1, '{\"host\":\"smtp.exmail.qq.com\",\"port\":465,\"username\":\"ibg-fund@we.cn\",\"password\":\"3dkrkpzPecq59kZL\",\"ssl\":1}', 'ChannelEmailHandler', 1);

-- ----------------------------
-- Table structure for t_alert_conf
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_conf`;
CREATE TABLE `t_alert_conf`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `conf_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '300' COMMENT '报警间隔，单位秒',
  `conf_val` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '模板',
  `conf_desc` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '配置描述',
  `conf_type` int(4) NOT NULL COMMENT '配置类型：1数字，2字符串',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uniq_conf_key`(`conf_key`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警配置' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_alert_conf
-- ----------------------------
INSERT INTO `t_alert_conf` VALUES (1462795963172130816, '2021-11-23 06:51:50', '2022-05-13 14:42:24', 0, 'alertSpan', '300', '报警间隔', 1);
INSERT INTO `t_alert_conf` VALUES (1462797872809381888, '2021-11-26 14:59:25', '2022-05-13 14:42:24', 0, 'template', 'spider系统预警\n\n{{range .items}}\n{{.TaskName}}：{{.HitRule}}\n{{end}}\n\n', '报警模板', 2);

-- ----------------------------
-- Table structure for t_alert_group
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_group`;
CREATE TABLE `t_alert_group`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '报警组名称',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警组' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_alert_group
-- ----------------------------
INSERT INTO `t_alert_group` VALUES (1465196552401195008, '2021-12-02 05:50:55', '2022-05-13 14:42:30', 0, '核心系统组');

-- ----------------------------
-- Table structure for t_alert_group_user
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_group_user`;
CREATE TABLE `t_alert_group_user`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `user_id` bigint(20) NOT NULL COMMENT '用户id',
  `group_id` bigint(20) NOT NULL COMMENT '分组id',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_group`(`group_id`) USING BTREE,
  INDEX `idx_user`(`user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警组用户' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_alert_group_user
-- ----------------------------
INSERT INTO `t_alert_group_user` VALUES (1465213686179172352, '2021-11-29 14:59:01', '2022-05-13 14:42:35', 0, 1, 1465196552401195008);

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
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '监控主板表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_monitor_dashboard
-- ----------------------------
INSERT INTO `t_monitor_dashboard` VALUES (1524657288403488769, '2022-05-12 15:46:39', '2022-05-13 14:42:42', 0, 'zi-jin-da-pan', '3UWRc3_7k', '资金大盘', '/d/3UWRc3_7k/zi-jin-da-pan', 1);
INSERT INTO `t_monitor_dashboard` VALUES (1524668164594470914, '2022-05-12 16:29:52', '2022-05-12 16:29:52', 0, 'zi-jin-da-pan-2', '5iNo23lnz', '资金大盘2', '/d/5iNo23lnz/zi-jin-da-pan-2', 2);
INSERT INTO `t_monitor_dashboard` VALUES (1525004012573691906, '2022-05-13 14:44:24', '2022-05-13 14:44:24', 0, 'zi-jin-da-pan-4', '_8u__RX7z', '资金大盘4', '/d/_8u__RX7z/zi-jin-da-pan-4', 4);

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
  `sort` int(10) NOT NULL DEFAULT 1 COMMENT '排序，大的靠前',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '监控主板任务关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_monitor_dashboard_task
-- ----------------------------

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
  `type` tinyint(10) NOT NULL COMMENT '数据库类型：1-mongodb，0-mysql',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '监控数据源表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_monitor_database
-- ----------------------------
INSERT INTO `t_monitor_database` VALUES (1478994420463308801, '2022-01-06 15:38:43', '2022-05-13 14:42:57', 0, 'butterfly_monitor', 'Mysql', 'root', 'root', 'localhost:3306', 0);

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
  `pre_sample_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上一次样本时间',
  `task_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务key，对应influxdb表',
  `task_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务名称',
  `step_span` int(10) NOT NULL DEFAULT 10 COMMENT '跨步间隔，s为单位',
  `time_span` int(10) NOT NULL COMMENT '时间间隔，s为单位',
  `command` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指令，可以是url，也可以是sql',
  `task_type` int(4) NOT NULL COMMENT '任务类型，1数据库，2url',
  `exec_params` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '执行参数',
  `task_status` tinyint(2) NOT NULL DEFAULT 1 COMMENT '任务状态，2关闭，1开启',
  `alert_status` tinyint(2) NOT NULL DEFAULT 1 COMMENT '报警状态，2关闭，1开启',
  `collect_err_msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '收集错误信息',
  `sample_err_msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '样本错误信息',
  `sampled` tinyint(2) NOT NULL DEFAULT 1 COMMENT '是否需要样本，2不需要，1需要',
  `recall_status` int(4) NOT NULL DEFAULT 1 COMMENT '是否支持回溯：1支持，2不支持',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uniq_task_key`(`task_key`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '监控任务表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_monitor_task
-- ----------------------------

-- ----------------------------
-- Table structure for t_monitor_task_alert
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task_alert`;
CREATE TABLE `t_monitor_task_alert`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标记',
  `task_id` bigint(20) NOT NULL COMMENT '任务id',
  `alert_channels` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '报警渠道列表',
  `alert_groups` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '报警组列表',
  `time_span` int(10) NOT NULL COMMENT '检测间隔，30s起步，以30为倍数',
  `duration` int(10) NOT NULL DEFAULT 0 COMMENT '持续时间，秒为单位',
  `params` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '规则参数：[{比较方式，值，关系，比较值类型}]',
  `alert_status` int(4) NOT NULL DEFAULT 1 COMMENT '报警状态：1正常，2出现异常，3达到报警条件',
  `deal_status` int(4) NOT NULL DEFAULT 1 COMMENT '处理状态：1正常，2处理中',
  `pre_check_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上一次的检测时间',
  `first_flag_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '首次标记时间, 如果未出现异常, 则此值持续更新, 如果出现异常, 则这个值不再更新',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uniq_task_alert`(`task_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警规则' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_monitor_task_alert
-- ----------------------------

-- ----------------------------
-- Table structure for t_monitor_task_event
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task_event`;
CREATE TABLE `t_monitor_task_event`  (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted` tinyint(1) NOT NULL DEFAULT 2 COMMENT '删除标记',
  `alert_id` bigint(20) NOT NULL COMMENT '报警规则id',
  `task_id` bigint(20) NOT NULL COMMENT '任务id',
  `alert_msg` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `deal_time` datetime NULL DEFAULT NULL COMMENT '处理时间',
  `complete_time` datetime NULL DEFAULT NULL COMMENT '完成时间',
  `content` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '事件经过',
  `deal_status` int(4) NOT NULL DEFAULT 1 COMMENT '处理状态：1待认领，2处理中，3处理完成，4误报忽略',
  `deal_user` bigint(20) NULL DEFAULT NULL COMMENT '认领人',
  `pre_alert_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上次报警时间',
  `next_alert_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '下次报警时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_alert_id`(`alert_id`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '报警事件表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of t_monitor_task_event
-- ----------------------------

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
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统菜单表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_menu
-- ----------------------------
INSERT INTO `t_sys_menu` VALUES (1332302770434215918, '2020-11-30 20:38:28', '2022-05-13 14:43:25', '系统管理', '/sys', 'crown', '', 200, '[]', 0, 0, '/1332302770434215918', 'sys');
INSERT INTO `t_sys_menu` VALUES (1332302770434215920, '2021-10-14 22:58:22', '2022-05-13 14:43:25', '菜单管理', '/sys/sysMenu', 'smile', './SysMenu', 1, '[]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215920', 'sysMenu');
INSERT INTO `t_sys_menu` VALUES (1332302770434215922, '2021-10-13 02:11:28', '2022-05-13 14:43:25', '用户管理', '/sys/sysUser', 'smile', './SysUser', 2, '[]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215922', 'sysUser');
INSERT INTO `t_sys_menu` VALUES (1332302770434215924, '2021-10-12 02:11:56', '2022-05-13 14:43:25', '角色管理', '/sys/sysRole', 'smile', './SysRole', 1, '[{\"id\":405,\"name\":\"xxx\",\"value\":\"xxxx\",\"method\":\"POST\",\"path\":\"/test\"}]', 1332302770434215918, 0, '/1332302770434215918/1332302770434215924', 'sysRole');
INSERT INTO `t_sys_menu` VALUES (1332302770434215926, '2021-10-11 18:12:47', '2022-05-13 14:43:25', '监控管理', '/monitor', 'smile', '', 100, '[]', 0, 0, '/1332302770434215926', 'monitor');
INSERT INTO `t_sys_menu` VALUES (1332302770434215928, '2021-10-12 02:13:06', '2022-05-13 14:43:25', '数据源管理', '/monitor/database', 'smile', './MonitorDatabase', 2, '[]', 1332302770434215926, 0, '/1332302770434215926/1332302770434215928', 'monitorDatabase');
INSERT INTO `t_sys_menu` VALUES (1332302770434215930, '2021-10-18 02:13:51', '2022-05-13 14:43:25', '任务管理', '/monitor/task', 'table', './MonitorTask', 1, '[]', 1332302770434215926, 0, '/1332302770434215926/1332302770434215930', 'monitorTask');
INSERT INTO `t_sys_menu` VALUES (1452284009022230528, '2021-10-27 22:41:05', '2022-05-13 14:43:25', '面板管理', '/monitor/dashboard', 'table', './MonitorDashboard', 3, '', 1332302770434215926, 0, '/1332302770434215926/1452284009022230528', 'monitorDashboard');
INSERT INTO `t_sys_menu` VALUES (1462709329521020928, '2021-11-24 01:07:35', '2022-05-13 14:43:25', '报警配置', '/alert/alertConf', 'smile', './AlertConf', 4, '', 1465652495161233408, 0, '/1465652495161233408/1462709329521020928', 'alertConf');
INSERT INTO `t_sys_menu` VALUES (1465165133809455104, '2021-11-30 03:46:04', '2022-05-13 14:43:25', '报警组管理', '/alert/alertGroup', 'smile', './AlertGroup', 4, '', 1465652495161233408, 0, '/1465652495161233408/1465165133809455104', 'alertGroup');
INSERT INTO `t_sys_menu` VALUES (1465561401698291712, '2021-12-01 14:00:42', '2022-05-13 14:43:25', '报警通道', '/alert/alertChannel', 'smile', './AlertChannel', 5, '', 1465652495161233408, 0, '/1465652495161233408/1465561401698291712', 'alertChannel');
INSERT INTO `t_sys_menu` VALUES (1465652495161233408, '2021-12-01 12:02:40', '2022-05-13 14:43:25', '报警管理', '/alert', 'crown', '', 50, '', 0, 0, '/1465652495161233408', 'alert');
INSERT INTO `t_sys_menu` VALUES (1472888326758338560, '2021-12-21 19:15:17', '2022-05-13 14:43:25', '异常事件', '/monitor/taskEvent', 'smile', './MonitorTaskEvent', 4, '', 1332302770434215926, 0, '/1332302770434215926/1472888326758338560', 'monitorTaskEvent');

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
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统菜单操作表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_menu_option
-- ----------------------------
INSERT INTO `t_sys_menu_option` VALUES (1447759564626726912, '2021-10-12 18:47:33', '2022-05-13 14:43:33', 0, '菜单查看', 'sys:menu:query', 'GET', '/api/sys/menu', 'caa126a343b0e1cef0774b637c246af3', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448238719118086145, '2021-10-12 19:04:17', '2022-05-13 14:43:33', 0, '菜单新增', 'sys:menu:create', 'POST', '/api/sys/menu', '79102b6efd1174afdf1732d9e7e80629', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927012900866, '2021-10-14 23:16:02', '2022-05-13 14:43:33', 0, '菜单修改', 'sys:menu:modify', 'PUT', '/api/sys/menu', '6a7b949b19c27a1f9ee3753e06c3ecf5', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095168, '2021-10-14 23:16:02', '2022-05-13 14:43:33', 0, '菜单删除', 'sys:menu:delete', 'DELETE', '/api/sys/menu/:id', 'c249795688bf6e62fe7b16ba1d539540', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095169, '2021-10-14 23:16:02', '2022-05-13 14:43:33', 0, '菜单操作', 'sys:menu:option', 'GET', '/api/sys/menu/option/:id', 'dfb98e82d0666d936314879ab3cbe37d', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448668927017095170, '2021-10-14 23:16:02', '2022-05-13 14:43:33', 0, '菜单获取', 'sys:menu:queryWithOption', 'GET', '/api/sys/menu/withOption', '3bf2be85387a98b219c49921dab0c28c', 1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497152, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '角色查询', 'sys:role:query', 'GET', '/api/sys/role', '92cd13408fbd6512a4e5328c800d5439', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497153, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '角色创建', 'sys:role:create', 'POST', '/api/sys/role', '0f1f61330f3ef4b03bf5632bbcc5737f', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497154, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '角色修改', 'sys:role:modify', 'PUT', '/api/sys/role', 'f83fd0b3bd1902b67676b1a64a78b309', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497155, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '角色删除', 'sys:role:delete', 'DELETE', '/api/sys/role/:id', '0f14a64e7dff47937164f64dfa01dbbf', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497156, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '查询全部角色', 'sys:role:queryAll', 'GET', '/api/sys/role/all', '5e987b4ec4dd5da6eb30d86e87e54f67', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1448670082216497157, '2021-10-14 23:20:37', '2022-05-13 14:43:33', 0, '角色权限查询', 'sys:role:queryPermission', 'GET', '/api/sys/role/permission/:roleId', '381b1a498606e82226a0604d8c853e65', 1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389440, '2021-10-15 07:21:59', '2022-05-13 14:43:33', 0, '数据源查询', 'monitor:database:query', 'GET', '/api/monitor/database', 'e9fdf345326fce9103dffe4b62c648f3', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389441, '2021-10-15 07:21:59', '2022-05-13 14:43:33', 0, '数据源查看', 'monitor:database:create', 'POST', '/api/monitor/database', '534eab021f3a2451281fff1d1767a0cc', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389442, '2021-10-15 07:21:59', '2022-05-13 14:43:33', 0, '数据源更新', 'monitor:database:modify', 'PUT', '/api/monitor/database', 'b4708d30ae76818d5d3f7ea355e51f65', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380992, '2021-10-17 20:50:48', '2022-05-13 14:43:33', 0, '任务查询', 'monitor:task:query', 'GET', '/api/monitor/task', 'b122c53237e751115ce1ecc913ec6865', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380993, '2021-10-17 20:50:48', '2022-05-13 14:43:33', 0, '任务更新', 'monitor:task:modify', 'PUT', '/api/monitor/task', 'a9837ad678785aaf1a5f8d806a0304bb', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380994, '2021-10-17 20:50:48', '2022-05-13 14:43:33', 0, '任务创建', 'monitor:task:create', 'POST', '/api/monitor/task', '1284964e9851ff3d3393c804c76100df', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449747431762694147, '2021-10-17 22:41:37', '2022-05-13 14:43:33', 0, '全部数据源', 'monitor:database:queryAll', 'GET', '/api/monitor/database/all', '92690787f68af00a309627e5bdadf55f', 1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605955, '2021-10-18 04:46:35', '2022-05-13 14:43:33', 0, '任务状态修改', 'monitor:task:modifyTaskStatus', 'PUT', '/api/monitor/task/taskStatus/:id/:status', '540164c1917f24d11d7359973d6d67e0', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605956, '2021-10-18 04:46:35', '2022-05-13 14:43:33', 0, '报警状态修改', 'monitor:task:modifyAlertStatus', 'PUT', '/api/monitor/task/alertStatus/:id/:status', 'f76e02a14e173e00e633f43797e321e5', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999616, '2021-10-25 07:00:37', '2022-05-13 14:43:33', 0, '面板查询', 'monitor:dashboard:query', 'GET', '/api/monitor/dashboard', '510ac77819b00cd71805f27509d7eb6e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999617, '2021-10-25 07:00:37', '2022-05-13 14:43:33', 0, '面板创建', 'monitor:dashboard:create', 'POST', '/api/monitor/dashboard', 'd15f34654e3b6541a097230f41650071', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999618, '2021-10-25 07:00:37', '2022-05-13 14:43:33', 0, '面板更新', 'monitor:dashboard:modify', 'PUT', '/api/monitor/dashboard', 'acd15a7fc886a223f8afb15ee779587e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999619, '2021-10-25 07:00:37', '2022-05-13 14:43:33', 0, '面板全部', 'monitor:dashboard:queryAll', 'GET', '/api/monitor/dashboard/all', '32914b427951e2a2e88d2a35a5c5891f', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453286781452554244, '2021-10-27 19:28:01', '2022-05-13 14:43:33', 0, '全部任务', 'monitor:dashboard:queryAll', 'GET', '/api/monitor/dashboard/task/:id', '0095c870578fdf68dbae73e02c85e95e', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453349603091943429, '2021-10-28 01:05:44', '2022-05-13 14:43:34', 0, '任务排序', 'monitor:dashboard:sort', 'PUT', '/api/monitor/dashboard/taskSort', 'ef422a4fd11cdbb1f416c72567da10a6', 1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1462721145227710464, '2021-11-23 01:21:58', '2022-05-13 14:43:34', 0, '读取配置', 'alert:conf:query', 'GET', '/api/alert/conf', '8457f70f619ae309571253e7f81213ef', 1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1462721145227710465, '2021-11-23 01:21:58', '2022-05-13 14:43:34', 0, '修改配置', 'alert:conf:modify', 'PUT', '/api/alert/conf', '052a70165846e9a96c726d82a71d683e', 1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1463448357279109125, '2021-10-26 02:06:30', '2022-05-13 14:43:34', 0, '收集状态修改', 'monitor:task:modifySampled', 'PUT', '/api/monitor/task/sampled/:id/:status', '517025152bd9ec4062976bc681923418', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1463451882771976198, '2021-11-25 02:04:13', '2022-05-13 14:43:34', 0, '任务回溯', 'monitor:task:recall', 'POST', '/api/monitor/task/execForTimeRange/:id', '5a8f57afb59643d60a7173c0e7d1c691', 1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637184, '2021-10-15 07:17:21', '2022-05-13 14:43:34', 0, '用户查询', 'sys:user:query', 'GET', '/api/sys/user', '0924d00bac6e4d1b9e10040e095a980f', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637185, '2021-10-15 07:17:21', '2022-05-13 14:43:34', 0, '用户修改', 'sys:user:modify', 'PUT', '/api/sys/user', 'e75e13959a8f6577fee78e1bc61d3e10', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637186, '2021-10-15 07:17:21', '2022-05-13 14:43:34', 0, '用户创建', 'sys:menu:create', 'POST', '/api/sys/user', '80b0eae883b924868c44df5295b8ee33', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637187, '2021-11-25 22:55:04', '2022-05-13 14:43:34', 0, '全部用户', 'sys:user:queryAll', 'GET', '/api/sys/user/all', '2b202b723b15b6b2db71b9cf0079c904', 1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843712, '2021-11-29 11:46:04', '2022-05-13 14:43:34', 0, '报警组查询', 'alert:group:query', 'GET', '/api/alert/group', '3d3511b26417d92fdb9a6c705e769af8', 1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843713, '2021-11-29 11:46:04', '2022-05-13 14:43:34', 0, '报警组创建', 'alert:group:create', 'POST', '/api/alert/group', '2b7c5300c8efa0671a0b584e1a76986a', 1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843714, '2021-11-29 11:46:04', '2022-05-13 14:43:34', 0, '报警组修改', 'alert:group:modify', 'PUT', '/api/alert/group', '27b37f08e60468864f6d2c74d583daf2', 1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843715, '2021-11-29 11:46:04', '2022-05-13 14:43:34', 0, '报警组查全部', 'alert:group:queryall', 'GET', '/api/alert/group/all', '84138424922938bf2c6fda3b29557e96', 1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843716, '2021-11-29 11:46:04', '2022-05-13 14:43:34', 0, '报警组下用户', 'alert:group:users', 'GET', '/api/alert/group/groupUser/:id', 'b8b09ecfedbd9bf2282a5400131b8eb6', 1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068928, '2021-11-30 14:00:42', '2022-05-13 14:43:34', 0, '报警通道查询', 'alert:channel:query', 'GET', '/api/alert/channel', 'a9fb2fa8c82717cb47209781b0b017db', 1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068929, '2021-11-30 14:00:42', '2022-05-13 14:43:34', 0, '报警通道创建', 'alert:channel:save', 'POST', '/api/alert/channel', '2428c926e4bdaa278532ff34c7797d6f', 1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068930, '2021-11-30 14:00:42', '2022-05-13 14:43:34', 0, '报警通道修改', 'alert:channel:update', 'PUT', '/api/alert/channel', 'aea7ff8fd9ad8c8fb939f4b560e9138a', 1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068931, '2021-11-30 14:00:42', '2022-05-13 14:43:34', 0, '报警通道处理器', 'alert:channel:handlers', 'GET', '/api/alert/channel/handlers', 'cfa624de6151bd2f614f24ad88417009', 1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465652723025186818, '2021-11-23 01:54:32', '2022-05-13 14:43:34', 0, '创建配置', 'alert:conf:create', 'POST', '/api/alert/conf', '1e5446c29206d4da16838b45a3601e43', 1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1467698655321395204, '2021-12-06 11:33:23', '2022-05-13 14:43:34', 0, '报警通道全部', 'alert:channel:queryAll', 'GET', '/api/alert/channel/all', '90f3cf9c6ad1d18013b95397719fabd4', 1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1472889239543746560, '2021-12-21 03:16:12', '2022-05-13 14:43:34', 0, '事件查询', 'monitor:taskEvent:query', 'GET', '/api/monitor/task/event', '1ebaf6a664546061edeedeabcedf2f1f', 1472888326758338560);
INSERT INTO `t_sys_menu_option` VALUES (1473138104612163585, '2021-12-21 11:47:48', '2022-05-13 14:43:34', 0, '事件处理', 'monitor:taskEvent:deal', 'POST', '/api/monitor/task/event/deal/:id', 'c0b8e609bfd3af3f287b5916d4f543ed', 1472888326758338560);
INSERT INTO `t_sys_menu_option` VALUES (1473138104612163586, '2021-12-21 11:47:48', '2022-05-13 14:43:34', 0, '事件完成', 'monitor:taskEvent:complete', 'POST', '/api/monitor/task/event/complete/:id', '336c174b06c39a1b49640c8e2175b8c6', 1472888326758338560);

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
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统权限表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_permission
-- ----------------------------
INSERT INTO `t_sys_permission` VALUES (1473138135784230912, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215922, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230913, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215920, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230914, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215924, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230915, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1452284009022230528, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230916, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215928, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230917, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215930, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230918, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1465561401698291712, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230919, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1465165133809455104, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230920, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1462709329521020928, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230921, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215918, 1, '', 0, 1, 0, 1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230922, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1465652495161233408, 1, '', 0, 1, 0, 1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230923, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1472888326758338560, 1, '', 0, 1, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230924, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215926, 1, '', 0, 1, 0, 1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230925, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215920, 1, '1447759564626726912,1448238719118086145,1448668927012900866,1448668927017095168,1448668927017095169,1448668927017095170', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230926, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215924, 1, '1448670082216497152,1448670082216497153,1448670082216497154,1448670082216497156,1448670082216497157,1448670082216497155', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230927, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215928, 1, '1449574882915389440,1449574882915389441,1449574882915389442,1449747431762694147', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230928, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215930, 1, '1449718480839380992,1449718480839380993,1449718480839380994,1452577297956605956,1452577297956605955,1463448357279109125,1463451882771976198', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230929, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1452284009022230528, 1, '1453201790357999616,1453201790357999617,1453201790357999618,1453201790357999619,1453286781452554244,1453349603091943429', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230930, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1462709329521020928, 1, '1462721145227710464,1462721145227710465,1465652723025186818', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230931, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1332302770434215922, 1, '1463883940220637184,1463883940220637185,1463883940220637186,1463883940220637187', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230932, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1465165133809455104, 1, '1465165133817843712,1465165133817843713,1465165133817843714,1465165133817843715,1465165133817843716', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230933, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1465561401698291712, 1, '1465561401715068928,1465561401715068929,1465561401715068930,1465561401715068931,1467698655321395204', 0, 0, 0, 0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230934, '2021-12-21 11:47:56', '2022-05-13 14:43:41', 1472888326758338560, 1, '1472889239543746560,1473138104612163585,1473138104612163586', 0, 0, 0, 0);

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
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统角色表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_role
-- ----------------------------
INSERT INTO `t_sys_role` VALUES (1, '2021-11-05 04:51:40', '2022-05-13 14:43:50', 'super_admin', 0);

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
) ENGINE = InnoDB AUTO_INCREMENT = 190 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统令牌表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_token
-- ----------------------------
INSERT INTO `t_sys_token` VALUES (188, '2022-05-12 15:20:11', '2022-05-12 15:20:11', 'fa679b61-f79d-473d-9c7f-1a21d1a718a9', 1, '39e36e27-9fa2-4739-8a3b-5b7d0fd3f384', 0);
INSERT INTO `t_sys_token` VALUES (189, '2022-05-13 14:40:45', '2022-05-13 14:40:45', '20912588-ed08-4598-89f2-73d83d4ef0e7', 1, 'e785d341-a003-4457-9a59-f01c9ef22159', 0);

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
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `mobile` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uniq_username`(`username`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统用户表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of t_sys_user
-- ----------------------------
INSERT INTO `t_sys_user` VALUES (1, '2020-11-24 15:49:07', '2022-05-13 14:44:02', 'admin', '593d4632a8c70251d0e9be4b1799bcc1', '54099a65-a235-158c-d610-74d2ff4c789b', 0, '王小二', 'https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png', '1', 'pengweihuang@we.cn', '18650036719');

SET FOREIGN_KEY_CHECKS = 1;
