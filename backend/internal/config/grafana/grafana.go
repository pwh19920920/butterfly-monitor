package grafana

import (
	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const defaultAddr = "http://127.0.0.1:3000"

// Config Grafana 配置
// 指标命名与查询方言已迁至 domain/handler.MetricQueryDialect
// type 由 timeseries.backend 对应的 dialect 决定（prometheus / tdengine-datasource）
type Config struct {
	Addr   string `yaml:"addr"`
	ApiKey string `yaml:"apiKey"`
	// DatasourceUid 可选：直接绑定的数据源 UID。
	// 未配置时会调 Grafana /api/datasources 按 dialect 类型自动解析（该接口通常需要 Admin）。
	// 若 API Key 只有 Editor，建议在此写死真实 UID。
	DatasourceUid string `yaml:"datasourceUid"`
}

type grafanaConf struct {
	Grafana Config `yaml:"grafana"`
}

// InitGrafanaConfig 加载 Grafana 配置
func InitGrafanaConfig() *Config {
	viper.SetDefault("grafana.addr", defaultAddr)
	gfConf := new(grafanaConf)
	config.LoadConf(&gfConf)
	return &gfConf.Grafana
}
