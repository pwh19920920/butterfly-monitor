package grafana

import (
	"github.com/pwh19920920/butterfly/config"
	"github.com/spf13/viper"
)

const defaultAddr = "http://127.0.0.1:3000"

type Config struct {
	Addr   string `yaml:"addr"`   // 地址
	ApiKey string `yaml:"apiKey"` // 密钥
}

type grafanaConf struct {
	Grafana Config `yaml:"grafana"`
}

// InitGrafanaConfig 获取连接
func InitGrafanaConfig() *Config {
	// 默认值
	viper.SetDefault("grafana.addr", defaultAddr)

	// 加载配置
	gfConf := new(grafanaConf)
	config.LoadConf(&gfConf, config.GetOptions().ConfigFilePath)

	return &gfConf.Grafana
}
