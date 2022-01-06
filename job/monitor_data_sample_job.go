package job

import (
	"butterfly-monitor/domain/entity"
	"context"
	"encoding/json"
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
	"sort"
	"sync"
	"time"
)

// ExecRemoveDataSampling 定时删除最大最小
func (job *MonitorDataCollectJob) ExecRemoveDataSampling(cxt context.Context, param *xxl.RunReq) (msg string) {
	var lastId int64 = 0
	return job.ExecRemoveDataSamplingForPage(lastId, cxt, param)
}

// ExecRemoveDataSamplingForPage 递归执行
func (job *MonitorDataCollectJob) ExecRemoveDataSamplingForPage(lastId int64, cxt context.Context, param *xxl.RunReq) (msg string) {
	const pageSize = 50

	// 获取任务分片数据
	tasks, err := job.repository.MonitorTaskRepository.FindSamplingJobBySharding(pageSize, lastId, param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		logrus.Error("从数据库获取任务失败", err)
		return fmt.Sprintf("exec failure, 从数据库获取任务失败")
	}

	// 循环执行command, 并行执行
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go job.doRemoveDataSampling(task, &wg, task.PreSampleTime.Time, time.Now())
	}

	wg.Wait()

	// 结束条件
	if len(tasks) < pageSize {
		return "execute sampling complete"
	}

	// 继续递归
	lastTask := tasks[pageSize-1]
	return job.ExecRemoveDataSamplingForPage(lastTask.Id, cxt, param)
}

// doExecDataSampling
func (job *MonitorDataCollectJob) doRemoveDataSampling(task entity.MonitorTask, wg *sync.WaitGroup, beginTime, endTime time.Time) {
	// 执行标记
	defer wg.Done()

	// 延迟调用匿名函数 (匿名函数在主函数结束之前最后调用，可以捕获主函数中的异常)
	defer func() {
		if errInfo := recover(); errInfo != nil {
			logrus.Errorf("sample剔除样本发送异常, %v", errInfo)
			return
		}
	}()

	cli := job.influxDbOption.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Errorf("influxdb ping 失败, reason： %v", err.Error())
		return
	}

	//  TODO 后续替换入口
	sampleMeasurementName := job.grafana.GetSampleMeasurementName(task.TaskKey)
	sampleMeasurementNewName := job.grafana.GetSampleMeasurementNewName(task.TaskKey)

	// TODO 后续替换入口
	preSampleTime, err := job.doRecursiveRemoveDataSampling(task, "", sampleMeasurementName, beginTime, endTime)
	_, _ = job.doRecursiveRemoveDataSampling(task, job.grafana.SampleRpName, sampleMeasurementNewName, beginTime, endTime)

	errMsg := " "
	if err != nil {
		logrus.Error("doRecursiveRemoveDataSampling exec fail, taskId: ", task.Id, err)
		errMsg = err.Error()

		if len(errMsg) > 100 {
			errMsg = errMsg[1:100]
		}
	}

	_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		SampleErrMsg:  errMsg,
		PreSampleTime: &common.LocalTime{Time: preSampleTime},
	})
}

// doExecDataSampling
func (job *MonitorDataCollectJob) doRecursiveRemoveDataSampling(task entity.MonitorTask, rpName, sampleMeasurementName string, beginTime, maxTime time.Time) (time.Time, error) {
	duration, _ := time.ParseDuration(fmt.Sprintf("%vs", task.TimeSpan))
	endTime := beginTime.Add(duration)

	// 执行结束
	if endTime.UnixMilli() > maxTime.UnixMilli() {
		logrus.Info("task执行结束, taskId: ", task.Id)
		return beginTime, nil
	}

	cli := job.influxDbOption.GetClient()
	logrus.Info(task.TaskKey, "：剔除数据执行范围：", beginTime.Format("2006-01-02 15:04:05"), "至", endTime.Format("2006-01-02 15:04:05"))
	querySql := fmt.Sprintf("select * from %s where time >= %v and time < %v", sampleMeasurementName, beginTime.UnixNano(), endTime.UnixNano())
	query := client.NewQueryWithRP(querySql, job.influxDbOption.DbConf.Influx.Database, rpName, "s")

	response, err := cli.Query(query)
	if err != nil {
		logrus.Error("执行查询样本失败 -> ", sampleMeasurementName, err.Error())
		return beginTime, err
	}

	result := response.Results
	if result[0].Err != "" {
		logrus.Error("执行查询样本失败 -> ", sampleMeasurementName, result[0].Err)
		return beginTime, err
	}

	// 代表样本没有, 或者样本数据低于3天
	if len(result[0].Series) == 0 || len(result[0].Series[0].Values) <= 5 {
		time.Sleep(time.Duration(200) * time.Millisecond)
		return job.doRecursiveRemoveDataSampling(task, rpName, sampleMeasurementName, endTime, maxTime)
	}

	columns := make(map[string]int)
	for i, column := range result[0].Series[0].Columns {
		columns[column] = i
	}

	dayIndex, _ := columns["day"]
	valueIndex, _ := columns["value"]

	valueDayMap := make(map[float64]string, 0)
	values := make(MyList, 0)
	for _, item := range result[0].Series[0].Values {
		day := item[dayIndex].(string)
		value, _ := item[valueIndex].(json.Number).Float64()

		valueDayMap[value] = day
		values = append(values, value)
	}

	// 对值排序
	sort.Sort(values)

	minValue := values[0]
	maxValue := values[values.Len()-1]

	minValueDay := valueDayMap[minValue]
	maxValueDay := valueDayMap[maxValue]

	// 删除最大，最小数据, 然后继续收集数据
	deleteSql := fmt.Sprintf("delete from %s where time >= %v and time < %v and (day = '%s' or day = '%s')", sampleMeasurementName, beginTime.UnixNano(), endTime.UnixNano(), minValueDay, maxValueDay)
	deleteQuery := client.NewQueryWithRP(deleteSql, job.influxDbOption.DbConf.Influx.Database, rpName, "s")

	response, err = cli.Query(deleteQuery)
	if err != nil {
		logrus.Error("执行查询样本删除失败 -> ", sampleMeasurementName, err.Error())
		return beginTime, err
	}

	if response.Results[0].Err != "" {
		logrus.Error("执行查询样本删除失败 -> ", sampleMeasurementName, response.Results[0].Err)
		return beginTime, err
	}

	time.Sleep(time.Duration(200) * time.Millisecond)
	return job.doRecursiveRemoveDataSampling(task, rpName, sampleMeasurementName, endTime, maxTime)
}

type MyList []float64

func (m MyList) Len() int {
	return len(m)
}

func (m MyList) Less(i, j int) bool {
	return m[i] < m[j]
}

func (m MyList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
