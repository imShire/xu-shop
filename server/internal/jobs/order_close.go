// Package jobs 实现 asynq 后台任务处理器。
package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskOrderClose 超时关单任务名。
const TaskOrderClose = "order:close"

// OrderClosePayload 超时关单任务 payload。
type OrderClosePayload struct {
	OrderID int64 `json:"order_id"`
}

// OrderCloser 供 job 调用的订单操作接口。
type OrderCloser interface {
	// GetRaw 裸查询订单（不校验归属）。
	GetRaw(ctx context.Context, orderID int64) (OrderInfo, error)
	// Transition 状态机变更。
	Transition(ctx context.Context, orderID int64, trigger, operatorType string, operatorID int64, reason string) error
	// ReleaseStock 释放库存。
	ReleaseStock(ctx context.Context, orderID int64, orderNo string)
	// RefundUserBalance 退还订单余额（订单已扣余额时调用）。
	RefundUserBalance(ctx context.Context, orderID int64) error
}

// OrderInfo 订单关键信息（供 job 使用）。
type OrderInfo struct {
	Status   string
	OrderNo  string
	ExpireAt time.Time
}

// NewOrderCloseHandler 构造超时关单 Handler。
func NewOrderCloseHandler(svc OrderCloser) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p OrderClosePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("order:close unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		return handleOrderClose(ctx, p.OrderID, svc)
	}
}

func handleOrderClose(ctx context.Context, orderID int64, svc OrderCloser) error {
	info, err := svc.GetRaw(ctx, orderID)
	if err != nil {
		logger.L().Error("order:close get order", zap.Int64("order_id", orderID), zap.Error(err))
		return err
	}

	// 幂等：非 pending 直接跳过
	if info.Status != "pending" {
		return nil
	}

	// 未到期：不应该提前执行，直接跳过（asynq 会按 ProcessIn 调度，正常不会出现）
	if info.ExpireAt.After(time.Now()) {
		return nil
	}

	if err := svc.Transition(ctx, orderID, "expire", "system", 0, "支付超时自动关单"); err != nil {
		logger.L().Error("order:close transition", zap.Int64("order_id", orderID), zap.Error(err))
		return err
	}

	// 释放 Redis + DB 库存锁
	svc.ReleaseStock(ctx, orderID, info.OrderNo)

	// 退还已扣余额（如有）
	if err := svc.RefundUserBalance(ctx, orderID); err != nil {
		logger.L().Error("order:close refund balance failed", zap.Int64("order_id", orderID), zap.Error(err))
	}

	logger.L().Info("order:close done", zap.Int64("order_id", orderID))
	return nil
}
