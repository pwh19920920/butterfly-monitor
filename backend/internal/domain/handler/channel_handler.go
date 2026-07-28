package handler

import "dragonfly-monitor/internal/domain/entity"

// ChannelHandler 告警通道发送抽象
type ChannelHandler interface {
	// GetClassName 处理器类名，与 AlertChannel.Handler 匹配
	GetClassName() string
	// DispatchMessage 向分组用户发送消息
	DispatchMessage(channel entity.AlertChannel, users []entity.SysUser, message string) error
	// TestDispatchMessage 测试发送，testTarget 为测试接收人（邮件地址等），message 为测试模板内容
	TestDispatchMessage(channel entity.AlertChannel, testTarget string, message string) error
}
