package support

import (
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/domain/entity"
	"context"
	"errors"
	"fmt"
	sdk "github.com/pwh19920920/grafanasdk"
)

type GrafanaOptionHandler struct {
	Grafana *grafana.Config
}

func NewGrafanaOptionHandler(conf *grafana.Config) *GrafanaOptionHandler {
	return &GrafanaOptionHandler{conf}
}

// CreateDashboard 创建空面板
func (handler *GrafanaOptionHandler) CreateDashboard(name string) (*sdk.StatusMessage, error) {
	// 基本参数
	board := sdk.NewBoard(name)
	board.Time.From = "now-30m"
	board.Time.To = "now"

	autoCount := 30
	var boolInt int64 = 2
	board.Templating.List = []sdk.TemplateVar{{
		Auto:      true,
		AutoCount: &autoCount,
		AutoMin:   "30s",
		Current: sdk.Current{
			Value: "$__auto_interval_my_interval",
			Text:  &sdk.StringSliceString{Value: []string{"auto"}},
		},
		Hide:    2,
		Refresh: sdk.BoolInt{Value: &boolInt},
		Query:   "1m,5m,10m,30m,1h,6h,12h,1d,7d,14d,30d",
		Type:    "interval",
		Name:    "my_interval",
		Options: []sdk.Option{
			{Selected: true, Text: "auto", Value: "$__auto_interval_my_interval"},
			{Selected: false, Text: "1m", Value: "1m"},
			{Selected: false, Text: "5m", Value: "5m"},
			{Selected: false, Text: "10m", Value: "10m"},
			{Selected: false, Text: "30m", Value: "30m"},
			{Selected: false, Text: "1h", Value: "1h"},
			{Selected: false, Text: "6h", Value: "6h"},
			{Selected: false, Text: "12h", Value: "12h"},
			{Selected: false, Text: "1d", Value: "1d"},
			{Selected: false, Text: "7d", Value: "7d"},
			{Selected: false, Text: "14d", Value: "14d"},
			{Selected: false, Text: "30d", Value: "30d"},
		},
	}}

	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return nil, err
	}

	// 发送http
	resp, err := client.SetDashboard(context.TODO(), *board, sdk.SetDashboardParams{
		Overwrite: false,
	})
	return &resp, err
}

func (handler *GrafanaOptionHandler) ModifyDashboardName(uid, name string) (*sdk.StatusMessage, error) {
	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return nil, err
	}

	board, _, err := client.GetDashboardByUID(context.TODO(), uid)
	if err != nil {
		return nil, err
	}

	// 发送http
	board.Title = name
	resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
		Overwrite: true,
	})
	return &resp, err
}

func (handler *GrafanaOptionHandler) ModifySampleTarget(dashboardUIDs []string, task *entity.MonitorTask) error {
	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return err
	}

	for _, dashboardUID := range dashboardUIDs {
		board, _, err := client.GetDashboardByUID(context.TODO(), dashboardUID)
		if err != nil {
			return err
		}

		// 重新赋值
		panels := board.Panels
		for index, panel := range panels {
			// 找到具体的panel
			if panel.Description != nil && *panel.Description == task.TaskKey {
				newPanel := handler.buildPanel(*task)
				newPanel.GridPos = panel.GridPos
				newPanel.CustomPanel = panel.CustomPanel
				panels[index] = newPanel
			}
		}
		board.Panels = panels

		// http
		resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
			Overwrite: true,
		})

		if err != nil || *resp.Status != "success" {
			return errors.New(*resp.Status)
		}
	}
	return nil
}

func (handler *GrafanaOptionHandler) AddPanel(uid string, task entity.MonitorTask) (*sdk.StatusMessage, error) {
	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return nil, err
	}

	board, _, err := client.GetDashboardByUID(context.TODO(), uid)
	if err != nil {
		return nil, err
	}

	panels := board.Panels
	panels = append(panels, handler.buildPanel(task))
	board.Panels = handler.sortPanels(panels)

	// http
	resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
		Overwrite: true,
	})
	return &resp, err
}

func (handler *GrafanaOptionHandler) buildPanel(task entity.MonitorTask) *sdk.Panel {
	graph := sdk.NewGraph(task.TaskName)
	target := handler.createTarget("A", task.TaskKey, "实时")
	graph.AddTarget(&target)

	// 是否加入样本对比
	if task.Sampled == entity.MonitorSampledStatusOpen {
		target := handler.createTarget("B", fmt.Sprintf("%s.%s_sample", handler.Grafana.SampleRpName, task.TaskKey), "样本")
		graph.AddTarget(&target)
	}

	graph.Type = "timeseries"
	graph.Description = &task.TaskKey
	graph.GraphPanel.FieldConfig = &sdk.FieldConfig{}
	graph.GraphPanel.FieldConfig.Defaults.Custom.FillOpacity = 10

	toolTip := struct {
		Mode string `json:"mode,omitempty"`
	}{Mode: "multi"}

	options := struct {
		Tooltip struct {
			Mode string `json:"mode,omitempty"`
		} `json:"tooltip,omitempty"`
	}{toolTip}

	graph.GraphPanel.Options = options
	return graph
}

