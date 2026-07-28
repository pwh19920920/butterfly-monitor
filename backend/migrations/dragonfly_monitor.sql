/*
 JookDB DUMP

 Source Server         : local
 Source Server Type    : mysql
 Source Server Version : 5.7.44
 Source Host           : localhost:3306
 Source DBName         : dragonfly_monitor

 File Encoding         : UTF-8

 Date: 2026-07-27 20:02:45
*/

SET NAMES utf8mb4;
SET UNIQUE_CHECKS = 0;
SET FOREIGN_KEY_CHECKS = 0;


CREATE DATABASE `dragonfly_monitor` CHARACTER SET utf8mb4;
USE `dragonfly_monitor`;

-- ----------------------------
-- Table structure for t_alert_channel
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_channel`;
CREATE TABLE `t_alert_channel` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL COMMENT '渠道名称',
  `type` int(10) NOT NULL DEFAULT '1' COMMENT '1邮件 2webhook 3短信',
  `params` varchar(2000) NOT NULL DEFAULT '' COMMENT '渠道参数JSON',
  `handler` varchar(64) NOT NULL COMMENT '处理器类名key',
  `fail_route` int(10) NOT NULL DEFAULT '1' COMMENT '失败路由 1否 2是',
  `template` text NOT NULL COMMENT '通道告警模板，空则用 alertConf 按 handler 的默认模板',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警通道';


-- ----------------------------
-- Table structure for t_alert_conf
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_conf`;
CREATE TABLE `t_alert_conf` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `conf_key` varchar(100) NOT NULL,
  `conf_val` text NOT NULL,
  `conf_desc` varchar(255) NOT NULL DEFAULT '',
  `conf_type` int(4) NOT NULL COMMENT '1数字 2字符串',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_conf_key` (`conf_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警配置';

INSERT INTO `t_alert_conf` VALUES (1462795963172130816,'2026-07-25 19:08:10','2026-07-26 13:31:52',0,'alertSpan','300','报警间隔',1);
INSERT INTO `t_alert_conf` VALUES (1462797872809381888,'2026-07-25 19:08:10','2026-07-27 18:59:14',0,'template.ChannelWechatHandler','# 业务监控平台系统预警\n{{- range .items}}\n<font color=\"info\">{{.TaskName}}：<\/font><font color=\"comment\">{{.HitRule}}<\/font>\n{{- end}}\n\n{{ if .relationTaskNames }}\n <font color=\"warning\">关联告警：<\/font>\n{{- end}}\n{{- range .relationTaskNames}}\n    <font color=\"warning\">{{.}};<\/font>\n{{- end}}','企微(ChannelWechatHandler)默认告警模板',2);
INSERT INTO `t_alert_conf` VALUES (1462797872809381889,'2026-07-25 19:08:10','2026-07-25 19:08:10',0,'firstDelay','60','首次报警延迟(秒)',1);
INSERT INTO `t_alert_conf` VALUES (1462797872809381890,'2026-07-26 15:30:49','2026-07-27 18:59:01',0,'template.ChannelEmailHandler','<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>业务监控平台系统预警<\/title>\n    <style>\n        body {\n            margin: 0;\n            padding: 0;\n            background-color: #f5f5f5;\n            font-family: \"Microsoft YaHei\", \"微软雅黑\", Arial, sans-serif;\n            color: #333;\n        }\n        .container {\n            max-width: 600px;\n            margin: 0px auto;\n            background-color: #ffffff;\n            border-radius: 8px;\n            box-shadow: 0 4px 12px rgba(0,0,0,0.08);\n            overflow: hidden;\n        }\n        .header {\n            background-color: #2c3e50;\n            color: #ffffff;\n            padding: 20px;\n            text-align: center;\n            font-size: 24px;\n            font-weight: bold;\n            letter-spacing: 1px;\n        }\n        .content {\n            padding: 30px;\n        }\n        .alert-title {\n            font-size: 18px;\n            font-weight: bold;\n            margin-bottom: 20px;\n            color: #e74c3c;\n        }\n        .alert-item {\n            background-color: #f8f9fa;\n            border-left: 5px solid #e74c3c;\n            padding: 15px;\n            margin-bottom: 15px;\n            border-radius: 4px;\n            font-size: 16px;\n            line-height: 1.6;\n        }\n        .footer {\n            text-align: center;\n            padding: 20px;\n            font-size: 14px;\n            color: #7f8c8d;\n            border-top: 1px solid #ecf0f1;\n            background-color: #fafafa;\n        }\n        .highlight {\n            color: #e67e22;\n            font-weight: bold;\n        }\n    <\/style>\n<\/head>\n<body>\n    <div class=\"container\">\n        <div class=\"header\">\n            业务监控平台系统预警\n        <\/div>\n        <div class=\"content\">\n            <div class=\"alert-title\">🚨 监控告警通知<\/div>\n<div class=\"alert-item\">\n{{- range .items}}\n<div> <span class=\"highlight\">{{.TaskName}}：<\/span>{{.HitRule}}<\/div>\n{{- end}}\n\n{{ if .relationTaskNames }}\n<\/br>\n <span class=\"highlight\">关联告警：<\/span>\n{{- end}}\n{{- range .relationTaskNames}}\n    <div class=\"highlight\">{{.}};<\/div>\n{{- end}}\n<\/div>\n            \n        <div class=\"footer\">\n            © 2026 业务监控平台\n        <\/div>\n    <\/div>\n<\/body>\n<\/html>','邮件(ChannelEmailHandler)默认告警模板',2);
INSERT INTO `t_alert_conf` VALUES (1462797872809381891,'2026-07-25 19:08:10','2026-07-25 19:08:10',0,'simplePageSize','50','样本生成每页任务数',1);
INSERT INTO `t_alert_conf` VALUES (1462797872809381892,'2026-07-25 19:08:10','2026-07-25 19:08:10',0,'simpleMaxSecond','600','样本生成最大回溯秒数',1);

