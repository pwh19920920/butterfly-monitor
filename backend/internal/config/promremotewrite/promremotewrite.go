package promremotewrite

import (
	"strings"
	"time"

	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/spf13/viper"
)

const defaultRemoteWriteURL = "http://127.0.0.1:9090/api/v1/write"

// Config 远端 Prometheus 兼容 remote_write 接收端配置
//
// remote_write 协议要求：
//   - HTTP POST {URL}
//   - Content-Type: application/x-protobuf
//   - Content-Encoding: snappy
//   - Body: snappy-compressed prompb.WriteRequest
//
// 远端兼容对象：Prometheus 1.x/2.x 自身不直接接收 remote_write，
// 通常由 vmagent / Mimir / Thanos Receive / Grafana Cloud / OpenObserve 等前置组件接收，
// 故 endpoint 命名上不局限于 Prometheus。
type Config struct {
	// URL remote_write 端点，例如 http://prom-remote:8428/api/v1/write
	URL string `yaml:"url"`

	// QueryBase 可选；查询端点 base（平台读路径会请求 {QueryBase}/api/v1/query[_range]）。
	// 未配置时自动从 URL 剥掉尾部 /api/v1/write 推导——vmagent / Mimir / VM receiver 等
	// 标准部署 write 与 query 同源，推导正确。
	// 非标准路径（如 Grafana Cloud write=/api/prom/push、query=/api/prom）必须显式配置。
	QueryBase string `yaml:"queryBase"`

	// BasicAuth 可选；多数云端 / Mimir / Grafana Cloud 走 Bearer Token，
	// 因此额外暴露两个字段而非复用 Username/Password。
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	// BearerToken 可选；非空时优先于 BasicAuth 设置 Authorization: Bearer <token>
	BearerToken string `yaml:"bearerToken"`

	// TenantID 可选；Mimir / Cortex 等多租户后端使用 X-Scope-OrgID 头透传租户。
	TenantID string `yaml:"tenantId"`

	// RequestTimeout 单次 HTTP 请求超时（秒）；0 或负值默认 30s。
	RequestTimeout int `yaml:"requestTimeout"`

	// MaxRetries 写入失败重试次数（不含首次）；0 表示不重试。
	// 每次重试间隔按指数退避：1s, 2s, 4s ... 最多 30s。
	MaxRetries int `yaml:"maxRetries"`

	// ExternalLabels 静态附加到所有 series 的标签（X-Prometheus-Remote-Write-Version 之外的业务标签）。
	// remote_write 协议里通常用作集群/环境标记，例如 {cluster="prod", region="cn-east"}。
	// 注意：不要在 ExternalLabels 里放 __name__，指标名由 Metric 字段承担。
	ExternalLabels map[string]string `yaml:"externalLabels"`
}

type wrap struct {
	PromRemoteWrite Config `yaml:"promRemoteWrite"`
}

// Load 加载配置；缺省时给出可在本地起一个 vmagent 调试的默认值。
func Load() Config {
	viper.SetDefault("promRemoteWrite.url", defaultRemoteWriteURL)
	viper.SetDefault("promRemoteWrite.requestTimeout", 30)
	viper.SetDefault("promRemoteWrite.maxRetries", 2)
	w := new(wrap)
	config.LoadConf(&w)
	c := w.PromRemoteWrite
	if strings.TrimSpace(c.URL) == "" {
		c.URL = defaultRemoteWriteURL
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	return c
}

// Timeout 返回解析后的请求超时
func (c Config) Timeout() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}
