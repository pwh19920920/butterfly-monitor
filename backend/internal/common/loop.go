package common

import (
	"context"
	"time"

	"github.com/pwh19920920/butterfly/pkg/logger"
)

// RunSafeLoop 周期性执行 fn 的后台循环骨架。
//
// 顶层 recover 兜底防单次 panic 拖垮整个进程；每次迭代独立 recover，
// 保证某次 fn 异常不会让循环 goroutine 永久退出。
// 进循环前先立即执行一次，避免消费者打开页面看到空历史。
func RunSafeLoop(name string, interval time.Duration, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.ErrorFormat(context.Background(), "%s panic exit: %v", name, r)
		}
	}()

	fn()
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.ErrorFormat(context.Background(), "%s iteration panic: %v", name, r)
				}
			}()
			fn()
		}()
		time.Sleep(interval)
	}
}