-- ----------------------------
-- Table structure for t_alert_group
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_group`;
CREATE TABLE `t_alert_group` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警组';

INSERT INTO `t_alert_group` VALUES (2080990458799984641,'2026-07-25 20:16:03','2026-07-25 20:16:03',0,' 测试组');

-- ----------------------------
-- Table structure for t_alert_group_user
-- ----------------------------
DROP TABLE IF EXISTS `t_alert_group_user`;
CREATE TABLE `t_alert_group_user` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `user_id` bigint(20) NOT NULL,
  `group_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group` (`group_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警组用户';

INSERT INTO `t_alert_group_user` VALUES (2080990458799984642,'2026-07-25 20:16:03','2026-07-25 20:16:03',0,1,2080990458799984641);

-- ----------------------------
-- Table structure for t_monitor_dashboard
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_dashboard`;
CREATE TABLE `t_monitor_dashboard` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `slug` varchar(64) NOT NULL DEFAULT '',
  `uid` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(64) NOT NULL,
  `url` varchar(255) NOT NULL DEFAULT '',
  `board_id` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控面板';


-- ----------------------------
-- Table structure for t_monitor_dashboard_task
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_dashboard_task`;
CREATE TABLE `t_monitor_dashboard_task` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `task_id` bigint(20) NOT NULL,
  `dashboard_id` bigint(20) NOT NULL,
  `sort` int(10) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_task` (`task_id`),
  KEY `idx_dashboard` (`dashboard_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='面板-任务关联';

-- ----------------------------
-- Table structure for t_monitor_database
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_database`;
CREATE TABLE `t_monitor_database` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `database` varchar(255) NOT NULL DEFAULT '',
  `name` varchar(255) NOT NULL,
  `username` varchar(255) NOT NULL DEFAULT '',
  `password` varchar(255) NOT NULL DEFAULT '',
  `salt` varchar(64) NOT NULL DEFAULT '',
  `url` varchar(255) NOT NULL,
  `type` tinyint(10) NOT NULL COMMENT '0=MysqlOld 1=Mongo 2=Mysql',
  `params` varchar(1000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控数据源';


-- ----------------------------
-- Table structure for t_monitor_group
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_group`;
CREATE TABLE `t_monitor_group` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL DEFAULT '',
  `route` varchar(512) NOT NULL DEFAULT '',
  `parent` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_parent` (`parent`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控依赖分组';


-- ----------------------------
-- Table structure for t_monitor_task
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task`;
CREATE TABLE `t_monitor_task` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `pre_execute_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `pre_sample_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `task_key` varchar(255) NOT NULL,
  `task_name` varchar(255) NOT NULL,
  `step_span` int(10) NOT NULL DEFAULT '10',
  `time_span` int(10) NOT NULL,
  `command` varchar(2000) NOT NULL,
  `task_type` int(4) NOT NULL COMMENT '1=Database 2=URL 3=Push',
  `exec_params` varchar(2000) NOT NULL DEFAULT '',
  `task_status` tinyint(2) NOT NULL DEFAULT '1',
  `alert_status` tinyint(2) NOT NULL DEFAULT '1',
  `collect_err_msg` varchar(255) NOT NULL DEFAULT '',
  `sample_err_msg` varchar(255) NOT NULL DEFAULT '',
  `sampled` tinyint(2) NOT NULL DEFAULT '1' COMMENT '样本展示开关(仅Grafana)',
  `monitor_group` varchar(255) NOT NULL DEFAULT '',
  `labels` varchar(2000) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_task_key` (`task_key`),
  KEY `idx_task_status` (`task_status`),
  KEY `idx_pre_execute_time` (`pre_execute_time`),
  KEY `idx_pre_sample_time` (`pre_sample_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控任务';


