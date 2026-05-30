package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskShareClickFlush 分享点击批量入库（每分钟）。
const TaskShareClickFlush = "share:click-flush"

// ShareClickFlusher 由 jobs 调用。
type ShareClickFlusher interface {
	ShareClickFlush(ctx context.Context) (int, error)
}

// NewShareClickFlushHandler 构造分享点击 flush Handler。
func NewShareClickFlushHandler(svc ShareClickFlusher) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.ShareClickFlush(ctx)
		if err != nil {
			logger.L().Error(TaskShareClickFlush+" failed", zap.Error(err))
			return err
		}
		logger.L().Debug(TaskShareClickFlush+" done", zap.Int("flushed", n))
		return nil
	}
}
