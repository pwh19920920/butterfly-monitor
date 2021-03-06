package job

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/config/influxdb"
	"butterfly-monitor/domain/entity"
	handler "butterfly-monitor/domain/handler"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"

	"sync"
	"text/template"
	"time"
)

type MonitorDataCollectJob struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	xxlExec        xxl.Executor
	influxDbOption *influxdb.DbOption
	grafana        *grafana.Config
	commonMap      application.CommonMapApplication
}

// ExecDataCollectForTimeRange 执行特定时间范围内得数据收集
func (job *MonitorDataCollectJob) ExecDataCollectForTimeRange(taskId int64, req *types.MonitorTaskExecForRangeRequest) error {
	task, err := job.repository.MonitorTaskRepository.GetById(taskId)
	if err != nil || task == nil {
		logrus.Error("ExecDataCollectForTimeRange下任务获取失败", err)
		return errors.New("任务获取失败")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go job.executeCommand(*task, &wg, req.BeginDate.Time, req.EndDate.Time, false)

	wg.Wait()
	return nil
}

// ExecDataCollect 通过xxl的index, 到数据库中取task, 然后批量执行塞入channel, 批量插入influxdb
func (job *MonitorDataCollectJob) ExecDataCollect(cxt context.Context, param *xxl.RunReq) (msg string) {
	// 获取任务分片数据
	tasks, err := job.repository.MonitorTaskRepository.FindJobByShardingNoPaging(param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		logrus.Error("从数据库获取任务失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取任务失败")
	}

	// 循环执行command, 并行执行
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go job.executeCommand(task, &wg, task.PreExecuteTime.Time, time.Now(), true)
	}

	wg.Wait()
	return "execute complete"
}

func (job *MonitorDataCollectJob) doExecuteCommand(commandHandler handler.CommandHandler, task entity.MonitorTask) (result interface{}, err error) {
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
	case <-time.After(25 * time.Second):
		return 0, errors.New("exec timeout")
	}
}

// recursiveExecuteCommand 递归执行
func (job *MonitorDataCollectJob) recursiveExecuteCommand(commandHandler handler.CommandHandler, task entity.MonitorTask, points, samplePoints []*client.Point, beginTime, maxTime time.Time, jobHandler bool) ([]*client.Point, []*client.Point, time.Time, error) {
	duration, _ := time.ParseDuration(fmt.Sprintf("%vs", task.TimeSpan))
	endTime := beginTime.Add(duration)

	// 执行结束
	if endTime.UnixMilli() > maxTime.UnixMilli() {
		logrus.Info("task执行结束, taskId: ", task.Id)
		return points, samplePoints, beginTime, nil
	}

	// 如果是job调用，就只执行一次, 直接返回就可以了
	if jobHandler {
		duration, _ := time.ParseDuration(fmt.Sprintf("-%vs", task.TimeSpan))
		beginTime = maxTime.Add(duration)
		endTime = maxTime
	}

	// 跨步时间，即真正的startTime
	stepDuration, _ := time.ParseDuration(fmt.Sprintf("-%vs", task.StepSpan))
	startTime := endTime.Add(stepDuration)

	// 执行
	logrus.Infof("%s: 执行范围: beginTime: %s - endTime: %s, startTime: %s", task.TaskKey, beginTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"))
	command, err := job.RenderTaskCommandForRange(task, beginTime, endTime, startTime)
	if err != nil {
		logrus.Errorf("commandHandler任务处理器模板引擎渲染失败, [%v]-[%v], reason: %v", task.Id, task.TaskKey, err.Error())
		return points, samplePoints, beginTime, errors.New(fmt.Sprintf("commandHandler任务处理器模板引擎渲染失败, taskId: %v", task.Id))
	}

	task.Command = command
	logrus.Infof("[%v]-[%v] 执行指令：%v", task.Id, task.TaskKey, command)
	result, err := job.doExecuteCommand(commandHandler, task)
	if err != nil {
		logrus.Errorf("[%v]-[%v] 执行指令commandHandler执行失败, err: %v", task.Id, task.TaskKey, err.Error())
		return points, samplePoints, beginTime, err
	}

	logrus.Infof("任务%v收集到的数据值为: %v", task.TaskKey, result)
	newPoints, newSamplePoints, err := job.buildPoint(task.TaskKey, result, endTime)
	if err != nil {
		return points, samplePoints, beginTime, err
	}

	// 加入列表
	points = append(points, newPoints...)
	samplePoints = append(samplePoints, newSamplePoints...)

	// 继续发起下次执行
	return job.recursiveExecuteCommand(commandHandler, task, points, samplePoints, endTime, maxTime, jobHandler)
}

func (job *MonitorDataCollectJob) buildPoint(taskKey string, result interface{}, time time.Time) ([]*client.Point, []*client.Point, error) {
	points := make([]*client.Point, 0)
	samplePoints := make([]*client.Point, 0)

	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": result,
	}

	// 创建记录
	point, err := client.NewPoint(taskKey, tags, fields, time)
	if err != nil {
		return points, samplePoints, err
	}
	points = append(points, point)
	points = append(points, point)

	sampleMeasurementNewName := job.grafana.GetSampleMeasurementNewNameForCreate(taskKey)
	for i := 1; i <= 8; i++ {
		// 创建记录
		fields := map[string]interface{}{
			"value": result,
		}

		tags := map[string]string{
			"day": fmt.Sprintf("%v", i),
		}

		samplePoint, err := client.NewPoint(sampleMeasurementNewName, tags, fields, time.AddDate(0, 0, i))
		if err != nil {
			return points, samplePoints, err
		}

		samplePoints = append(samplePoints, samplePoint)
	}
	return points, samplePoints, err
}