-- ----------------------------
-- Table structure for t_monitor_task_alert
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task_alert`;
CREATE TABLE `t_monitor_task_alert` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `task_id` bigint(20) NOT NULL,
  `alert_channels` varchar(200) NOT NULL DEFAULT '',
  `alert_groups` varchar(200) NOT NULL DEFAULT '',
  `time_span` int(10) NOT NULL DEFAULT '30',
  `duration` int(10) NOT NULL DEFAULT '0',
  `params` varchar(2000) NOT NULL DEFAULT '',
  `alert_status` int(4) NOT NULL DEFAULT '1' COMMENT '1正常 2Pending 3Firing',
  `deal_status` int(4) NOT NULL DEFAULT '1' COMMENT '1正常 2处理中',
  `pre_check_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `first_flag_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_task_alert` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警规则';


-- ----------------------------
-- Table structure for t_monitor_task_event
-- ----------------------------
DROP TABLE IF EXISTS `t_monitor_task_event`;
CREATE TABLE `t_monitor_task_event` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `alert_id` bigint(20) NOT NULL,
  `task_id` bigint(20) NOT NULL,
  `alert_msg` varchar(2000) NOT NULL DEFAULT '',
  `deal_time` datetime DEFAULT NULL,
  `complete_time` datetime DEFAULT NULL,
  `content` text NOT NULL,
  `deal_status` int(4) NOT NULL DEFAULT '1' COMMENT '1待处理 2处理中 3已完成 4已忽略',
  `deal_user` bigint(20) DEFAULT NULL,
  `pre_alert_time` datetime DEFAULT NULL,
  `next_alert_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `event_level` int(4) NOT NULL DEFAULT '0',
  `alert_count` int(10) NOT NULL DEFAULT '0' COMMENT '已成功告警次数',
  PRIMARY KEY (`id`),
  KEY `idx_alert_id` (`alert_id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_next_alert` (`deal_status`,`next_alert_time`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报警事件';


-- ----------------------------
-- Table structure for t_sys_menu
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_menu`;
CREATE TABLE `t_sys_menu` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(255) NOT NULL,
  `path` varchar(255) NOT NULL,
  `icon` varchar(255) NOT NULL DEFAULT '',
  `component` varchar(255) NOT NULL DEFAULT '',
  `sort` int(10) NOT NULL DEFAULT '0',
  `option` varchar(1024) NOT NULL DEFAULT '',
  `parent` bigint(20) NOT NULL DEFAULT '0',
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `route` varchar(255) NOT NULL DEFAULT '',
  `code` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统菜单';

