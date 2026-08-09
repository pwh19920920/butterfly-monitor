package system

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"butterfly-monitor/internal/common"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// 环形缓冲容量：30s 采一次，60 点 ≈ 30 分钟趋势
const historySize = 60

// 采集间隔
const collectInterval = 30 * time.Second

// CPU 采样窗口（阻塞时长），gopsutil 用 interval 计算两次采样差值
const cpuSampleInterval = 500 * time.Millisecond

// ringBuffer 定长环形缓冲，并发安全（采集 goroutine 写，HTTP 请求读）
type ringBuffer struct {
	mu   sync.RWMutex
	buf  []float64
	head int // 下一个写入位置
	size int // 已写入数量（< cap 时为实际数量）
}

func newRing(cap int) ringBuffer {
	return ringBuffer{buf: make([]float64, cap)}
}

// push 写入一个采样点，满则覆盖最旧
func (r *ringBuffer) push(v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = v
	r.head = (r.head + 1) % len(r.buf)
	if r.size < len(r.buf) {
		r.size++
	}
}

// snapshot 返回老→新有序的历史切片（拷贝，避免调用方误改）
func (r *ringBuffer) snapshot() []float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]float64, r.size)
	if r.size < len(r.buf) {
		// 未满：0..head-1 即老→新
		copy(out, r.buf[:r.size])
		return out
	}
	// 已满：head..end + start..head-1 即老→新
	n := copy(out, r.buf[r.head:])
	copy(out[n:], r.buf[:r.head])
	return out
}

// current 返回最近一次采样值，无数据返回 0
func (r *ringBuffer) current() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.size == 0 {
		return 0
	}
	// 最近写入位置 = head-1
	last := r.head - 1
	if last < 0 {
		last = len(r.buf) - 1
	}
	return r.buf[last]
}

// series 取出 MetricSeries（current + history）
func (r *ringBuffer) series() MetricSeries {
	return MetricSeries{
		Current: r.current(),
		History: r.snapshot(),
	}
}

// Collector 系统指标采集器：后台 goroutine 低频采样，内存环形缓冲，不入库
type Collector struct {
	cpu       ringBuffer
	mem       ringBuffer
	disk      ringBuffer
	goroutine ringBuffer
	gc        ringBuffer
	rss       ringBuffer

	pid int32
}

func NewCollector() *Collector {
	return &Collector{
		cpu:       newRing(historySize),
		mem:       newRing(historySize),
		disk:      newRing(historySize),
		goroutine: newRing(historySize),
		gc:        newRing(historySize),
		rss:       newRing(historySize),
		pid:       int32(os.Getpid()),
	}
}

// Run 后台采集循环，照搬 refreshDatabaseConnections 的双层 recover 范式：
// Run 启动后台采集循环：进循环前先采一次，避免首页打开 history 为空。
// 双层 recover + 周期执行由 common.RunSafeLoop 提供
func (c *Collector) Run() {
	common.RunSafeLoop("system metrics collector", collectInterval, c.collect)
}

// collect 一次性采集所有指标，push 进各 ring；单项失败不影响其余
func (c *Collector) collect() {
	ctx := context.Background()

	// CPU 使用率（阻塞 cpuSampleInterval 采样）
	if percents, err := cpu.Percent(cpuSampleInterval, false); err == nil && len(percents) > 0 {
		c.cpu.push(percents[0])
	} else {
		logger.Warn(ctx, "collect cpu percent fail", err)
	}

	// 内存使用率
	if vm, err := mem.VirtualMemory(); err == nil {
		c.mem.push(vm.UsedPercent)
	} else {
		logger.Warn(ctx, "collect mem fail", err)
	}

	// 磁盘使用率：Linux 根分区 /，Windows C 盘
	mount := "/"
	if runtime.GOOS == "windows" {
		mount = "C:"
	}
	if du, err := disk.Usage(mount); err == nil {
		c.disk.push(du.UsedPercent)
	} else {
		logger.Warn(ctx, "collect disk fail", err)
	}

	// Go goroutine 数
	c.goroutine.push(float64(runtime.NumGoroutine()))

	// Go GC 累计次数
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c.gc.push(float64(ms.NumGC))

	// 进程 RSS 内存 MB
	if p, err := process.NewProcess(c.pid); err == nil {
		if rss, err := p.MemoryInfo(); err == nil && rss != nil {
			c.rss.push(float64(rss.RSS) / 1024 / 1024)
		} else {
			logger.Warn(ctx, "collect process rss fail", err)
		}
	} else {
		logger.Warn(ctx, "new process fail", err)
	}
}

// Snapshot 返回所有指标的当前值 + 历史(老→新)
func (c *Collector) Snapshot() MetricsResponse {
	return MetricsResponse{
		CpuPercent:     c.cpu.series(),
		MemPercent:     c.mem.series(),
		DiskPercent:    c.disk.series(),
		GoroutineCount: c.goroutine.series(),
		GoGcCount:      c.gc.series(),
		ProcessRssMB:   c.rss.series(),
	}
}
