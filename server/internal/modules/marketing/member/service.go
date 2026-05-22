// Package member 实现会员等级与重算逻辑。
package member

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Level 等级配置。
type Level struct {
	Code               string    `gorm:"column:code;primaryKey"`
	Name               string    `gorm:"column:name"`
	GMVThresholdCents  int64     `gorm:"column:gmv_threshold_cents"`
	PointMultiplier    float64   `gorm:"column:point_multiplier"`
	BirthdayCouponTplID *int64   `gorm:"column:birthday_coupon_tpl_id"`
	Enabled            bool      `gorm:"column:enabled"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Level) TableName() string { return "member_level" }

// Repo 会员仓储。
type Repo interface {
	ListLevels(ctx context.Context) ([]Level, error)
	FindLevel(ctx context.Context, code string) (*Level, error)
	GetUserGMVAndLevel(ctx context.Context, userID int64) (gmv int64, levelCode string, expireAt *time.Time, err error)
	UpdateUserLevel(ctx context.Context, userID int64, levelCode string, expireAt time.Time) error
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造仓储。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) ListLevels(ctx context.Context) ([]Level, error) {
	var list []Level
	err := r.db.WithContext(ctx).Where("enabled = true").Order("gmv_threshold_cents ASC").Find(&list).Error
	return list, err
}

func (r *repoImpl) FindLevel(ctx context.Context, code string) (*Level, error) {
	var lv Level
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&lv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lv, nil
}

func (r *repoImpl) GetUserGMVAndLevel(ctx context.Context, userID int64) (int64, string, *time.Time, error) {
	var row struct {
		Gmv      int64      `gorm:"column:recent_365d_gmv_cents"`
		Level    string     `gorm:"column:member_level_code"`
		ExpireAt *time.Time `gorm:"column:member_level_expire_at"`
	}
	err := r.db.WithContext(ctx).Table("user").
		Select("recent_365d_gmv_cents, member_level_code, member_level_expire_at").
		Where("id = ?", userID).Take(&row).Error
	if err != nil {
		return 0, "", nil, err
	}
	return row.Gmv, row.Level, row.ExpireAt, nil
}

func (r *repoImpl) UpdateUserLevel(ctx context.Context, userID int64, levelCode string, expireAt time.Time) error {
	return r.db.WithContext(ctx).Table("user").Where("id = ?", userID).Updates(map[string]any{
		"member_level_code":      levelCode,
		"member_level_expire_at": expireAt,
	}).Error
}

// Service 会员等级服务。
type Service struct {
	repo Repo
}

// NewService 构造。
func NewService(repo Repo) *Service { return &Service{repo: repo} }

// Recompute 重算单用户等级。
//
// 规则（详见 docs/arch/16-membership.md §3）：
//   1. 取近 365 日累计 GMV（订单完成 - 退款），匹配 GMV 阈值最高的 enabled 等级
//   2. 升级：立即生效，expire_at = now + 365d
//   3. 同级：刷新 expire_at = now + 365d
//   4. 降级保护：当前等级未到期前不降级；只有 expire_at <= now 才能降到目标等级
func (s *Service) Recompute(ctx context.Context, userID int64) (string, error) {
	gmv, currentCode, currentExpire, err := s.repo.GetUserGMVAndLevel(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("load user: %w", err)
	}
	levels, err := s.repo.ListLevels(ctx)
	if err != nil {
		return "", err
	}
	// 找到匹配的最高等级（gmv >= threshold）
	var matched *Level
	for i := range levels {
		if gmv >= levels[i].GMVThresholdCents {
			matched = &levels[i]
		}
	}
	if matched == nil {
		// 没有任何等级 → normal
		matched = &Level{Code: "normal"}
	}

	now := time.Now()
	newExpire := now.AddDate(1, 0, 0) // 365d

	currentLevel, _ := s.repo.FindLevel(ctx, currentCode)
	currentThreshold := int64(0)
	if currentLevel != nil {
		currentThreshold = currentLevel.GMVThresholdCents
	}

	// 降级保护
	if matched.GMVThresholdCents < currentThreshold {
		if currentExpire != nil && now.Before(*currentExpire) {
			// 不降级，保留当前等级和 expire_at
			return currentCode, nil
		}
	}

	if matched.Code == currentCode {
		// 仅刷新 expire
		if err := s.repo.UpdateUserLevel(ctx, userID, currentCode, newExpire); err != nil {
			return "", err
		}
		return currentCode, nil
	}

	if err := s.repo.UpdateUserLevel(ctx, userID, matched.Code, newExpire); err != nil {
		return "", err
	}
	return matched.Code, nil
}

// MultiplierFor 取指定等级的积分倍率（找不到则 1.0）。
func (s *Service) MultiplierFor(ctx context.Context, levelCode string) float64 {
	if levelCode == "" {
		return 1.0
	}
	lv, err := s.repo.FindLevel(ctx, levelCode)
	if err != nil || lv == nil {
		return 1.0
	}
	return lv.PointMultiplier
}

// ListLevels 返回全部启用等级。
func (s *Service) ListLevels(ctx context.Context) ([]Level, error) {
	return s.repo.ListLevels(ctx)
}
