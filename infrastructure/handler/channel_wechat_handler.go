package handler

import (
	"butterfly-monitor/domain/entity"
	"encoding/json"
	"github.com/kirinlabs/HttpRequest"
	sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"
)

type ChannelWechatHandler struct {
}

type ChannelWechatHandlerParams struct {
	Addr string `json:"addr"`
}

type ChannelWechatHandlerRequestBody struct {
	Content string `json:"content"`
}

type ChannelWechatHandlerRequest struct {
	MsgType string                          `json:"msgtype"`
	Text    ChannelWechatHandlerRequestBody `json:"text"`
}

// DispatchMessage 分发消息【特殊参数，分发对象】
func (channelHandler ChannelWechatHandler) DispatchMessage(channel entity.AlertChannel, groupUsers []sysEntity.SysUser, message string) error {
	var params ChannelWechatHandlerParams
	err := json.Unmarshal([]byte(channel.Params), &params)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(ChannelWechatHandlerRequest{
		MsgType: "text",
		Text: ChannelWechatHandlerRequestBody{
			Content: message,
		},
	})
	req := HttpRequest.NewRequest()
	_, err = req.Post(params.Addr, string(data))
	return err
}
