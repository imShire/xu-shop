package coupon

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/marketing/shared"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// Service 优惠券业务服务。
type Service struct {
	repo Repo
	db   *gorm.DB
}

// NewService 构造服务。
func NewService(repo Repo, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// ===== 状态机 =====
//
// 状态流转表（详见 docs/arch/16-membership.md §1.4）：
//   领取(无)        → unused           insert
//   下单锁定        unused → locked    update
//   下单成功支付    locked → used      update
//   订单取消        locked → unused    update
//   订单退款(全)    used   → unused    update（如未过期）
//   过期扫描        unused → expired   update（每日）
//   过期扫描        locked → expired   update（保险，正常不应出现）
type ucTrans struct {
	From, To string
	Trigger  string
}

var ucTransitions = []ucTrans{
	{UCStatusUnused, UCStatusLocked, "lock"},
	{UCStatusLocked, UCStatusUsed, "consume"},
	{UCStatusLocked, UCStatusUnused, "release"},
	{UCStatusUsed, UCStatusUnused, "refund_restore"},
	{UCStatusUnused, UCStatusExpired, "expire"},
	{UCStatusLocked, UCStatusExpired, "expire"},
}

func findUCTransition(from, trigger string) (string, bool) {
	for _, t := range ucTransitions {
		if t.From == from && t.Trigger == trigger {
			return t.To, true
		}
	}
	return "", false
}

// Transition 用户券状态变更入口（强制走此函数，禁止散点 UPDATE）。
//
// orderID/usedAt 等业务字段通过 fields 透传，函数仅负责状态转换合法性。
func (s *Service) Transition(ctx context.Context, tx *gorm.DB, ucID int64, trigger string, fields map[string]any) error {
	uc, err := s.repo.FindUserCouponForUpdate(ctx, tx, ucID)
	if err != nil {
		return fmt.Errorf("coupon transition: load: %w", err)
	}
	to, ok := findUCTransition(uc.Status, trigger)
	if !ok {
		return shared.ErrInvalidStateTransition.WithMsg(fmt.Sprintf("券状态 %s 不支持触发 %s", uc.Status, trigger))
	}
	rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, ucID, uc.Status, to, fields)
	if err != nil {
		return fmt.Errorf("coupon transition: update: %w", err)
	}
	if rows == 0 {
		return shared.ErrInvalidStateTransition.WithMsg("券状态已被并发修改")
	}
	return nil
}

// ===== 领券 =====

// Claim 领取一张券（来自活动列表 / 推送 / 兑换码 / 后台批量）。
//
// 幂等：通过 (user_id, template_id, source) 计数 + per_user_limit 兜底。
// 并发：模板行 SELECT FOR UPDATE，确保 claimed_count 不超量。
func (s *Service) Claim(ctx context.Context, userID, templateID int64, source string, sourceRef map[string]any) (*UserCoupon, error) {
	if userID <= 0 || templateID <= 0 {
		return nil, errs.ErrParam
	}
	now := time.Now()

	var created *UserCoupon
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tpl, err := s.repo.FindTemplateForUpdate(ctx, tx, templateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return shared.ErrCouponNotFound
			}
			return err
		}
		if tpl.Status != TplStatusOnline {
			return shared.ErrCouponTemplateOffline
		}
		// 领取窗口
		if tpl.ClaimEndAt != nil && now.After(*tpl.ClaimEndAt) {
			return shared.ErrCouponTemplateOffline.WithMsg("活动领取已结束")
		}
		if tpl.ClaimStartAt != nil && now.Before(*tpl.ClaimStartAt) {
			return shared.ErrCouponTemplateOffline.WithMsg("活动还未开始")
		}
		// 名额（0 表示不限）
		if tpl.TotalQuota > 0 && tpl.ClaimedCount >= tpl.TotalQuota {
			return shared.ErrCouponQuotaExhausted
		}
		// 单用户限领
		if tpl.PerUserLimit > 0 {
			cnt, err := s.repo.CountUserClaim(ctx, userID, templateID)
			if err != nil {
				return err
			}
			if cnt >= int64(tpl.PerUserLimit) {
				return shared.ErrCouponClaimLimit
			}
		}

		// 计算到期时间
		expireAt := computeExpireAt(tpl, now)

		// 快照
		snapshot := JSONMap{
			"name":               tpl.Name,
			"type":               tpl.Type,
			"value_cents":        tpl.ValueCents,
			"discount_rate":      tpl.DiscountRate,
			"max_discount_cents": tpl.MaxDiscountCents,
			"min_amount_cents":   tpl.MinAmountCents,
			"scope_type":         tpl.ScopeType,
			"scope_targets":      tpl.ScopeTargets,
			"include_freight":    tpl.IncludeFreight,
			"per_order_limit":    tpl.PerOrderLimit,
			"stack_with_points":  tpl.StackWithPoints,
		}

		uc := &UserCoupon{
			ID:               snowflake.NextID(),
			UserID:           userID,
			CouponTemplateID: templateID,
			Source:           source,
			SourceRef:        JSONMap(sourceRef),
			Status:           UCStatusUnused,
			ClaimedAt:        now,
			ExpireAt:         expireAt,
			Snapshot:         snapshot,
		}
		if err := s.repo.CreateUserCoupon(ctx, tx, uc); err != nil {
			return fmt.Errorf("create user coupon: %w", err)
		}
		if err := s.repo.IncrTemplateClaimed(ctx, tx, templateID, 1); err != nil {
			return fmt.Errorf("incr claimed: %w", err)
		}
		created = uc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ClaimByCode 兑换码领券（事务内：标记码已用 + Claim）。
