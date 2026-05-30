package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPointExpireScan 积分过期扫描（每日 02:30）。
const TaskPointExpireScan = "point:expire-scan"

// PointExpireScanner 由 jobs 调用。
type PointExpireScanner interface {
	ExpireScan(ctx context.Context, batchSize int) (int, error)
}

// NewPointExpireScanHandler 构造积分过期扫描 Handler。
func NewPointExpireScanHandler(svc PointExpireScanner) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.ExpireScan(ctx, 500)
		if err != nil {
			logger.L().Error(TaskPointExpireScan+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskPointExpireScan+" done", zap.Int("processed", n))
		return nil
	}
}
