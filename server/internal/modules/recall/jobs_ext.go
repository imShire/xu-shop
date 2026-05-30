package recall

import (
	"context"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// ExecuteForUser 单用户执行召回（jobs 分发使用）。
//
// 由 recall:execute 任务调用：先取活动，校验状态在线，复用 executeForUser 内部节流/动作逻辑。
func (s *Service) ExecuteForUser(ctx context.Context, campaignID, userID int64) error {
	if campaignID <= 0 || userID <= 0 {
		return errs.ErrParam
	}
	c, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return err
	}
	if c.Status != StatusOnline {
		return nil
	}
	return s.executeForUser(ctx, c, userID)
}
