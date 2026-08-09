package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/gorm"
)

// mongoConn 封装 mongo 客户端与目标数据库名
type mongoConn struct {
	client *mongo.Client
	dbName string
}

// DatabaseMongoHandler MongoDB 数据源
// task.Command 语义：Extended JSON 聚合管道，形如 [{"$match":...},{"$group":...}]
// 管道结果取第一条文档中的第一个数值字段
type DatabaseMongoHandler struct{}

func (h *DatabaseMongoHandler) TestConnect(database entity.MonitorDatabase) error {
	client, err := buildMongoClient(database)
	if err != nil {
		return err
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("%s - mongo connect failure: %s", database.GetUrl(), err.Error())
	}
	return nil
}

func (h *DatabaseMongoHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	client, err := buildMongoClient(database)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		logrus.Errorf("mongo NewInstance ping fail: %v", err)
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return &mongoConn{client: client, dbName: database.Database}, nil
}

func (h *DatabaseMongoHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	docs, err := h.runAggregation(ctx, db, task)
	if err != nil {
		return 0, err
	}
	if len(docs) == 0 {
		// 查询成功但无文档：以 gorm.ErrRecordNotFound 哨兵表示"无数据"，由上层回落默认值
		return 0, gorm.ErrRecordNotFound
	}
	return extractFirstNumeric(docs[0])
}

// Close 关闭 MongoDB 客户端
func (h *DatabaseMongoHandler) Close(db interface{}) error {
	conn, ok := db.(*mongoConn)
	if !ok || conn == nil || conn.client == nil {
		return nil
	}
	return conn.client.Disconnect(context.Background())
}

// ExecuteQueryMultiRows 分组聚合采集：执行聚合管道并返回全部文档，每个文档映射为一个 RowResult。
func (h *DatabaseMongoHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	docs, err := h.runAggregation(ctx, db, task)
	if err != nil {
		return nil, err
	}
	results := make([]domainHandler.RowResult, 0, len(docs))
	for _, doc := range docs {
		row := domainHandler.RowResult{Columns: make(map[string]interface{}, len(doc))}
		for k, v := range doc {
			row.Columns[k] = v
		}
		results = append(results, row)
	}
	return results, nil
}

// runAggregation 公共聚合执行：解析管道 → 设超时 → 执行 Aggregate → 返回全部文档。
// ExecuteQuery 取首文档提取数值，ExecuteQueryMultiRows 映射全部文档为 RowResult。
func (h *DatabaseMongoHandler) runAggregation(ctx context.Context, db interface{}, task entity.MonitorTask) ([]bson.M, error) {
	conn, ok := db.(*mongoConn)
	if !ok || conn == nil || conn.client == nil {
		return nil, errors.New("invalid mongo connection")
	}

	var pipelineSpec struct {
		Collection string   `bson:"collection" json:"collection"`
		Pipeline   []bson.D `bson:"pipeline" json:"pipeline"`
	}
	if err := bson.UnmarshalExtJSON([]byte(task.Command), true, &pipelineSpec); err != nil || len(pipelineSpec.Pipeline) == 0 {
		var pipeline []bson.D
		if err2 := bson.UnmarshalExtJSON([]byte(task.Command), true, &pipeline); err2 != nil {
			return nil, fmt.Errorf("parse mongo aggregation pipeline: %w", err)
		}
		pipelineSpec.Pipeline = pipeline
	}
	if pipelineSpec.Collection == "" {
		pipelineSpec.Collection = "metrics"
	}

	// ctx 由采集层任务级超时控制；若上游未设超时（如手动调用），保留 30s 兜底
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	stages := make(mongo.Pipeline, 0, len(pipelineSpec.Pipeline))
	for _, stage := range pipelineSpec.Pipeline {
		stages = append(stages, stage)
	}

	// 防御性追加 $limit：若用户未在管道末指定 $limit，自动追加以防返回海量文档 OOM
	hasLimit := len(stages) > 0 && len(stages[len(stages)-1]) > 0 && stages[len(stages)-1][0].Key == "$limit"
	if !hasLimit {
		stages = append(stages, bson.D{{Key: "$limit", Value: constant.MaxAggregateRows}})
	}

	cursor, err := conn.client.Database(conn.dbName).Collection(pipelineSpec.Collection).Aggregate(ctx, stages)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func buildMongoClient(database entity.MonitorDatabase) (*mongo.Client, error) {
	uri := buildMongoURI(database)
	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(5)
	return mongo.Connect(opts)
}

func buildMongoURI(database entity.MonitorDatabase) string {
	url := database.GetUrl()
	plain, _ := database.GetDecodePassword()
	params := database.GetParams()

	var uri string
	if database.Username != "" && plain != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s", database.Username, plain, url)
	} else if database.Username != "" {
		uri = fmt.Sprintf("mongodb://%s@%s", database.Username, url)
	} else {
		uri = fmt.Sprintf("mongodb://%s", url)
	}

	// 若 Database 字段有值，拼到 path
	if database.Database != "" {
		uri += "/" + database.Database
	}

	if params != "" {
		uri += "?" + params
	}
	return uri
}

// extractFirstNumeric 从 BSON 结果中提取第一个数值
func extractFirstNumeric(m bson.M) (float64, error) {
	// 优先取常见字段
	for _, key := range []string{"value", "n", "count", "total", "result"} {
		if v, ok := m[key]; ok {
			if f, ok := bsonToFloat64(v); ok {
				return f, nil
			}
		}
	}
	// 回退：跳过 _id / 元数据，取第一个数值
	for k, v := range m {
		if k == "_id" || k == "ok" || k == "operationTime" || k == "$clusterTime" {
			continue
		}
		if f, ok := bsonToFloat64(v); ok {
			return f, nil
		}
	}
	return 0, errors.New("no numeric value found in mongo result")
}

func bsonToFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case bson.Decimal128:
		f, err := strconv.ParseFloat(n.String(), 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}
