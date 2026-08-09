package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"butterfly-monitor/internal/config/grafana"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/pwh19920920/butterfly/pkg/logger"
)

// GrafanaHandler 通过 Grafana HTTP API 管理 dashboard / panel
// panel 锚点：Description == taskKey
// 查询表达式 / 数据源类型由 MetricQueryDialect 决定
// panel/target 直接绑定 Grafana 真实数据源 UID（按 dialect 类型解析，不再用 ${datasource}）
type GrafanaHandler struct {
	conf    *grafana.Config
	dialect domainHandler.MetricQueryDialect
	client  *http.Client

	// 按 dialect 类型解析出的真实数据源，懒加载缓存
	dsMu     sync.Mutex
	dsType   string
	dsUID    string
	dsName   string
	dsCached bool
}

// NewGrafanaHandler 创建；dialect 可为空（无面板查询时跳过 target 构建细节）
func NewGrafanaHandler(conf *grafana.Config, dialect domainHandler.MetricQueryDialect) *GrafanaHandler {
	return &GrafanaHandler{
		conf:    conf,
		dialect: dialect,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *GrafanaHandler) enabled() bool {
	return h != nil && h.conf != nil && strings.TrimSpace(h.conf.Addr) != "" && strings.TrimSpace(h.conf.ApiKey) != ""
}

func (h *GrafanaHandler) baseURL() string {
	return strings.TrimRight(h.conf.Addr, "/")
}

// CreateDashboard 创建空大盘，回填 url/slug/uid/boardId
func (h *GrafanaHandler) CreateDashboard(ctx context.Context, name string) (url, slug, uid string, boardId int64, err error) {
	if !h.enabled() {
		logger.Warn(ctx, "Grafana not configured (addr/apiKey empty), CreateDashboard skipped")
		return "", "", "", 0, nil
	}
	board := map[string]interface{}{
		"title":         name,
		"timezone":      "browser",
		"schemaVersion": 30,
		"version":       0,
		"refresh":       "30s",
		"time": map[string]string{
			"from": "now-30m",
			"to":   "now",
		},
		// panel 直接绑定真实数据源 UID；templating 放分组时间跨度变量
		"templating": map[string]interface{}{
			"list": h.intervalTemplateVars(),
		},
		"panels": []interface{}{},
	}
	payload := map[string]interface{}{
		"dashboard": board,
		"overwrite": false,
		"folderId":  0,
	}
	var resp struct {
		ID      int64  `json:"id"`
		UID     string `json:"uid"`
		URL     string `json:"url"`
		Slug    string `json:"slug"`
		Status  string `json:"status"`
		Version int    `json:"version"`
	}
	if err = h.doJSON(http.MethodPost, "/api/dashboards/db", payload, &resp); err != nil {
		return "", "", "", 0, err
	}
	if resp.Status != "" && resp.Status != "success" {
		return "", "", "", 0, fmt.Errorf("grafana create dashboard status=%s", resp.Status)
	}
	return resp.URL, resp.Slug, resp.UID, resp.ID, nil
}

// ModifyDashboardName 修改大盘标题
func (h *GrafanaHandler) ModifyDashboardName(uid, name string) error {
	if !h.enabled() || uid == "" {
		return nil
	}
	board, meta, err := h.getDashboard(uid)
	if err != nil {
		return err
	}
	board["title"] = name
	h.ensureIntervalTemplate(board)
	return h.saveDashboard(board, meta, true)
}

// AddPanel 向大盘追加 timeseries panel
func (h *GrafanaHandler) AddPanel(uid, taskKey, taskName string, sampled bool, related []RelatedMetric) error {
	if !h.enabled() || uid == "" || taskKey == "" {
		return nil
	}
	board, meta, err := h.getDashboard(uid)
	if err != nil {
		return err
	}
	h.ensureIntervalTemplate(board)
	panels := getPanels(board)
	// 已存在则改为 update（保留原 panel id）
	for i, p := range panels {
		if panelDesc(p) == taskKey {
			panel, err := h.buildPanel(taskKey, taskName, sampled, related, panelGridPos(p), panelID(p))
			if err != nil {
				return err
			}
			panels[i] = panel
			board["panels"] = sortPanels(panels)
			return h.saveDashboard(board, meta, true)
		}
	}
	panel, err := h.buildPanel(taskKey, taskName, sampled, related, nil, nextPanelID(panels))
	if err != nil {
		return err
	}
	panels = append(panels, panel)
	board["panels"] = sortPanels(panels)
	return h.saveDashboard(board, meta, true)
}

// ModifyDashBoardPanel 对单个 dashboard 做 panel 增/删/改
// add/del/update 三选一语义（与 application 调用一致）
func (h *GrafanaHandler) ModifyDashBoardPanel(uid, taskKey, taskName string, sampled bool, related []RelatedMetric, add, del, update bool) error {
	if !h.enabled() || uid == "" || taskKey == "" {
		return nil
	}
	board, meta, err := h.getDashboard(uid)
	if err != nil {
		return err
	}
	h.ensureIntervalTemplate(board)
	panels := getPanels(board)

	if del {
		next := make([]map[string]interface{}, 0, len(panels))
		for _, p := range panels {
			if panelDesc(p) != taskKey {
				next = append(next, p)
			}
		}
		board["panels"] = sortPanels(next)
		return h.saveDashboard(board, meta, true)
	}

	if update {
		found := false
		for i, p := range panels {
			if panelDesc(p) == taskKey {
				panel, err := h.buildPanel(taskKey, taskName, sampled, related, panelGridPos(p), panelID(p))
				if err != nil {
					return err
				}
				panels[i] = panel
				found = true
			}
		}
		if !found {
			// 没有则当新增
			panel, err := h.buildPanel(taskKey, taskName, sampled, related, nil, nextPanelID(panels))
			if err != nil {
				return err
			}
			panels = append(panels, panel)
		}
		board["panels"] = sortPanels(panels)
		return h.saveDashboard(board, meta, true)
	}

	if add {
		// 已存在则更新
		for i, p := range panels {
			if panelDesc(p) == taskKey {
				panel, err := h.buildPanel(taskKey, taskName, sampled, related, panelGridPos(p), panelID(p))
				if err != nil {
					return err
				}
				panels[i] = panel
				board["panels"] = sortPanels(panels)
				return h.saveDashboard(board, meta, true)
			}
		}
		panel, err := h.buildPanel(taskKey, taskName, sampled, related, nil, nextPanelID(panels))
		if err != nil {
			return err
		}
		panels = append(panels, panel)
		board["panels"] = sortPanels(panels)
		return h.saveDashboard(board, meta, true)
	}
	return nil
}

// ReSortDashboard 按 taskKeys 顺序重排 panel
func (h *GrafanaHandler) ReSortDashboard(uid string, taskKeys []string) error {
	if !h.enabled() || uid == "" {
		return nil
	}
	board, meta, err := h.getDashboard(uid)
	if err != nil {
		return err
	}
	panels := getPanels(board)
	panelMap := make(map[string]map[string]interface{}, len(panels))
	for _, p := range panels {
		if d := panelDesc(p); d != "" {
			panelMap[d] = p
		}
	}
	ordered := make([]map[string]interface{}, 0, len(taskKeys))
	used := make(map[string]bool)
	for _, k := range taskKeys {
		if p, ok := panelMap[k]; ok {
			ordered = append(ordered, p)
			used[k] = true
		}
	}
	// 未在排序列表中的 panel 追加末尾
	for _, p := range panels {
		d := panelDesc(p)
		if d == "" || used[d] {
			continue
		}
		ordered = append(ordered, p)
	}
	board["panels"] = sortPanels(ordered)
	return h.saveDashboard(board, meta, true)
}

// ---------- panel 构建（方言由 MetricQueryDialect 提供） ----------

func (h *GrafanaHandler) datasourceType() string {
	if h.dialect != nil {
		return h.dialect.DatasourceType()
	}
	return "prometheus"
}

// resolveDatasource 解析 panel 要绑定的真实数据源
// type 始终来自 dialect（timeseries.backend）；uid 优先用配置 grafana.datasourceUid，
// 未配则 GET /api/datasources 按 type 匹配（通常需 Admin）
// 优先 isDefault，否则取该类型第一个；结果缓存
func (h *GrafanaHandler) resolveDatasource() (dsType, dsUID, dsName string, err error) {
	h.dsMu.Lock()
	defer h.dsMu.Unlock()

	if h.dsCached && h.dsUID != "" {
		return h.dsType, h.dsUID, h.dsName, nil
	}

	wantType := h.datasourceType()
	if h.conf != nil {
		if uid := strings.TrimSpace(h.conf.DatasourceUid); uid != "" {
			h.dsType = wantType
			h.dsUID = uid
			h.dsName = uid
			h.dsCached = true
			return h.dsType, h.dsUID, h.dsName, nil
		}
	}

	var list []struct {
		UID       string `json:"uid"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		IsDefault bool   `json:"isDefault"`
	}
	if err = h.doJSON(http.MethodGet, "/api/datasources", nil, &list); err != nil {
		return "", "", "", fmt.Errorf(
			"解析 Grafana 数据源失败（GET /api/datasources 通常需要 Admin；也可在配置 grafana.datasourceUid 写死 UID）: %w",
			err,
		)
	}

	var firstUID, firstName string
	for _, ds := range list {
		if ds.Type != wantType || strings.TrimSpace(ds.UID) == "" {
			continue
		}
		if ds.IsDefault {
			h.dsType = ds.Type
			h.dsUID = ds.UID
			h.dsName = ds.Name
			h.dsCached = true
			return h.dsType, h.dsUID, h.dsName, nil
		}
		if firstUID == "" {
			firstUID = ds.UID
			firstName = ds.Name
		}
	}
	if firstUID == "" {
		return "", "", "", fmt.Errorf(
			"grafana 中未找到 type=%s 的数据源；请在 Grafana 配置该类型数据源，或设置 grafana.datasourceUid",
			wantType,
		)
	}

	h.dsType = wantType
	h.dsUID = firstUID
	h.dsName = firstName
	h.dsCached = true
	return h.dsType, h.dsUID, h.dsName, nil
}

// intervalTemplateVars 分组时间跨度下拉（Grafana interval 变量）
// 面板查询通过 $interval 引用，例如 avg_over_time(metric[$interval])
func (h *GrafanaHandler) intervalTemplateVars() []map[string]interface{} {
	options := []map[string]interface{}{
		{"text": "15s", "value": "15s", "selected": false},
		{"text": "30s", "value": "30s", "selected": false},
		{"text": "1m", "value": "1m", "selected": true},
		{"text": "5m", "value": "5m", "selected": false},
		{"text": "10m", "value": "10m", "selected": false},
		{"text": "30m", "value": "30m", "selected": false},
		{"text": "1h", "value": "1h", "selected": false},
	}
	return []map[string]interface{}{
		{
			"name":        "interval",
			"label":       "分组跨度",
			"type":        "interval",
			"query":       "15s,30s,1m,5m,10m,30m,1h",
			"auto":        false,
			"auto_count":  30,
			"auto_min":    "10s",
			"current":     map[string]interface{}{"text": "1m", "value": "1m", "selected": true},
			"options":     options,
			"refresh":     2,
			"hide":        0,
			"includeAll":  false,
			"multi":       false,
			"skipUrlSync": false,
		},
	}
}

// ensureIntervalTemplate 保证 dashboard 上存在 interval 变量（不覆盖其它变量）
func (h *GrafanaHandler) ensureIntervalTemplate(board map[string]interface{}) {
	if board == nil {
		return
	}
	want := h.intervalTemplateVars()[0]
	raw, _ := board["templating"].(map[string]interface{})
	if raw == nil {
		board["templating"] = map[string]interface{}{"list": []interface{}{want}}
		return
	}
	listRaw, ok := raw["list"]
	if !ok || listRaw == nil {
		raw["list"] = []interface{}{want}
		board["templating"] = raw
		return
	}
	var list []interface{}
	switch v := listRaw.(type) {
	case []interface{}:
		list = v
	case []map[string]interface{}:
		list = make([]interface{}, 0, len(v))
		for _, m := range v {
			list = append(list, m)
		}
	default:
		list = []interface{}{}
	}
	found := false
	for i, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		if name == "interval" {
			if cur, ok := m["current"].(map[string]interface{}); ok {
				want["current"] = cur
			}
			list[i] = want
			found = true
			break
		}
	}
	if !found {
		list = append(list, want)
	}
	raw["list"] = list
	board["templating"] = raw
}

// bindDatasource 覆盖 target 内的数据源引用为真实 UID
func bindDatasource(target map[string]interface{}, dsType, dsUID string) {
	if target == nil {
		return
	}
	target["datasource"] = map[string]string{
		"type": dsType,
		"uid":  dsUID,
	}
}

// RelatedMetric 关联任务在面板上的取数信息
// 关联任务的实时/样本曲线会叠加到主任务面板，图例命名为 任务名_实时 / 任务名_样本
type RelatedMetric struct {
	TaskKey  string
	TaskName string
	Sampled  bool
}

func (h *GrafanaHandler) buildPanel(taskKey, taskName string, sampled bool, related []RelatedMetric, grid map[string]interface{}, id int) (map[string]interface{}, error) {
	if taskName == "" {
		taskName = taskKey
	}
	if grid == nil {
		grid = map[string]interface{}{"x": 0, "y": 0, "w": 8, "h": 8}
	}

	dsType, dsUID, _, err := h.resolveDatasource()
	if err != nil {
		return nil, err
	}

	targets := make([]map[string]interface{}, 0, 2+len(related)*2)
	if h.dialect != nil {
		// refID 按 A/B/.../Z/AA/AB... 递增，主任务占前两个，关联任务依次往后
		refIdx := 0
		nextRef := func() string {
			refIdx++
			n := refIdx
			s := ""
			for n > 0 {
				n--
				s = string(rune('A'+n%26)) + s
				n /= 26
			}
			return s
		}
		tA := h.dialect.BuildPanelTarget(nextRef(), "实时", h.dialect.RealtimeExpr(taskKey))
		bindDatasource(tA, dsType, dsUID)
		targets = append(targets, tA)
		if sampled {
			tB := h.dialect.BuildPanelTarget(nextRef(), "样本", h.dialect.SmoothExpr(taskKey))
			bindDatasource(tB, dsType, dsUID)
			targets = append(targets, tB)
		}
		// 关联任务：每个叠加 实时（+样本，若该任务开了样本展示）
		for _, r := range related {
			name := r.TaskName
			if name == "" {
				name = r.TaskKey
			}
			tr := h.dialect.BuildPanelTarget(nextRef(), name+"_实时", h.dialect.RealtimeExpr(r.TaskKey))
			bindDatasource(tr, dsType, dsUID)
			targets = append(targets, tr)
			if r.Sampled {
				ts := h.dialect.BuildPanelTarget(nextRef(), name+"_样本", h.dialect.SmoothExpr(r.TaskKey))
				bindDatasource(ts, dsType, dsUID)
				targets = append(targets, ts)
			}
		}
	}
	return map[string]interface{}{
		"id":          id,
		"type":        "timeseries",
		"title":       taskName,
		"description": taskKey,
		"gridPos":     grid,
		"datasource": map[string]string{
			"type": dsType,
			"uid":  dsUID,
		},
		"targets": targets,
		"fieldConfig": map[string]interface{}{
			"defaults": map[string]interface{}{
				"custom": map[string]interface{}{
					"fillOpacity": 10,
					"pointSize":   3,
					"drawStyle":   "line",
					"lineWidth":   1,
					"showPoints":  "auto",
				},
			},
			"overrides": []interface{}{},
		},
		"options": map[string]interface{}{
			"tooltip": map[string]string{"mode": "multi"},
			"legend":  map[string]interface{}{"displayMode": "list", "placement": "bottom"},
		},
	}, nil
}

// ---------- HTTP helpers ----------

func (h *GrafanaHandler) getDashboard(uid string) (board map[string]interface{}, meta map[string]interface{}, err error) {
	var resp struct {
		Dashboard map[string]interface{} `json:"dashboard"`
		Meta      map[string]interface{} `json:"meta"`
	}
	if err = h.doJSON(http.MethodGet, "/api/dashboards/uid/"+uid, nil, &resp); err != nil {
		return nil, nil, err
	}
	if resp.Dashboard == nil {
		return nil, nil, errors.New("dashboard not found: " + uid)
	}
	return resp.Dashboard, resp.Meta, nil
}

func (h *GrafanaHandler) saveDashboard(board, meta map[string]interface{}, overwrite bool) error {
	// 去掉 id 冲突风险时保留 version
	payload := map[string]interface{}{
		"dashboard": board,
		"overwrite": overwrite,
	}
	if meta != nil {
		if fid, ok := meta["folderId"]; ok {
			payload["folderId"] = fid
		}
	}
	var resp struct {
		Status string `json:"status"`
		UID    string `json:"uid"`
		URL    string `json:"url"`
	}
	if err := h.doJSON(http.MethodPost, "/api/dashboards/db", payload, &resp); err != nil {
		return err
	}
	if resp.Status != "" && resp.Status != "success" {
		return fmt.Errorf("grafana save dashboard status=%s", resp.Status)
	}
	return nil
}

func (h *GrafanaHandler) doJSON(method, path string, body interface{}, out interface{}) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, h.baseURL()+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+h.conf.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("grafana %s %s status=%d body=%s", method, path, resp.StatusCode, string(respBody))
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

// ---------- panel utils ----------

func getPanels(board map[string]interface{}) []map[string]interface{} {
	raw, ok := board["panels"]
	if !ok || raw == nil {
		return []map[string]interface{}{}
	}
	arr, ok := raw.([]interface{})
	if !ok {
		// 可能已是 []map
		if m, ok := raw.([]map[string]interface{}); ok {
			return m
		}
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

func panelDesc(p map[string]interface{}) string {
	if p == nil {
		return ""
	}
	if v, ok := p["description"].(string); ok {
		return v
	}
	return ""
}

func panelID(p map[string]interface{}) int {
	if p == nil {
		return 0
	}
	switch v := p["id"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	}
	return 0
}

func panelGridPos(p map[string]interface{}) map[string]interface{} {
	if p == nil {
		return nil
	}
	if g, ok := p["gridPos"].(map[string]interface{}); ok {
		// 拷贝
		cp := make(map[string]interface{}, len(g))
		for k, v := range g {
			cp[k] = v
		}
		return cp
	}
	return nil
}

func nextPanelID(panels []map[string]interface{}) int {
	maxID := 0
	for _, p := range panels {
		if id := panelID(p); id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

// sortPanels 3 列网格，每格 8x8
func sortPanels(panels []map[string]interface{}) []map[string]interface{} {
	for i, p := range panels {
		x := (i % 3) * 8
		y := (i / 3) * 8
		p["gridPos"] = map[string]interface{}{
			"x": x,
			"y": y,
			"w": 8,
			"h": 8,
		}
		panels[i] = p
	}
	return panels
}
