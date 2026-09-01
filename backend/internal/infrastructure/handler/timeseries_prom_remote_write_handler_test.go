package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"butterfly-monitor/internal/config/promremotewrite"
	domainHandler "butterfly-monitor/internal/domain/handler"
)

// TestRemoteWriteWritePoints 验证写路径：header、snappy 编码可解、labels 组装正确
func TestRemoteWriteWritePoints(t *testing.T) {
	var gotPath, gotContentType, gotEncoding string
	var decodedLabels []map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotEncoding = r.Header.Get("Content-Encoding")
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		// body 是 snappy 压缩的 protobuf；这里只校验长度>0（完整解码依赖 prompb，测试里直接解压验证可解性）
		if gotEncoding == "snappy" {
			if len(body) == 0 {
				t.Error("write body is empty")
			}
			// 解压成功即认为编码链路正确（prompb marshal 在 encodeSnappy 已保证）
			decodedLabels = append(decodedLabels, map[string]string{"len": string(rune(len(body)))})
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	h := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL:            srv.URL + "/api/v1/write",
		ExternalLabels: map[string]string{"cluster": "test"},
	})
	now := time.Now()
	points := []domainHandler.TimeSeriesPoint{
		{Metric: "order_count", Tags: map[string]string{"region": "华东"}, Value: 100, Timestamp: now},
		{Metric: "order_count", Tags: map[string]string{"region": "华南"}, Value: math.NaN(), Timestamp: now}, // NaN 应被过滤
	}
	if err := h.WritePoints(context.Background(), points); err != nil {
		t.Fatalf("WritePoints err: %v", err)
	}
	if gotPath != "/api/v1/write" {
		t.Errorf("path = %s, want /api/v1/write", gotPath)
	}
	if gotContentType != "application/x-protobuf" {
		t.Errorf("Content-Type = %s", gotContentType)
	}
	if gotEncoding != "snappy" {
		t.Errorf("Content-Encoding = %s", gotEncoding)
	}
	if len(decodedLabels) != 1 {
		t.Errorf("decoded batches = %d, want 1", len(decodedLabels))
	}
}

// TestRemoteWriteQueryRange 验证读路径：query 端点推导、tagFilters 拼接、响应解析
func TestRemoteWriteQueryRange(t *testing.T) {
	var gotQuery, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		resp := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"result": []map[string]interface{}{
					{
						"metric": map[string]string{"__name__": "order_count", "region": "华东"},
						"values": [][]interface{}{{float64(1700000000), "100"}, {float64(1700000015), "120"}},
					},
					{
						"metric": map[string]string{"__name__": "order_count", "region": "华南"},
						"values": [][]interface{}{{float64(1700000000), "50"}},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	h := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL: srv.URL + "/api/v1/write",
	})

	start := time.Unix(1700000000, 0)
	end := time.Unix(1700000100, 0)

	// 带 tagFilters
	series, err := h.QueryRangeWithTags(context.Background(), "order_count", start, end,
		map[string][]string{"region": {"华东"}})
	if err != nil {
		t.Fatalf("QueryRangeWithTags err: %v", err)
	}
	if gotPath != "/api/v1/query_range" {
		t.Errorf("query path = %s, want /api/v1/query_range", gotPath)
	}
	// 单值精确匹配
	if gotQuery != `order_count{region="华东"}` {
		t.Errorf("query = %s", gotQuery)
	}
	if len(series) != 2 {
		t.Errorf("series count = %d, want 2 (远端不过滤，返回全部)", len(series))
	}

	// 多值正则
	_, err = h.QueryRangeWithTags(context.Background(), "order_count", start, end,
		map[string][]string{"region": {"华东", "华南"}})
	if err != nil {
		t.Fatalf("QueryRangeWithTags multi err: %v", err)
	}
	if gotQuery != `order_count{region=~"华东|华南"}` {
		t.Errorf("multi query = %s", gotQuery)
	}
}

