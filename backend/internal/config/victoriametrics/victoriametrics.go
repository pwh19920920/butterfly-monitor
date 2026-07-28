package victoriametrics

import (
	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const defaultAddr = "http://127.0.0.1:8428"

// Config VictoriaMetrics 连接配置（仅 DTO，业务实现在 infrastructure/handler）
type Config struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type vmConf struct {
	VictoriaMetrics Config `yaml:"victoriaMetrics"`
}

// Load 加载配置
func Load() Config {
	viper.SetDefault("victoriaMetrics.addr", defaultAddr)
	conf := new(vmConf)
	config.LoadConf(&conf)
	return conf.VictoriaMetrics
}
