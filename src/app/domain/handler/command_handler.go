package handler

import "butterfly-monitor/src/app/domain/entity"

type CommandHandler interface {

	// ExecuteCommand 通过任务得到结果
	ExecuteCommand(task entity.JobTask) (int64, error)
}
