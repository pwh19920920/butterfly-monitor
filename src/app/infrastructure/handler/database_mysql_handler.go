package handler

import (
	"butterfly-monitor/src/app/domain/entity"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

const defaultMaxIdleConnect = 10
const defaultMaxOpenConnect = 100
const defaultConnMaxLifeTimeSecond = 3600

type DatabaseMysqlHandler struct {
	db *gorm.DB
}

// NewInstance username:password@tcp(127.0.0.1:3306)/butterfly_admin?charset=utf8mb4&parseTime=True&loc=Local
func (dbHandler *DatabaseMysqlHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	// 创建连接
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		database.Username, database.Password,
		database.Url, database.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil || db == nil {
		logrus.Error("db connect open failure")
		return nil, errors.New("db connect open failure")
	}

	// 关闭sql log
	db.Logger = logger.Default.LogMode(logger.Silent)

	// 打开连接
	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		logrus.Error("db open failure")
		return nil, errors.New("db open failure")
	}

	// 连接池设置
	// SetMaxIdleCons 设置连接池中的最大闲置连接数。
	sqlDB.SetMaxIdleConns(defaultMaxIdleConnect)

	// SetMaxOpenCons 设置数据库的最大连接数量。
	sqlDB.SetMaxOpenConns(defaultMaxOpenConnect)

	// SetConnMaxLifeTime 设置连接的最大可复用时间。
	sqlDB.SetConnMaxLifetime(time.Duration(defaultConnMaxLifeTimeSecond) * time.Second)

	// 返回对象
	return &DatabaseMysqlHandler{db: db}, nil
}

// ExecuteQuery 执行查询
func (dbHandler *DatabaseMysqlHandler) ExecuteQuery(task entity.MonitorTask) (interface{}, error) {
	var result interface{} = 0
	err := dbHandler.db.Raw(task.Command).Scan(&result).Error
	return result, err
}
