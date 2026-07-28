package application

import (
	"dragonfly-monitor/internal/infrastructure/system"
)

// SystemApplication 系统指标应用服务：包装 Collector，对外提供快照查询
type SystemApplication struct {
	collector *system.Collector
}

// NewSystemApplication 创建系统指标应用服务
func NewSystemApplication(collector *system.Collector) SystemApplication {
	return SystemApplication{collector: collector}
}

// Metrics 返回当前系统性能指标快照（当前值 + 历史趋势）
func (a *SystemApplication) Metrics() system.MetricsResponse {
	return a.collector.Snapshot()
}
