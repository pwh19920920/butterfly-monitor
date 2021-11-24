package application

import (
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/config/influxdb"
	"butterfly-monitor/domain/entity"
	handler "butterfly-monitor/domain/handler"
	handlerImpl "butterfly-monitor/infrastructure/handler"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"sync"
	"text/template"
	"time"
)

var commandHandlerMap = make(map[entity.MonitorTaskType]handler.CommandHandler, 0)
var databaseHandlerMap = make(map[entity.DataSourceType]handler.DatabaseHandler, 0)

var databaseLoadTime *common.LocalTime
var databaseMap = make(map[int64]interface{}, 0)

type MonitorExecApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	xxlExec        xxl.Executor
	influxDbOption *influxdb.DbOption
	grafana        *grafana.Config
}

func NewMonitorExecApplication(sequence *snowflake.Node, repository *persistence.Repository, xxlExec xxl.Executor, influxDbOption *influxdb.DbOption, grafana *grafana.Config) MonitorExecApplication {
	// 初始化数据库源
	go initDatabaseConnect(repository)
	return MonitorExecApplication{
		sequence:       sequence,
		repository:     repository,
		xxlExec:        xxlExec,
		influxDbOption: influxDbOption,
		grafana:        grafana,
	}
}

func initDatabaseConnect(repository *persistence.Repository) {
	databaseList, err := repository.MonitorDatabaseRepository.SelectAll(databaseLoadTime)
	if err != nil {
		return
	}

	for _, database := range databaseList {
		databaseHandler, ok := databaseHandlerMap[database.Type]
		if !ok {
			continue
		}

		dbHandler, err := databaseHandler.NewInstance(database)
		if err != nil {
			// 失败得情况下需要更新一下，以便下一次定时扫新连接得时候重新再连接
			_ = repository.MonitorDatabaseRepository.UpdateById(database.Id, &database)
			continue
		}
		databaseMap[database.Id] = dbHandler
	}

	// 初始化执行类型
	commandHandlerMap[entity.TaskTypeDatabase] = &handlerImpl.CommandDataBaseHandler{DatabaseMap: databaseMap}

	// 睡眠后继续执行
	databaseLoadTime = &common.LocalTime{Time: time.Now()}
	time.Sleep(time.Duration(1) * time.Minute)
	go initDatabaseConnect(repository)
}

// 默认参数
func init() {
	// 命令类型
	commandHandlerMap[entity.TaskTypeURL] = new(handlerImpl.CommandUrlHandler)
	commandHandlerMap[entity.TaskTypeDatabase] = new(handlerImpl.CommandDataBaseHandler)

	// 数据库类型
	databaseHandlerMap[entity.DataSourceTypeMysql] = new(handlerImpl.DatabaseMysqlHandler)
}

// ExecDataCollectForTimeRange 执行特定时间范围内得数据收集
func (job *MonitorExecApplication) ExecDataCollectForTimeRange(taskId int64, req *types.MonitorTaskExecForRangeRequest) error {
	task, err := job.repository.MonitorTaskRepository.GetById(taskId)
	if err != nil || task == nil {
		logrus.Error("ExecDataCollectForTimeRange下任务获取失败", err)
		return errors.New("任务获取失败")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go job.executeCommand(*task, &wg, req.BeginDate.Time, req.EndDate.Time)

	wg.Wait()
	return nil
}

// ExecDataCollect 通过xxl的index, 到数据库中取task, 然后批量执行塞入channel, 批量插入influxdb
func (job *MonitorExecApplication) ExecDataCollect(cxt context.Context, param *xxl.RunReq) (msg string) {
	var lastId int64 = 0
	return job.ExecDataCollectForPage(lastId, cxt, param)
}

// ExecDataCollectForPage 递归执行
func (job *MonitorExecApplication) ExecDataCollectForPage(lastId int64, cxt context.Context, param *xxl.RunReq) (msg string) {
	const pageSize = 50

	// 获取任务分片数据
	tasks, err := job.repository.MonitorTaskRepository.FindJobBySharding(pageSize, lastId, param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		logrus.Error("从数据库获取任务失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取任务失败")
	}

	// 循环执行command, 并行执行
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go job.executeCommand(task, &wg, task.PreExecuteTime.Time, time.Now())
	}

	wg.Wait()

	// 结束条件
	if len(tasks) < pageSize {
		return "execute complete"
	}

	// 继续递归
	lastTask := tasks[pageSize-1]
	return job.ExecDataCollectForPage(lastTask.Id, cxt, param)
}

