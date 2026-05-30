package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCommissionRecompute distributor 维度佣金累计统计（每日 05:30）。
const TaskCommissionRecompute = "commission:recompute"

// CommissionStatsRecomputer 由 jobs 调用。
type CommissionStatsRecomputer interface {
	CommissionRecomputeStats(ctx context.Context) error
}

// NewCommissionRecomputeHandler 构造佣金重算 Handler。
func NewCommissionRecomputeHandler(svc CommissionStatsRecomputer) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := svc.CommissionRecomputeStats(ctx); err != nil {
			logger.L().Error(TaskCommissionRecompute+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCommissionRecompute + " done")
		return nil
	}
}
