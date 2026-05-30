package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPointRollback 售后退款时冲销已发积分（事件触发）。
const TaskPointRollback = "point:rollback"

// PointRollbackPayload payload。
type PointRollbackPayload struct {
	UserID  int64 `json:"user_id"`
	OrderID int64 `json:"order_id"`
	// PointChange 待返还/冲销的积分绝对值；正数表示返还（refund）。
	PointChange int64 `json:"point_change"`
}

// PointRefunder 由 jobs 调用：冲销订单已发积分。
type PointRefunder interface {
	Refund(ctx context.Context, userID int64, change int64, refID int64, idemKey string) error
}

// NewPointRollbackHandler 构造积分冲销 Handler。
func NewPointRollbackHandler(svc PointRefunder) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p PointRollbackPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error(TaskPointRollback+" unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		if p.UserID <= 0 || p.OrderID <= 0 || p.PointChange == 0 {
			return nil
		}
		idem := fmt.Sprintf("point_rollback:%d", p.OrderID)
		if err := svc.Refund(ctx, p.UserID, p.PointChange, p.OrderID, idem); err != nil {
			logger.L().Warn(TaskPointRollback+" refund failed",
				zap.Int64("user_id", p.UserID), zap.Int64("order_id", p.OrderID), zap.Error(err))
			return err
		}
		return nil
	}
}