func (job *MonitorExecApplication) doExecuteCommand(commandHandler handler.CommandHandler, task entity.MonitorTask) (result interface{}, err error) {
	ctx := context.Background()
	done := make(chan bool, 1)

	// 执行耗时任务
	go func(ctx context.Context) {
		// 延迟调用匿名函数 (匿名函数在主函数结束之前最后调用，可以捕获主函数中的异常)
		defer func() {
			if errInfo := recover(); errInfo != nil {
				result = nil
				err = errInfo.(error)
				done <- true
			}
		}()

		// 正常执行
		result, err = commandHandler.ExecuteCommand(task)
		done <- true
	}(ctx)

	select {
	case <-done:
		return result, err
	case <-time.After(30 * time.Second):
		return 0, errors.New("exec timeout")
	}
}

// recursiveExecuteCommand 递归执行
func (job *MonitorExecApplication) recursiveExecuteCommand(commandHandler handler.CommandHandler, task entity.MonitorTask,
	points []*client.Point, beginTime, maxTime time.Time) ([]*client.Point, time.Time, error) {
	duration, _ := time.ParseDuration(fmt.Sprintf("%vs", task.TimeSpan))
	endTime := beginTime.Add(duration)

	// 如果不支持回溯，就只执行一次, 直接返回就可以了
	if *task.RecallStatus == entity.MonitorRecallStatusNotSupport {
		duration, _ := time.ParseDuration(fmt.Sprintf("-%vs", task.TimeSpan))
		beginTime = maxTime.Add(duration)
		endTime = maxTime
	}

	// 执行结束
	if endTime.UnixMilli() > maxTime.UnixMilli() {
		logrus.Info("task执行结束, taskId: ", task.Id)
		return points, beginTime, nil
	}

	// 执行
	logrus.Info(task.TaskKey, "：执行范围：", beginTime.Format("2006-01-02 15:04:05"), "至", endTime.Format("2006-01-02 15:04:05"))
	command, err := job.RenderTaskCommandForRange(task, beginTime, endTime)
	if err != nil {
		logrus.Error("commandHandler任务处理器模板引擎渲染失败", task, err.Error())
		return points, beginTime, errors.New(fmt.Sprintf("commandHandler任务处理器模板引擎渲染失败, taskId: %v", task.Id))
	}

	task.Command = command
	logrus.Info("执行指令：", command)
	result, err := job.doExecuteCommand(commandHandler, task)
	if err != nil {
		logrus.Error("commandHandler执行失败", err.Error())
		return points, beginTime, err
	}

	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": result,
	}

	// 创建记录
	point, err := client.NewPoint(task.TaskKey, tags, fields, endTime)
	if err != nil {
		return points, beginTime, err
	}

	// 样本数据
	samplePoints := make([]*client.Point, 0)
	sampleMeasurementName := fmt.Sprintf("%s.%s_sample", job.grafana.SampleRpName, task.TaskKey)
	for i := 1; i <= 7; i++ {
		// 创建记录
		fields := map[string]interface{}{
			"value": result,
		}

		tags := map[string]string{
			"day": fmt.Sprintf("%v", i),
		}

		samplePoint, err := client.NewPoint(sampleMeasurementName, tags, fields, endTime.AddDate(0, 0, i))
		if err != nil {
			return points, beginTime, err
		}

		samplePoints = append(samplePoints, samplePoint)
	}

	// 添加结果
	points = append(points, point)
	for _, samplePoint := range samplePoints {
		points = append(points, samplePoint)
	}

	// 如果不支持回溯，就只执行一次, 直接返回就可以了
	if *task.RecallStatus == entity.MonitorRecallStatusNotSupport {
		return points, endTime, nil
	}

	// 继续发起下次执行
	return job.recursiveExecuteCommand(commandHandler, task, points, endTime, maxTime)
}

