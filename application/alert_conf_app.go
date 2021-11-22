package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/config/sequence"
	"github.com/sirupsen/logrus"
	"strconv"
)

type AlertConfApplication struct {
	repository *persistence.Repository
}

type AlertConfObject struct {
	AlertSpan int64  // 报警间隔
	Template  string // 报警模板
}

// Cover2AlertConf 转换配置
func (application *AlertConfApplication) Cover2AlertConf(dataMap map[string]string) (*AlertConfObject, error) {
	const defaultAlertSpan int64 = 120
	alertSpan := defaultAlertSpan
	alertSpanVal, ok := dataMap["alertSpan"]
	if ok {
		// 校验转换结果, 其次判断是否大于最小值
		alertSpan, err := strconv.ParseInt(alertSpanVal, 10, 64)
		if err != nil || alertSpan < defaultAlertSpan {
			alertSpan = defaultAlertSpan
		}
	}

	return &AlertConfObject{
		AlertSpan: alertSpan,
		Template:  dataMap["template"],
	}, nil
}

// SelectConf 分页查询
func (application *AlertConfApplication) SelectConf() (*AlertConfObject, error) {
	data, err := application.repository.AlertConfRepository.SelectAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, 0)
	for _, item := range data {
		result[item.ConfKey] = item.ConfVal
	}
	return application.Cover2AlertConf(result)
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
	alertConf.Id = sequence.GetSequence().Generate().Int64()
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
