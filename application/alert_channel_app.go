package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

type AlertChannelApplication struct {
	repository *persistence.Repository
	sequence   *snowflake.Node
	commonMap  CommonMapApplication
}

func (app *AlertChannelApplication) Handlers() []types.AlertChannelHandlerResponse {
	alertChannelHandlers := make([]types.AlertChannelHandlerResponse, 0)

	for channelType, handlers := range app.commonMap.GetAlertChannelHandlerMap() {
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
func (app *AlertChannelApplication) Query(request *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error) {
	total, data, err := app.repository.AlertChannelRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Select() happen error for", err)
	}

	return total, data, err
}

// QueryAll 分页查询
func (app *AlertChannelApplication) QueryAll() ([]entity.AlertChannel, error) {
	data, err := app.repository.AlertChannelRepository.SelectAll()

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.SelectAll() happen error for", err)
	}
	return data, err
}

// Create 创建
func (app *AlertChannelApplication) Create(request *types.AlertChannelCreateRequest) error {
	alertChannel := request.AlertChannel
	alertChannel.Id = app.sequence.Generate().Int64()

	handle, ok := app.commonMap.GetAlertChannelHandlerNameMap()[alertChannel.Handler]
	if !ok {
		return errors.New("处理器不存在")
	}

	// 测试校验
	if err := handle.TestDispatchMessage(alertChannel, request.TestParams); err != nil {
		return err
	}

	err := app.repository.AlertChannelRepository.Save(&alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改
func (app *AlertChannelApplication) Modify(request *types.AlertChannelModifyRequest) error {
	alertChannel := request.AlertChannel

	handle, ok := app.commonMap.GetAlertChannelHandlerNameMap()[alertChannel.Handler]
	if !ok {
		return errors.New("处理器不存在")
	}

	// 测试校验
	if err := handle.TestDispatchMessage(alertChannel, request.TestParams); err != nil {
		return err
	}

	err := app.repository.AlertChannelRepository.Modify(alertChannel.Id, &alertChannel)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Modify() happen error", err)
	}
	return err
}

// Delete 修改
func (app *AlertChannelApplication) Delete(id int64) error {
	err := app.repository.AlertChannelRepository.Delete(id)

	// 错误记录
	if err != nil {
		logrus.Error("AlertChannelRepository.Delete() happen error", err)
	}
	return err
}