INSERT INTO `t_sys_menu` VALUES (1332302770434215918,'2026-07-25 19:08:12','2026-07-25 19:08:12','系统管理','\/sys','crown','',200,'[]',0,0,'\/1332302770434215918','sys');
INSERT INTO `t_sys_menu` VALUES (1332302770434215920,'2026-07-25 19:08:12','2026-07-25 19:08:12','菜单管理','\/sys\/sysMenu','smile','.\/SysMenu',1,'[]',1332302770434215918,0,'\/1332302770434215918\/1332302770434215920','sysMenu');
INSERT INTO `t_sys_menu` VALUES (1332302770434215922,'2026-07-25 19:08:12','2026-07-25 19:08:12','用户管理','\/sys\/sysUser','smile','.\/SysUser',2,'[]',1332302770434215918,0,'\/1332302770434215918\/1332302770434215922','sysUser');
INSERT INTO `t_sys_menu` VALUES (1332302770434215924,'2026-07-25 19:08:12','2026-07-25 19:08:12','角色管理','\/sys\/sysRole','smile','.\/SysRole',1,'[]',1332302770434215918,0,'\/1332302770434215918\/1332302770434215924','sysRole');
INSERT INTO `t_sys_menu` VALUES (1332302770434215926,'2026-07-25 19:08:12','2026-07-25 19:08:12','监控管理','\/monitor','dashboard','',100,'[]',0,0,'\/1332302770434215926','monitor');
INSERT INTO `t_sys_menu` VALUES (1332302770434215928,'2026-07-25 19:08:12','2026-07-26 19:01:21','数据源管理','\/monitor\/database','database','.\/MonitorDatabase',2,'[]',1465652495161233408,0,'\/1465652495161233408\/1332302770434215928','monitorDatabase');
INSERT INTO `t_sys_menu` VALUES (1332302770434215930,'2026-07-25 19:08:12','2026-07-25 22:30:36','任务管理','\/monitor\/task','table','.\/MonitorTask',1,'[]',1332302770434215926,0,'\/1332302770434215926\/1332302770434215930','monitorTask');
INSERT INTO `t_sys_menu` VALUES (1452284009022230528,'2026-07-25 19:08:12','2026-07-25 19:08:12','面板管理','\/monitor\/dashboard','fund','.\/MonitorDashboard',3,'[]',1332302770434215926,0,'\/1332302770434215926\/1452284009022230528','monitorDashboard');
INSERT INTO `t_sys_menu` VALUES (1462709329521020928,'2026-07-25 19:08:12','2026-07-25 19:08:12','报警配置','\/alert\/alertConf','setting','.\/AlertConf',4,'[]',1465652495161233408,0,'\/1465652495161233408\/1462709329521020928','alertConf');
INSERT INTO `t_sys_menu` VALUES (1465165133809455104,'2026-07-25 19:08:12','2026-07-25 19:08:12','报警组管理','\/alert\/alertGroup','team','.\/AlertGroup',4,'[]',1465652495161233408,0,'\/1465652495161233408\/1465165133809455104','alertGroup');
INSERT INTO `t_sys_menu` VALUES (1465561401698291712,'2026-07-25 19:08:12','2026-07-25 19:08:12','报警通道','\/alert\/alertChannel','notification','.\/AlertChannel',5,'[]',1465652495161233408,0,'\/1465652495161233408\/1465561401698291712','alertChannel');
INSERT INTO `t_sys_menu` VALUES (1465652495161233408,'2026-07-25 19:08:12','2026-07-27 14:24:29','报警管理','\/alert','smile','',50,'[]',0,0,'\/1465652495161233408','alert');
INSERT INTO `t_sys_menu` VALUES (1472888326758338560,'2026-07-25 19:08:12','2026-07-27 19:40:43','异常事件','\/monitor\/taskEvent','alert','.\/MonitorTaskEvent',4,'[]',1332302770434215926,0,'\/1332302770434215926\/1472888326758338560','monitorTaskEvent');
INSERT INTO `t_sys_menu` VALUES (1500000000000000001,'2026-07-25 19:08:12','2026-07-25 19:08:12','监控分组','\/monitor\/group','apartment','.\/MonitorGroup',5,'[]',1332302770434215926,0,'\/1332302770434215926\/1500000000000000001','monitorGroup');

-- ----------------------------
-- Table structure for t_sys_menu_option
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_menu_option`;
CREATE TABLE `t_sys_menu_option` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL DEFAULT '',
  `value` varchar(255) NOT NULL DEFAULT '',
  `method` varchar(10) NOT NULL DEFAULT '',
  `path` varchar(255) NOT NULL DEFAULT '',
  `code` varchar(64) NOT NULL DEFAULT '',
  `menu_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unq_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统菜单操作';

