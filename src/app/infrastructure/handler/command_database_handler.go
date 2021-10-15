package handler

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/domain/handler"
	"errors"
	"fmt"
)

type CommandDataBaseHandler struct {
	DatabaseMap map[int64]interface{}
}

func (dbHandler *CommandDataBaseHandler) ExecuteCommand(task entity.JobTask) (int64, error) {
	// 从map获取数据连接
	dbConn, ok := dbHandler.DatabaseMap[task.DatabaseId]
	if !ok {
		return 0, errors.New(fmt.Sprintf("dbMap is not contain %v", task.DatabaseId))
	}
	return dbConn.(handler.DatabaseHandler).ExecuteQuery(task)
}
