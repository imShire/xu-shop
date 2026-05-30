package member

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// UserSource 由 jobs 注入：迭代所有 active 用户 ID。
type UserSource interface {
	ListAllUserIDs(ctx context.Context, lastID int64, batchSize int) ([]int64, error)
}

// LevelChangeNotifier 等级变更时触发（可为 nil）。
type LevelChangeNotifier interface {
	OnLevelChanged(ctx context.Context, userID int64, oldCode, newCode string)
}

// RecomputeAll 每日 03:00：遍历所有 active 用户重算等级。
//
// 简化实现：分批拉用户，逐个 Recompute；产生等级变更时回调 notifier。
// notifier == nil 时不发通知。
//
// TODO(backend-dev): 性能优化方向：用 SQL 聚合一次取 (userID, gmv) → 按等级阈值做 CASE WHEN 批量 update。
// 该优化应在 member.Service.RecomputeAll 内部完成，避免散点写入。
func (s *Service) RecomputeAll(ctx context.Context, src UserSource, notifier LevelChangeNotifier, batchSize int) (int, error) {
	if src == nil {
		return 0, fmt.Errorf("member.RecomputeAll: nil user source")
	}
	if batchSize <= 0 {
		batchSize = 500
	}
	var lastID int64
	processed := 0
	for {
		ids, err := src.ListAllUserIDs(ctx, lastID, batchSize)
		if err != nil {
			return processed, err
		}
		if len(ids) == 0 {
			return processed, nil
		}
		for _, uid := range ids {
			lastID = uid
			oldGmv, oldCode, _, gerr := s.repo.GetUserGMVAndLevel(ctx, uid)
			_ = oldGmv
			if gerr != nil {
				continue
			}
			newCode, err := s.Recompute(ctx, uid)
			if err != nil {
				continue
			}
			processed++
			if notifier != nil && newCode != "" && newCode != oldCode {
				notifier.OnLevelChanged(ctx, uid, oldCode, newCode)
			}
		}
		if len(ids) < batchSize {
			return processed, nil
		}
	}
}

var _ = gorm.ErrRecordNotFound
