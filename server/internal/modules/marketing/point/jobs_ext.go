package point

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// NotificationDispatcher 由 jobs 注入，发积分到期提醒。
type NotificationDispatcher interface {
	Dispatch(ctx context.Context, eventType string, userID int64, refID string, params map[string]any) error
}

// ExpireWarning 扫描即将过期（now < expire_at < now+withinDays）的 earn 流水，逐条派发提醒。
//
// MVP 简化：按 transaction 单条触发；同用户当日去重交由 notification 模块。
func (s *Service) ExpireWarning(ctx context.Context, dispatcher NotificationDispatcher, withinDays int, limit int) (int, error) {
	if withinDays <= 0 {
		withinDays = 7
	}
	if limit <= 0 {
		limit = 1000
	}
	now := time.Now()
	upper := now.Add(time.Duration(withinDays) * 24 * time.Hour)

	var list []Transaction
	if err := s.db.WithContext(ctx).
		Where("type = ? AND consumed = false AND expire_at IS NOT NULL AND expire_at > ? AND expire_at <= ?",
			TxnTypeEarn, now, upper).
		Order("expire_at ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return 0, err
	}
	if dispatcher == nil {
		return len(list), nil
	}
	sent := 0
	for _, t := range list {
		expStr := ""
		if t.ExpireAt != nil {
			expStr = t.ExpireAt.Format("2006-01-02")
		}
		params := map[string]any{
			"txn_id":    t.ID,
			"amount":    t.Change,
			"expire_at": expStr,
		}
		refID := fmt.Sprintf("point_warn:%d:%s", t.ID, expStr)
		if err := dispatcher.Dispatch(ctx, "point_expire_warning", t.UserID, refID, params); err != nil {
			continue
		}
		sent++
	}
	return sent, nil
}

// OrderEarnSource 由 jobs 注入：返回已完成但未发积分的订单（兜底）。
type OrderEarnSource interface {
	// ListPendingEarnOrders 返回订单 (id, userID, payCents)；已通过左关联 point_transaction(ref_type=order, ref_id=order_id) 过滤掉已入账的。
	ListPendingEarnOrders(ctx context.Context, lastOrderID int64, limit int) ([]PendingEarnOrder, error)
}

// PendingEarnOrder 等待入账的订单视图。
type PendingEarnOrder struct {
	OrderID    int64
	UserID     int64
	PayCents   int64
	CompletedAt time.Time
}

// EarnGrantFromOrders 每日兜底入账：扫已完成但未发积分的订单。
//
// MVP 计算规则：1 分 = 1 积分（PayCents），不应用等级倍率（倍率在订单完成事件入账时已应用）。
// 本 cron 仅作"漏单兜底"，正常情况 0 条。
//
// 幂等：以 idem_key = fmt.Sprintf("earn_order:%d", orderID) 保证；重复请求 Earn 内部返回首次结果。
//
// TODO(backend-dev): 当 order 完成事件已经实时触发 point.Earn 后，此 cron 只是兜底。如果想严格落地等级倍率，
// 应改为 OrderEarnSource 返回 levelCode 并在此处调 member.MultiplierFor 计算 final change。
// 该计算建议放在 marketing.Service 层而非 cron handler。
func (s *Service) EarnGrantFromOrders(ctx context.Context, src OrderEarnSource, limit int) (int, error) {
	if src == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 500
	}
	granted := 0
	var lastID int64
	for {
		orders, err := src.ListPendingEarnOrders(ctx, lastID, limit)
		if err != nil {
			return granted, err
		}
		if len(orders) == 0 {
			return granted, nil
		}
		for _, o := range orders {
			lastID = o.OrderID
			if o.UserID <= 0 || o.PayCents <= 0 {
				continue
			}
			refType := "order"
			refID := o.OrderID
			idem := fmt.Sprintf("earn_order:%d", o.OrderID)
			// 默认过期 365 天
			exp := time.Now().AddDate(1, 0, 0)
			if _, err := s.Earn(ctx, EarnReq{
				UserID:   o.UserID,
				Change:   o.PayCents,
				RefType:  refType,
				RefID:    &refID,
				Reason:   "订单完成兜底入账",
				IdemKey:  idem,
				ExpireAt: &exp,
			}); err != nil {
				continue
			}
			granted++
		}
		if len(orders) < limit {
			return granted, nil
		}
	}
}

// markdown placeholder
var _ = gorm.ErrRecordNotFound
