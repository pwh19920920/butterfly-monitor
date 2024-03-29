package entity

import "github.com/pwh19920920/butterfly-admin/common"

type DataSourceType int32

const (
	DataSourceTypeMongo  DataSourceType = 1
	DataSourceTypeMysql  DataSourceType = 2
	DataSourceTypeInflux DataSourceType = 3
)

type MonitorDatabase struct {
	common.BaseEntity

	Database string         `json:"database" gorm:"column:database"` // 数据库
	Name     string         `json:"name" gorm:"column:name"`         // 数据源名称
	Username string         `json:"username" gorm:"column:username"` // 数据库用户
	Password string         `json:"password" gorm:"column:password"` // 数据库密码
	Url      string         `json:"url" gorm:"column:url"`           // 数据库地址
	Type     DataSourceType `json:"type" gorm:"column:type"`         // 数据库类型
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorDatabase) TableName() string {
	return "t_monitor_database"
}
