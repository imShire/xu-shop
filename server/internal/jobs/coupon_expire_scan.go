package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCouponExpireScan 优惠券过期扫描（每日 02:00）。
const TaskCouponExpireScan = "coupon:expire-scan"

// CouponExpireScanner 由 jobs 调用，扫描并把过期 unused 券标记为 expired。
type CouponExpireScanner interface {
	ExpireScan(ctx context.Context, batchSize int) (int, error)
}

// NewCouponExpireScanHandler 构造券过期扫描 Handler。
func NewCouponExpireScanHandler(svc CouponExpireScanner) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.ExpireScan(ctx, 500)
		if err != nil {
			logger.L().Error(TaskCouponExpireScan+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCouponExpireScan+" done", zap.Int("processed", n))
		return nil
	}
}
