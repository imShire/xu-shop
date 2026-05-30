package coupon

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// NotificationDispatcher 由 jobs 注入，发券到期提醒等通知。
type NotificationDispatcher interface {
	Dispatch(ctx context.Context, eventType string, userID int64, refID string, params map[string]any) error
}

// ExpireWarning 扫描即将过期（now < expire_at < now+withinDays）的 available 券，逐张派发提醒。
//
// MVP 简化：按 user_coupon 单条触发；同用户当日去重由 notification 模块 cooldown 控制。
// dispatcher 为 nil 时仅扫描计数，不发通知。
func (s *Service) ExpireWarning(ctx context.Context, dispatcher NotificationDispatcher, withinDays int, limit int) (int, error) {
	if withinDays <= 0 {
		withinDays = 3
	}
	if limit <= 0 {
		limit = 1000
	}
	now := time.Now()
	upper := now.Add(time.Duration(withinDays) * 24 * time.Hour)

	var list []UserCoupon
	if err := s.db.WithContext(ctx).
		Where("status = ? AND expire_at > ? AND expire_at <= ?", UCStatusUnused, now, upper).
		Order("expire_at ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return 0, err
	}

	if dispatcher == nil {
		return len(list), nil
	}
	sent := 0
	for _, uc := range list {
		params := map[string]any{
			"coupon_id":   uc.ID,
			"template_id": uc.CouponTemplateID,
			"expire_at":   uc.ExpireAt.Format("2006-01-02 15:04"),
		}
		refID := fmt.Sprintf("coupon_warn:%d:%s", uc.ID, uc.ExpireAt.Format("20060102"))
		if err := dispatcher.Dispatch(ctx, "coupon_expire_warning", uc.UserID, refID, params); err != nil {
			continue
		}
		sent++
	}
	return sent, nil
}

// BirthdayUserSource 由 jobs 注入，返回今日生日的用户 + 等级配置。
type BirthdayUserSource interface {
	// ListTodayBirthday 返回今日（按月日匹配）birthday 非空的 active 用户列表。
	ListTodayBirthday(ctx context.Context, lastUserID int64, limit int) ([]BirthdayUser, error)
}

// BirthdayUser 生日券派发所需的最小用户信息。
type BirthdayUser struct {
	UserID    int64
	LevelCode string
}

// BirthdayTemplateResolver 根据等级 code 查 birthday_coupon_tpl_id。
type BirthdayTemplateResolver interface {
	BirthdayCouponTplFor(ctx context.Context, levelCode string) (int64, error)
}

// BirthdayGrant 每日 09:00 扫今日生日用户并按等级派发生日券。
//
// 幂等：通过 (user_id, template_id, source=birthday, source_ref.year) Claim 内部按 per_user_limit 保护，
// 加上 Redis SETNX 兜底防重复发券（key 包含年份）。
//
// 任一依赖为 nil 直接 no-op 跳过（开发环境）。
func (s *Service) BirthdayGrant(ctx context.Context, src BirthdayUserSource, resolver BirthdayTemplateResolver, rdbSetNX func(ctx context.Context, key string, ttl time.Duration) (bool, error)) (int, error) {
	if src == nil || resolver == nil {
		return 0, nil
	}
	const batch = 500
	var lastID int64
	granted := 0
	year := time.Now().Year()
	for {
		users, err := src.ListTodayBirthday(ctx, lastID, batch)
		if err != nil {
			return granted, err
		}
		if len(users) == 0 {
			return granted, nil
		}
		for _, u := range users {
			lastID = u.UserID
			tplID, err := resolver.BirthdayCouponTplFor(ctx, u.LevelCode)
			if err != nil || tplID <= 0 {
				continue
			}
			// 幂等：Redis 抢号 + Claim 自身限领兜底
			if rdbSetNX != nil {
				key := fmt.Sprintf("coupon:bday:%d:%d:%d", u.UserID, tplID, year)
				ok, _ := rdbSetNX(ctx, key, 48*time.Hour)
				if !ok {
					continue
				}
			}
			if _, err := s.Claim(ctx, u.UserID, tplID, "birthday", map[string]any{"year": year}); err != nil {
				continue
			}
			granted++
		}
		if len(users) < batch {
			return granted, nil
		}
	}
}

// markdownTouch 占位（仅用于让编译器/未使用 import 静默；保留以备扩展）。
var _ = gorm.ErrRecordNotFound
