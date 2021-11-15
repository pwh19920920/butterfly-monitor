package config

import (
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/config/influxdb"
	"butterfly-monitor/config/xxljob"
	"github.com/bwmarrin/snowflake"
	adminConfig "github.com/pwh19920920/butterfly-admin/config"
	"github.com/xxl-job/xxl-job-executor-go"
	"gorm.io/gorm"
)

type Config struct {
	XxlJobExec      xxl.Executor       // xxl
	InfluxDbOption  *influxdb.DbOption // influx数据操作
	DatabaseForGorm *gorm.DB           // 数据库
	Sequence        *snowflake.Node    // 数据库序列化工具
	Grafana         *grafana.Config    // grafana配置
}

func InitAll(butterflyAdminConfig adminConfig.Config) Config {
	xxlJobExec := xxljob.GetXxlJobExec()
	dbOption := influxdb.NewInfluxDbOption()
	grafanaConf := grafana.InitGrafanaConfig()
	sequence := butterflyAdminConfig.Sequence
	databaseForGorm := butterflyAdminConfig.DatabaseForGorm
	return Config{
		xxlJobExec,
		dbOption,
		databaseForGorm,
		sequence,
		grafanaConf,
	}
}
