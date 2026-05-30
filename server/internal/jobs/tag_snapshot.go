package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskTagSnapshot 标签快照（每日 04:30）。
const TaskTagSnapshot = "tag:snapshot"

// TagSnapshotter 由 jobs 调用。
type TagSnapshotter interface {
	MonthlySnapshot(ctx context.Context) error
}

// NewTagSnapshotHandler 构造标签快照 Handler。
func NewTagSnapshotHandler(svc TagSnapshotter) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := svc.MonthlySnapshot(ctx); err != nil {
			logger.L().Error(TaskTagSnapshot+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskTagSnapshot + " done")
		return nil
	}
}