// executeCommand 执行命令
func (job *MonitorExecApplication) executeCommand(task entity.MonitorTask, wg *sync.WaitGroup, beginTime, endTime time.Time) {
	// 执行标记
	defer wg.Done()

	commandHandler, ok := commandHandlerMap[*task.TaskType]
	if !ok {
		logrus.Error("commandHandler任务处理器不存在, 或者处理器类型有误")
		return
	}

	cli := job.influxDbOption.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Error("influxdb ping 失败")
		return
	}

	// 执行开始
	points := make([]*client.Point, 0)
	points, preExecuteTime, err := job.recursiveExecuteCommand(commandHandler, task, points, beginTime, endTime)

	if err != nil {
		logrus.Error("recursiveExecuteCommand exec fail, taskId: ", task.Id, err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: err.Error()}, nil)
		return
	}

	// 收集数据得结果为0条
	if len(points) == 0 {
		logrus.Error("收集数据为0条, taskId: ", task.Id)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "采集数据结果为0条"}, nil)
		return
	}

	// 切割保存
	pageCount := 5000
	sliceLen := len(points) / pageCount
	if len(points)%pageCount != 0 {
		sliceLen += 1
	}

	successOps := make(chan bool, sliceLen)
	var writeWg sync.WaitGroup
	for i := 0; i < sliceLen; i++ {
		end := (i+1)*pageCount - 1
		start := i * pageCount
		if len(points) <= end {
			end = len(points) - 1
		}

		ps := points[start:end]
		writeWg.Add(1)
		go job.WritingForInfluxDb(cli, task, ps, &writeWg, successOps)
		time.Sleep(time.Duration(2) * time.Second)
	}

	// 等待全部执行完毕
	writeWg.Wait()
	_ = cli.Close()

	// 判断是否全部保存完毕
	for i := 0; i < sliceLen; i++ {
		op := <-successOps
		if !op {
			return
		}
	}

	// 更新时间
	err = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		CollectErrMsg:  " ",
		PreExecuteTime: &common.LocalTime{Time: preExecuteTime}}, nil)
	if err != nil {
		logrus.Error("insert failure", err)
		return
	}
}

func (job *MonitorExecApplication) WritingForInfluxDb(cli client.Client, task entity.MonitorTask, points []*client.Point, wg *sync.WaitGroup, ops chan bool) {
	defer wg.Done()

	bp, err := job.influxDbOption.CreateBatchPoint()
	if err != nil {
		logrus.Error("exec fail, createBatchPoint is error", err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "createBatchPoint失败"}, nil)
		ops <- false
		return
	}

	// 存数据, 更新task的时间
	bp.AddPoints(points)
	err = cli.Write(bp)
	if err != nil {
		logrus.Error("write to influxdb fail", err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "插入influxdb失败"}, nil)
		ops <- false
		return
	}

	ops <- true
}

// RenderTaskCommandForRange 模板渲染
func (job *MonitorExecApplication) RenderTaskCommandForRange(task entity.MonitorTask, beginTime, endTime time.Time) (string, error) {
	params := make(map[string]interface{}, 0)
	params["endTime"] = endTime.Format("2006-01-02 15:04:05")
	params["beginTime"] = beginTime.Format("2006-01-02 15:04:05")

	// 创建模板对象, parse关联模板
	tmpl, err := template.New(task.TaskKey).Parse(task.Command)
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
func (job *MonitorExecApplication) RegisterExecJob() {
	job.xxlExec.RegTask("dataCollect", job.ExecDataCollect)
	job.xxlExec.RegTask("dataSampling", job.ExecRemoveDataSampling)
}
