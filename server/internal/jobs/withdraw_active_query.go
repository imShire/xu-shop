package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskWithdrawActiveQuery 提现主动查单（每 5 分钟）。
const TaskWithdrawActiveQuery = "withdraw:active-query"

// WithdrawActiveQuerier 由 jobs 调用。
type WithdrawActiveQuerier interface {
	WithdrawActiveQuery(ctx context.Context, limit int) (int, error)
}

// NewWithdrawActiveQueryHandler 构造提现查单 Handler。
func NewWithdrawActiveQueryHandler(svc WithdrawActiveQuerier) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.WithdrawActiveQuery(ctx, 200)
		if err != nil {
			logger.L().Error(TaskWithdrawActiveQuery+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskWithdrawActiveQuery+" done", zap.Int("queried", n))
		return nil
	}
}
