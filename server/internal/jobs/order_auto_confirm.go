package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskOrderAutoConfirm 自动确认收货任务名。
const TaskOrderAutoConfirm = "order:auto_confirm"

// OrderAutoConfirmPayload 自动确认任务 payload。
type OrderAutoConfirmPayload struct {
	OrderID int64 `json:"order_id"`
}

// OrderAutoConfirmer 供 job 调用的自动确认接口。
type OrderAutoConfirmer interface {
	GetRaw(ctx context.Context, orderID int64) (OrderInfo, error)
	Transition(ctx context.Context, orderID int64, trigger, operatorType string, operatorID int64, reason string) error
}

// NewOrderAutoConfirmHandler 构造自动确认收货 Handler。
func NewOrderAutoConfirmHandler(svc OrderAutoConfirmer) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p OrderAutoConfirmPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("order:auto_confirm unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		return handleOrderAutoConfirm(ctx, p.OrderID, svc)
	}
}

func handleOrderAutoConfirm(ctx context.Context, orderID int64, svc OrderAutoConfirmer) error {
	info, err := svc.GetRaw(ctx, orderID)
	if err != nil {
		logger.L().Error("order:auto_confirm get order", zap.Int64("order_id", orderID), zap.Error(err))
		return err
	}

	// 幂等：仅 shipped 状态触发
	if info.Status != "shipped" {
		return nil
	}

	if err := svc.Transition(ctx, orderID, "auto_confirm", "system", 0, "系统自动确认收货"); err != nil {
		logger.L().Error("order:auto_confirm transition", zap.Int64("order_id", orderID), zap.Error(err))
		return err
	}

	logger.L().Info("order:auto_confirm done", zap.Int64("order_id", orderID))
	return nil
}
