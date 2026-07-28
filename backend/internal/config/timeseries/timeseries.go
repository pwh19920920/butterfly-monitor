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