INSERT INTO `t_sys_menu_option` VALUES (1447759564626726912,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'菜单查看','sys:menu:query','GET','\/api\/sys\/menu','caa126a343b0e1cef0774b637c246af3',1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1448238719118086145,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'菜单新增','sys:menu:create','POST','\/api\/sys\/menu','79102b6efd1174afdf1732d9e7e80629',1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389440,'2026-07-25 19:08:12','2026-07-26 19:01:20',0,'数据源查询','monitor:database:query','GET','\/api\/monitor\/database','e9fdf345326fce9103dffe4b62c648f3',1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389441,'2026-07-25 19:08:12','2026-07-26 19:01:20',0,'数据源创建','monitor:database:create','POST','\/api\/monitor\/database','534eab021f3a2451281fff1d1767a0cc',1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449574882915389442,'2026-07-25 19:08:12','2026-07-26 19:01:20',0,'数据源更新','monitor:database:modify','PUT','\/api\/monitor\/database','b4708d30ae76818d5d3f7ea355e51f65',1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380992,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'任务查询','monitor:task:query','GET','\/api\/monitor\/task','b122c53237e751115ce1ecc913ec6865',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380993,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'任务更新','monitor:task:modify','PUT','\/api\/monitor\/task','a9837ad678785aaf1a5f8d806a0304bb',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449718480839380994,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'任务创建','monitor:task:create','POST','\/api\/monitor\/task','1284964e9851ff3d3393c804c76100df',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1449747431762694147,'2026-07-25 19:08:12','2026-07-26 19:01:20',0,'全部数据源','monitor:database:queryAll','GET','\/api\/monitor\/database\/all','92690787f68af00a309627e5bdadf55f',1332302770434215928);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605955,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'任务状态修改','monitor:task:modifyTaskStatus','PUT','\/api\/monitor\/task\/taskStatus\/:id\/:status','540164c1917f24d11d7359973d6d67e0',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1452577297956605956,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'报警状态修改','monitor:task:modifyAlertStatus','PUT','\/api\/monitor\/task\/alertStatus\/:id\/:status','f76e02a14e173e00e633f43797e321e5',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999616,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'面板查询','monitor:dashboard:query','GET','\/api\/monitor\/dashboard','510ac77819b00cd71805f27509d7eb6e',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999617,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'面板创建','monitor:dashboard:create','POST','\/api\/monitor\/dashboard','d15f34654e3b6541a097230f41650071',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999618,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'面板更新','monitor:dashboard:modify','PUT','\/api\/monitor\/dashboard','acd15a7fc886a223f8afb15ee779587e',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453201790357999619,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'面板全部','monitor:dashboard:queryAll','GET','\/api\/monitor\/dashboard\/all','32914b427951e2a2e88d2a35a5c5891f',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453286781452554244,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'面板任务','monitor:dashboard:task','GET','\/api\/monitor\/dashboard\/task\/:id','0095c870578fdf68dbae73e02c85e95e',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1453349603091943429,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'任务排序','monitor:dashboard:sort','PUT','\/api\/monitor\/dashboard\/taskSort','ef422a4fd11cdbb1f416c72567da10a6',1452284009022230528);
INSERT INTO `t_sys_menu_option` VALUES (1462721145227710464,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'读取配置','alert:conf:query','GET','\/api\/alert\/conf','8457f70f619ae309571253e7f81213ef',1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1462721145227710465,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'修改配置','alert:conf:modify','PUT','\/api\/alert\/conf','052a70165846e9a96c726d82a71d683e',1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1463448357279109125,'2026-07-25 19:08:12','2026-07-25 22:30:35',0,'样本展示修改','monitor:task:modifySampled','PUT','\/api\/monitor\/task\/sampled\/:id\/:status','517025152bd9ec4062976bc681923418',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637184,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'用户查询','sys:user:query','GET','\/api\/sys\/user','0924d00bac6e4d1b9e10040e095a980f',1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637185,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'用户修改','sys:user:modify','PUT','\/api\/sys\/user','e75e13959a8f6577fee78e1bc61d3e10',1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637186,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'用户创建','sys:user:create','POST','\/api\/sys\/user','80b0eae883b924868c44df5295b8ee33',1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1463883940220637187,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'全部用户','sys:user:queryAll','GET','\/api\/sys\/user\/all','2b202b723b15b6b2db71b9cf0079c904',1332302770434215922);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843712,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警组查询','alert:group:query','GET','\/api\/alert\/group','3d3511b26417d92fdb9a6c705e769af8',1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843713,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警组创建','alert:group:create','POST','\/api\/alert\/group','2b7c5300c8efa0671a0b584e1a76986a',1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843714,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警组修改','alert:group:modify','PUT','\/api\/alert\/group','27b37f08e60468864f6d2c74d583daf2',1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843715,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警组查全部','alert:group:queryall','GET','\/api\/alert\/group\/all','84138424922938bf2c6fda3b29557e96',1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465165133817843716,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警组下用户','alert:group:users','GET','\/api\/alert\/group\/groupUser\/:id','b8b09ecfedbd9bf2282a5400131b8eb6',1465165133809455104);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068928,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警通道查询','alert:channel:query','GET','\/api\/alert\/channel','a9fb2fa8c82717cb47209781b0b017db',1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068929,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警通道创建','alert:channel:save','POST','\/api\/alert\/channel','2428c926e4bdaa278532ff34c7797d6f',1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068930,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警通道修改','alert:channel:update','PUT','\/api\/alert\/channel','aea7ff8fd9ad8c8fb939f4b560e9138a',1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465561401715068931,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警通道处理器','alert:channel:handlers','GET','\/api\/alert\/channel\/handlers','cfa624de6151bd2f614f24ad88417009',1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1465652723025186818,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'创建配置','alert:conf:create','POST','\/api\/alert\/conf','1e5446c29206d4da16838b45a3601e43',1462709329521020928);
INSERT INTO `t_sys_menu_option` VALUES (1467698655321395204,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'报警通道全部','alert:channel:queryAll','GET','\/api\/alert\/channel\/all','90f3cf9c6ad1d18013b95397719fabd4',1465561401698291712);
INSERT INTO `t_sys_menu_option` VALUES (1472889239543746560,'2026-07-25 19:08:12','2026-07-27 19:40:43',0,'事件查询','monitor:taskEvent:query','GET','\/api\/monitor\/task\/event','1ebaf6a664546061edeedeabcedf2f1f',1472888326758338560);
INSERT INTO `t_sys_menu_option` VALUES (1473138104612163585,'2026-07-25 19:08:12','2026-07-27 19:40:43',0,'事件处理','monitor:taskEvent:deal','POST','\/api\/monitor\/task\/event\/deal\/:id','c0b8e609bfd3af3f287b5916d4f543ed',1472888326758338560);
INSERT INTO `t_sys_menu_option` VALUES (1473138104612163586,'2026-07-25 19:08:12','2026-07-27 19:40:43',0,'事件完成','monitor:taskEvent:complete','POST','\/api\/monitor\/task\/event\/complete\/:id','336c174b06c39a1b49640c8e2175b8c6',1472888326758338560);
INSERT INTO `t_sys_menu_option` VALUES (1484094198515765282,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'角色查询','sys:role:query','GET','\/api\/sys\/role','92cd13408fbd6512a4e5328c800d5439',1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1484094198515765283,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'角色创建','sys:role:create','POST','\/api\/sys\/role','0f1f61330f3ef4b03bf5632bbcc5737f',1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1484094198515765284,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'角色修改','sys:role:modify','PUT','\/api\/sys\/role','f83fd0b3bd1902b67676b1a64a78b309',1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1484094198515765285,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'角色删除','sys:role:delete','DELETE','\/api\/sys\/role\/:id','0f14a64e7dff47937164f64dfa01dbbf',1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1484094198515765286,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'角色权限查询','sys:role:queryPermission','GET','\/api\/sys\/role\/permission\/:roleId','381b1a498606e82226a0604d8c853e65',1332302770434215924);
INSERT INTO `t_sys_menu_option` VALUES (1484097160667467832,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'菜单修改','sys:menu:modify','PUT','\/api\/sys\/menu','6a7b949b19c27a1f9ee3753e06c3ecf5',1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1484097160667467833,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'菜单删除','sys:menu:delete','DELETE','\/api\/sys\/menu\/:id','c249795688bf6e62fe7b16ba1d539540',1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1484097160667467834,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'菜单操作','sys:menu:option','GET','\/api\/sys\/menu\/option\/:id','dfb98e82d0666d936314879ab3cbe37d',1332302770434215920);
INSERT INTO `t_sys_menu_option` VALUES (1500000000000000010,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'分组查询','monitor:group:query','GET','\/api\/monitor\/group','mgroupquery000000000000000000001',1500000000000000001);
INSERT INTO `t_sys_menu_option` VALUES (1500000000000000011,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'分组创建','monitor:group:create','POST','\/api\/monitor\/group','mgroupcreate00000000000000000001',1500000000000000001);
INSERT INTO `t_sys_menu_option` VALUES (1500000000000000012,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'分组修改','monitor:group:modify','PUT','\/api\/monitor\/group','mgroupmodify00000000000000000001',1500000000000000001);
INSERT INTO `t_sys_menu_option` VALUES (1500000000000000013,'2026-07-25 19:08:12','2026-07-25 19:08:12',0,'分组全部','monitor:group:queryAll','GET','\/api\/monitor\/group\/all','mgroupall00000000000000000000001',1500000000000000001);
INSERT INTO `t_sys_menu_option` VALUES (2081024319395205127,'2026-07-25 22:30:36','2026-07-25 22:30:36',0,'任务详情','monitor:task:detail','GET','\/api\/monitor\/task\/:id','9ad5a3972994d0228d30609a0f0217d1',1332302770434215930);
INSERT INTO `t_sys_menu_option` VALUES (2081706344011797211,'2026-07-27 19:40:43','2026-07-27 19:40:43',0,'事件忽略','monitor:taskEvent:ignore','POST','\/api\/monitor\/task\/event\/ignore\/:id','5ce0faddf6911ef61beb83448b0a995d',1472888326758338560);

