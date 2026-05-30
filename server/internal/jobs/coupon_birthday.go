package jobs

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	mkcoupon "github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCouponBirthdayCron 生日券发放（每日 09:00）。
const TaskCouponBirthdayCron = "coupon:birthday-cron"

// CouponBirthdayGranter 由 jobs 调用：派发今日生日用户的生日券。
type CouponBirthdayGranter interface {
	BirthdayGrant(ctx context.Context,
		src mkcoupon.BirthdayUserSource,
		resolver mkcoupon.BirthdayTemplateResolver,
		rdbSetNX func(ctx context.Context, key string, ttl time.Duration) (bool, error)) (int, error)
}

// NewCouponBirthdayCronHandler 构造生日券 Handler。
//
// src/resolver 任一为 nil → Service 内 no-op；适合开发环境。
func NewCouponBirthdayCronHandler(
	svc CouponBirthdayGranter,
	src mkcoupon.BirthdayUserSource,
	resolver mkcoupon.BirthdayTemplateResolver,
	rdbSetNX func(ctx context.Context, key string, ttl time.Duration) (bool, error),
) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.BirthdayGrant(ctx, src, resolver, rdbSetNX)
		if err != nil {
			logger.L().Error(TaskCouponBirthdayCron+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCouponBirthdayCron+" done", zap.Int("granted", n))
		return nil
	}
}
