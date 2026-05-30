package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskWithdrawReconcile 提现对账（每日 06:00）。
const TaskWithdrawReconcile = "withdraw:reconcile"

// WithdrawReconciler 由 jobs 调用。
type WithdrawReconciler interface {
	WithdrawReconcile(ctx context.Context) error
}

// NewWithdrawReconcileHandler 构造提现对账 Handler。
func NewWithdrawReconcileHandler(svc WithdrawReconciler) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := svc.WithdrawReconcile(ctx); err != nil {
			logger.L().Error(TaskWithdrawReconcile+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskWithdrawReconcile + " done")
		return nil
	}
}
