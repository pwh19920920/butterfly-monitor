package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"dragonfly-monitor/internal/domain/entity"

	"github.com/thedevsaddam/gojsonq/v2"
)

// CommandUrlHandler URL 探测采集
type CommandUrlHandler struct{}

// urlClient 复用单个 HTTP 客户端；超时由请求 ctx 控制（采集层任务级超时），
// 不在 client 上设 Timeout，避免与 ctx 超时叠加导致提前断连。
var urlClient = &http.Client{Timeout: 0}

func (h *CommandUrlHandler) ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error) {
	var params struct {
		ResultFieldPath string   `json:"resultFieldPath"`
		DefaultValue    *float64 `json:"defaultValue"`
	}
	if task.ExecParams != "" {
		_ = json.Unmarshal([]byte(task.ExecParams), &params)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.Command, nil)
	if err != nil {
		return 0, err
	}
	resp, err := urlClient.Do(req)
	if err != nil {
		// 区分 ctx 超时与网络错误，便于上层记录
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, fmt.Errorf("url collect canceled by task timeout: %w", ctxErr)
		}
		return 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if params.ResultFieldPath == "" {
		// 整个 body 当数字
		f, err := strconv.ParseFloat(strings.TrimSpace(string(body)), 64)
		if err != nil && params.DefaultValue != nil {
			return *params.DefaultValue, nil
		}
		return f, err
	}

	var root interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		if params.DefaultValue != nil {
			return *params.DefaultValue, nil
		}
		return 0, err
	}

	val := gojsonq.New().FromString(string(body)).Find(params.ResultFieldPath)
	if nil == val {
		return 0, errors.New("请求成功, 但取不到结果")
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}
