package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	mkpoint "github.com/xushop/xu-shop/internal/modules/marketing/point"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPointEarnGrant 积分入账兜底（每日 04:00）。
const TaskPointEarnGrant = "point:earn-grant"

// PointEarnGranter 由 jobs 调用：兜底入账已完成但未发积分的订单。
type PointEarnGranter interface {
	EarnGrantFromOrders(ctx context.Context, src mkpoint.OrderEarnSource, limit int) (int, error)
}

// NewPointEarnGrantHandler 构造积分入账 Handler。src 为 nil 时 service 内 no-op。
func NewPointEarnGrantHandler(svc PointEarnGranter, src mkpoint.OrderEarnSource) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.EarnGrantFromOrders(ctx, src, 500)
		if err != nil {
			logger.L().Error(TaskPointEarnGrant+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskPointEarnGrant+" done", zap.Int("granted", n))
		return nil
	}
}
