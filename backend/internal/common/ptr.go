package common

// Ptr 返回入参的指针，避免在调用处直接对局部变量取地址导致逃逸到堆。
// Go 1.26 vet 建议用此方式替代 v := expr; &v 的写法。
func Ptr[T any](v T) *T {
	return &v
}
