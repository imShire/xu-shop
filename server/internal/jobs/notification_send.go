package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskNotificationSend 通知发送任务名。
const TaskNotificationSend = "notification:send"

// NotificationSendPayload 通知发送任务 payload。
type NotificationSendPayload struct {
	TaskID int64 `json:"task_id"`
}

// NotificationSender 供 job 调用的通知发送接口。
type NotificationSender interface {
	HandleSend(ctx context.Context, taskID int64) error
}

// NewNotificationSendHandler 构造通知发送 Handler（3次重试，间隔10s/1m/5m）。
func NewNotificationSendHandler(svc NotificationSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p NotificationSendPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("notification:send unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}

		if err := svc.HandleSend(ctx, p.TaskID); err != nil {
			logger.L().Warn("notification:send failed",
				zap.Int64("task_id", p.TaskID),
				zap.Error(err))
			return err
		}
		return nil
	}
}
