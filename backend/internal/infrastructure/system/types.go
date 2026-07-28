package system

// MetricSeries 单个指标的当前值 + 历史
// History 为环形缓冲里的历史点，老→新有序，最多 maxHistory 点(30s×60≈30min)
type MetricSeries struct {
	Current float64   `json:"current"`
	History []float64 `json:"history"`
}

// MetricsResponse 首页系统性能指标响应
type MetricsResponse struct {
	CpuPercent     MetricSeries `json:"cpuPercent"`     // CPU 使用率 %
	MemPercent     MetricSeries `json:"memPercent"`     // 内存使用率 %
	DiskPercent    MetricSeries `json:"diskPercent"`    // 磁盘使用率 %
	GoroutineCount MetricSeries `json:"goroutineCount"` // Go goroutine 数
	GoGcCount      MetricSeries `json:"goGcCount"`      // Go GC 累计次数
	ProcessRssMB   MetricSeries `json:"processRssMB"`   // 进程 RSS 内存 MB
}
