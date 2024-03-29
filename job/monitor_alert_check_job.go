package job

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/config/influxdb"
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MonitorAlertCheckJob struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	influxdb   *influxdb.DbOption
	xxlExec    xxl.Executor
	grafana    *grafana.Config
	alertConf  application.AlertConfApplication
}

func (app *MonitorAlertCheckJob) alertCheck(cxt context.Context, param *xxl.RunReq) (msg string) {
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
func (app *MonitorAlertCheckJob) execCheck(conf application.AlertConfObject, check entity.MonitorTaskAlert, wg *sync.WaitGroup) {
	// 执行标记
	defer wg.Done()

	// 延迟调用匿名函数 (匿名函数在主函数结束之前最后调用，可以捕获主函数中的异常)
	defer func() {
		if errInfo := recover(); errInfo != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			logrus.Errorf("报警规则检查发送异常, %v", string(buf[:n]))
			return
		}
	}()

	// 判断任务状态
	task, err := app.repository.MonitorTaskRepository.GetById(check.TaskId)
	if err != nil || task == nil {
		logrus.Errorf("execCheck 获取任务[%v]失败", check.TaskId)
		return
	}

	if task.AlertStatus == entity.MonitorAlertStatusClose || task.TaskStatus == entity.MonitorTaskStatusClose {
		return
	}

	// 检查规则
	if check.Params == "" {
		logrus.Infof("execCheck 任务[%v]-[%v]未配置检查规则", check.TaskId, task.TaskKey)
		return
	}

	var params []entity.MonitorAlertCheckParams
	if err := json.Unmarshal([]byte(check.Params), &params); err != nil {
		logrus.Infof("execCheck 任务[%v]-[%v]规则反序列化失败", check.TaskId, task.TaskKey)
		return
	}

	// 时间倒退
	currentTime := time.Now()
	startDuration, _ := time.ParseDuration(fmt.Sprintf("-%vs", 2*task.TimeSpan))
	endDuration, _ := time.ParseDuration(fmt.Sprintf("-%vs", task.TimeSpan))
	startTime := currentTime.Add(startDuration)
	endTime := currentTime.Add(endDuration)

	// 检查influxdb是否正常
	cli := app.influxdb.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Errorf("influxdb ping 失败, 任务[%v]-[%v]执行跳过, reason： %v", check.TaskId, task.TaskKey, err.Error())
		return
	}

	// 执行结束关闭cli
	defer func(cli client.Client) {
		_ = cli.Close()
	}(cli)

	// 新入口, 查询样本平均, 以及实时数据, 只要有一个不存在, 则忽略, 无法判定为错误
	sampleMeasurementName := app.grafana.GetSampleMeasurementNewNameForQuery(task.TaskKey)
	sampleVal, err := app.getInfluxdbMeanVal(cli, app.grafana.SampleRpName, sampleMeasurementName, startTime, endTime)
	if err != nil || sampleVal == -1 {
		logrus.Infof("[%v]-[%v]任务样本数据没有数据点, 将被忽略, 错误原因：%v", task.Id, task.TaskKey, err)
		return
	}

	realVal, err := app.getInfluxdbMeanVal(cli, "", fmt.Sprintf("\"%s\"", task.TaskKey), startTime, endTime)
	if err != nil {
		logrus.Infof("[%v]-[%v]任务实时数据没有数据点, 将被忽略, 错误原因：%v", task.Id, task.TaskKey, err)
		return
	}

	// 规则校验 [[rule, rule], [], []], 是否被检查出来命中了检测规则, 如果没数据，直接认为是异常
	checkResult, hitMsg := app.checkForParam(currentTime, check, params, sampleVal, realVal)
	if !checkResult {
		// 没命中, 则要更新event为误报, 更新check任务的FirstFlagTime,PreCheckTime=current, AlertStatus=Normal
		_ = app.repository.MonitorTaskAlertRepository.ModifyForNormal(check.Id, currentTime)
		return
	}

	checkDuration, _ := time.ParseDuration(fmt.Sprintf("-%vs", check.Duration))
	beforeTime := currentTime.Add(checkDuration)

	// 超出5分钟得报警了
	if beforeTime.Unix() > check.FirstFlagTime.Unix() {
		logrus.Infof("alertCheck【%v】 - |(%v - %v)|/%v = %v, %s", task.Id, realVal, sampleVal, sampleVal, checkResult, strings.Join(hitMsg, ";"))

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
	logrus.Infof("execCheck [%v]-[%v]出现异常, 样本：%v, 实时: %v", task.Id, task.TaskKey, sampleVal, realVal)
	_ = app.repository.MonitorTaskAlertRepository.ModifyByPending(check.Id, currentTime)
}

func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return num
}

// *** 组于组之间为or ***
func (app *MonitorAlertCheckJob) checkForParam(currentTime time.Time, check entity.MonitorTaskAlert, paramGroups []entity.MonitorAlertCheckParams, sampleVal, realVal float64) (bool, []string) {
	// 如果实时不存在，直接认为是异常
	if realVal == -1 {
		return true, []string{fmt.Sprintf("样本值: %v, 当前值为空, 持续发生超过%v秒", Decimal(sampleVal), check.Duration)}
	}

	for _, param := range paramGroups {
		// 不在时间段里，直接跳过
		if param.EffectTimes != nil && app.checkTimeRange(param.EffectTimes, currentTime) {
			continue
		}

		// 一个成功即命中
		arrCheckResult, hitMsg := app.checkForParamArr(check, param, sampleVal, realVal)
		if arrCheckResult {
			return arrCheckResult, hitMsg
		}
	}
	return false, make([]string, 0)
}

