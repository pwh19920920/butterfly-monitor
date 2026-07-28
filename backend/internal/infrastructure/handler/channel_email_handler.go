package handler

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"dragonfly-monitor/internal/domain/entity"
)

// ChannelEmailHandler 邮件通道（HTML 正文）
type ChannelEmailHandler struct{}

// emailParams 通道参数（接口约定小写）
type emailParams struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	SSL      bool   `json:"ssl"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *ChannelEmailHandler) GetClassName() string { return "ChannelEmailHandler" }

func (h *ChannelEmailHandler) TestDispatchMessage(channel entity.AlertChannel, testTarget string, message string) error {
	// 空收件人时仅校验参数，避免创建通道时误发
	if strings.TrimSpace(testTarget) == "" {
		_, err := parseEmailParams(channel.Params)
		return err
	}
	if strings.TrimSpace(message) == "" {
		message = "<p>dragonfly-monitor 通道测试消息</p>"
	}
	return h.DispatchMessage(channel, []entity.SysUser{{Email: testTarget}}, message)
}

func (h *ChannelEmailHandler) DispatchMessage(channel entity.AlertChannel, users []entity.SysUser, message string) error {
	p, err := parseEmailParams(channel.Params)
	if err != nil {
		return err
	}
	to := collectEmails(users)
	if len(to) == 0 {
		return errors.New("无收件人")
	}

	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	msg := buildHTMLMail(p.Username, to, "报警提醒", message)
	auth := smtp.PlainAuth("", p.Username, p.Password, p.Host)
	if !p.SSL {
		return smtp.SendMail(addr, auth, p.Username, to, msg)
	}
	return sendMailTLS(addr, p.Host, p.Username, to, auth, msg)
}

// buildHTMLMail 组装 HTML 邮件：UTF-8 主题编码 + text/html 正文
func buildHTMLMail(from string, to []string, subject, htmlBody string) []byte {
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	// 正文按 HTML 发送，模板侧配置完整 HTML 片段即可
	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s",
		from, strings.Join(to, ","), encodedSubject, htmlBody,
	))
}

func parseEmailParams(raw string) (emailParams, error) {
	var p emailParams
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return p, err
	}
	if p.Host == "" || p.Port == 0 || p.Username == "" {
		return p, errors.New("邮件参数不完整")
	}
	return p, nil
}

func collectEmails(users []entity.SysUser) []string {
	to := make([]string, 0, len(users))
	for _, u := range users {
		if u.Email != "" {
			to = append(to, u.Email)
		}
	}
	return to
}

// sendMailTLS 隐式 TLS（465 常见）发送：拨号 → AUTH → MAIL/RCPT → DATA
func sendMailTLS(addr, host, from string, to []string, auth smtp.Auth, msg []byte) error {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
