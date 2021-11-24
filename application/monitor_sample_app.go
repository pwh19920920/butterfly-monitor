package application

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
func (job *MonitorExecApplication) ExecRemoveDataSampling(cxt context.Context, param *xxl.RunReq) (msg string) {
	var lastId int64 = 0
	return job.ExecRemoveDataSamplingForPage(lastId, cxt, param)
}

// ExecRemoveDataSamplingForPage 递归执行
func (job *MonitorExecApplication) ExecRemoveDataSamplingForPage(lastId int64, cxt context.Context, param *xxl.RunReq) (msg string) {
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
func (job *MonitorExecApplication) doRemoveDataSampling(task entity.MonitorTask, wg *sync.WaitGroup, beginTime, endTime time.Time) {
	// 执行标记
	defer wg.Done()

	cli := job.influxDbOption.GetClient()
	pingTime, version, err := cli.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ping返回 - ", pingTime, " - ", version)
	if err != nil {
		logrus.Error("influxdb ping 失败")
		return
	}

	// 执行开始
	preSampleTime, err := job.doRecursiveRemoveDataSampling(task, beginTime, endTime)

	errMsg := " "
	if err != nil {
		logrus.Error("doRecursiveRemoveDataSampling exec fail, taskId: ", task.Id, err)
		errMsg = err.Error()
	}

	_ = job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		SampleErrMsg:  errMsg,
		PreSampleTime: &common.LocalTime{Time: preSampleTime},
	}, nil)
}

// doExecDataSampling
func (job *MonitorExecApplication) doRecursiveRemoveDataSampling(task entity.MonitorTask, beginTime, maxTime time.Time) (time.Time, error) {
	// measurement
	sampleMeasurementName := fmt.Sprintf("\"%s.%s_sample\"", job.grafana.SampleRpName, task.TaskKey)
	duration, _ := time.ParseDuration(fmt.Sprintf("%vs", task.TimeSpan))
	endTime := beginTime.Add(duration)

	// 执行结束
	if endTime.UnixMilli() > maxTime.UnixMilli() {
		logrus.Info("task执行结束, taskId: ", task.Id)
		return beginTime, nil
	}

	// 执行
	cli := job.influxDbOption.GetClient()
	logrus.Info(task.TaskKey, "：剔除数据执行范围：", beginTime.Format("2006-01-02 15:04:05"), "至", endTime.Format("2006-01-02 15:04:05"))
	querySql := fmt.Sprintf("select * from %s where time >= %v and time < %v", sampleMeasurementName, beginTime.UnixNano(), endTime.UnixNano())
	query := client.NewQuery(querySql, job.influxDbOption.DbConf.Influx.Database, "s")

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
		return job.doRecursiveRemoveDataSampling(task, endTime, maxTime)
	}

	columns := make(map[string]int)
	for i, column := range result[0].Series[0].Columns {
		columns[column] = i
	}

	dayIndex, _ := columns["day"]
	valueIndex, _ := columns["value"]

	valueDayMap := make(map[int64]string, 0)
	values := make(MyList, 0)
	for _, item := range result[0].Series[0].Values {
		day := item[dayIndex].(string)
		value, _ := item[valueIndex].(json.Number).Int64()

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
	deleteQuery := client.NewQuery(deleteSql, job.influxDbOption.DbConf.Influx.Database, "s")

	response, err = cli.Query(deleteQuery)
	if err != nil {
		logrus.Error("执行查询样本删除失败 -> ", sampleMeasurementName, err.Error())
		return beginTime, err
	}

	if response.Results[0].Err != "" {
		logrus.Error("执行查询样本删除失败 -> ", sampleMeasurementName, response.Results[0].Err)
		return beginTime, err
	}
	return job.doRecursiveRemoveDataSampling(task, endTime, maxTime)
}

type MyList []int64

func (m MyList) Len() int {
	return len(m)
}

func (m MyList) Less(i, j int) bool {
	return m[i] < m[j]
}

func (m MyList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
