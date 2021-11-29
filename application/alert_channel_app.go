package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/domain/handler"
	handlerImpl "butterfly-monitor/infrastructure/handler"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var alertChannelHandlerMap = make(map[entity.AlertChannelType][]handler.ChannelHandler, 0)

func init() {
	webHookHandlers := []handler.ChannelHandler{handlerImpl.ChannelWechatHandler{}}
	alertChannelHandlerMap[entity.AlertChannelTypeWebhook] = webHookHandlers

	emailHandlers := []handler.ChannelHandler{handlerImpl.ChannelEmailHandler{}}
	alertChannelHandlerMap[entity.AlertChannelTypeEmail] = emailHandlers
}

type AlertChannelApplication struct {
	repository *persistence.Repository
	sequence   *snowflake.Node
}

// Query 分页查询
func (application *AlertChannelApplication) Query(request *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error) {
	total, data, err := application.repository.AlertChannelRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Select() happen error for", err)
	}

	return total, data, err
}

// Create 创建数据源
func (application *AlertChannelApplication) Create(request *types.AlertChannelCreateRequest) error {
	alertChannel := request.AlertChannel
	alertChannel.Id = application.sequence.Generate().Int64()
	err := application.repository.AlertChannelRepository.Save(&alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *AlertChannelApplication) Modify(request *types.AlertModifyModifyRequest) error {
	alertChannel := request.AlertChannel
	err := application.repository.AlertChannelRepository.Modify(alertChannel.Id, &alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Modify() happen error", err)
	}
	return err
}