// *** 使用此算法的前提是每个组里只有or或者and ***
func (app *MonitorAlertCheckJob) checkForParamArr(check entity.MonitorTaskAlert, param entity.MonitorAlertCheckParams, sampleVal, realVal float64) (bool, []string) {
	result := false
	hitMsg := make([]string, 0)
	for _, params := range param.Rules {
		itemResult := app.checkForParamItem(params, sampleVal, realVal)
		if itemResult {
			hitMsg = append(hitMsg, fmt.Sprintf("样本值: %v, 当前值: %v, %v样本阈值%v%s, 持续发生超过%v秒",
				Decimal(sampleVal), Decimal(realVal), params.CompareType.GetTransferMsg(), params.Value, params.ValueType.GetTransferMsg(), check.Duration))
		}

		// 1. 当表达式为or，其中一个为true, 就整个都是true
		if entity.MonitorAlertCheckParamsRelationOr == param.Relation && itemResult {
			return true, hitMsg
		}

		// 2. 当表达式为and, 其中一个为false，那整个都是false
		if entity.MonitorAlertCheckParamsRelationAnd == param.Relation && !itemResult {
			return false, hitMsg
		}

		// 3. 为and，遇到true, 直接修改标记为true, 如果遇到false直接走表达式2
		if entity.MonitorAlertCheckParamsRelationAnd == param.Relation && itemResult {
			result = true
		}

		// 4. 为or，遇到false, 啥也不处理, 本身标记为就是false
	}
	return result, hitMsg
}

func (app *MonitorAlertCheckJob) checkForParamItem(param entity.MonitorAlertCheckParamsItem, sampleVal, realVal float64) bool {
	// 计算是否符合表达式, 符合即代表异常了, 百分比和样本值有关，实际值和样本无关
	if param.ValueType == entity.MonitorAlertCheckParamsValueTypePercent {
		// 绝对值处理
		diff := realVal - sampleVal
		if sampleVal == 0 {
			// 如果样本为0，就得伪装成1，防止报错
			return app.compare(param, diff*100.0)
		}
		diffPercent := diff * 100.0 / sampleVal
		return app.compare(param, diffPercent)
	}

	// 走这个方式，不需要样本
	return app.compareForValue(param, realVal)
}

func (app *MonitorAlertCheckJob) compareForValue(param entity.MonitorAlertCheckParamsItem, realVal float64) bool {
	compareType := param.CompareType
	value := param.Value

	switch compareType {
	case entity.MonitorAlertCheckParamsCompareTypeGt:
		return realVal > value
	case entity.MonitorAlertCheckParamsCompareTypeLt:
		return realVal < value
	case entity.MonitorAlertCheckParamsCompareTypeEq:
		return realVal == value
	case entity.MonitorAlertCheckParamsCompareTypeEgt:
		return realVal >= value
	case entity.MonitorAlertCheckParamsCompareTypeElt:
		return realVal <= value
	}
	return false
}

func (app *MonitorAlertCheckJob) compare(param entity.MonitorAlertCheckParamsItem, diff float64) bool {
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

// 从influxdb获取不到数据，不代表是异常，也有可能本身就没数据
func (app *MonitorAlertCheckJob) getInfluxdbMeanVal(cli client.Client, policyName, measurementName string, startTime, endTime time.Time) (float64, error) {
	querySql := fmt.Sprintf("select mean(value) from %s where time >= %v and time < %v", measurementName, startTime.UnixNano(), endTime.UnixNano())
	query := client.NewQueryWithRP(querySql, app.influxdb.DbConf.Influx.Database, policyName, "s")

	response, err := cli.Query(query)
	if err != nil {
		errMsg := fmt.Sprintf("执行查询%s失败, reason: %s", measurementName, err.Error())
		logrus.Errorf(errMsg)
		return -1, errors.New(errMsg)
	}

	result := response.Results
	if result == nil || len(result) == 0 {
		errMsg := fmt.Sprintf("执行查询%s失败, reason: %s", measurementName, "返回result数据为nil")
		logrus.Errorf(errMsg)
		return -1, nil
	}

	// 内层错误校验
	if result[0].Err != "" {
		errMsg := fmt.Sprintf("执行查询%s失败, reason: %s", measurementName, result[0].Err)
		logrus.Errorf(errMsg)
		return -1, errors.New(errMsg)
	}

	if len(result[0].Series) == 0 || len(result[0].Series[0].Values) == 0 {
		errMsg := fmt.Sprintf("执行查询%s成功, 但是没有数据点", measurementName)
		logrus.Errorf(errMsg)
		return -1, nil
	}

	columns := make(map[string]int)
	for i, column := range result[0].Series[0].Columns {
		columns[column] = i
	}

	// 解析返回
	meanIndex, _ := columns["mean"]
	row := result[0].Series[0].Values[0]
	meanVal, _ := row[meanIndex].(json.Number).Float64()
	return meanVal, nil
}

func (app *MonitorAlertCheckJob) checkTimeRange(effectTimes []string, currentTime time.Time) bool {
	startTimeStr := effectTimes[0]
	endTimeStr := effectTimes[1]

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
func (app *MonitorAlertCheckJob) RegisterExecJob() {
	app.xxlExec.RegTask("alertCheck", app.alertCheck)
}
