package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dragonfly-monitor/internal/domain/entity"
)

// ChannelDingtalkHandler 钉钉机器人 webhook
type ChannelDingtalkHandler struct{}

// dingtalkParams 通道参数（接口约定小写）
type dingtalkParams struct {
	Addr   string `json:"addr"`   // 钉钉机器人 webhook 地址
	Secret string `json:"secret"` // 可选，配了启用加签
}

func (h *ChannelDingtalkHandler) GetClassName() string { return "ChannelDingtalkHandler" }

func (h *ChannelDingtalkHandler) TestDispatchMessage(channel entity.AlertChannel, _ string, message string) error {
	if strings.TrimSpace(message) == "" {
		message = "dragonfly-monitor 通道测试消息"
	}
	return h.DispatchMessage(channel, nil, message)
}

func (h *ChannelDingtalkHandler) DispatchMessage(channel entity.AlertChannel, _ []entity.SysUser, message string) error {
	var p dingtalkParams
	if err := json.Unmarshal([]byte(channel.Params), &p); err != nil {
		return err
	}
	if p.Addr == "" {
		return errors.New("钉钉 webhook 地址为空")
	}
	// 钉钉机器人 markdown 消息：title 为通知栏标题，text 为正文（支持 markdown）
	body, _ := json.Marshal(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": "报警提醒",
			"text":  message,
		},
	})
	requestURL := p.Addr
	if p.Secret != "" {
		requestURL = dingtalkSignURL(p.Addr, p.Secret)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(requestURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errors.New("钉钉发送失败: " + resp.Status)
	}
	// 钉钉响应 {"errcode":0,"errmsg":"ok"}，非 0 视为失败（如 IP 白名单/加签错误）
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.ErrCode != 0 {
		return errors.New("钉钉发送失败: " + result.ErrMsg)
	}
	return nil
}

// dingtalkSignURL 钉钉加签：毫秒时间戳 + "\n" + secret 做 HmacSHA256，base64 后 URL 编码拼到 webhook
func dingtalkSignURL(addr, secret string) string {
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sep := "&"
	if !strings.Contains(addr, "?") {
		sep = "?"
	}
	return fmt.Sprintf("%s%stimestamp=%d&sign=%s", addr, sep, timestamp, url.QueryEscape(sign))
}
