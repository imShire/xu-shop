package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	mkcoupon "github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCouponExpireWarning 优惠券到期提醒（每日 10:00）。
const TaskCouponExpireWarning = "coupon:expire-warning"

// CouponExpireWarner 由 jobs 调用，触发即将过期的券提醒。
type CouponExpireWarner interface {
	ExpireWarning(ctx context.Context, dispatcher mkcoupon.NotificationDispatcher, withinDays int, limit int) (int, error)
}

// NewCouponExpireWarningHandler 构造券到期提醒 Handler。
//
// dispatcher 为 nil 时仅扫描计数（开发环境）。
func NewCouponExpireWarningHandler(svc CouponExpireWarner, dispatcher mkcoupon.NotificationDispatcher) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.ExpireWarning(ctx, dispatcher, 3, 1000)
		if err != nil {
			logger.L().Error(TaskCouponExpireWarning+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCouponExpireWarning+" done", zap.Int("sent", n))
		return nil
	}
}
