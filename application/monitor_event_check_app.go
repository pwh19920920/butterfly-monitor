package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type MonitorEventCheckApplication struct {
	sequence     *snowflake.Node
	repository   *persistence.Repository
	xxlExec      xxl.Executor
	alertConf    AlertConfApplication
	alertChannel AlertChannelApplication
}

type MonitorEventTemplateParam struct {
	TaskName   string            `json:"taskName"`   // 任务名
	HitRule    string            `json:"hitRule"`    // 命中规则
	HappenTime *common.LocalTime `json:"happenTime"` // 发生事件
	EventId    int64             `json:"eventId"`    // 时间id
}

func (app *MonitorEventCheckApplication) eventCheck(cxt context.Context, param *xxl.RunReq) (msg string) {
	// 获取任务分片数据
	events, err := app.repository.MonitorTaskEventRepository.FindEventJob()
	if err != nil {
		logrus.Error("从数据库获取报警事件失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取报警事件失败")
	}

	if events == nil || len(events) == 0 {
		logrus.Info("报警事件为空")
		return fmt.Sprintf("exec complete, 报警事件为空")
	}

	alertConfInstance, err := app.alertConf.SelectConf()
	if err != nil {
		logrus.Error("从数据库获取报警配置失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取报警配置失败")
	}

	// 消息聚合, 将相同分组的消息聚合为一笔
	alertIds := make([]int64, 0)
	taskIds := make([]int64, 0)
	alertIdForEventMap := make(map[int64]entity.MonitorTaskEvent, 0)
	for _, event := range events {
		alertIds = append(alertIds, event.AlertId)
		taskIds = append(taskIds, event.TaskId)
		alertIdForEventMap[event.AlertId] = event
	}

	// 查询task_alert表
	alerts, alertsErr := app.repository.MonitorTaskAlertRepository.BatchGetByIds(alertIds)
	tasks, tasksErr := app.repository.MonitorTaskRepository.SelectByIds(taskIds)
	if alertsErr != nil || tasksErr != nil {
		logrus.Error("从数据库获取批量获取报警检查任务失败")
		return "exec fail, 从数据库获取任务失败"
	}

	if len(alertIds) != len(taskIds) {
		logrus.Info("taskAlert长度与task长度不匹配")
		return fmt.Sprintf("exec fail, taskAlert长度与task长度不匹配")
	}

	// 将任务转为map
	taskIdForTaskMap := make(map[int64]entity.MonitorTask, 0)
	for _, task := range tasks {
		taskIdForTaskMap[task.Id] = task
	}

	// 分组下的任务聚合, 筛选出每个分组下拥有的taskAlertIds
	groupForChannelForParamsMap := app.BuildGroupForChannelForParamsMap(alerts, taskIdForTaskMap, alertIdForEventMap)

	// 分组发送
	successEventIds := make([]int64, 0)
	for group, channelForParamsMap := range groupForChannelForParamsMap {
		for channel, params := range channelForParamsMap {
			groupId, _ := strconv.ParseInt(group, 10, 64)
			channelId, _ := strconv.ParseInt(channel, 10, 64)
			text, err := app.RenderTemplate(params, alertConfInstance.Alert.Template)
			if err != nil {
				logrus.Error("模板渲染失败")
				continue
			}

			err = app.DispatchMessage(groupId, text, channelId)
			if err == nil {
				for _, param := range params {
					successEventIds = append(successEventIds, param.EventId)
				}
			}
		}
	}

	if len(successEventIds) == 0 {
		return "execute complete, success count = 0"
	}

	currentTime := time.Now()
	duration, _ := time.ParseDuration(fmt.Sprintf("%vs", alertConfInstance.Alert.AlertSpan))
	nextTime := currentTime.Add(duration)

	// 批量更新下次日期, 本次报警发送日期
	err = app.repository.MonitorTaskEventRepository.BatchModifyByEvents(successEventIds, &entity.MonitorTaskEvent{
		PreAlertTime:  &common.LocalTime{Time: currentTime},
		NextAlertTime: &common.LocalTime{Time: nextTime},
	})

	if err != nil {
		println(err.Error())
	}
	return "execute complete"
}

func (app *MonitorEventCheckApplication) BuildGroupForChannelForParamsMap(alerts []entity.MonitorTaskAlert, taskIdForTaskMap map[int64]entity.MonitorTask, alertIdForEventMap map[int64]entity.MonitorTaskEvent) map[string]map[string][]MonitorEventTemplateParam {
	groupForChannelForParamsMap := make(map[string]map[string][]MonitorEventTemplateParam, 0)
	for _, alert := range alerts {
		groups := strings.Split(alert.AlertGroups, ",")
		for _, group := range groups {
			channelForParamsMap, ok := groupForChannelForParamsMap[group]
			if !ok {
				// 不存在则创建
				channelForParamsMap = make(map[string][]MonitorEventTemplateParam, 0)
				groupForChannelForParamsMap[group] = channelForParamsMap
			}

			channels := strings.Split(alert.AlertChannels, ",")
			for _, channel := range channels {
				params, ok := channelForParamsMap[channel]
				if !ok {
					// 不存在则创建
					params = make([]MonitorEventTemplateParam, 0)
					groupForChannelForParamsMap[group][channel] = params
				}

				params = append(params, MonitorEventTemplateParam{
					TaskName:   taskIdForTaskMap[alert.TaskId].TaskName,
					HitRule:    alertIdForEventMap[alert.Id].AlertMsg,
					HappenTime: alertIdForEventMap[alert.Id].CreatedAt,
					EventId:    alertIdForEventMap[alert.Id].Id,
				})

				groupForChannelForParamsMap[group][channel] = params
			}
		}
	}
	return groupForChannelForParamsMap
}

func (app *MonitorEventCheckApplication) DispatchMessage(groupId int64, text string, channelId int64) error {
	channel, err := app.repository.AlertChannelRepository.GetById(channelId)
	if err != nil {
		return errors.New("数据库获取通道失败")
	}

	alertChannelHandler, ok := alertChannelHandlerNameMap[channel.Handler]
	if !ok {
		return errors.New("处理器不存在")
	}

	groupUserIds, err := app.repository.AlertGroupUserRepository.SelectByGroupId(groupId)
	if err != nil {
		return errors.New("数据库获取GroupUser失败")
	}

	groupUsers, err := app.repository.AlertGroupUserRepository.SelectUsersByUserIds(groupUserIds)
	if err != nil {
		return errors.New("数据库获取用户信息失败")
	}

	if groupUsers == nil || len(groupUsers) == 0 {
		logrus.Infof("分组下的可用用户为空, groupId: %v", groupId)
		return nil
	}
	return alertChannelHandler.DispatchMessage(channel, groupUsers, text)
}

func (app *MonitorEventCheckApplication) RenderTemplate(paramsArr []MonitorEventTemplateParam, templateStr string) (string, error) {
	params := make(map[string]interface{}, 0)
	params["items"] = paramsArr

	// 创建模板对象, parse关联模板
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", err
	}

	// 渲染动态数据
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RegisterExecJob 注册执行
func (app *MonitorEventCheckApplication) RegisterExecJob() {
	app.xxlExec.RegTask("eventCheck", app.eventCheck)
}
