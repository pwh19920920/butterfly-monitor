package handler

import (
	"butterfly-monitor/domain/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strconv"
	"time"
)

const defaultMongoMaxPoolSize = 50
const defaultMongoTimeout = 20

type DatabaseMongoHandler struct {
	connect *mongo.Database
}

func (dbHandler *DatabaseMongoHandler) TestConnect(database entity.MonitorDatabase) error {
	// 创建连接
	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s?w=majority", database.Username, database.Password, database.Url, database.Database)

	// 设置连接超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultMongoTimeout))
	defer cancel()

	// 通过传进来的uri连接相关的配置, 设置最大连接数 - 默认是100 ，不设置就是最大 max 64
	option := options.Client().ApplyURI(dsn).SetMaxPoolSize(uint64(defaultMongoMaxPoolSize))
	client, err := mongo.Connect(ctx, option)
	if err != nil || client == nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure", database.Url, database.Database))
	}

	// 判断服务是不是可用
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		return errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}

	// 最后释放连接
	_ = client.Disconnect(ctx)
	return nil
}

// NewInstance mongodb://username:password@url/dbName?w=majority
func (dbHandler *DatabaseMongoHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	// 创建连接
	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s?w=majority", database.Username, database.Password, database.Url, database.Database)

	// 设置连接超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultMongoTimeout))
	defer cancel()

	// 通过传进来的uri连接相关的配置, 设置最大连接数 - 默认是100 ，不设置就是最大 max 64
	option := options.Client().ApplyURI(dsn).SetMaxPoolSize(uint64(defaultMongoMaxPoolSize))
	client, err := mongo.Connect(ctx, option)
	if err != nil || client == nil {
		logrus.Errorf("%s - %s db connect open failure", database.Url, database.Database)
		return nil, errors.New(fmt.Sprintf("%s - %s db connect open failure", database.Url, database.Database))
	}

	// 判断服务是不是可用
	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.Url, database.Database, err.Error()))
	}

	// 返回对象
	connect := client.Database(database.Database)
	return &DatabaseMongoHandler{connect: connect}, nil
}

type DatabaseMongoParams struct {
	ResultFieldPath string  `json:"resultFieldPath"` // 支持对象.属性
	CollectName     string  `json:"collectName"`     // 集合名称
	DefaultValue    float64 `json:"defaultValue"`    // 默认值
}

// ExecuteQuery 执行查询
func (dbHandler *DatabaseMongoHandler) ExecuteQuery(task entity.MonitorTask) (interface{}, error) {
	var params DatabaseMongoParams
	if err := json.Unmarshal([]byte(task.ExecParams), &params); err != nil {
		return nil, err
	}

	var bdoc interface{}
	if err := bson.UnmarshalExtJSON([]byte(task.Command), true, &bdoc); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregate, err := dbHandler.connect.Collection(params.CollectName).Aggregate(ctx, bdoc)
	defer aggregate.Close(ctx)

	// 判断错误
	if err != nil {
		return nil, errors.New("请求错误, 或者取不到结果 -> [" + err.Error() + "]")
	}

	// 从结果中取数据
	for aggregate.Next(ctx) {
		var doc map[string]interface{}
		aggregate.Decode(&doc)

		result := gojsonq.New().FromInterface(doc).Find(params.ResultFieldPath)
		if nil == result {
			return nil, errors.New("请求成功, 但取不到结果")
		}

		return strconv.ParseFloat(fmt.Sprintf("%v", result), 64)
	}

	// 返回默认值, 没有错误, 规定无数据比例100%，数量位0
	return params.DefaultValue, nil
}
