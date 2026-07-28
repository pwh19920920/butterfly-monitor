package types

import (
	"dragonfly-monitor/internal/domain/entity"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type AlertChannelQueryRequest struct {
	response.RequestPaging
	Name string                   `form:"name"`
	Type *entity.AlertChannelType `form:"type"`
}

// AlertChannelHandlerVO 通道类型与可用处理器的绑定关系
type AlertChannelHandlerVO struct {
	ChannelType int32    `json:"channelType"`
	Handlers    []string `json:"handlers"`
}

// AlertChannelTestParams 测试发送参数（不入库，仅本次请求用于触发测试发送）
// 消息内容由通道 template / handler 默认模板 + 假参数渲染，无需再传测试模板
type AlertChannelTestParams struct {
	Email string `json:"email"` // 测试接收人邮箱（仅邮件通道）
}

// AlertChannelSaveRequest 创建/修改通道请求体，携带临时测试参数
type AlertChannelSaveRequest struct {
	entity.AlertChannel
	TestParams AlertChannelTestParams `json:"testParams"`
}
