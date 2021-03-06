package handler

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"
	"net/smtp"
)

type ChannelEmailHandler struct {
}

type ChannelEmailHandlerParamsSSL int32

const ChannelEmailHandlerParamsSSLOpen ChannelEmailHandlerParamsSSL = 1
const ChannelEmailHandlerParamsSSLClose ChannelEmailHandlerParamsSSL = 2

type ChannelEmailHandlerParams struct {
	Host     string                       `json:"host"`
	Port     int32                        `json:"port"`
	SSL      ChannelEmailHandlerParamsSSL `json:"ssl"`
	Username string                       `json:"username"`
	Password string                       `json:"password"`
}

func (channelHandler ChannelEmailHandler) GetClassName() string {
	return "ChannelEmailHandler"
}

func (channelHandler ChannelEmailHandler) TestDispatchMessage(channel entity.AlertChannel, params types.AlertChannelTestParams) error {
	groupUsers := []sysEntity.SysUser{sysEntity.SysUser{Email: params.Email}}
	return channelHandler.DispatchMessage(channel, groupUsers, params.Template)
}

// DispatchMessage 分发消息【特殊参数，分发对象】
func (channelHandler ChannelEmailHandler) DispatchMessage(channel entity.AlertChannel, groupUsers []sysEntity.SysUser, message string) error {
	var params ChannelEmailHandlerParams
	err := json.Unmarshal([]byte(channel.Params), &params)
	if err != nil {
		return err
	}

	emails := make([]string, 0)
	for _, item := range groupUsers {
		emails = append(emails, item.Email)
	}

	if emails == nil || len(emails) == 0 {
		return errors.New("待发送邮箱不存在")
	}

	em := email.NewEmail()
	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	em.From = fmt.Sprintf("spider-monitor监控系统 <%s>", params.Username)

	// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
	em.To = emails

	// 设置主题
	em.Subject = "报警提醒"

	// 简单设置文件发送的内容，暂时设置成纯文本
	em.Text = []byte(message)

	//设置服务器相关的配置
	addr := fmt.Sprintf("%s:%v", params.Host, params.Port)
	auth := smtp.PlainAuth("", params.Username, params.Password, params.Host)
	if params.SSL != ChannelEmailHandlerParamsSSLOpen {
		return em.Send(addr, auth)
	}
	return em.SendWithTLS(addr, auth, &tls.Config{ServerName: params.Host})
}
