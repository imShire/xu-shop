package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	mkpoint "github.com/xushop/xu-shop/internal/modules/marketing/point"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPointExpireWarning 积分到期提醒（每日 10:00）。
const TaskPointExpireWarning = "point:expire-warning"

// PointExpireWarner 由 jobs 调用。
type PointExpireWarner interface {
	ExpireWarning(ctx context.Context, dispatcher mkpoint.NotificationDispatcher, withinDays int, limit int) (int, error)
}

// NewPointExpireWarningHandler 构造积分到期提醒 Handler。
func NewPointExpireWarningHandler(svc PointExpireWarner, dispatcher mkpoint.NotificationDispatcher) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.ExpireWarning(ctx, dispatcher, 7, 1000)
		if err != nil {
			logger.L().Error(TaskPointExpireWarning+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskPointExpireWarning+" done", zap.Int("sent", n))
		return nil
	}
}
