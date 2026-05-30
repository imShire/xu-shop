package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskRecallExecute 召回单用户执行（由 cron-scan / event-handler 分发）。
const TaskRecallExecute = "recall:execute"

// RecallExecutePayload payload。
type RecallExecutePayload struct {
	CampaignID int64 `json:"campaign_id"`
	UserID     int64 `json:"user_id"`
}

// RecallExecutor 由 jobs 调用。
type RecallExecutor interface {
	ExecuteForUser(ctx context.Context, campaignID, userID int64) error
}

// NewRecallExecuteHandler 构造召回执行 Handler。
func NewRecallExecuteHandler(svc RecallExecutor) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p RecallExecutePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error(TaskRecallExecute+" unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		if p.CampaignID <= 0 || p.UserID <= 0 {
			return nil
		}
		if err := svc.ExecuteForUser(ctx, p.CampaignID, p.UserID); err != nil {
			logger.L().Warn(TaskRecallExecute+" failed",
				zap.Int64("campaign_id", p.CampaignID),
				zap.Int64("user_id", p.UserID), zap.Error(err))
			return err
		}
		return nil
	}
}
