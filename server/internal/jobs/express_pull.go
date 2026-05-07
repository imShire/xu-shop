package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskExpressPull 快递轨迹兜底拉取任务名（每 2 小时执行）。
const TaskExpressPull = "express:pull_stale"

// StaleShipmentPuller 兜底拉取接口（由 shipping.Service 实现）。
type StaleShipmentPuller interface {
	PullStaleShipments(ctx context.Context) error
}

// NewExpressPullHandler 构造快递兜底拉取 Handler。
func NewExpressPullHandler(puller StaleShipmentPuller) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		logger.L().Info("express:pull_stale start")
		if err := puller.PullStaleShipments(ctx); err != nil {
			logger.L().Error("express:pull_stale failed", zap.Error(err))
			return err
		}
		logger.L().Info("express:pull_stale done")
		return nil
	}
}
