package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPaymentAutoRefund 自动退款任务名（金额不一致等场景触发）。
const TaskPaymentAutoRefund = "payment:auto_refund"

// AutoRefundPayload 自动退款任务 payload。
type AutoRefundPayload struct {
	PaymentID  int64  `json:"payment_id"`
	OrderID    int64  `json:"order_id"`
	OrderNo    string `json:"order_no"`
	Amount     int64  `json:"amount"`
	TotalCents int64  `json:"total_cents"`
	Reason     string `json:"reason"`
}

// AutoRefunder 自动退款接口（由 payment.Service 实现）。
type AutoRefunder interface {
	ApplyRefund(ctx context.Context, orderID, adminID int64, amtCents int64, reason string) error
}

// NewPaymentAutoRefundHandler 构造自动退款 Handler。
func NewPaymentAutoRefundHandler(refunder AutoRefunder) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p AutoRefundPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("payment:auto_refund unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}

		logger.L().Info("payment:auto_refund start",
			zap.Int64("order_id", p.OrderID),
			zap.Int64("amount", p.Amount),
			zap.String("reason", p.Reason))

		if err := refunder.ApplyRefund(ctx, p.OrderID, 0, p.Amount, p.Reason); err != nil {
			logger.L().Error("payment:auto_refund failed",
				zap.Int64("order_id", p.OrderID), zap.Error(err))
			return err
		}

		logger.L().Info("payment:auto_refund done", zap.Int64("order_id", p.OrderID))
		return nil
	}
}