-- ----------------------------
-- Table structure for t_sys_permission
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_permission`;
CREATE TABLE `t_sys_permission` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `menu_id` bigint(20) NOT NULL,
  `role_id` bigint(20) NOT NULL,
  `option` varchar(1024) NOT NULL DEFAULT '',
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `independent` tinyint(2) NOT NULL DEFAULT '0',
  `half` tinyint(2) NOT NULL DEFAULT '0',
  `root` tinyint(2) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统权限';

INSERT INTO `t_sys_permission` VALUES (1473138135784230912,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215922,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230913,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215920,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230914,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215924,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230915,'2026-07-25 19:08:12','2026-07-27 19:40:53',1452284009022230528,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230916,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215928,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230917,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215930,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230918,'2026-07-25 19:08:12','2026-07-27 19:40:53',1465561401698291712,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230919,'2026-07-25 19:08:12','2026-07-27 19:40:53',1465165133809455104,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230920,'2026-07-25 19:08:12','2026-07-27 19:40:53',1462709329521020928,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230921,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215918,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230922,'2026-07-25 19:08:12','2026-07-27 19:40:53',1465652495161233408,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230923,'2026-07-25 19:08:12','2026-07-27 19:40:53',1472888326758338560,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230924,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215926,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (1473138135784230925,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215920,1,'1447759564626726912,1448238719118086145,1484097160667467832,1484097160667467833,1484097160667467834',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230926,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215924,1,'1484094198515765282,1484094198515765283,1484094198515765284,1484094198515765285,1484094198515765286',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230927,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215928,1,'1449574882915389440,1449574882915389441,1449574882915389442,1449747431762694147',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230928,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215930,1,'1449718480839380992,1449718480839380993,1449718480839380994,1452577297956605955,1452577297956605956,1463448357279109125',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230929,'2026-07-25 19:08:12','2026-07-27 19:40:53',1452284009022230528,1,'1453201790357999616,1453201790357999617,1453201790357999618,1453201790357999619,1453286781452554244,1453349603091943429',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230930,'2026-07-25 19:08:12','2026-07-27 19:40:53',1462709329521020928,1,'1462721145227710464,1462721145227710465,1465652723025186818',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230931,'2026-07-25 19:08:12','2026-07-27 19:40:53',1332302770434215922,1,'1463883940220637184,1463883940220637185,1463883940220637186,1463883940220637187',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230932,'2026-07-25 19:08:12','2026-07-27 19:40:53',1465165133809455104,1,'1465165133817843712,1465165133817843713,1465165133817843714,1465165133817843715,1465165133817843716',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230933,'2026-07-25 19:08:12','2026-07-27 19:40:53',1465561401698291712,1,'1465561401715068928,1465561401715068929,1465561401715068930,1465561401715068931,1467698655321395204',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1473138135784230934,'2026-07-25 19:08:12','2026-07-27 19:40:53',1472888326758338560,1,'1472889239543746560,1473138104612163585,1473138104612163586',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (1500000000000000020,'2026-07-25 19:08:12','2026-07-27 19:40:53',1500000000000000001,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (1500000000000000021,'2026-07-25 19:08:12','2026-07-27 19:40:53',1500000000000000001,1,'1500000000000000010,1500000000000000011,1500000000000000012,1500000000000000013',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714120,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215922,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714121,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215920,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714122,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215924,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714123,'2026-07-25 22:30:46','2026-07-27 19:40:53',1500000000000000001,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714124,'2026-07-25 22:30:46','2026-07-27 19:40:53',1472888326758338560,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714125,'2026-07-25 22:30:46','2026-07-27 19:40:53',1452284009022230528,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714126,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215928,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714127,'2026-07-25 22:30:46','2026-07-27 19:40:53',1465561401698291712,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714128,'2026-07-25 22:30:46','2026-07-27 19:40:53',1465165133809455104,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714129,'2026-07-25 22:30:46','2026-07-27 19:40:53',1462709329521020928,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714130,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215918,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081024361434714131,'2026-07-25 22:30:46','2026-07-27 19:40:53',1465652495161233408,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081024361434714132,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215930,1,'',1,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714133,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215926,1,'',1,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081024361434714134,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215920,1,'1447759564626726912,1448238719118086145,1484097160667467832,1484097160667467833,1484097160667467834',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714135,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215924,1,'1484094198515765282,1484094198515765283,1484094198515765284,1484094198515765285,1484094198515765286',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714136,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215928,1,'1449574882915389440,1449574882915389441,1449574882915389442,1449747431762694147',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714137,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215930,1,'1449718480839380992,1449718480839380993,1449718480839380994,1452577297956605955,1452577297956605956,1463448357279109125,2081024319395205127',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714138,'2026-07-25 22:30:46','2026-07-27 19:40:53',1452284009022230528,1,'1453201790357999616,1453201790357999617,1453201790357999618,1453201790357999619,1453286781452554244,1453349603091943429',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714139,'2026-07-25 22:30:46','2026-07-27 19:40:53',1462709329521020928,1,'1462721145227710464,1462721145227710465,1465652723025186818',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714140,'2026-07-25 22:30:46','2026-07-27 19:40:53',1332302770434215922,1,'1463883940220637184,1463883940220637185,1463883940220637186,1463883940220637187',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714141,'2026-07-25 22:30:46','2026-07-27 19:40:53',1465165133809455104,1,'1465165133817843712,1465165133817843713,1465165133817843714,1465165133817843715,1465165133817843716',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714142,'2026-07-25 22:30:46','2026-07-27 19:40:53',1465561401698291712,1,'1465561401715068928,1465561401715068929,1465561401715068930,1465561401715068931,1467698655321395204',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714143,'2026-07-25 22:30:46','2026-07-27 19:40:53',1472888326758338560,1,'1472889239543746560,1473138104612163585,1473138104612163586',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081024361434714144,'2026-07-25 22:30:46','2026-07-27 19:40:53',1500000000000000001,1,'1500000000000000010,1500000000000000011,1500000000000000012,1500000000000000013',1,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210652,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215922,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210653,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215920,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210654,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215924,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210655,'2026-07-27 19:40:53','2026-07-27 19:40:53',1500000000000000001,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210656,'2026-07-27 19:40:53','2026-07-27 19:40:53',1452284009022230528,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210657,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215930,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210658,'2026-07-27 19:40:53','2026-07-27 19:40:53',1465561401698291712,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210659,'2026-07-27 19:40:53','2026-07-27 19:40:53',1465165133809455104,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210660,'2026-07-27 19:40:53','2026-07-27 19:40:53',1462709329521020928,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210661,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215928,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210662,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215918,1,'',0,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081706386416210663,'2026-07-27 19:40:53','2026-07-27 19:40:53',1465652495161233408,1,'',0,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081706386416210664,'2026-07-27 19:40:53','2026-07-27 19:40:53',1472888326758338560,1,'',0,1,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210665,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215926,1,'',0,1,0,1);
INSERT INTO `t_sys_permission` VALUES (2081706386416210666,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215920,1,'1447759564626726912,1448238719118086145,1484097160667467832,1484097160667467833,1484097160667467834',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210667,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215924,1,'1484094198515765282,1484094198515765283,1484094198515765284,1484094198515765285,1484094198515765286',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210668,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215928,1,'1449574882915389440,1449574882915389441,1449574882915389442,1449747431762694147',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210669,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215930,1,'1449718480839380992,1449718480839380993,1449718480839380994,1452577297956605955,1452577297956605956,1463448357279109125,2081024319395205127',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210670,'2026-07-27 19:40:53','2026-07-27 19:40:53',1452284009022230528,1,'1453201790357999616,1453201790357999617,1453201790357999618,1453201790357999619,1453286781452554244,1453349603091943429',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210671,'2026-07-27 19:40:53','2026-07-27 19:40:53',1462709329521020928,1,'1462721145227710464,1462721145227710465,1465652723025186818',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210672,'2026-07-27 19:40:53','2026-07-27 19:40:53',1332302770434215922,1,'1463883940220637184,1463883940220637185,1463883940220637186,1463883940220637187',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210673,'2026-07-27 19:40:53','2026-07-27 19:40:53',1465165133809455104,1,'1465165133817843712,1465165133817843713,1465165133817843714,1465165133817843715,1465165133817843716',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210674,'2026-07-27 19:40:53','2026-07-27 19:40:53',1465561401698291712,1,'1465561401715068928,1465561401715068929,1465561401715068930,1465561401715068931,1467698655321395204',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210675,'2026-07-27 19:40:53','2026-07-27 19:40:53',1472888326758338560,1,'1472889239543746560,1473138104612163585,1473138104612163586,2081706344011797211',0,0,0,0);
INSERT INTO `t_sys_permission` VALUES (2081706386416210676,'2026-07-27 19:40:53','2026-07-27 19:40:53',1500000000000000001,1,'1500000000000000010,1500000000000000011,1500000000000000012,1500000000000000013',0,0,0,0);

-- ----------------------------
-- Table structure for t_sys_role
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_role`;
CREATE TABLE `t_sys_role` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(255) NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色';