// executeCommand 执行命令
func (job *MonitorDataCollectJob) executeCommand(task entity.MonitorTask, wg *sync.WaitGroup, beginTime, endTime time.Time, isJobHandler bool) {
	// 执行标记
	defer wg.Done()

	// 延迟调用匿名函数 (匿名函数在主函数结束之前最后调用，可以捕获主函数中的异常)
	defer func() {
		if errInfo := recover(); errInfo != nil {
			logrus.Errorf("execCollect发生异常, [%v]-[%v], reason: %v", task.Id, task.TaskKey, errInfo)
			return
		}
	}()

	commandHandler, ok := job.commonMap.GetCommandHandlerMap()[*task.TaskType]
	if !ok {
		logrus.Errorf("commandHandler任务处理器不存在, 或者处理器类型有误, 任务执行失败：[%v]-[%v]", task.Id, task.TaskKey)
		return
	}

	cli := job.influxDbOption.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Errorf("[%v]-[%v] 任务influxdb ping 失败, reason： %v", task.Id, task.TaskKey, err.Error())
		return
	}

	// 执行开始
	points := make([]*client.Point, 0)
	samplePoints := make([]*client.Point, 0)
	points, samplePoints, preExecuteTime, err := job.recursiveExecuteCommand(commandHandler, task, points, samplePoints, beginTime, endTime, isJobHandler)

	if err != nil {
		logrus.Error("recursiveExecuteCommand exec fail, taskId: ", task.Id, err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: err.Error()})
		return
	}

	// 添加结果
	sampleMeasurementNewName := job.grafana.GetSampleMeasurementNewNameForCreate(task.TaskKey)
	logrus.Infof("生成记录数：%v - %v", sampleMeasurementNewName, len(samplePoints))
	logrus.Infof("生成记录数：%v - %v", task.TaskKey, len(points))

	// 收集数据得结果为0条
	if len(points) == 0 || len(samplePoints) == 0 {
		logrus.Error("收集数据为0条, taskId: ", task.Id)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "采集数据结果为0条"})
		return
	}

	// 等待实际数据，样本全部保存完毕
	if err := job.BatchWritingForInfluxDb(cli, task, points, ""); err != nil {
		logrus.Error("保存实时数据发生错误, taskId: ", task.Id, err)
		return
	}
	if err := job.BatchWritingForInfluxDb(cli, task, samplePoints, job.grafana.SampleRpName); err != nil {
		logrus.Error("保存样本数据发生错误, taskId: ", task.Id, err)
		return
	}
	_ = cli.Close()

	// 不需要更新, 直接返回了
	if !isJobHandler {
		return
	}

	// 更新时间
	err = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		CollectErrMsg:  " ",
		PreExecuteTime: &common.LocalTime{Time: preExecuteTime}})
	if err != nil {
		logrus.Error("insert failure", err)
		return
	}
}

