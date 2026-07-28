package tdengine

import (
	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const (
	defaultAddr     = "http://127.0.0.1:6041"
	defaultUsername = "root"
	defaultPassword = "taosdata"
	defaultDatabase = "monitor"
)

// Config TDEngine 连接配置（REST，仅 DTO；实现见 infrastructure/handler）
type Config struct {
	Addr     string `yaml:"addr"`     // REST 地址，默认 http://127.0.0.1:6041
	Username string `yaml:"username"` // 默认 root
	Password string `yaml:"password"` // 默认 taosdata
	Database string `yaml:"database"` // 默认 monitor
}

type wrap struct {
	TDEngine Config `yaml:"tdEngine"`
}

// Load 加载配置
func Load() Config {
	viper.SetDefault("tdEngine.addr", defaultAddr)
	viper.SetDefault("tdEngine.username", defaultUsername)
	viper.SetDefault("tdEngine.password", defaultPassword)
	viper.SetDefault("tdEngine.database", defaultDatabase)
	conf := new(wrap)
	config.LoadConf(&conf)
	cfg := conf.TDEngine
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if cfg.Username == "" {
		cfg.Username = defaultUsername
	}
	if cfg.Database == "" {
		cfg.Database = defaultDatabase
	}
	return cfg
}
