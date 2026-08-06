package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type AlertConfQueryRequest struct {
	response.RequestPaging
	ConfKey string `form:"confKey"`
}

// 告警配置默认值。唯一来源：Cover2AlertConf 的 SetDefault 与各兜底均引用此处，避免散落魔法数。
const (
	DefaultAlertSpan                = int64(300)  // 事件下次报警间隔基数（秒），发送后按告警次数翻倍
	DefaultFirstDelay               = int64(60)   // 首次报警延迟（秒）
	DefaultSimplePageSize           = int64(50)   // 样本生成每页任务数
	DefaultSimpleMaxSecond          = int64(600)  // 样本生成最大回溯秒数
	DefaultCollectMaxSecond         = int64(25)   // 采集单任务最大执行秒数
	DefaultFreezeSampleLookBackDays = int64(3)    // 大促冻结基线回溯普通日最大天数
	DefaultMaxAlertSpan             = int64(1800) // 告警间隔翻倍退避上限（秒）
	DefaultPromoPeakRatio           = float64(1)  // 高峰（peak）上偏阈值放大系数，默认不放大
	DefaultPromoTroughRatio         = float64(1)  // 低谷（trough）样本差下偏阈值放大系数，默认不放大
	DefaultAlertCheckConcurrency    = int64(100)  // 告警检查单实例同时执行的规则数上限（防瞬时打爆 VM）
	DefaultSamplingConcurrency      = int64(80)   // 样本生成单实例同时执行的任务数上限（串行改并发，控 VM QPS）
	DefaultSampleRawDays            = int64(8)    // 样本原料投射未来天数：每次采集写入未来 N 天的原料点
	DefaultBatchWriteChunkSize      = int64(3000) // 时序写入单批最大点数
	DefaultMaxAlertShift            = int64(5)    // 告警间隔翻倍最大位移 2^N
)

// AlertConfObject 运行期装配对象，提供两种取数能力：
//   - 字段：常用项由 viper.Unmarshal 填充（AlertSpan / FirstDelay / ...）
//   - key：任意配置项通过 Get / GetFloat 按 confKey 取，用于模板等动态 key
type AlertConfObject struct {
	AlertSpan                int64             `json:"alertSpan"`
	FirstDelay               int64             `json:"firstDelay"`
	DefaultTemplate          string            `json:"defaultTemplate"`
	SimplePageSize           int64             `json:"simplePageSize"`
	SimpleMaxSecond          int64             `json:"simpleMaxSecond"`          // SimpleMaxSecond 样本生成最大回溯秒数
	CollectMaxSecond         int64             `json:"collectMaxSecond"`         // CollectMaxSecond 采集单任务最大执行秒数。超时即 ctx 截断，防止拖垮下一个采集批次。
	FreezeSampleLookBackDays int64             `json:"freezeSampleLookBackDays"` // FreezeSampleLookBackDays 大促冻结基线向前回溯普通日的最大天数（从「前一日」起算）。
	PromoPeakRatio           float64           `json:"promoPeakRatio"`           // PromoPeakRatio 敏感任务在特殊日对上偏阈值的放大系数；1=不放大，默认 1。
	PromoTroughRatio         float64           `json:"promoTroughRatio"`         // PromoTroughRatio 低谷（trough）对样本差类下偏阈值的放大系数；1=不放大，默认 1。
	AlertCheckConcurrency    int64             `json:"alertCheckConcurrency"`    // AlertCheckConcurrency 告警检查单实例并发上限
	SamplingConcurrency      int64             `json:"samplingConcurrency"`      // SamplingConcurrency 样本生成单实例并发上限
	SampleRawDays            int64             // SampleRawDays 样本原料投射未来天数
	BatchWriteChunkSize      int64             // BatchWriteChunkSize 时序写入单批最大点数
	MaxAlertShift            int64             // MaxAlertShift 告警间隔翻倍最大位移
	kv                       map[string]string // kv 库里实际配置的原始 KV，供按 key 取任意配置项（含模板等动态 key）
}

// Put 写入一条原始 KV（由 Cover2AlertConf 装配时逐条调用）。
func (o *AlertConfObject) Put(key, value string) {
	if o.kv == nil {
		o.kv = make(map[string]string)
	}
	o.kv[key] = value
}

// Get 按 confKey 取字符串；不存在返回 ""。适用于模板等动态 key。
func (o *AlertConfObject) Get(key string) string {
	if o == nil || o.kv == nil {
		return ""
	}
	return o.kv[key]
}

// GetFloat 按 confKey 取 float64；不存在 / 非法返回 0。
func (o *AlertConfObject) GetFloat(key string) float64 {
	s := strings.TrimSpace(o.Get(key))
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ResolveTemplate 解析通道实际使用的模板：
// 1. 通道自身 Template 非空 → 用通道模板
// 2. 否则按 handler 取 confKey default<Handler>Template，如 defaultChannelEmailHandlerTemplate
// 3. 再兜底全局 DefaultTemplate
func (o *AlertConfObject) ResolveTemplate(channelTemplate, handler string) string {
	channelTemplate = strings.TrimSpace(channelTemplate)
	handler = strings.TrimSpace(handler)
	if o == nil {
		return channelTemplate
	}

	// 取通道的模板
	if channelTemplate != "" {
		return channelTemplate
	}

	// 取通道默认模板
	if handler == "" {
		return o.DefaultTemplate
	}

	// 回退
	fullName := fmt.Sprintf("default%sTemplate", handler)
	if defaultSpecTemplate := strings.TrimSpace(o.Get(fullName)); defaultSpecTemplate != "" {
		return defaultSpecTemplate
	}
	return o.DefaultTemplate
}
