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

var dbConf *influxConf

type influxConf struct {
	Influx Config `yaml:"influx"`
}

type DbOption struct {
	Client client.Client // 操作客户端
}

func NewInfluxDbOption() *DbOption {
	return &DbOption{
		getDbConn(),
	}
}

// getConn 获取连接
func getDbConn() client.Client {
	// 默认值
	viper.SetDefault("influx.addr", defaultAddr)
	viper.SetDefault("influx.precision", defaultPrecision)
	viper.SetDefault("influx.username", defaultUsername)
	viper.SetDefault("influx.password", defaultPassword)

	// 加载配置
	dbConf = new(influxConf)
	config.LoadConf(&dbConf)

	// 创建client
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     dbConf.Influx.Addr,
		Username: dbConf.Influx.Username,
		Password: dbConf.Influx.Password,
	})

	// 判断错误
	if err != nil {
		logrus.Panic("influx connect open failure", err)
	}
	return cli
}

// CreateBatchPoint 获取批量保存点
func (op *DbOption) CreateBatchPoint() (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbConf.Influx.Database,
		Precision: dbConf.Influx.Precision, //精度，默认ns
	})

	// 判断错误
	if err != nil {
		logrus.Error("influx create batch point failure", err)
		return nil, err
	}
	return bp, nil
}