func (job *MonitorDataCollectJob) BatchWritingForInfluxDb(cli client.Client, task entity.MonitorTask, points []*client.Point, rpName string) error {
	// 切割保存
	pageCount := 3000
	sliceLen := len(points) / pageCount
	if len(points)%pageCount != 0 {
		sliceLen += 1
	}

	successOps := make(chan bool, sliceLen)
	var writeWg sync.WaitGroup
	for i := 0; i < sliceLen; i++ {
		end := (i + 1) * pageCount
		start := i * pageCount
		if len(points) <= end {
			end = len(points)
		}

		ps := points[start:end]
		writeWg.Add(1)
		go job.WritingForInfluxDb(cli, task, ps, &writeWg, successOps, rpName)
		time.Sleep(time.Duration(200) * time.Millisecond)
	}

	// 等待全部执行完毕
	writeWg.Wait()

	// 判断是否全部保存完毕
	for i := 0; i < sliceLen; i++ {
		op := <-successOps
		if !op {
			return errors.New("save failure")
		}
	}
	return nil
}

func (job *MonitorDataCollectJob) CreateBatchPoint(rpName string) (client.BatchPoints, error) {
	if rpName == "" {
		return job.influxDbOption.CreateBatchPoint()
	}
	return job.influxDbOption.CreateBatchPointWithRP(rpName)
}

func (job *MonitorDataCollectJob) WritingForInfluxDb(cli client.Client, task entity.MonitorTask, points []*client.Point, wg *sync.WaitGroup, ops chan bool, rpName string) {
	defer wg.Done()

	bp, err := job.CreateBatchPoint(rpName)
	if err != nil {
		logrus.Error("exec fail, createBatchPoint is error", err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "createBatchPoint失败"})
		ops <- false
		return
	}

	// 存数据, 更新task的时间
	bp.AddPoints(points)
	err = cli.Write(bp)
	if err != nil {
		logrus.Error("write to influxdb fail", err)
		_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: "插入influxdb失败"})
		ops <- false
		return
	}

	ops <- true
}

// RenderTaskCommandForRange 模板渲染
func (job *MonitorDataCollectJob) RenderTaskCommandForRange(task entity.MonitorTask, beginTime, endTime, startTime time.Time) (string, error) {
	params := make(map[string]interface{}, 0)
	params["endTime"] = endTime.Format("2006-01-02 15:04:05")
	params["beginTime"] = beginTime.Format("2006-01-02 15:04:05")
	params["startTime"] = startTime.Format("2006-01-02 15:04:05")
	params["startTimeMilli"] = startTime.UnixMilli()
	params["endTimeMilli"] = endTime.UnixMilli()
	params["beginTimeMilli"] = beginTime.UnixMilli()

	// 创建模板对象, parse关联模板
	tmpl, err := template.New(task.TaskKey).Funcs(sprig.TxtFuncMap()).Parse(task.Command)
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
func (job *MonitorDataCollectJob) RegisterExecJob() {
	job.xxlExec.RegTask("dataCollect", job.ExecDataCollect)
	job.xxlExec.RegTask("dataSampling", job.ExecRemoveDataSampling)
}