func (s *Service) ClaimByCode(ctx context.Context, userID int64, code string) (*UserCoupon, error) {
	if userID <= 0 || code == "" {
		return nil, errs.ErrParam
	}
	rc, err := s.repo.FindRedeemCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrRedeemCodeInvalid
		}
		return nil, err
	}
	if rc.Status != RCStatusUnused {
		return nil, shared.ErrRedeemCodeUsed
	}
	if rc.ExpireAt != nil && time.Now().After(*rc.ExpireAt) {
		return nil, shared.ErrRedeemCodeInvalid.WithMsg("兑换码已过期")
	}

	var created *UserCoupon
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := s.repo.UseRedeemCode(ctx, tx, rc.ID, userID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return shared.ErrRedeemCodeUsed
		}
		uc, err := s.claimInTx(ctx, tx, userID, rc.TemplateID, "redeem_code", map[string]any{"code": code})
		if err != nil {
			return err
		}
		created = uc
		return nil
	})
	return created, err
}

// claimInTx 内部用：在已有事务里发券（不再开事务）。
func (s *Service) claimInTx(ctx context.Context, tx *gorm.DB, userID, templateID int64, source string, sourceRef map[string]any) (*UserCoupon, error) {
	now := time.Now()
	tpl, err := s.repo.FindTemplateForUpdate(ctx, tx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrCouponNotFound
		}
		return nil, err
	}
	if tpl.Status != TplStatusOnline {
		return nil, shared.ErrCouponTemplateOffline
	}
	expireAt := computeExpireAt(tpl, now)

	uc := &UserCoupon{
		ID:               snowflake.NextID(),
		UserID:           userID,
		CouponTemplateID: templateID,
		Source:           source,
		SourceRef:        JSONMap(sourceRef),
		Status:           UCStatusUnused,
		ClaimedAt:        now,
		ExpireAt:         expireAt,
		Snapshot: JSONMap{
			"name":               tpl.Name,
			"type":               tpl.Type,
			"value_cents":        tpl.ValueCents,
			"discount_rate":      tpl.DiscountRate,
			"max_discount_cents": tpl.MaxDiscountCents,
			"min_amount_cents":   tpl.MinAmountCents,
			"scope_type":         tpl.ScopeType,
			"scope_targets":      tpl.ScopeTargets,
			"include_freight":    tpl.IncludeFreight,
			"per_order_limit":    tpl.PerOrderLimit,
			"stack_with_points":  tpl.StackWithPoints,
		},
	}
	if err := s.repo.CreateUserCoupon(ctx, tx, uc); err != nil {
		return nil, fmt.Errorf("create user coupon: %w", err)
	}
	if err := s.repo.IncrTemplateClaimed(ctx, tx, templateID, 1); err != nil {
		return nil, err
	}
	return uc, nil
}

// computeExpireAt 计算用户券到期时间。
func computeExpireAt(tpl *CouponTemplate, now time.Time) time.Time {
	switch tpl.ValidityMode {
	case ValidityRelative:
		days := 30
		if tpl.ValidDays != nil && *tpl.ValidDays > 0 {
			days = *tpl.ValidDays
		}
		return now.AddDate(0, 0, days)
	default: // absolute
		if tpl.ValidTo != nil {
			return *tpl.ValidTo
		}
		return now.AddDate(0, 1, 0) // 默认 1 个月兜底
	}
}

