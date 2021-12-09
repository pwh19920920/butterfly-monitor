package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/domain/handler"
	handlerImpl "butterfly-monitor/infrastructure/handler"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var alertChannelHandlerMap = make(map[entity.AlertChannelType][]handler.ChannelHandler, 0)
var alertChannelHandlerNameMap = make(map[string]handler.ChannelHandler, 0)

func init() {
	wechatHandler := handlerImpl.ChannelWechatHandler{}
	webHookHandlers := []handler.ChannelHandler{wechatHandler}
	alertChannelHandlerNameMap[wechatHandler.GetClassName()] = wechatHandler
	alertChannelHandlerMap[entity.AlertChannelTypeWebhook] = webHookHandlers

	emailHandler := handlerImpl.ChannelEmailHandler{}
	emailHandlers := []handler.ChannelHandler{emailHandler}
	alertChannelHandlerNameMap[emailHandler.GetClassName()] = emailHandler
	alertChannelHandlerMap[entity.AlertChannelTypeEmail] = emailHandlers
}

type AlertChannelApplication struct {
	repository *persistence.Repository
	sequence   *snowflake.Node
}

func (application *AlertChannelApplication) Handlers() []types.AlertChannelHandlerResponse {
	alertChannelHandlers := make([]types.AlertChannelHandlerResponse, 0)

	for channelType, handlers := range alertChannelHandlerMap {
		// 转换名字
		handlerNames := make([]string, 0)
		for _, channelHandler := range handlers {
			handlerNames = append(handlerNames, channelHandler.GetClassName())
		}

		alertChannelHandlers = append(alertChannelHandlers, types.AlertChannelHandlerResponse{
			ChannelType: channelType,
			Handlers:    handlerNames,
		})
	}
	return alertChannelHandlers
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

// QueryAll 分页查询
func (application *AlertChannelApplication) QueryAll() ([]entity.AlertChannel, error) {
	data, err := application.repository.AlertChannelRepository.SelectAll()

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.SelectAll() happen error for", err)
	}
	return data, err
}

// Create 创建
func (application *AlertChannelApplication) Create(request *types.AlertChannelCreateRequest) error {
	alertChannel := request.AlertChannel
	alertChannel.Id = application.sequence.Generate().Int64()

	handle, ok := alertChannelHandlerNameMap[alertChannel.Handler]
	if !ok {
		return errors.New("处理器不存在")
	}

	// 测试校验
	if err := handle.TestDispatchMessage(alertChannel, request.TestParams); err != nil {
		return err
	}

	err := application.repository.AlertChannelRepository.Save(&alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改
func (application *AlertChannelApplication) Modify(request *types.AlertChannelModifyRequest) error {
	alertChannel := request.AlertChannel

	handle, ok := alertChannelHandlerNameMap[alertChannel.Handler]
	if !ok {
		return errors.New("处理器不存在")
	}

	// 测试校验
	if err := handle.TestDispatchMessage(alertChannel, request.TestParams); err != nil {
		return err
	}

	err := application.repository.AlertChannelRepository.Modify(alertChannel.Id, &alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Modify() happen error", err)
	}
	return err
}

// Delete 修改
func (application *AlertChannelApplication) Delete(id int64) error {
	err := application.repository.AlertChannelRepository.Delete(id)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Delete() happen error", err)
	}
	return err
}
