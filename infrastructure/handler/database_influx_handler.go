package handler

import (
	"butterfly-monitor/domain/entity"
	"encoding/json"
	"errors"
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type DatabaseInfluxHandler struct {
	client client.Client
}

type CommandInfluxParams struct {
	Database        string `json:"database"`
	RetentionPolicy string `json:"retentionPolicy"`
	Column          string `json:"column"`
}

func (dbHandler *DatabaseInfluxHandler) TestConnect(database entity.MonitorDatabase) error {
	// 创建client
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     database.Url,
		Username: database.Username,
		Password: database.Password,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}

	// 判断服务是不是可用
	pingTime, version, err := httpClient.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ", database.Url, " ping返回 - ", pingTime, " - ", version)
	if err != nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}

	_ = httpClient.Close()
	return nil
}

// NewInstance 创建实例
func (dbHandler *DatabaseInfluxHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	// 创建client
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     database.Url,
		Username: database.Username,
		Password: database.Password,
	})

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}

	// 判断服务是不是可用
	pingTime, version, err := httpClient.Ping(time.Duration(10) * time.Second)
	logrus.Info("influxdb ", database.Url, " ping返回 - ", pingTime, " - ", version)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}
	return &DatabaseInfluxHandler{client: httpClient}, nil
}

// ExecuteQuery 执行查询
func (dbHandler *DatabaseInfluxHandler) ExecuteQuery(task entity.MonitorTask) (interface{}, error) {
	if task.ExecParams == "" {
		return 0, errors.New("执行参数有误")
	}

	var params CommandInfluxParams
	err := json.Unmarshal([]byte(task.ExecParams), &params)
	if err != nil {
		return 0, err
	}

	query := client.NewQueryWithRP(task.Command, params.Database, params.RetentionPolicy, "s")
	response, err := dbHandler.client.Query(query)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("taskId:%v - %v, influx执行发生异常: %v", task.Id, task.TaskName, err.Error()))
	}

	result := response.Results
	if result == nil || len(result) == 0 {
		return 0, errors.New(fmt.Sprintf("taskId:%v - %v, influx执行Results查无结果", task.Id, task.TaskName))
	}

	// 内层错误校验
	if result[0].Err != "" {
		return 0, errors.New(fmt.Sprintf("taskId:%v - %v, influx执行发生异常: %v", task.Id, task.TaskName, result[0].Err))
	}

	if len(result[0].Series) == 0 || len(result[0].Series[0].Values) == 0 {
		return 0, errors.New(fmt.Sprintf("taskId:%v - %v, influx执行Series查无数据", task.Id, task.TaskName))
	}

	columns := make(map[string]int)
	for i, column := range result[0].Series[0].Columns {
		columns[column] = i
	}

	columnIndex, _ := columns[params.Column]
	row := result[0].Series[0].Values[0]
	columnResult := row[columnIndex]
	if columnResult == nil {
		return 0, errors.New(fmt.Sprintf("InfluxDB执行查询成功, 但列[%v]无数据", params.Column))
	}

	meanVal, _ := row[columnIndex].(json.Number).Float64()
	return strconv.ParseFloat(fmt.Sprintf("%v", meanVal), 64)
}
