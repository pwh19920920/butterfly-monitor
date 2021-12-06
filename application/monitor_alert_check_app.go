package application

import (
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/config/influxdb"
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"strings"
	"sync"
	"time"
)

type MonitorAlertCheckApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	influxdb   *influxdb.DbOption
	xxlExec    xxl.Executor
	grafana    *grafana.Config
	alertConf  AlertConfApplication
}

func (app *MonitorAlertCheckApplication) alertCheck(cxt context.Context, param *xxl.RunReq) (msg string) {
	// 获取任务分片数据
	checks, err := app.repository.MonitorTaskAlertRepository.FindCheckJob(param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		logrus.Error("从数据库获取报警检查任务失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取报警检查任务失败")
	}

	alertConfInstance, err := app.alertConf.SelectConf()
	if err != nil {
		logrus.Error("从数据库获取报警配置失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取报警配置失败")
	}

	// 循环执行command, 并行执行
	var wg sync.WaitGroup
	for _, check := range checks {
		wg.Add(1)
		go app.execCheck(alertConfInstance.Alert, check, &wg)
	}

	wg.Wait()
	return "execute complete"
}

//
func (app *MonitorAlertCheckApplication) execCheck(conf AlertConfObject, check entity.MonitorTaskAlert, wg *sync.WaitGroup) {
	// 执行标记
	defer wg.Done()

	// 判断任务状态
	task, err := app.repository.MonitorTaskRepository.GetById(check.TaskId)
	if err != nil || task == nil {
		logrus.Errorf("execCheck 获取任务[%v]失败", check.TaskId)
		return
	}

	if task.AlertStatus == entity.MonitorAlertStatusClose || task.TaskStatus == entity.MonitorTaskStatusClose {
		logrus.Infof("execCheck 任务[%v]未开启或者报警未开启", check.TaskId)
		return
	}

	// 检查规则
	if check.Params == "" {
		logrus.Infof("execCheck 任务[%v]未配置检查规则", check.TaskId)
		return
	}

	var params []entity.MonitorAlertCheckParams
	if err := json.Unmarshal([]byte(check.Params), &params); err != nil {
		logrus.Infof("execCheck 任务[%v]规则反序列化失败", check.TaskId)
		return
	}

	// 检查时间
	currentTime := time.Now()
	if check.EffectTime != nil && app.checkTimeRange(*check.EffectTime, currentTime) {
		logrus.Infof("execCheck 任务[%v]不在检测时间内, 自动忽略", check.TaskId)
		return
	}

	// 时间倒退
	duration, _ := time.ParseDuration(fmt.Sprintf("-%vs", task.TimeSpan))
	startTime := currentTime.Add(duration)

	// 检查influxdb是否正常
	cli := app.influxdb.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Error("influxdb ping 失败")
		return
	}

	// 执行结束关闭cli
	defer func(cli client.Client) {
		_ = cli.Close()
	}(cli)

	// 查询样本平均, 以及实时数据, 只要有一个不存在, 则忽略, 无法判定为错误
	sampleMeasurementName := fmt.Sprintf("\"%s.%s_sample\"", app.grafana.SampleRpName, task.TaskKey)
	sampleVal, err := app.getInfluxdbMeanVal(cli, sampleMeasurementName, startTime, currentTime)
	if err != nil || sampleVal == 0 {
		return
	}

	realVal, err := app.getInfluxdbMeanVal(cli, task.TaskKey, startTime, currentTime)
	if err != nil {
		return
	}

	// 规则校验 [[rule, rule], [], []], 是否被检查出来命中了检测规则
	checkResult, hitMsg := app.checkForParam(check, params, sampleVal, realVal)
	logrus.Infof("|(%v - %v)|/%v = %v, %s", sampleVal, realVal, sampleVal, checkResult, strings.Join(hitMsg, ";"))
	if !checkResult {
		// 没命中, 则要更新event为误报, 更新check任务的FirstFlagTime,PreCheckTime=current, AlertStatus=Normal
		_ = app.repository.MonitorTaskAlertRepository.ModifyForNormal(check.Id, currentTime)
		return
	}

	checkDuration, _ := time.ParseDuration(fmt.Sprintf("-%vs", check.Duration))
	beforeTime := currentTime.Add(checkDuration)

	// 超出5分钟得报警了
	if beforeTime.Unix() > check.FirstFlagTime.Unix() {
		duration, _ := time.ParseDuration(fmt.Sprintf("%vs", conf.FirstDelay))
		nextAlertTime := currentTime.Add(duration)

		// 更新AlertStatus状态为3达到报警条件, PreCheckTime=current
		// 查询event表是否存在此纪录, 不存在则插入
		_ = app.repository.MonitorTaskAlertRepository.ModifyByFiring(check.Id, currentTime, &entity.MonitorTaskEvent{
			BaseEntity:    common.BaseEntity{Id: app.sequence.Generate().Int64()},
			AlertId:       check.Id,
			TaskId:        check.TaskId,
			AlertMsg:      strings.Join(hitMsg, ";"),
			DealStatus:    entity.MonitorTaskEventDealStatusPending,
			NextAlertTime: &common.LocalTime{Time: nextAlertTime},
		})
		return
	}

	// 更新AlertStatus状态为2出现异常, PreCheckTime=current
	_ = app.repository.MonitorTaskAlertRepository.ModifyByPending(check.Id, currentTime)

}

