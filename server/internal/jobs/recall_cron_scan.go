package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskRecallCronScan 召回 cron 扫描（每 10 分钟）。
const TaskRecallCronScan = "recall:cron-scan"

// RecallScanner 由 jobs 调用。
type RecallScanner interface {
	ScheduleScan(ctx context.Context) error
}

// NewRecallCronScanHandler 构造召回 cron 扫描 Handler。
func NewRecallCronScanHandler(svc RecallScanner) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := svc.ScheduleScan(ctx); err != nil {
			logger.L().Error(TaskRecallCronScan+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskRecallCronScan + " done")
		return nil
	}
}
