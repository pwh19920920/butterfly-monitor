package handler

import (
	"butterfly-monitor/domain/entity"
)

type CommandHandler interface {

	// ExecuteCommand 通过任务得到结果
	ExecuteCommand(task entity.MonitorTask) (interface{}, error)
}
