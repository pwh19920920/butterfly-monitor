-- 分组聚合采集支持：t_monitor_task 增加 data_type 列
-- 1=单值采集(默认) 2=分组聚合采集

ALTER TABLE `t_monitor_task`
  ADD COLUMN `data_type` int(4) NOT NULL DEFAULT '1' COMMENT '1=单值采集 2=分组聚合采集'
  AFTER `labels`;

CREATE INDEX `idx_data_type` ON `t_monitor_task` (`data_type`);

-- 关联任务：把其它任务的实时/样本曲线叠加到本任务面板（逗号分隔的任务 ID）
ALTER TABLE `t_monitor_task`
  ADD COLUMN `related_task_ids` varchar(1024) NOT NULL DEFAULT '' COMMENT '关联任务ID，逗号分隔'
  AFTER `data_type`;
