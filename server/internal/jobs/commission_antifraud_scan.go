package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskCommissionAntifraudScan 佣金反作弊扫描（每小时）。
const TaskCommissionAntifraudScan = "commission:antifraud-scan"

// CommissionAntifraudScanner 由 jobs 调用。
type CommissionAntifraudScanner interface {
	AntifraudScan(ctx context.Context) (int, error)
}

// NewCommissionAntifraudScanHandler 构造反作弊扫描 Handler。
func NewCommissionAntifraudScanHandler(svc CommissionAntifraudScanner) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.AntifraudScan(ctx)
		if err != nil {
			logger.L().Error(TaskCommissionAntifraudScan+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskCommissionAntifraudScan+" done", zap.Int("marked", n))
		return nil
	}
}
