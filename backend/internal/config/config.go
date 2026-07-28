package config

import (
	"fmt"

	"dragonfly-monitor/internal/config/auth"
	"dragonfly-monitor/internal/config/database"
	"dragonfly-monitor/internal/config/grafana"
	"dragonfly-monitor/internal/config/sequence"
	"dragonfly-monitor/internal/config/tdengine"
	"dragonfly-monitor/internal/config/timeseries"
	"dragonfly-monitor/internal/config/victoriametrics"
	domainHandler "dragonfly-monitor/internal/domain/handler"
	infraHandler "dragonfly-monitor/internal/infrastructure/handler"

	"github.com/pwh19920920/snowflake"
	"github.com/xxl-job/xxl-job-executor-go"
	"gorm.io/gorm"
)

// Config 全局配置聚合
type Config struct {
	DatabaseForGorm *gorm.DB        // 数据库
	Sequence        *snowflake.Node // 雪花 ID
	AuthConfig      *auth.Config    // 权限配置
	Grafana         *grafana.Config // Grafana
	// TimeSeries 时序存储（读写）
	TimeSeries domainHandler.TimeSeriesStore
	// MetricQuery 指标命名 + Grafana 查询方言（与 TimeSeries 通常同一实现）
	MetricQuery domainHandler.MetricQueryDialect
	XxlJobExec  xxl.Executor // XXL-JOB 执行器（starter 注入）
}

// InitAll 初始化全部配置
func InitAll() Config {
	tsCfg := timeseries.Load()
	grafanaConf := grafana.InitGrafanaConfig()
	store, dialect := newTimeSeriesStack(tsCfg)

	return Config{
		DatabaseForGorm: database.GetConn(),
		Sequence:        sequence.GetSequence(),
		AuthConfig:      auth.GetAuthConf(),
		Grafana:         grafanaConf,
		TimeSeries:      store,
		MetricQuery:     dialect,
	}
}

// newTimeSeriesStack 按 backend 装配 Store + Dialect
func newTimeSeriesStack(tsCfg timeseries.Config) (domainHandler.TimeSeriesStore, domainHandler.MetricQueryDialect) {
	switch tsCfg.Backend {
	case timeseries.BackendVictoriaMetrics, "":
		vm := infraHandler.NewTimeSeriesVmHandler(victoriametrics.Load())
		return vm, vm
	case timeseries.BackendTDengine:
		td := infraHandler.NewTimeSeriesTDengineHandler(tdengine.Load())
		return td, td
	default:
		panic(fmt.Sprintf(
			"unknown timeseries.backend=%q, supported: %s, %s",
			tsCfg.Backend,
			timeseries.BackendVictoriaMetrics,
			timeseries.BackendTDengine,
		))
	}
}
