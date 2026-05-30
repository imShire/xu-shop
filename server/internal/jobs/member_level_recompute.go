package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	mkmember "github.com/xushop/xu-shop/internal/modules/marketing/member"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskMemberLevelRecompute 会员等级日重算（每日 03:00）。
const TaskMemberLevelRecompute = "member:level-recompute"

// MemberRecomputer 由 jobs 调用。
type MemberRecomputer interface {
	RecomputeAll(ctx context.Context, src mkmember.UserSource, notifier mkmember.LevelChangeNotifier, batchSize int) (int, error)
}

// NewMemberLevelRecomputeHandler 构造会员等级重算 Handler。
func NewMemberLevelRecomputeHandler(
	svc MemberRecomputer,
	src mkmember.UserSource,
	notifier mkmember.LevelChangeNotifier,
) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		n, err := svc.RecomputeAll(ctx, src, notifier, 500)
		if err != nil {
			logger.L().Error(TaskMemberLevelRecompute+" failed", zap.Error(err))
			return err
		}
		logger.L().Info(TaskMemberLevelRecompute+" done", zap.Int("processed", n))
		return nil
	}
}
