package security

import (
	"time"
)

type TokenService interface {
	// GenericToken 生成令牌
	GenericToken(secret, subject string, expireTime time.Time) (string, error)

	// GetSubjectFromToken 获取Subject
	GetSubjectFromToken(token string) (string, error)

	// CheckToken 校验令牌
	CheckToken(token, secret string) bool
}