// ===== Quote / Lock / Consume / Release / RefundRestore =====

// Quote 预算抵扣金额（不修改 DB）。
func (s *Service) Quote(ctx context.Context, req QuoteReq) (int64, error) {
	if req.UserCouponID == 0 {
		return 0, nil
	}
	uc, err := s.repo.FindUserCoupon(ctx, req.UserCouponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, shared.ErrCouponNotFound
		}
		return 0, err
	}
	if uc.UserID != req.UserID {
		return 0, errs.ErrForbidden
	}
	if uc.Status != UCStatusUnused {
		if uc.Status == UCStatusLocked {
			return 0, shared.ErrCouponLocked
		}
		return 0, shared.ErrCouponNotEligible.WithMsg(fmt.Sprintf("券状态:%s", uc.Status))
	}
	now := time.Now()
	if now.After(uc.ExpireAt) {
		return 0, shared.ErrCouponExpired
	}
	return calcDeduct(uc.Snapshot, req)
}

// calcDeduct 根据快照与订单数据计算实际抵扣（分）。
func calcDeduct(snap JSONMap, req QuoteReq) (int64, error) {
	if snap == nil {
		return 0, errs.ErrInternal.WithMsg("券快照丢失")
	}
	minAmt, _ := snap["min_amount_cents"].(float64)
	if int64(minAmt) > 0 && req.OrderAmountCents < int64(minAmt) {
		return 0, shared.ErrCouponNotEligible.WithMsg("不满足最低订单金额")
	}

	// 范围校验（暂只校验 all/category/product/sku，brand 留 TODO）
	scope, _ := snap["scope_type"].(string)
	if scope != "" && scope != ScopeAll {
		targetsAny, _ := snap["scope_targets"].([]any)
		targets := make(map[int64]struct{}, len(targetsAny))
		for _, v := range targetsAny {
			if f, ok := v.(float64); ok {
				targets[int64(f)] = struct{}{}
			}
		}
		ok := false
		switch scope {
		case ScopeCategory:
			for _, id := range req.ItemCategoryIDs {
				if _, hit := targets[id]; hit {
					ok = true
					break
				}
			}
		case ScopeProduct:
			for _, id := range req.ItemProductIDs {
				if _, hit := targets[id]; hit {
					ok = true
					break
				}
			}
		case ScopeSKU:
			for _, id := range req.ItemSKUIDs {
				if _, hit := targets[id]; hit {
					ok = true
					break
				}
			}
		case ScopeBrand:
			for _, id := range req.ItemBrandIDs {
				if _, hit := targets[id]; hit {
					ok = true
					break
				}
			}
		}
		if !ok {
			return 0, shared.ErrCouponNotEligible.WithMsg("适用范围不匹配")
		}
	}

	typ, _ := snap["type"].(string)
	switch typ {
	case TypeAmount, TypeNoThreshold:
		val, _ := snap["value_cents"].(float64)
		deduct := int64(val)
		if deduct > req.OrderAmountCents {
			deduct = req.OrderAmountCents
		}
		return deduct, nil
	case TypeDiscount:
		rate, _ := snap["discount_rate"].(float64)
		if rate <= 0 || rate >= 1 {
			return 0, errs.ErrInternal.WithMsg("折扣券费率无效")
		}
		deduct := int64(math.Round(float64(req.OrderAmountCents) * (1 - rate)))
		maxDeduct, _ := snap["max_discount_cents"].(float64)
		if int64(maxDeduct) > 0 && deduct > int64(maxDeduct) {
			deduct = int64(maxDeduct)
		}
		return deduct, nil
	case TypeExchange:
		return 0, nil // 兑换券走单独流程
	default:
		return 0, errs.ErrInternal.WithMsg("未知券类型")
	}
}

