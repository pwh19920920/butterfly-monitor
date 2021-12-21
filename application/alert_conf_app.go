package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"fmt"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AlertConfApplication struct {
	repository *persistence.Repository
	sequence   *snowflake.Node
}

const DefaultAlertSpan = 600 // 默认报警间隔 10分钟
const DefaultFirstDelay = 60 // 默认首次延迟60s

type AlertConfObject struct {
	AlertSpan  int64  `json:"alertSpan"`  // 报警间隔
	FirstDelay int64  `json:"firstDelay"` // 首次延迟
	Template   string `json:"template"`   // 报警模板
}

type AlertConfObjectInstance struct {
	Alert AlertConfObject `json:"alert"`
}

// Cover2AlertConf 转换配置
func (application *AlertConfApplication) Cover2AlertConf(data []entity.AlertConf) (*AlertConfObjectInstance, error) {
	conf := viper.New()
	conf.SetDefault("alert.firstDelay", DefaultFirstDelay)
	conf.SetDefault("alert.alertSpan", DefaultAlertSpan)

	for _, item := range data {
		conf.Set(fmt.Sprintf("alert.%s", item.ConfKey), item.ConfVal)
	}

	confInstance := new(AlertConfObjectInstance)
	err := conf.Unmarshal(&confInstance)
	return confInstance, err
}

// SelectConf 分页查询
func (application *AlertConfApplication) SelectConf() (*AlertConfObjectInstance, error) {
	data, err := application.repository.AlertConfRepository.SelectAll()
	if err != nil {
		return nil, err
	}
	return application.Cover2AlertConf(data)
}

// Query 分页查询
func (application *AlertConfApplication) Query(request *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error) {
	total, data, err := application.repository.AlertConfRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("AlertConfRepository.Select() happen error for", err)
	}

	return total, data, err
}

// Create 创建数据源
func (application *AlertConfApplication) Create(request *types.AlertConfCreateRequest) error {
	alertConf := request.AlertConf
	alertConf.Id = application.sequence.Generate().Int64()
	err := application.repository.AlertConfRepository.Save(&alertConf)

	// 错误记录
	if err != nil {
		logrus.Error("AlertConfRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *AlertConfApplication) Modify(request *types.AlertConfModifyRequest) error {
	alertConf := request.AlertConf
	err := application.repository.AlertConfRepository.Modify(alertConf.Id, &alertConf)

	// 错误记录
	if err != nil {
		logrus.Error("AlertConfRepository.Modify() happen error", err)
	}
	return err
}
