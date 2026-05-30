package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskTagRecompute 用户标签日重算（每日 04:00 + 手动）。
const TaskTagRecompute = "tag:recompute"

// TagRecomputer 由 jobs 调用，全量重算 RFM + 生命周期标签。
type TagRecomputer interface {
	RecomputeAll(ctx context.Context) error
}

// NewTagRecomputeHandler 构造标签重算 Handler。
func NewTagRecomputeHandler(svc TagRecomputer) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := svc.RecomputeAll(ctx); err != nil {
			logger.L().Error(TaskTagRecompute+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskTagRecompute + " done")
		return nil
	}
}
