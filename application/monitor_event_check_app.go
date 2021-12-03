package application

import (
	"butterfly-monitor/infrastructure/persistence"
	"github.com/bwmarrin/snowflake"
	"github.com/xxl-job/xxl-job-executor-go"
)

type MonitorEventCheckApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	xxlExec    xxl.Executor
}
