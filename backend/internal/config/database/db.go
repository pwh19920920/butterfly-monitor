package database

import (
	"dragonfly-monitor/internal/common"
	"time"

	"github.com/pwh19920920/butterfly/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultMaxIdleConnect = 10
const defaultMaxOpenConnect = 100
const defaultConnMaxLifeTimeSecond = 3600

type Config struct {
	Dsn                   string `yaml:"dsn"`                   // 数据库链接
	MaxIdleConnect        int    `yaml:"maxIdleConnect"`        // 最大空闲时间
	MaxOpenConnect        int    `yaml:"maxOpenConnect"`        // 最大连接时间
	ConnMaxLifeTimeSecond int    `yaml:"connMaxLifeTimeSecond"` //
}

type dbConf struct {
	Db Config `yaml:"db"`
}

// GetConn 初始化db
func GetConn() *gorm.DB {
	// 默认值
	viper.SetDefault("db.maxIdleConnect", defaultMaxIdleConnect)
	viper.SetDefault("db.maxOpenConnect", defaultMaxOpenConnect)
	viper.SetDefault("db.ConnMaxLifeTimeSecond", defaultConnMaxLifeTimeSecond)

	// 加载配置
	databaseConf := new(dbConf)
	config.LoadConf(&databaseConf)

	// 创建连接
	db, err := gorm.Open(mysql.Open(databaseConf.Db.Dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 common.NewGormLogger(),
	})

	if err != nil || db == nil {
		logrus.Error("db connect open failure")
		return nil
	}

	// 关闭sql log
	db.Logger.LogMode(logger.Silent)

	// 打开连接
	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		logrus.Error("db open failure")
		return nil
	}

	// 连接池设置
	// SetMaxIdleCons 设置连接池中的最大闲置连接数。
	sqlDB.SetMaxIdleConns(databaseConf.Db.MaxIdleConnect)

	// SetMaxOpenCons 设置数据库的最大连接数量。
	sqlDB.SetMaxOpenConns(databaseConf.Db.MaxOpenConnect)

	// SetConnMaxLifeTime 设置连接的最大可复用时间。
	sqlDB.SetConnMaxLifetime(time.Duration(databaseConf.Db.ConnMaxLifeTimeSecond) * time.Second)

	// 初始化全局
	return db
}