// Lock 在订单创建事务内调用，锁定一张券。
func (s *Service) Lock(ctx context.Context, tx *gorm.DB, orderID int64, userCouponID int64, userID int64, orderAmountCents int64) (int64, error) {
	if userCouponID == 0 {
		return 0, nil
	}
	now := time.Now()
	uc, err := s.repo.FindUserCouponForUpdate(ctx, tx, userCouponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, shared.ErrCouponNotFound
		}
		return 0, err
	}
	if uc.UserID != userID {
		return 0, errs.ErrForbidden
	}
	if uc.Status != UCStatusUnused {
		if uc.Status == UCStatusLocked {
			return 0, shared.ErrCouponLocked
		}
		return 0, shared.ErrCouponNotEligible.WithMsg(fmt.Sprintf("券状态:%s", uc.Status))
	}
	if now.After(uc.ExpireAt) {
		return 0, shared.ErrCouponExpired
	}
	// 简化：order 已通过 Quote 校验过适用范围/最低额，这里仅做基础校验
	deduct, err := calcDeduct(uc.Snapshot, QuoteReq{
		UserID:           userID,
		UserCouponID:     userCouponID,
		OrderAmountCents: orderAmountCents,
	})
	if err != nil {
		return 0, err
	}

	rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, userCouponID, UCStatusUnused, UCStatusLocked, map[string]any{
		"order_id":   orderID,
		"locked_at":  now,
	})
	if err != nil {
		return 0, fmt.Errorf("lock coupon: %w", err)
	}
	if rows == 0 {
		return 0, shared.ErrCouponLocked
	}
	return deduct, nil
}

// Consume 订单支付成功，券 locked → used。
func (s *Service) Consume(ctx context.Context, tx *gorm.DB, orderID int64) error {
	uc, err := s.repo.FindUserCouponByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if uc == nil || uc.Status != UCStatusLocked {
		return nil // 没有券或不在 locked，无需处理
	}
	now := time.Now()
	rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, uc.ID, UCStatusLocked, UCStatusUsed, map[string]any{
		"used_at": now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return shared.ErrInvalidStateTransition.WithMsg("券状态已变化")
	}
	return s.repo.IncrTemplateUsed(ctx, tx, uc.CouponTemplateID, 1)
}

// Release 订单取消，券 locked → unused。
func (s *Service) Release(ctx context.Context, tx *gorm.DB, orderID int64) error {
	uc, err := s.repo.FindUserCouponByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if uc == nil || uc.Status != UCStatusLocked {
		return nil
	}
	rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, uc.ID, UCStatusLocked, UCStatusUnused, map[string]any{
		"order_id":  nil,
		"locked_at": nil,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return shared.ErrInvalidStateTransition.WithMsg("券状态已变化")
	}
	return s.repo.IncrTemplateClaimed(ctx, tx, uc.CouponTemplateID, -1)
}

// RefundRestore 全额退款时把已用券恢复（如未过期），部分退款不恢复。
func (s *Service) RefundRestore(ctx context.Context, tx *gorm.DB, orderID int64, isFullRefund bool) error {
	if !isFullRefund {
		return nil
	}
	uc, err := s.repo.FindUserCouponByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if uc == nil || uc.Status != UCStatusUsed {
		return nil
	}
	if time.Now().After(uc.ExpireAt) {
		return nil // 已过期，不恢复
	}
	rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, uc.ID, UCStatusUsed, UCStatusUnused, map[string]any{
		"order_id":  nil,
		"used_at":   nil,
		"locked_at": nil,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return shared.ErrInvalidStateTransition.WithMsg("券状态已变化")
	}
	return s.repo.IncrTemplateUsed(ctx, tx, uc.CouponTemplateID, -1)
}

// ===== 我的券 / 活动列表 =====

// MyList 我的券列表。
func (s *Service) MyList(ctx context.Context, userID int64, status string, page, size int) ([]UserCoupon, int64, error) {
	return s.repo.ListMyCoupons(ctx, userID, status, page, size)
}

// PublicList 公开活动列表（C 端可领）。
func (s *Service) PublicList(ctx context.Context, page, size int) ([]CouponTemplate, int64, error) {
	return s.repo.ListOnlineTemplates(ctx, page, size)
}

// ===== 过期扫描 =====

// ExpireScan 扫描已过期但仍为 unused 的券，批量改为 expired。
//
// 由 asynq 每日任务调用。返回处理数量。
func (s *Service) ExpireScan(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	now := time.Now()
	list, err := s.repo.ScanExpire(ctx, now, batchSize)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, uc := range list {
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			rows, err := s.repo.UpdateUserCouponStatus(ctx, tx, uc.ID, UCStatusUnused, UCStatusExpired, nil)
			if err != nil {
				return err
			}
			if rows > 0 {
				processed++
			}
			return nil
		})
		if err != nil {
			return processed, err
		}
	}
	return processed, nil
}
