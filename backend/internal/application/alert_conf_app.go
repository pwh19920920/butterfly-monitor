package application

import (
	"context"
	"strconv"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

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
		logger.Error(ctx, "AlertConfRepository.Select() happen error for", err)
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

// Cover2AlertConf 用独立 viper 实例将 KV 配置装配为运行期对象
func (app *AlertConfApplication) Cover2AlertConf(ctx context.Context) (*types.AlertConfObject, error) {
	data, err := app.repository.AlertConfRepository.SelectAll()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	// 默认值
	v.SetDefault("alert.firstDelay", int64(60))
	v.SetDefault("alert.alertSpan", int64(300))
	v.SetDefault("alert.simplePageSize", int64(50))
	v.SetDefault("alert.simpleMaxSecond", int64(600))
	v.SetDefault("alert.collectMaxSecond", int64(25))

	// 按 handler 维度收集默认模板：confKey 形如 template.ChannelEmailHandler
	templates := make(map[string]string)
	for _, item := range data {
		key := "alert." + item.ConfKey
		// 数字类型尝试转 int64
		if item.ConfType == entity.AlertConfTypeNumber {
			if n, err := strconv.ParseInt(item.ConfVal, 10, 64); err == nil {
				v.Set(key, n)
				continue
			}
		}
		v.Set(key, item.ConfVal)

		// template.<HandlerClassName> → Templates[HandlerClassName]
		if strings.HasPrefix(item.ConfKey, "template.") {
			handlerName := strings.TrimPrefix(item.ConfKey, "template.")
			if handlerName != "" {
				templates[handlerName] = item.ConfVal
			}
		}
	}

	obj := &types.AlertConfObject{}
	// 手动读取，避免 mapstructure 依赖路径问题
	obj.AlertSpan = v.GetInt64("alert.alertSpan")
	obj.FirstDelay = v.GetInt64("alert.firstDelay")
	// 兼容旧全局 key「template」
	obj.Template = v.GetString("alert.template")
	obj.Templates = templates
	obj.SimplePageSize = v.GetInt64("alert.simplePageSize")
	obj.SimpleMaxSecond = v.GetInt64("alert.simpleMaxSecond")
	obj.CollectMaxSecond = v.GetInt64("alert.collectMaxSecond")
	return obj, nil
}
