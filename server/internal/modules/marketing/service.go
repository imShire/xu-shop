// Package marketing 是营销聚合模块（coupon + point + member）。
//
// 对外提供：
//   - Service：实现 shared.HookService，供 order 注入
//   - RegisterRoutes：挂载 c 端 / 后台 子路由
package marketing

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	"github.com/xushop/xu-shop/internal/modules/marketing/member"
	"github.com/xushop/xu-shop/internal/modules/marketing/point"
	"github.com/xushop/xu-shop/internal/modules/marketing/shared"
)

// Service 营销聚合服务，实现 shared.HookService。
type Service struct {
	Coupon *coupon.Service
	Point  *point.Service
	Member *member.Service
	db     *gorm.DB
}

// NewService 构造。
func NewService(db *gorm.DB) *Service {
	return &Service{
		Coupon: coupon.NewService(coupon.NewRepo(db), db),
		Point:  point.NewService(point.NewRepo(db), db),
		Member: member.NewService(member.NewRepo(db)),
		db:     db,
	}
}

var _ shared.HookService = (*Service)(nil)

// Lock 实现 HookService.Lock。
func (s *Service) Lock(ctx context.Context, tx *gorm.DB, req shared.LockReq) (*shared.LockResp, error) {
	resp := &shared.LockResp{}
	if req.UserCouponID != nil {
		deduct, err := s.Coupon.Lock(ctx, tx, req.OrderID, *req.UserCouponID, req.UserID, req.OrderAmountCents)
		if err != nil {
			return nil, err
		}
		resp.CouponDeductCents = deduct
	}
	if req.PointUsed > 0 {
		// 积分抵扣：1 积分 = 1 分（详见 docs/arch/16）
		// 同事务内执行 Spend；幂等键以 order_id 保证
		idem := fmt.Sprintf("order_lock_point:%d", req.OrderID)
		// 调用 service 内部 Spend，但需要 tx；简化为：不在 marketing.Spend 暴露 tx，
		// 改在此处直接执行扣减并写流水（保证事务一致）
		if err := s.spendInTx(ctx, tx, req.UserID, req.PointUsed, "order_pay", req.OrderID, idem); err != nil {
			return nil, err
		}
		resp.PointDeductCents = req.PointUsed
	}
	return resp, nil
}

// spendInTx 在指定 tx 内消耗积分（避免嵌套事务）。
func (s *Service) spendInTx(ctx context.Context, tx *gorm.DB, userID, change int64, refType string, refID int64, idem string) error {
	// 简化版：直接 update 余额 + 写流水，不做 FIFO（FIFO 在 ExpireScan 与 Spend 单独入口走）
	// 检查幂等
	var existing point.Transaction
	if err := tx.WithContext(ctx).Where("idem_key = ?", idem).First(&existing).Error; err == nil {
		return nil
	}
	var acc point.Account
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&acc).Error; err != nil {
		return shared.ErrPointAccountNotFound
	}
	if acc.Balance < change {
		return shared.ErrPointInsufficient
	}
	idemPtr := idem
	refTypePtr := refType
	refIDPtr := refID
	t := point.Transaction{
		ID:           snowID(),
		UserID:       userID,
		Change:       -change,
		Type:         point.TxnTypeSpend,
		RefType:      &refTypePtr,
		RefID:        &refIDPtr,
		BalanceAfter: acc.Balance - change,
		Reason:       "订单支付抵扣",
		IdemKey:      &idemPtr,
	}
	if err := tx.WithContext(ctx).Create(&t).Error; err != nil {
		return err
	}
	return tx.WithContext(ctx).Model(&point.Account{}).Where("user_id = ?", userID).
		Updates(map[string]any{
			"balance":     gorm.Expr("balance - ?", change),
			"total_spent": gorm.Expr("total_spent + ?", change),
		}).Error
}

// Consume 实现 HookService.Consume。
func (s *Service) Consume(ctx context.Context, tx *gorm.DB, req shared.ConsumeReq) error {
	if err := s.Coupon.Consume(ctx, tx, req.OrderID); err != nil {
		return err
	}
	// 入账积分异步处理（订单完成 + T+7 由 asynq 延迟任务处理），此处仅触发券更新。
	return nil
}

// Release 实现 HookService.Release。
func (s *Service) Release(ctx context.Context, tx *gorm.DB, req shared.ReleaseReq) error {
	if err := s.Coupon.Release(ctx, tx, req.OrderID); err != nil {
		return err
	}
	// 退还已扣积分
	idem := fmt.Sprintf("order_release_point:%d", req.OrderID)
	return s.refundPointInTx(ctx, tx, req.UserID, req.OrderID, idem)
}

// RefundRestore 实现 HookService.RefundRestore。
func (s *Service) RefundRestore(ctx context.Context, tx *gorm.DB, req shared.RefundRestoreReq) error {
	if err := s.Coupon.RefundRestore(ctx, tx, req.OrderID, req.IsFullRefund); err != nil {
		return err
	}
	if req.IsFullRefund {
		idem := fmt.Sprintf("order_refund_full:%d", req.OrderID)
		return s.refundPointInTx(ctx, tx, req.UserID, req.OrderID, idem)
	}
	return nil
}

// refundPointInTx 找到 order 关联的 spend 流水，做反向 refund（幂等）。
func (s *Service) refundPointInTx(ctx context.Context, tx *gorm.DB, userID, orderID int64, idem string) error {
	var existing point.Transaction
	if err := tx.WithContext(ctx).Where("idem_key = ?", idem).First(&existing).Error; err == nil {
		return nil
	}
	// 找到 order 锁定时的 spend 流水
	var spendTxn point.Transaction
	err := tx.WithContext(ctx).
		Where("user_id = ? AND ref_type = ? AND ref_id = ? AND type = ?", userID, "order_pay", orderID, point.TxnTypeSpend).
		First(&spendTxn).Error
	if err != nil {
		return nil // 没有抵扣过，无需返还
	}
	change := -spendTxn.Change // 转正
	if change <= 0 {
		return nil
	}
	var acc point.Account
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&acc).Error; err != nil {
		return err
	}
	idemPtr := idem
	refType := "order_refund"
	ref := orderID
	t := point.Transaction{
		ID:           snowID(),
		UserID:       userID,
		Change:       change,
		Type:         point.TxnTypeRefund,
		RefType:      &refType,
		RefID:        &ref,
		BalanceAfter: acc.Balance + change,
		Reason:       "订单退款返还积分",
		IdemKey:      &idemPtr,
	}
	if err := tx.WithContext(ctx).Create(&t).Error; err != nil {
		return err
	}
	return tx.WithContext(ctx).Model(&point.Account{}).Where("user_id = ?", userID).
		Updates(map[string]any{
			"balance":     gorm.Expr("balance + ?", change),
			"total_spent": gorm.Expr("total_spent - ?", change),
			"updated_at":  time.Now(),
		}).Error
}
