package job

import (
	"context"
	"sync"
	"time"
)

// Runner 后台任务（阶段 H · 3.3）。
// 最小实现：ticker + 日志；进阶：扫 pending 超时订单自动取消。
type Runner struct {
	Interval time.Duration
	// TODO(H3): 注入 OrderStore / DB 等依赖

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Start 在独立 goroutine 跑循环；main 里调用。
func (r *Runner) Start(parent context.Context) {
	// TODO(H3):
	//  ctx, r.cancel = context.WithCancel(parent)
	//  r.wg.Add(1)
	//  go func() { defer r.wg.Done(); ticker ... select <-ctx.Done() }
	_ = parent
}

// Stop 优雅停机时调用，等待当前 tick 结束。
func (r *Runner) Stop() {
	// TODO(H3): r.cancel(); r.wg.Wait()
}
