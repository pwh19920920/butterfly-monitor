package security

type EncodeService interface {
	// Encode 数据加密
	Encode(text, salt string) string
}
