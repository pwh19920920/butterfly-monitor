package config

import (
	"fmt"

	"butterfly-monitor/internal/config/auth"
	"butterfly-monitor/internal/config/database"
	"butterfly-monitor/internal/config/grafana"
	"butterfly-monitor/internal/config/promremotewrite"
	"butterfly-monitor/internal/config/sequence"
	"butterfly-monitor/internal/config/tdengine"
	"butterfly-monitor/internal/config/timeseries"
	"butterfly-monitor/internal/config/victoriametrics"
	domainHandler "butterfly-monitor/internal/domain/handler"
	infraHandler "butterfly-monitor/internal/infrastructure/handler"

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
	case timeseries.BackendPromRemoteWrite:
		// remote_write 是出站写入协议，不带本地查询能力。
		// Dialect 仍沿用 VictoriaMetrics 的 PromQL 方言——指标命名 (taskKey / _sampling / _sample)
		// 与 VM 完全一致，远端任何兼容 PromQL 的存储都能用同一套 Grafana 表达式。
		rw := infraHandler.NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Load())
		vmDialect := infraHandler.NewTimeSeriesVmHandler(victoriametrics.Load())
		return rw, vmDialect
	default:
		panic(fmt.Sprintf(
			"unknown timeseries.backend=%q, supported: %s, %s, %s",
			tsCfg.Backend,
			timeseries.BackendVictoriaMetrics,
			timeseries.BackendTDengine,
			timeseries.BackendPromRemoteWrite,
		))
	}
}
