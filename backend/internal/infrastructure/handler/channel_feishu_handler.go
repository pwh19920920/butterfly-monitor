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
	"strings"
	"time"

	"dragonfly-monitor/internal/domain/entity"
)

// ChannelFeishuHandler 飞书机器人 webhook
type ChannelFeishuHandler struct{}

// feishuParams 通道参数（接口约定小写）
type feishuParams struct {
	Addr   string `json:"addr"`   // 飞书机器人 webhook 地址
	Secret string `json:"secret"` // 可选，配了启用加签
}

func (h *ChannelFeishuHandler) GetClassName() string { return "ChannelFeishuHandler" }

func (h *ChannelFeishuHandler) TestDispatchMessage(channel entity.AlertChannel, _ string, message string) error {
	if strings.TrimSpace(message) == "" {
		message = "dragonfly-monitor 通道测试消息"
	}
	return h.DispatchMessage(channel, nil, message)
}

func (h *ChannelFeishuHandler) DispatchMessage(channel entity.AlertChannel, _ []entity.SysUser, message string) error {
	var p feishuParams
	if err := json.Unmarshal([]byte(channel.Params), &p); err != nil {
		return err
	}
	if p.Addr == "" {
		return errors.New("飞书 webhook 地址为空")
	}
	// 飞书机器人 text 消息
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}
	if p.Secret != "" {
		timestamp, sign := feishuSign(p.Secret)
		payload["timestamp"] = timestamp
		payload["sign"] = sign
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(p.Addr, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errors.New("飞书发送失败: " + resp.Status)
	}
	// 飞书响应 {"code":0,"msg":"success"}，非 0 视为失败（如加签校验失败）
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return errors.New("飞书发送失败: " + result.Msg)
	}
	return nil
}

// feishuSign 飞书加签：秒级时间戳 + "\n" + secret 做 HmacSHA256，base64 标准编码
func feishuSign(secret string) (string, string) {
	timestamp := time.Now().Unix()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d", timestamp), sign
}
