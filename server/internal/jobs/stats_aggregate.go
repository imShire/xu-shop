package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskStatsAggregate 每日 01:00 聚合前一天数据。
const TaskStatsAggregate = "stats:aggregate"

// StatsAggregatePayload 聚合任务 payload。
type StatsAggregatePayload struct {
	// Date 聚合日期（格式 2006-01-02，默认为昨天）。
	Date string `json:"date"`
}

// StatsAggregator 供 job 调用的聚合接口。
type StatsAggregator interface {
	AggregateDaily(ctx context.Context, date time.Time) error
}

// NewStatsAggregateHandler 构造每日统计聚合 Handler。
func NewStatsAggregateHandler(svc StatsAggregator) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p StatsAggregatePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("stats:aggregate unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}

		dateStr := p.Date
		if dateStr == "" {
			dateStr = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logger.L().Error("stats:aggregate parse date", zap.String("date", dateStr), zap.Error(err))
			return asynq.SkipRetry
		}

		logger.L().Info("stats:aggregate start", zap.String("date", dateStr))
		if err := svc.AggregateDaily(ctx, date); err != nil {
			logger.L().Error("stats:aggregate failed", zap.String("date", dateStr), zap.Error(err))
			return err
		}
		logger.L().Info("stats:aggregate done", zap.String("date", dateStr))
		return nil
	}
}
