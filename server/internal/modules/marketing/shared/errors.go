package shared

import "github.com/xushop/xu-shop/internal/pkg/errs"

// 业务错误码：60000-69999 段分配给 marketing 模块。
var (
	// 通用
	ErrInvalidStateTransition = errs.New(60001, "状态变更不被允许", 409)

	// 优惠券
	ErrCouponTemplateOffline = errs.New(61001, "活动已下线", 400)
	ErrCouponClaimLimit      = errs.New(61002, "已达领取上限", 409)
	ErrCouponQuotaExhausted  = errs.New(61003, "活动已被领完", 410)
	ErrCouponNotEligible     = errs.New(61004, "本单不满足券使用条件", 400)
	ErrCouponExpired         = errs.New(61005, "优惠券已过期", 410)
	ErrCouponLocked          = errs.New(61006, "优惠券已被占用，请先解除", 409)
	ErrCouponNotFound        = errs.New(61007, "优惠券不存在", 404)
	ErrRedeemCodeInvalid     = errs.New(61101, "兑换码无效", 400)
	ErrRedeemCodeUsed        = errs.New(61102, "兑换码已被使用", 409)

	// 积分
	ErrPointInsufficient    = errs.New(62001, "积分余额不足", 400)
	ErrPointDeductOverLimit = errs.New(62002, "本单可抵扣积分超限", 400)
	ErrPointAdjustPending   = errs.New(62003, "存在待审批的积分调整", 409)
	ErrPointAccountNotFound = errs.New(62004, "积分账户不存在", 404)

	// 等级
	ErrMemberLevelNotFound = errs.New(63001, "会员等级不存在", 404)
)
