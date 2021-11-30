package handler

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)
import sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"

type ChannelHandler interface {
	// GetClassName 获取实现名称
	GetClassName() string

	// DispatchMessage 分发消息【特殊参数，分发对象】
	DispatchMessage(channel entity.AlertChannel, groupUsers []sysEntity.SysUser, message string) error

	// TestDispatchMessage 测试消息分发
	TestDispatchMessage(channel entity.AlertChannel, params types.AlertChannelTestParams) error
}