// *** 使用此算法的前提是每个组里只有or或者and ***
func (app *MonitorAlertCheckApplication) checkForParam(check entity.MonitorTaskAlert, params []entity.MonitorAlertCheckParams, sampleVal, realVal int64) (bool, []string) {
	hitMsgList := make([]string, 0)
	result := false
	for _, param := range params {
		arrCheckResult, hitMsg := app.checkForParamArr(check, param.Params, sampleVal, realVal)
		hitMsgList = append(hitMsgList, hitMsg...)

		// 当表达式为or，其中一个为true, 就整个都是true
		if entity.MonitorAlertCheckParamsRelationOr == param.Relation && arrCheckResult {
			return true, hitMsg
		}

		// 当表达式为and, 其中一个为false，那整个都是false
		if entity.MonitorAlertCheckParamsRelationAnd == param.Relation && !arrCheckResult {
			return false, hitMsg
		}

		// 其他情况，为or，然后都是false，为and，都是true
		if entity.MonitorAlertCheckParamsRelationAnd == param.Relation && arrCheckResult {
			result = true
		}
	}
	return result, hitMsgList
}

// *** 使用此算法的前提是每个组里只有or或者and ***
func (app *MonitorAlertCheckApplication) checkForParamArr(check entity.MonitorTaskAlert, paramArr []entity.MonitorAlertCheckParamsItem, sampleVal, realVal int64) (bool, []string) {
	result := false
	hitMsg := make([]string, 0)
	for _, params := range paramArr {
		itemResult := app.checkForParamItem(params, sampleVal, realVal)
		if itemResult {
			hitMsg = append(hitMsg, fmt.Sprintf("样本值: %v, 当前值: %v, %v样本阈值%v%s, 持续发生超过%v秒",
				sampleVal, realVal, params.CompareType.GetTransferMsg(), params.Value, params.ValueType.GetTransferMsg(), check.Duration))
		}

		// 当表达式为or，其中一个为true, 就整个都是true
		if entity.MonitorAlertCheckParamsRelationOr == params.Relation && itemResult {
			return true, hitMsg
		}

		// 当表达式为and, 其中一个为false，那整个都是false
		if entity.MonitorAlertCheckParamsRelationAnd == params.Relation && !itemResult {
			return false, hitMsg
		}

		// 为and，遇到true
		if entity.MonitorAlertCheckParamsRelationAnd == params.Relation && itemResult {
			result = true
		}
	}
	return result, hitMsg
}

func (app *MonitorAlertCheckApplication) checkForParamItem(param entity.MonitorAlertCheckParamsItem, sampleVal, realVal int64) bool {
	// 绝对值处理
	diff := realVal - sampleVal
	diffPercent := diff * 100.0 / sampleVal

	// 计算是否符合表达式, 符合即代表异常了
	if param.ValueType == entity.MonitorAlertCheckParamsValueTypePercent {
		return app.compare(param, diffPercent)
	}
	return app.compare(param, diff)
}