INSERT INTO `t_sys_role` VALUES (1,'2026-07-25 19:08:13','2026-07-27 19:40:53','super_admin',0);

-- ----------------------------
-- Table structure for t_sys_token
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_token`;
CREATE TABLE `t_sys_token` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `secret` varchar(255) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `expire_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COMMENT='系统令牌';


-- ----------------------------
-- Table structure for t_sys_user
-- ----------------------------
DROP TABLE IF EXISTS `t_sys_user`;
CREATE TABLE `t_sys_user` (
  `id` bigint(20) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `salt` varchar(255) NOT NULL,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `name` varchar(255) NOT NULL,
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `roles` varchar(255) NOT NULL DEFAULT '',
  `email` varchar(255) NOT NULL DEFAULT '',
  `mobile` varchar(11) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户';

INSERT INTO `t_sys_user` VALUES (1,'2026-07-25 19:08:13','2026-07-26 11:34:32','admin','593d4632a8c70251d0e9be4b1799bcc1','54099a65-a235-158c-d610-74d2ff4c789b',0,'管理员','https:\/\/gw.alipayobjects.com\/zos\/antfincdn\/XAosXuNZyF\/BiazfanxmamNRoxxVxka.png','1','173186915@qq.com','13800000000');

SET UNIQUE_CHECKS = 1;
SET FOREIGN_KEY_CHECKS = 1;
