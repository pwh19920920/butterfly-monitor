package types

import (
	"strings"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type AlertConfQueryRequest struct {
	response.RequestPaging
	ConfKey string `form:"confKey"`
}

// AlertConfObject 运行期装配对象
type AlertConfObject struct {
	AlertSpan  int64 `mapstructure:"alertSpan"`
	FirstDelay int64 `mapstructure:"firstDelay"`
	// Template 兼容旧 key「template」的全局兜底模板
	Template string `mapstructure:"template"`
	// Templates 按通道 Handler 类名的默认模板，key 如 ChannelEmailHandler / ChannelWechatHandler
	// 对应 confKey: template.ChannelEmailHandler / template.ChannelWechatHandler
	Templates       map[string]string `mapstructure:"-"`
	SimplePageSize  int64             `mapstructure:"simplePageSize"`
	SimpleMaxSecond int64             `mapstructure:"simpleMaxSecond"`
	// CollectMaxSecond 采集单任务最大执行秒数。超时即 ctx 截断，防止拖垮下一个采集批次。
	CollectMaxSecond int64 `mapstructure:"collectMaxSecond"`
}

// ResolveTemplate 解析通道实际使用的模板：
// 1. 通道自身 Template 非空 → 用通道模板
// 2. 否则按 handler 取 Templates[handler]
// 3. 再兜底全局 Template
func (o *AlertConfObject) ResolveTemplate(channelTemplate, handler string) string {
	channelTemplate = strings.TrimSpace(channelTemplate)
	handler = strings.TrimSpace(handler)
	if o == nil {
		return channelTemplate
	}
	if channelTemplate != "" {
		return channelTemplate
	}
	if handler != "" && o.Templates != nil {
		if t, ok := o.Templates[handler]; ok && strings.TrimSpace(t) != "" {
			return t
		}
	}
	return o.Template
}
