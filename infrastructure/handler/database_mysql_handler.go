package handler

import (
	"butterfly-monitor/common"
	"butterfly-monitor/domain/entity"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

func (dbHandler *DatabaseMysqlHandler) TestConnect(database entity.MonitorDatabase) error {
	// 创建连接
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		database.Username, database.Password,
		database.Url, database.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil || db == nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure", database.Url, database.Database))
	}

	err = db.Ping()
	defer db.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}
	return nil
}

// NewInstance username:password@tcp(127.0.0.1:3306)/butterfly_admin?charset=utf8mb4&parseTime=True&loc=Local
func (dbHandler *DatabaseMysqlHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	// 创建连接
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		database.Username, database.Password,
		database.Url, database.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 common.NewGormLogger(),
	})

	if err != nil || db == nil {
		logrus.Errorf("%s - %s db connect open failure", database.Url, database.Database)
		return nil, errors.New(fmt.Sprintf("%s - %s db connect open failure", database.Url, database.Database))
	}

	// 关闭sql log
	//	db.Logger = common.NewGormLogger()
	db.Logger.LogMode(logger.Silent)

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
	var result float64 = 0
	err := dbHandler.db.Raw(task.Command).Scan(&result).Error
	return result, err
}
