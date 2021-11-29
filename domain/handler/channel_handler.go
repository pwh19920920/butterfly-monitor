package handler

import "butterfly-monitor/domain/entity"
import sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"

type ChannelHandler interface {

	// DispatchMessage 分发消息【特殊参数，分发对象】
	DispatchMessage(channel entity.AlertChannel, groupUsers []sysEntity.SysUser, message string) error
}
