package influxdb

import (
	"github.com/pwh19920920/butterfly/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)
import client "github.com/influxdata/influxdb1-client/v2"

const defaultPrecision = "s"
const defaultUsername = "admin"
const defaultPassword = ""
const defaultAddr = "http://127.0.0.1:8086"

type Config struct {
	Addr      string `yaml:"addr"`      // 地址
	Username  string `yaml:"username"`  // 用户名
	Password  string `yaml:"password"`  // 密码
	Database  string `yaml:"database"`  // 数据库
	Precision string `yaml:"precision"` // 精度
}

type influxConf struct {
	Influx Config `yaml:"influx"`
}

type DbOption struct {
	DbConf *influxConf
}

func (op *DbOption) GetClient() client.Client {
	// 创建client
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     op.DbConf.Influx.Addr,
		Username: op.DbConf.Influx.Username,
		Password: op.DbConf.Influx.Password,
	})

	// 判断错误
	if err != nil {
		logrus.Panic("influx connect open failure", err)
	}
	return cli
}

func NewInfluxDbOption() *DbOption {
	return &DbOption{
		getDbConf(),
	}
}

// getConn 获取连接
func getDbConf() *influxConf {

	// 默认值
	viper.SetDefault("influx.addr", defaultAddr)
	viper.SetDefault("influx.precision", defaultPrecision)
	viper.SetDefault("influx.username", defaultUsername)
	viper.SetDefault("influx.password", defaultPassword)

	// 加载配置
	dbConf := new(influxConf)
	config.LoadConf(&dbConf)
	return dbConf
}

// CreateBatchPoint 获取批量保存点
func (op *DbOption) CreateBatchPoint() (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  op.DbConf.Influx.Database,
		Precision: op.DbConf.Influx.Precision, //精度，默认ns
	})

	// 判断错误
	if err != nil {
		logrus.Error("influx create batch point failure", err)
		return nil, err
	}
	return bp, nil
}
