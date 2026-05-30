package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCommissionUnfreeze 佣金冻结期到期放行（每日 05:00）。
const TaskCommissionUnfreeze = "commission:unfreeze"

// CommissionUnfreezer 由 jobs 调用。
type CommissionUnfreezer interface {
	FreezeReleaseScan(ctx context.Context) (int, error)
}

// NewCommissionUnfreezeHandler 构造佣金解冻 Handler。
func NewCommissionUnfreezeHandler(svc CommissionUnfreezer) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.FreezeReleaseScan(ctx)
		if err != nil {
			logger.L().Error(TaskCommissionUnfreeze+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCommissionUnfreeze+" done", zap.Int("released", n))
		return nil
	}
}