func (app *MonitorAlertCheckApplication) reverse(compareType entity.MonitorAlertCheckParamsCompareType) entity.MonitorAlertCheckParamsCompareType {
	switch compareType {
	case entity.MonitorAlertCheckParamsCompareTypeGt:
		return entity.MonitorAlertCheckParamsCompareTypeLt
	case entity.MonitorAlertCheckParamsCompareTypeLt:
		return entity.MonitorAlertCheckParamsCompareTypeGt
	case entity.MonitorAlertCheckParamsCompareTypeEq:
		return entity.MonitorAlertCheckParamsCompareTypeEq
	case entity.MonitorAlertCheckParamsCompareTypeEgt:
		return entity.MonitorAlertCheckParamsCompareTypeElt
	case entity.MonitorAlertCheckParamsCompareTypeElt:
		return entity.MonitorAlertCheckParamsCompareTypeEgt
	default:
		return entity.MonitorAlertCheckParamsCompareTypeEq
	}
}

func (app *MonitorAlertCheckApplication) compare(param entity.MonitorAlertCheckParamsItem, diff int64) bool {
	compareType := param.CompareType
	value := param.Value

	switch compareType {
	case entity.MonitorAlertCheckParamsCompareTypeGt:
		return diff > value
	case entity.MonitorAlertCheckParamsCompareTypeLt:
		return diff < -value
	case entity.MonitorAlertCheckParamsCompareTypeEq:
		return diff == value
	case entity.MonitorAlertCheckParamsCompareTypeEgt:
		return diff >= value
	case entity.MonitorAlertCheckParamsCompareTypeElt:
		return diff <= -value
	}
	return false
}

func (app *MonitorAlertCheckApplication) getInfluxdbMeanVal(cli client.Client, measurementName string, startTime, endTime time.Time) (int64, error) {
	querySql := fmt.Sprintf("select mean(value) from %s where time >= %v and time < %v", measurementName, startTime.UnixNano(), endTime.UnixNano())
	query := client.NewQuery(querySql, app.influxdb.DbConf.Influx.Database, "s")

	response, err := cli.Query(query)
	if err != nil {
		errMsg := fmt.Sprintf("执行查询%s失败, reason: %s", measurementName, err.Error())
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}

	result := response.Results
	if result[0].Err != "" {
		errMsg := fmt.Sprintf("执行查询%s失败, reason: %s", measurementName, result[0].Err)
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}

	// 代表样本没有
	if len(result[0].Series) == 0 || len(result[0].Series[0].Values) == 0 {
		errMsg := fmt.Sprintf("执行查询%s成功, 但是没有数据点", measurementName)
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}

	columns := make(map[string]int)
	for i, column := range result[0].Series[0].Columns {
		columns[column] = i
	}

	// 解析返回
	meanIndex, _ := columns["mean"]
	row := result[0].Series[0].Values[0]
	meanVal, _ := row[meanIndex].(json.Number).Int64()
	return meanVal, nil
}

func (app *MonitorAlertCheckApplication) checkTimeRange(effectTime string, currentTime time.Time) bool {
	idx := strings.LastIndex(effectTime, "-")
	startTimeStr := effectTime[0:idx]
	endTimeStr := effectTime[idx+1 : len(effectTime)-1]

	// 转换开始时间
	dateStr := currentTime.Format("2006-01-02")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", dateStr, startTimeStr), time.Local)
	if err != nil {
		return false
	}

	// 转换结束时间
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", dateStr, endTimeStr), time.Local)
	if err != nil {
		return false
	}
	return currentTime.Unix() < startTime.Unix() || currentTime.Unix() > endTime.Unix()
}

// RegisterExecJob 注册执行
func (app *MonitorAlertCheckApplication) RegisterExecJob() {
	app.xxlExec.RegTask("alertCheck", app.alertCheck)
}
