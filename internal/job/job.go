package job

import (
	"context"
	"sync"
	"time"

	"e_commerce_platform/internal/applog"
	"e_commerce_platform/internal/model"
	"e_commerce_platform/internal/repository"

	"gorm.io/gorm"
)

// Runner 后台任务（阶段 H · 3.3）：定期扫超时未支付订单并取消。
type Runner struct {
	// Timeout pending 超过此时长则取消（如 30m）
	Timeout time.Duration
	// Tick 扫描周期；为 0 时默认 1m
	Tick time.Duration

	Orders repository.OrderStore
	DB     *gorm.DB

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (r *Runner) cancelTimedOut(ctx context.Context) {
	if r.DB == nil || r.Orders == nil {
		return
	}

	var orders []model.Order
	if err := r.DB.WithContext(ctx).
		Where("status = ?", model.OrderPending).
		Find(&orders).Error; err != nil {
		applog.FromContext(ctx).Warn("job list pending failed", "err", err.Error())
		return
	}

	for _, order := range orders {
		if order.OrderAt == nil {
			continue
		}
		if time.Since(*order.OrderAt) <= r.Timeout {
			continue
		}
		if err := r.Orders.UpdateStatus(ctx, order.ID, model.OrderPending, model.OrderCancelled); err != nil {
			applog.FromContext(ctx).Warn("job cancel timeout order failed", "order_id", order.ID, "err", err.Error())
			continue
		}
		applog.FromContext(ctx).Info("job cancel timeout order ok", "order_id", order.ID)
	}
}

// Start 在独立 goroutine 跑循环；main 里调用。
func (r *Runner) Start(parent context.Context) {
	if r.Timeout <= 0 {
		r.Timeout = 30 * time.Minute
	}
	tick := r.Tick
	if tick <= 0 {
		tick = time.Minute
	}

	ctx, cancel := context.WithCancel(parent)
	r.cancel = cancel
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				applog.FromContext(ctx).Info("job tick", "timeout", r.Timeout.String())
				r.cancelTimedOut(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop 优雅停机时调用，等待当前循环退出。
func (r *Runner) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
}