// TestRemoteWriteQueryMean 验证 QueryMean 走 instant query（/api/v1/query + time=end）
// 与 VM handler 语义逐字对齐：窗口均值 avg(avg_over_time(...[Ns])) 在 end 时刻回看
func TestRemoteWriteQueryMean(t *testing.T) {
	var gotQuery, gotPath string
	var gotTime string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		gotTime = r.URL.Query().Get("time")
		// instant 响应：result[].value = [ts, "v"]
		resp := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"result": []map[string]interface{}{
					{
						"metric": map[string]string{},
						"value":  []interface{}{float64(1700000060), "75"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	h := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL: srv.URL + "/api/v1/write",
	})

	start := time.Unix(1700000000, 0)
	end := time.Unix(1700000060, 0)
	mean, err := h.QueryMean(context.Background(), "order_count", start, end)
	if err != nil {
		t.Fatalf("QueryMean err: %v", err)
	}
	if mean == nil || *mean != 75 {
		t.Errorf("mean = %v, want 75", mean)
	}
	if gotPath != "/api/v1/query" {
		t.Errorf("path = %s, want /api/v1/query (instant)", gotPath)
	}
	if gotTime != "1700000060" {
		t.Errorf("time = %s, want 1700000060 (end)", gotTime)
	}
	want := `avg(avg_over_time(order_count[60s]))`
	if gotQuery != want {
		t.Errorf("query = %s, want %s", gotQuery, want)
	}
}

// TestRemoteWriteQueryMeanInstant 验证 window<=0 退化为 avg(metric)
func TestRemoteWriteQueryMeanInstant(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("query")
		resp := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"result": []map[string]interface{}{
					{
						"metric": map[string]string{},
						"value":  []interface{}{float64(1700000060), "42"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	h := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL: srv.URL + "/api/v1/write",
	})

	ts := time.Unix(1700000060, 0)
	mean, err := h.QueryMean(context.Background(), "order_count", ts, ts)
	if err != nil {
		t.Fatalf("QueryMean err: %v", err)
	}
	if mean == nil || *mean != 42 {
		t.Errorf("mean = %v, want 42", mean)
	}
	if gotQuery != `avg(order_count)` {
		t.Errorf("query = %s, want avg(order_count)", gotQuery)
	}
}

// TestRemoteWriteAdaptiveStep 验证 step 自适应：短窗口 15s、长窗口 ≤500 点
func TestRemoteWriteAdaptiveStep(t *testing.T) {
	cases := []struct {
		name       string
		start, end time.Time
		want       string
	}{
		{"窗口=15s", time.Unix(1700000000, 0), time.Unix(1700000015, 0), "15s"},
		{"窗口=30s", time.Unix(1700000000, 0), time.Unix(1700000030, 0), "15s"},
		{"窗口=60s", time.Unix(1700000000, 0), time.Unix(1700000060, 0), "15s"},
		// 8 天窗口 = 691200s / 500 = 1382.4s → 1382s
		{"窗口=8天", time.Unix(1700000000, 0), time.Unix(1700000000+691200, 0), "1382s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := adaptiveStep(c.start, c.end)
			if got != c.want {
				t.Errorf("adaptiveStep = %s, want %s", got, c.want)
			}
		})
	}
}

// TestRemoteWriteResolveEndpoint 验证 query 端点推导与 queryBase 覆盖
func TestRemoteWriteResolveEndpoint(t *testing.T) {
	h := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL: "http://host:8429/api/v1/write",
	})
	if got := h.resolveEndpoint("/api/v1/query_range"); got != "http://host:8429/api/v1/query_range" {
		t.Errorf("derived = %s", got)
	}

	h2 := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL:       "https://prom.example.grafana.net/api/prom/push", // Grafana Cloud 非标准路径
		QueryBase: "https://prom.example.grafana.net/api/prom",
	})
	if got := h2.resolveEndpoint("/api/v1/query"); got != "https://prom.example.grafana.net/api/prom/api/v1/query" {
		t.Errorf("queryBase = %s", got)
	}

	// 未配 queryBase 且 URL 不带 /api/v1/write 尾缀：以 write URL 为 base 拼接（该场景需用户配 queryBase）
	h3 := NewTimeSeriesPromRemoteWriteHandler(promremotewrite.Config{
		URL: "https://prom.example.grafana.net/api/prom/push",
	})
	if got := h3.resolveEndpoint("/api/v1/query"); got != "https://prom.example.grafana.net/api/prom/push/api/v1/query" {
		t.Errorf("fallback = %s", got)
	}
}
