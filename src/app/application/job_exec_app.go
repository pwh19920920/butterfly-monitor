package application

import (
	"butterfly-monitor/src/app/config/influxdb"
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/domain/handler"
	handlerImpl "butterfly-monitor/src/app/infrastructure/handler"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/src/app/common"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"time"
)

var commandHandlerMap = make(map[entity.TaskType]handler.CommandHandler, 0)
var databaseHandlerMap = make(map[entity.DataSourceType]handler.DatabaseHandler, 0)

var databaseLoadTime *common.LocalTime
var databaseMap = make(map[int64]interface{}, 0)

type JobExecApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	xxlExec        xxl.Executor
	influxDbOption *influxdb.DbOption
}

func NewJobExecApplication(sequence *snowflake.Node, repository *persistence.Repository,
	xxlExec xxl.Executor, influxDbOption *influxdb.DbOption) JobExecApplication {
	// 初始化数据库源
	go initDatabaseConnect(repository)
	return JobExecApplication{
		sequence:       sequence,
		repository:     repository,
		xxlExec:        xxlExec,
		influxDbOption: influxDbOption,
	}
}

func initDatabaseConnect(repository *persistence.Repository) {
	databaseList, err := repository.JobDatabaseRepository.SelectAll(databaseLoadTime)
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
			_ = repository.JobDatabaseRepository.UpdateById(database.Id, &database)
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

// ExecDataCollect 通过xxl的index, 到数据库中取task, 然后批量执行塞入channel, 批量插入influxdb
func (job *JobExecApplication) ExecDataCollect(cxt context.Context, param *xxl.RunReq) (msg string) {
	var lastId int64 = 0
	const pageSize = 50

	// 获取任务分片数据
	tasks, err := job.repository.JobTaskRepository.FindJobBySharding(pageSize, lastId, param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		logrus.Error("从数据库获取任务失败", err)
		return
	}

	bp, err := job.influxDbOption.CreateBatchPoint()
	if err != nil {
		logrus.Error("exec fail, createBatchPoint is error", err)
		return
	}

	// 循环执行command, 并行执行, 通过chan做交互
	pointChan := make(chan *client.Point, len(tasks))
	for _, task := range tasks {
		go job.executeCommand(task, pointChan)
	}

	// 从chan中取结果
	for range tasks {
		point := <-pointChan
		if point == nil {
			continue
		}
		bp.AddPoint(point)
	}

	if len(bp.Points()) == 0 {
		logrus.Error("exec done, complete count is zero")
		return "exec done, complete count is zero"
	}

	writeErr := job.influxDbOption.Client.Write(bp)
	if err != nil {
		logrus.Error("exec fail", writeErr)
		return
	}
	return fmt.Sprintf("exec done, success count is %v", len(bp.Points()))
}

func (job *JobExecApplication) doExecuteCommand(commandHandler handler.CommandHandler, task entity.JobTask) (int64, error) {
	ctx := context.Background()
	done := make(chan bool, 1)

	var result int64
	var err error

	// 执行耗时任务
	go func(ctx context.Context) {
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

// executeCommand 执行命令
func (job *JobExecApplication) executeCommand(task entity.JobTask, pointChan chan *client.Point) {
	commandHandler, ok := commandHandlerMap[task.TaskType]
	if !ok {
		logrus.Error("commandHandler任务处理器不存在, 或者处理器类型有误")
		pointChan <- nil
		return
	}

	// 执行
	result, err := job.doExecuteCommand(commandHandler, task)
	if err != nil {
		logrus.Error("commandHandler执行失败")
		pointChan <- nil
		return
	}

	tags := map[string]string{}
	fields := map[string]interface{}{
		"value": result,
	}

	// 创建记录
	point, err := client.NewPoint(task.TaskKey, tags, fields, time.Now())
	if err != nil {
		pointChan <- nil
		return
	}
	pointChan <- point
}

// RegisterExecJob 注册执行
func (job *JobExecApplication) RegisterExecJob() {
	job.xxlExec.RegTask("dataCollect", job.ExecDataCollect)
}
