package grafana

import (
	"github.com/pwh19920920/butterfly/config"
	sdk "github.com/pwh19920920/grafanasdk"
	"github.com/spf13/viper"
)

const defaultAddr = "http://127.0.0.1:3000"
const defaultSampleRpName = "expire_7d"

type Config struct {
	Addr         string `yaml:"addr"`         // 地址
	ApiKey       string `yaml:"apiKey"`       // 密钥
	SampleRpName string `yaml:"sampleRpName"` // 样本rp名字
}

type grafanaConf struct {
	Grafana Config `yaml:"grafana"`
}

// InitGrafanaConfig 获取连接
func InitGrafanaConfig() *Config {
	// 默认值
	viper.SetDefault("grafana.addr", defaultAddr)
	viper.SetDefault("grafana.sampleRpName", defaultSampleRpName)

	// 加载配置
	gfConf := new(grafanaConf)
	config.LoadConf(&gfConf)

	return &gfConf.Grafana
}

func (conf *Config) GetGrafanaClient() (*sdk.Client, error) {
	return sdk.NewClient(conf.Addr, conf.ApiKey, sdk.DefaultHTTPClient)
}
