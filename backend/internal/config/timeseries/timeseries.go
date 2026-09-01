package timeseries

import (
	"strings"

	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const (
	// BackendVictoriaMetrics 默认后端
	BackendVictoriaMetrics = "victoriaMetrics"
	// BackendTDengine TDengine（REST SQL）
	BackendTDengine = "tdengine"
	// BackendPromRemoteWrite 通过 Prometheus remote_write 协议推送到远端
	// （云端 Prometheus / vmagent / Mimir / Thanos Receive / Grafana Cloud 等兼容端点）。
	// 注意：本后端仅负责出站写入，查询仍需由前端 Grafana 直接对接远端，
	// 因此平台内的 QueryMean/QueryRange 等读路径会落到由 MetricQuery Dialect 提供的实现上
	// （本后端 Dialect 与 VictoriaMetrics 等价——指标命名一致，PromQL 兼容）。
	BackendPromRemoteWrite = "promRemoteWrite"
)

// Config 时序后端选择
type Config struct {
	Backend string `yaml:"backend"`
}

type wrap struct {
	Timeseries Config `yaml:"timeseries"`
}

// Load 加载 timeseries 配置，默认 victoriaMetrics
func Load() Config {
	viper.SetDefault("timeseries.backend", BackendVictoriaMetrics)
	w := new(wrap)
	config.LoadConf(&w)
	backend := strings.TrimSpace(w.Timeseries.Backend)
	if backend == "" {
		backend = BackendVictoriaMetrics
	}
	return Config{Backend: backend}
}