// 创建target
func (handler *GrafanaOptionHandler) createTarget(refID, measurement, alias string) sdk.Target {
	return sdk.Target{
		RefID:       refID,
		Datasource:  "InfluxDB",
		Alias:       alias,
		Format:      "time_series",
		Measurement: measurement,
		Select: [][]struct {
			Params []string `json:"params,omitempty"`
			Type   string   `json:"type,omitempty"`
		}{{{Params: []string{"value"}, Type: "field"}, {Type: "mean"}}},
		GroupBy: []struct {
			Type   string   `json:"type,omitempty"`
			Params []string `json:"params,omitempty"`
		}{{Type: "time", Params: []string{"$my_interval"}}, {Type: "fill", Params: []string{"null"}}},
	}
}

// ModifyDashBoardPanel 交集, 删除，新增
func (handler *GrafanaOptionHandler) ModifyDashBoardPanel(intersectionBoardUIDs, removeDashboardUIDs, addDashboardUIDs []string, task entity.MonitorTask) error {
	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return err
	}

	// 修改已存在的, 只允许修改title
	for _, boardUID := range intersectionBoardUIDs {
		board, _, err := client.GetDashboardByUID(context.TODO(), boardUID)
		if err != nil {
			return err
		}

		// 重新赋值
		panels := board.Panels
		for index, panel := range panels {
			if panel.Description != nil && *panel.Description == task.TaskKey {
				panel.Title = task.TaskName

				if panel.CustomPanel != nil {
					customPanel := *panel.CustomPanel
					customPanel["title"] = task.TaskName
					panel.CustomPanel = &customPanel
				}
				panels[index] = panel
			}
		}
		board.Panels = panels

		// http
		resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
			Overwrite: true,
		})

		if err != nil || *resp.Status != "success" {
			return errors.New(*resp.Status)
		}
	}

	// 删除
	for _, boardUID := range removeDashboardUIDs {
		board, _, err := client.GetDashboardByUID(context.TODO(), boardUID)
		if err != nil {
			return err
		}

		panels := board.Panels

		// 删除当前这个panel
		nextPanels := make([]*sdk.Panel, 0)
		for _, panel := range panels {
			if panel.Description == nil || *panel.Description != task.TaskKey {
				nextPanels = append(nextPanels, panel)
			}
		}

		board.Panels = handler.sortPanels(nextPanels)

		// http
		resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
			Overwrite: true,
		})

		if err != nil || *resp.Status != "success" {
			return errors.New(*resp.Status)
		}
	}

	// 新增
	for _, dashboardUID := range addDashboardUIDs {
		resp, err := handler.AddPanel(dashboardUID, task)
		if err != nil || *resp.Status != "success" {
			return errors.New(*resp.Status)
		}
	}
	return nil
}

// 排序功能
func (handler *GrafanaOptionHandler) sortPanels(panels []*sdk.Panel) []*sdk.Panel {
	length := len(panels)
	lines := length / 3
	if length%3 != 0 {
		lines = lines + 1
	}

	// 重排界面
	for i := 0; i < lines; i++ {
		for j := 0; j < 3; j++ {
			index := i*3 + j
			if index >= length {
				break
			}

			panel := panels[index]
			x := j * 8
			y := i * 8
			hw := 8
			panel.GridPos.X = &x
			panel.GridPos.Y = &y
			panel.GridPos.H = &hw
			panel.GridPos.W = &hw

			if panel.CustomPanel != nil {
				gridPos := make(map[string]interface{})
				gridPos["x"] = x
				gridPos["y"] = y
				gridPos["w"] = hw
				gridPos["h"] = hw

				customPanel := *panel.CustomPanel
				customPanel["gridPos"] = gridPos
				panel.CustomPanel = &customPanel
			}
			panels[index] = panel
		}
	}
	return panels
}

func (handler *GrafanaOptionHandler) ReSortDashboard(boardUID string, taskIds []int64, taskMap map[int64]entity.MonitorTask) error {
	taskKeys := make([]string, 0)
	for _, taskId := range taskIds {
		taskKeys = append(taskKeys, taskMap[taskId].TaskKey)
	}

	client, err := handler.Grafana.GetGrafanaClient()
	if err != nil {
		return err
	}

	board, _, err := client.GetDashboardByUID(context.TODO(), boardUID)
	if err != nil {
		return err
	}

	panels := board.Panels
	panelMap := make(map[string]*sdk.Panel, 0)
	for _, panel := range panels {
		panelMap[*panel.Description] = panel
	}

	newPanels := make([]*sdk.Panel, 0)
	for _, taskKey := range taskKeys {
		newPanels = append(newPanels, panelMap[taskKey])
	}

	board.Panels = handler.sortPanels(newPanels)

	// http
	resp, err := client.SetDashboard(context.TODO(), board, sdk.SetDashboardParams{
		Overwrite: true,
	})

	if err != nil || *resp.Status != "success" {
		return errors.New(*resp.Status)
	}
	return nil
}
