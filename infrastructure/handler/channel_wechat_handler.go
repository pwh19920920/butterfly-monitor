package handler

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
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

func (channelHandler ChannelWechatHandler) GetClassName() string {
	return "ChannelWechatHandler"
}

func (channelHandler ChannelWechatHandler) TestDispatchMessage(channel entity.AlertChannel, params types.AlertChannelTestParams) error {
	groupUsers := make([]sysEntity.SysUser, 0)
	return channelHandler.DispatchMessage(channel, groupUsers, params.Template)
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
