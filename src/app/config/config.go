package config

import (
	"butterfly-monitor/src/app/config/influxdb"
	"butterfly-monitor/src/app/config/xxljob"
	"github.com/bwmarrin/snowflake"
	adminConfig "github.com/pwh19920920/butterfly-admin/src/app/config"
	"github.com/xxl-job/xxl-job-executor-go"
	"gorm.io/gorm"
)

type Config struct {
	XxlJobExec      xxl.Executor       // xxl
	InfluxDbOption  *influxdb.DbOption // influx数据操作
	DatabaseForGorm *gorm.DB           // 数据库
	Sequence        *snowflake.Node    // 数据库序列化工具
}

func InitAll(butterflyAdminConfig adminConfig.Config) Config {
	xxlJobExec := xxljob.GetXxlJobExec()
	dbOption := influxdb.NewInfluxDbOption()
	sequence := butterflyAdminConfig.Sequence
	databaseForGorm := butterflyAdminConfig.DatabaseForGorm
	return Config{
		xxlJobExec,
		dbOption,
		databaseForGorm,
		sequence,
	}
}
