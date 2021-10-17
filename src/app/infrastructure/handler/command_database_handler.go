package handler

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/domain/handler"
	"encoding/json"
	"errors"
	"fmt"
)

type CommandDataBaseHandler struct {
	DatabaseMap map[int64]interface{}
}

type CommandDataBaseParams struct {
	DatabaseId int64 `json:"databaseId"`
}

func (dbHandler *CommandDataBaseHandler) ExecuteCommand(task entity.MonitorTask) (int64, error) {
	if task.ExecParams == "" {
		return 0, errors.New("执行参数有误")
	}

	var params CommandDataBaseParams
	err := json.Unmarshal([]byte(task.ExecParams), &params)
	if err != nil {
		return 0, err
	}

	// 从map获取数据连接
	dbConn, ok := dbHandler.DatabaseMap[params.DatabaseId]
	if !ok {
		return 0, errors.New(fmt.Sprintf("dbMap is not contain %v", task.DatabaseId))
	}
	return dbConn.(handler.DatabaseHandler).ExecuteQuery(task)
}
