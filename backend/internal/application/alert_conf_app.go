package application

import (
	"context"

	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
	"github.com/spf13/viper"
)

// AlertConfApplication 告警配置应用服务
type AlertConfApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewAlertConfApplication 创建告警配置应用服务
func NewAlertConfApplication(sequence *snowflake.Node, repository *persistence.Repository) AlertConfApplication {
	return AlertConfApplication{sequence: sequence, repository: repository}
}

// Query 分页查询
func (app *AlertConfApplication) Query(ctx context.Context, req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error) {
	total, data, err := app.repository.AlertConfRepository.Select(req)
	if err != nil {
		return total, nil, err
	}
	return total, data, nil
}

// Create 新建配置项
func (app *AlertConfApplication) Create(ctx context.Context, conf *entity.AlertConf) error {
	conf.Id = app.sequence.Generate().Int64()
	return app.repository.AlertConfRepository.Save(conf)
}

// Modify 修改配置项
func (app *AlertConfApplication) Modify(ctx context.Context, conf *entity.AlertConf) error {
	return app.repository.AlertConfRepository.Modify(conf.Id, conf)
}

// Cover2AlertConf 用独立 viper 实例将 KV 配置装配为运行期对象。
// 常用项由 viper.Unmarshal 映射到字段；全部 KV（含模板等动态 key）通过 obj.Put 保留，供 Get/GetFloat 按 key 取。
func (app *AlertConfApplication) Cover2AlertConf(ctx context.Context) (*types.AlertConfObject, error) {
	data, err := app.repository.AlertConfRepository.SelectAll()
	if err != nil {
		return nil, err
	}

	conf := viper.New()
	// 默认值（key 与 conf_key 一致，不带 alert. 前缀；取值见 types.Default* 常量）
	conf.SetDefault("firstDelay", types.DefaultFirstDelay)
	conf.SetDefault("alertSpan", types.DefaultAlertSpan)
	conf.SetDefault("simplePageSize", types.DefaultSimplePageSize)
	conf.SetDefault("simpleMaxSecond", types.DefaultSimpleMaxSecond)
	conf.SetDefault("collectMaxSecond", types.DefaultCollectMaxSecond)
	conf.SetDefault("freezeSampleLookBackDays", types.DefaultFreezeSampleLookBackDays)
	conf.SetDefault("promoPeakRatio", types.DefaultPromoPeakRatio)
	conf.SetDefault("promoTroughRatio", types.DefaultPromoTroughRatio)
	conf.SetDefault("alertCheckConcurrency", types.DefaultAlertCheckConcurrency)
	conf.SetDefault("samplingConcurrency", types.DefaultSamplingConcurrency)
	conf.SetDefault("sampleRawDays", types.DefaultSampleRawDays)
	conf.SetDefault("batchWriteChunkSize", types.DefaultBatchWriteChunkSize)
	conf.SetDefault("maxAlertShift", types.DefaultMaxAlertShift)

	obj := &types.AlertConfObject{}
	for _, item := range data {
		obj.Put(item.ConfKey, item.ConfVal)
		conf.Set(item.ConfKey, item.ConfVal)
	}
	// viper.Unmarshal 默认弱类型输入：配置表里存的字符串可自动转 int64
	if err := conf.Unmarshal(obj); err != nil {
		return nil, err
	}

	// 字段规范化：显式配 0/负等非法值回退默认，保证下游直接取字段即可，无需各自防御
	obj.SimplePageSize = normalizeInt(obj.SimplePageSize, types.DefaultSimplePageSize)
	obj.SimpleMaxSecond = normalizeInt(obj.SimpleMaxSecond, types.DefaultSimpleMaxSecond)
	obj.CollectMaxSecond = normalizeInt(obj.CollectMaxSecond, types.DefaultCollectMaxSecond)
	obj.AlertSpan = normalizeInt(obj.AlertSpan, types.DefaultAlertSpan)
	obj.FirstDelay = normalizeInt(obj.FirstDelay, types.DefaultFirstDelay)
	obj.FreezeSampleLookBackDays = normalizeInt(obj.FreezeSampleLookBackDays, types.DefaultFreezeSampleLookBackDays)
	obj.PromoPeakRatio = normalizeFloat(obj.PromoPeakRatio, types.DefaultPromoPeakRatio)
	obj.PromoTroughRatio = normalizeFloat(obj.PromoTroughRatio, types.DefaultPromoTroughRatio)
	obj.AlertCheckConcurrency = normalizeInt(obj.AlertCheckConcurrency, types.DefaultAlertCheckConcurrency)
	obj.SamplingConcurrency = normalizeInt(obj.SamplingConcurrency, types.DefaultSamplingConcurrency)
	obj.SampleRawDays = normalizeInt(obj.SampleRawDays, types.DefaultSampleRawDays)
	obj.BatchWriteChunkSize = normalizeInt(obj.BatchWriteChunkSize, types.DefaultBatchWriteChunkSize)
	obj.MaxAlertShift = normalizeInt(obj.MaxAlertShift, types.DefaultMaxAlertShift)
	return obj, nil
}

// normalizeInt 非正数回退默认值（配置项显式配 0/负视为非法，装配期统一规整）
func normalizeInt(v, def int64) int64 {
	if v <= 0 {
		return def
	}
	return v
}

// normalizeFloat 非正数回退默认值（配置项显式配 0/负视为非法，装配期统一规整）
func normalizeFloat(v, def float64) float64 {
	if v <= 0 {
		return def
	}
	return v
}
