package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskRecallEventHandler 召回事件触发（事件触发时入队）。
const TaskRecallEventHandler = "recall:event-handler"

// RecallEventPayload payload。
type RecallEventPayload struct {
	// EventName 事件名（如 order_paid / cart_abandoned / favorite_added / no_order_30d）。
	EventName string `json:"event_name"`
	// TargetUserID >0 时只对单用户执行；=0 时按 audience_filter 全量执行。
	TargetUserID int64 `json:"target_user_id"`
}

// RecallEventDispatcher 由 jobs 调用。
type RecallEventDispatcher interface {
	OnEvent(ctx context.Context, eventName string, targetUserID int64) error
}

// NewRecallEventHandler 构造召回事件 Handler。
func NewRecallEventHandler(svc RecallEventDispatcher) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p RecallEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error(TaskRecallEventHandler+" unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		if p.EventName == "" {
			return nil
		}
		if err := svc.OnEvent(ctx, p.EventName, p.TargetUserID); err != nil {
			logger.L().Warn(TaskRecallEventHandler+" failed",
				zap.String("event", p.EventName), zap.Int64("user_id", p.TargetUserID), zap.Error(err))
			return err
		}
		return nil
	}
}
