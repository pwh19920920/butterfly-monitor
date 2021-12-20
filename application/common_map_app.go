package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/domain/handler"
	handlerImpl "butterfly-monitor/infrastructure/handler"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/infrastructure/support"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/common"
	"time"
)

var commandHandlerMap = make(map[entity.MonitorTaskType]handler.CommandHandler, 0)
var databaseHandlerMap = make(map[entity.DataSourceType]handler.DatabaseHandler, 0)

var databaseLoadTime *common.LocalTime
var databaseMap = make(map[int64]interface{}, 0)

var alertChannelHandlerMap = make(map[entity.AlertChannelType][]handler.ChannelHandler, 0)
var alertChannelHandlerNameMap = make(map[string]handler.ChannelHandler, 0)

type CommonMapApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaOptionHandler
}

// 默认参数
func init() {
	// 命令类型
	commandHandlerMap[entity.TaskTypeURL] = new(handlerImpl.CommandUrlHandler)
	commandHandlerMap[entity.TaskTypeDatabase] = new(handlerImpl.CommandDataBaseHandler)

	// 数据库类型
	databaseHandlerMap[entity.DataSourceTypeMysql] = new(handlerImpl.DatabaseMysqlHandler)

	// 报警通道
	wechatHandler := handlerImpl.ChannelWechatHandler{}
	webHookHandlers := []handler.ChannelHandler{wechatHandler}
	alertChannelHandlerNameMap[wechatHandler.GetClassName()] = wechatHandler
	alertChannelHandlerMap[entity.AlertChannelTypeWebhook] = webHookHandlers

	emailHandler := handlerImpl.ChannelEmailHandler{}
	emailHandlers := []handler.ChannelHandler{emailHandler}
	alertChannelHandlerNameMap[emailHandler.GetClassName()] = emailHandler
	alertChannelHandlerMap[entity.AlertChannelTypeEmail] = emailHandlers
}

func NewCommonMapApplication(repository *persistence.Repository) CommonMapApplication {
	go initDatabaseConnect(repository)
	return CommonMapApplication{}
}

func (commonMap CommonMapApplication) GetDatabaseHandlerMap() map[entity.DataSourceType]handler.DatabaseHandler {
	return databaseHandlerMap
}

func (commonMap CommonMapApplication) GetCommandHandlerMap() map[entity.MonitorTaskType]handler.CommandHandler {
	return commandHandlerMap
}

func (commonMap CommonMapApplication) GetAlertChannelHandlerNameMap() map[string]handler.ChannelHandler {
	return alertChannelHandlerNameMap
}

func (commonMap CommonMapApplication) GetAlertChannelHandlerMap() map[entity.AlertChannelType][]handler.ChannelHandler {
	return alertChannelHandlerMap
}

func initDatabaseConnect(repository *persistence.Repository) {
	databaseList, err := repository.MonitorDatabaseRepository.SelectAll(databaseLoadTime)
	if err != nil {
		return
	}

	for _, database := range databaseList {
		databaseHandler, ok := databaseHandlerMap[database.Type]
		if !ok {
			continue
		}

		dbHandler, err := databaseHandler.NewInstance(database)
		if err != nil {
			// 失败得情况下需要更新一下，以便下一次定时扫新连接得时候重新再连接
			_ = repository.MonitorDatabaseRepository.UpdateById(database.Id, &database)
			continue
		}
		databaseMap[database.Id] = dbHandler
	}

	// 初始化执行类型
	commandHandlerMap[entity.TaskTypeDatabase] = &handlerImpl.CommandDataBaseHandler{DatabaseMap: databaseMap}

	// 睡眠后继续执行
	databaseLoadTime = &common.LocalTime{Time: time.Now()}
	time.Sleep(time.Duration(1) * time.Minute)
	go initDatabaseConnect(repository)
}
