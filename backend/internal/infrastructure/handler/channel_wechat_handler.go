package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"butterfly-monitor/internal/domain/entity"
)

// ChannelWechatHandler 企业微信 webhook
type ChannelWechatHandler struct{}

// wechatParams 通道参数（接口约定小写）
type wechatParams struct {
	Addr string `json:"addr"`
}

func (h *ChannelWechatHandler) GetClassName() string { return "ChannelWechatHandler" }

func (h *ChannelWechatHandler) TestDispatchMessage(channel entity.AlertChannel, _ string, message string) error {
	if strings.TrimSpace(message) == "" {
		message = "butterfly-monitor 通道测试消息"
	}
	return h.DispatchMessage(channel, nil, message)
}

func (h *ChannelWechatHandler) DispatchMessage(channel entity.AlertChannel, _ []entity.SysUser, message string) error {
	var p wechatParams
	if err := json.Unmarshal([]byte(channel.Params), &p); err != nil {
		return err
	}
	if p.Addr == "" {
		return errors.New("企微 webhook 地址为空")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": message,
		},
	})
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(p.Addr, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return errors.New("企微发送失败")
	}
	return nil
}
