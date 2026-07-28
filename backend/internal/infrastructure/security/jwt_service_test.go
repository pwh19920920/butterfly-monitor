package security

import (
	"testing"
	"time"
)

func TestJwtGenericAndCheck(t *testing.T) {
	svc := NewJwtServiceImpl()
	secret := "unit-test-secret"
	subject := "user-123"
	exp := time.Now().Add(1 * time.Hour)

	token, err := svc.GenericToken(secret, subject, exp)
	if err != nil {
		t.Fatalf("GenericToken fail: %v", err)
	}
	if token == "" {
		t.Fatal("GenericToken returned empty token")
	}

	// 正确 secret 应校验通过
	if !svc.CheckToken(token, secret) {
		t.Fatal("CheckToken with correct secret should be true")
	}
	// 错误 secret 应失败
	if svc.CheckToken(token, "wrong-secret") {
		t.Fatal("CheckToken with wrong secret should be false")
	}
	// 过期 token 应失败
	expired, _ := svc.GenericToken(secret, subject, time.Now().Add(-1*time.Minute))
	if svc.CheckToken(expired, secret) {
		t.Fatal("CheckToken with expired token should be false")
	}

	// GetSubjectFromToken 应解出 subject（注意：该方法本身不验签，仅解析 payload）
	got, err := svc.GetSubjectFromToken(token)
	if err != nil {
		t.Fatalf("GetSubjectFromToken fail: %v", err)
	}
	if got != subject {
		t.Fatalf("GetSubjectFromToken subject mismatch: want=%s got=%s", subject, got)
	}
}
