// Package shared 包含 marketing 模块的跨子领域公共定义。
//
// HookService 是给 order 模块调用的接口，避免 order → marketing 的硬依赖；
// order 在创建/支付/取消/退款时调用对应 hook，marketing 内部分发到 coupon / point 子领域。
package shared

import (
	"context"

	"gorm.io/gorm"
)

// LockReq order 在 CreateOrder 中传入：锁定 1 张优惠券 + 一定数量积分。
type LockReq struct {
	OrderID         int64
	UserID          int64
	OrderAmountCents int64 // 商品金额（去运费/未抵扣前）
	UserCouponID    *int64
	PointUsed       int64
}

// LockResp 返回扣减金额，order 用于计算 final_pay_cents。
type LockResp struct {
	CouponDeductCents int64
	PointDeductCents  int64
}

// ConsumeReq 订单支付成功，把 locked → used；返还预计可获积分（待 T+N 入账）。
type ConsumeReq struct {
	OrderID         int64
	UserID          int64
	PaidAmountCents int64
	IsFirstOrder    bool
}

// ReleaseReq 订单取消时调用：locked → unused（券）/ unfreeze（积分）。
type ReleaseReq struct {
	OrderID int64
	UserID  int64
}

// RefundRestoreReq 订单退款（部分/全部）时调用。
//
// 全额退款：券恢复（如未过期）+ 已入账积分扣回；部分退款：仅扣回比例积分。
type RefundRestoreReq struct {
	OrderID            int64
	UserID             int64
	RefundAmountCents  int64
	OrderAmountCents   int64 // 用于按比例计算
	IsFullRefund       bool
}

// HookService 提供给 order 模块的统一接口。
//
// 所有方法都支持 tx 注入（使用同一事务），调用方在创建订单事务内传入 *gorm.DB。
// 返回错误时调用方应回滚事务。
type HookService interface {
	// Lock 在订单创建事务内调用：校验+锁定券、冻结积分。
	Lock(ctx context.Context, tx *gorm.DB, req LockReq) (*LockResp, error)
	// Consume 订单支付成功后调用：券 used、积分扣减确认、入账积分进入 frozen。
	Consume(ctx context.Context, tx *gorm.DB, req ConsumeReq) error
	// Release 订单取消时调用：券解锁、积分解冻。
	Release(ctx context.Context, tx *gorm.DB, req ReleaseReq) error
	// RefundRestore 订单退款时调用：券恢复（如可）、积分按比例扣回。
	RefundRestore(ctx context.Context, tx *gorm.DB, req RefundRestoreReq) error
}

// NoopHook 空实现，order 在 marketing 未注入时使用，确保不破坏既有流程。
type NoopHook struct{}

func (NoopHook) Lock(_ context.Context, _ *gorm.DB, _ LockReq) (*LockResp, error) {
	return &LockResp{}, nil
}
func (NoopHook) Consume(_ context.Context, _ *gorm.DB, _ ConsumeReq) error { return nil }
func (NoopHook) Release(_ context.Context, _ *gorm.DB, _ ReleaseReq) error { return nil }
func (NoopHook) RefundRestore(_ context.Context, _ *gorm.DB, _ RefundRestoreReq) error {
	return nil
}
