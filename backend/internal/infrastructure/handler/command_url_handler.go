package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"dragonfly-monitor/internal/common/constant"
	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

// maxResponseBodySize URL 采集响应体上限（10MB），防止超大响应耗尽内存
const maxResponseBodySize = 10 << 20

// CommandUrlHandler URL 探测采集
type CommandUrlHandler struct{}

// urlClient 复用单个 HTTP 客户端；超时由请求 ctx 控制（采集层任务级超时），
// 不在 client 上设 Timeout，避免与 ctx 超时叠加导致提前断连。
var urlClient = &http.Client{Timeout: 0}

// validateCommandURL 校验 URL 采集指令：仅允许 http/https 协议，拒绝私有/回环/链路本地 IP，
// 防止 SSRF 探测内网服务或云元数据端点（D-012）。
func validateCommandURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL 格式非法: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL 协议仅允许 http/https，当前为 %q", u.Scheme)
	}
	host := u.Hostname()
	// 域名先解析为 IP 再判断（防止 DNS 重绑定需配合 DialContext 校验，此处做基础拦截）
	ip := net.ParseIP(host)
	if ip == nil {
		// 非 IP 字面量（域名）：解析所有 A/AAAA 记录，任一命中私有段即拒绝
		ips, resolveErr := net.LookupIP(host)
		if resolveErr != nil {
			// 解析失败不阻断（可能内网 DNS），交由 HTTP 层处理
			return nil
		}
		for _, resolved := range ips {
			if isPrivateIP(resolved) {
				return fmt.Errorf("URL 目标 %s 解析到内网地址 %s，禁止采集", host, resolved)
			}
		}
		return nil
	}
	if isPrivateIP(ip) {
		return fmt.Errorf("URL 目标为内网地址 %s，禁止采集", ip)
	}
	return nil
}

// isPrivateIP 判断 IP 是否属于私有/回环/链路本地/未指定段
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"fc00::/7",
	}
	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (h *CommandUrlHandler) ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error) {
	if err := validateCommandURL(task.Command); err != nil {
		return 0, err
	}
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
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

// ExecuteMultiRows 分组聚合采集：HTTP GET 返回 JSON 数组，每个元素映射为一个 RowResult。
// 无 resultFieldPath 时用 json.Decoder 流式逐条读取，达到 10000 行上限即停止，避免全量 JSON 解析 OOM。
// resultFieldPath 场景仍需先全量解析再取嵌套数组（gojsonq 接管）。
func (h *CommandUrlHandler) ExecuteMultiRows(ctx context.Context, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	if err := validateCommandURL(task.Command); err != nil {
		return nil, err
	}
	var params struct {
		ResultFieldPath string `json:"resultFieldPath"`
	}
	if task.ExecParams != "" {
		_ = json.Unmarshal([]byte(task.ExecParams), &params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.Command, nil)
	if err != nil {
		return nil, err
	}
	resp, err := urlClient.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("url collect canceled by task timeout: %w", ctxErr)
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	// 指定 resultFieldPath 时需先全量解析再取嵌套数组（gojsonq 接管）
	if params.ResultFieldPath != "" {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
		if readErr != nil {
			return nil, readErr
		}
		return h.executeMultiRowsWithPath(body, params.ResultFieldPath, task)
	}

	// 无 resultFieldPath：body 即 JSON 数组，用 json.Decoder 流式逐条读取，达到上限即停
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize))

	tok, tokErr := decoder.Token()
	if tokErr != nil || tok != json.Delim('[') {
		return nil, fmt.Errorf("URL 响应非数组，无法用于聚合采集")
	}

	results := make([]domainHandler.RowResult, 0)
	for decoder.More() {
		var item map[string]interface{}
		if decErr := decoder.Decode(&item); decErr != nil {
			continue
		}
		row := domainHandler.RowResult{Columns: make(map[string]interface{}, len(item))}
		for k, v := range item {
			row.Columns[k] = v
		}
		results = append(results, row)
		if len(results) >= constant.MaxAggregateRows {
			logrus.Warnf("aggregate rows truncated at %d for task=%s", constant.MaxAggregateRows, task.TaskKey)
			break
		}
	}
	return results, nil
}

// executeMultiRowsWithPath resultFieldPath 场景：先全量解析 JSON 再由 gojsonq 取嵌套数组，
// 并在遍历时截断（嵌套数组通常已由上游分页控制大小）
func (h *CommandUrlHandler) executeMultiRowsWithPath(respBody []byte, resultFieldPath string, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	// 先校验 JSON 合法性（D-026：原 var raw 解析后未使用，改为轻量 json.Valid）
	if !json.Valid(respBody) {
		return nil, fmt.Errorf("解析 URL 响应 JSON 失败：非法 JSON")
	}

	nav := gojsonq.New().FromString(string(respBody)).Find(resultFieldPath)
	arr, ok := nav.([]interface{})
	if !ok {
		return nil, fmt.Errorf("resultFieldPath=%s 未指向数组", resultFieldPath)
	}

	results := make([]domainHandler.RowResult, 0, len(arr))
	for _, item := range arr {
		m, ok2 := item.(map[string]interface{})
		if !ok2 {
			continue
		}
		row := domainHandler.RowResult{Columns: make(map[string]interface{}, len(m))}
		for k, v := range m {
			row.Columns[k] = v
		}
		results = append(results, row)
		if len(results) >= constant.MaxAggregateRows {
			logrus.Warnf("aggregate rows truncated at %d for task=%s", constant.MaxAggregateRows, task.TaskKey)
			break
		}
	}
	return results, nil
}
