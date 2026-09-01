package application

//goland:noinspection GoUnsortedImport,GoUnsortedImport,GoUnsortedImport,GoUnsortedImport,GoUnsortedImport,GoUnsortedImport,GoUnsortedImport
import (
	"context"
	"errors"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"
	"github.com/pwh19920920/snowflake"
)

// AlertChannelApplication 告警通道应用服务
type AlertChannelApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	commonMap  *CommonMapApplication
	alertConf  AlertConfApplication
}

// NewAlertChannelApplication 创建告警通道应用服务
func NewAlertChannelApplication(
	sequence *snowflake.Node,
	repository *persistence.Repository,
	commonMap *CommonMapApplication,
	alertConf AlertConfApplication,
) AlertChannelApplication {
	return AlertChannelApplication{
		sequence:   sequence,
		repository: repository,
		commonMap:  commonMap,
		alertConf:  alertConf,
	}
}

// Query 分页查询
func (app *AlertChannelApplication) Query(ctx context.Context, req *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error) {
	total, data, err := app.repository.AlertChannelRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "AlertChannelRepository.Select() happen error for", err)
		return total, nil, err
	}
	return total, data, nil
}

// QueryAll 全量查询
func (app *AlertChannelApplication) QueryAll(ctx context.Context) ([]entity.AlertChannel, error) {
	data, err := app.repository.AlertChannelRepository.SelectAll()
	if err != nil {
		logger.Error(ctx, "AlertChannelRepository.SelectAll() happen error for", err)
	}
	return data, err
}

// Handlers 枚举通道类型与可用处理器的绑定关系
func (app *AlertChannelApplication) Handlers(ctx context.Context) []types.AlertChannelHandlerVO {
	return app.commonMap.GetAlertChannelHandlerMap(ctx)
}

// Create 创建通道，testParams 仅用于本次测试发送，不入库
func (app *AlertChannelApplication) Create(ctx context.Context, channel *entity.AlertChannel, testParams types.AlertChannelTestParams) error {
	if channel.Name == "" {
		return errors.New("通道名称不能为空")
	}
	if channel.Handler == "" {
		return errors.New("处理器不能为空")
	}
	nameMap := app.commonMap.GetAlertChannelHandlerNameMap(ctx)
	if !nameMap[channel.Handler] {
		return errors.New("处理器不存在")
	}
	// 若已注入真实 handler，尝试测试发送
	if err := app.tryTestDispatch(ctx, channel, testParams); err != nil {
		return err
	}
	channel.Id = app.sequence.Generate().Int64()
	return app.repository.AlertChannelRepository.Save(channel)
}

// Modify 修改通道，testParams 仅用于本次测试发送，不入库
func (app *AlertChannelApplication) Modify(ctx context.Context, channel *entity.AlertChannel, testParams types.AlertChannelTestParams) error {
	if channel.Id == 0 {
		return errors.New("id 不能为空")
	}
	if channel.Handler != "" {
		nameMap := app.commonMap.GetAlertChannelHandlerNameMap(ctx)
		if !nameMap[channel.Handler] {
			return errors.New("处理器不存在")
		}
		if h, ok := app.commonMap.GetChannelHandler(ctx, channel.Handler); ok {
			msg, err := app.buildTestMessage(ctx, channel)
			if err != nil {
				return err
			}
			if err := h.TestDispatchMessage(*channel, testParams.Email, msg); err != nil {
				return err
			}
		}
	}
	return app.repository.AlertChannelRepository.Modify(channel.Id, channel)
}

// tryTestDispatch 尝试发送测试消息，Create 和 Modify 共用。
func (app *AlertChannelApplication) tryTestDispatch(ctx context.Context, channel *entity.AlertChannel, testParams types.AlertChannelTestParams) error {
	if h, ok := app.commonMap.GetChannelHandler(ctx, channel.Handler); ok {
		msg, err := app.buildTestMessage(ctx, channel)
		if err != nil {
			return err
		}
		if err := h.TestDispatchMessage(*channel, testParams.Email, msg); err != nil {
			return err
		}
	}
	return nil
}

// buildTestMessage 用通道模板（或 handler 默认模板）+ 假参数渲染测试消息
func (app *AlertChannelApplication) buildTestMessage(ctx context.Context, channel *entity.AlertChannel) (string, error) {
	conf, err := app.alertConf.Cover2AlertConf(ctx)
	if err != nil {
		return "", err
	}
	tpl := conf.ResolveTemplate(channel.Template, channel.Handler)
	if tpl == "" {
		// 无模板时直接发固定文案
		return "butterfly-monitor 通道测试消息", nil
	}
	// 假参数：模拟一次真实告警，便于预览模板效果
	msg, err := common.RenderAlertTemplate(
		tpl,
		"【测试】示例监控任务",
		"实时数值99，高于阈值10，已持续发生60秒",
		"上游依赖A -> 上游依赖B",
	)
	if err != nil {
		return "", errors.New("模板渲染失败: " + err.Error())
	}
	return msg, nil
}

// Count 统计通道总数
func (app *AlertChannelApplication) Count(ctx context.Context) (int64, error) {
	return app.repository.AlertChannelRepository.Count()
}
