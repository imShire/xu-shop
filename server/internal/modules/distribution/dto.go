package distribution

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ===== 请求 DTO =====

// CreateShareLinkReq 创建分享链接。
type CreateShareLinkReq struct {
	Scene       string `json:"scene" binding:"required,oneof=product activity brand invite_register"`
	TargetID    *int64 `json:"target_id"`
	ChannelCode string `json:"channel_code"`
	TTLDays     int    `json:"ttl_days"`
}

// TrackShareReq 客户端调用：记录一次点击（短链 302 时也会自动记一次，本接口供 SPA 透传）。
type TrackShareReq struct {
	ShortToken         string  `json:"short_token" binding:"required,max=16"`
	TraceID            string  `json:"trace_id" binding:"required,max=64"`
	VisitorFingerprint *string `json:"visitor_fingerprint" binding:"omitempty,max=64"`
	Device             *string `json:"device" binding:"omitempty,max=8"`
	Referer            *string `json:"referer" binding:"omitempty,max=500"`
}

// ApplyDistributorReq 申请分销员。
type ApplyDistributorReq struct {
	Remark string `json:"remark" binding:"max=200"`
}

// AdjustRateReq 设置专属费率。
type AdjustRateReq struct {
	RateOverride *float64 `json:"rate_override"`
}

// AdjustLevelReq 调整等级。
type AdjustLevelReq struct {
	Level string `json:"level" binding:"required,oneof=normal senior"`
}

// RejectReq 拒绝/停用原因。
type RejectReq struct {
	Reason string `json:"reason" binding:"required,max=200"`
}

// AuditCommissionReq 人工审核佣金（release 解除 suspect / cancel 作废）。
type AuditCommissionReq struct {
	Action string `json:"action" binding:"required,oneof=release cancel"`
	Reason string `json:"reason" binding:"max=500"`
}

// WithdrawReq 申请提现。
type WithdrawReq struct {
	AmountCents int64  `json:"amount_cents" binding:"required,min=1000"`
	SmsCode     string `json:"sms_code" binding:"omitempty,max=8"`
	Password    string `json:"password" binding:"omitempty,max=128"`
}

// ===== 响应 DTO =====

// ShareLinkResp 分享链接响应。
type ShareLinkResp struct {
	ID          types.Int64Str `json:"id"`
	UserID      types.Int64Str `json:"user_id"`
	Scene       string         `json:"scene"`
	TargetID    *types.Int64Str `json:"target_id,omitempty"`
	ChannelCode string         `json:"channel_code"`
	ShortToken  string         `json:"short_token"`
	ShortURL    string         `json:"short_url"`
	ExpireAt    time.Time      `json:"expire_at"`
	ClickCount  int64          `json:"click_count"`
	OrderCount  int64          `json:"order_count"`
	GMVCents    int64          `json:"gmv_cents"`
	CreatedAt   time.Time      `json:"created_at"`
}

// DistributorResp 分销员响应。
type DistributorResp struct {
	ID           types.Int64Str `json:"id"`
	UserID       types.Int64Str `json:"user_id"`
	Level        string         `json:"level"`
	Rate         float64        `json:"rate"`
	RateOverride *float64       `json:"rate_override,omitempty"`
	Status       string         `json:"status"`
	ApplyAt      time.Time      `json:"apply_at"`
	ApprovedAt   *time.Time     `json:"approved_at,omitempty"`
}

// CommissionResp 佣金响应。
type CommissionResp struct {
	ID                types.Int64Str `json:"id"`
	OrderID           types.Int64Str `json:"order_id"`
	DistributorUserID types.Int64Str `json:"distributor_user_id"`
	Level             string         `json:"level"`
	Rate              float64        `json:"rate"`
	BaseAmountCents   int64          `json:"base_amount_cents"`
	AmountCents       int64          `json:"amount_cents"`
	Status            string         `json:"status"`
	SuspectReason     *string        `json:"suspect_reason,omitempty"`
	FreezeUntil       time.Time      `json:"freeze_until"`
	SettledAt         *time.Time     `json:"settled_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}

// WithdrawResp 提现工单响应。
type WithdrawResp struct {
	ID                types.Int64Str `json:"id"`
	DistributorUserID types.Int64Str `json:"distributor_user_id"`
	WithdrawNo        string         `json:"withdraw_no"`
	AmountCents       int64          `json:"amount_cents"`
	Status            string         `json:"status"`
	WxTransferNo      *string        `json:"wx_transfer_no,omitempty"`
	WxTransferState   *string        `json:"wx_transfer_state,omitempty"`
	FailReason        *string        `json:"fail_reason,omitempty"`
	AppliedAt         time.Time      `json:"applied_at"`
	FinishedAt        *time.Time     `json:"finished_at,omitempty"`
}

// SettlementResp 结算单响应。
type SettlementResp struct {
	ID                 types.Int64Str  `json:"id"`
	DistributorUserID  types.Int64Str  `json:"distributor_user_id"`
	PeriodYYYYMM       *string         `json:"period_yyyymm,omitempty"`
	RequestAmountCents int64           `json:"request_amount_cents"`
	WithdrawOrderID    *types.Int64Str `json:"withdraw_order_id,omitempty"`
	Status             string          `json:"status"`
	Channel            string          `json:"channel"`
	FailReason         *string         `json:"fail_reason,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

// FunnelReportResp 分销漏斗。
type FunnelReportResp struct {
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	ShareLinks        int64  `json:"share_links"`
	Clicks            int64  `json:"clicks"`
	Registers         int64  `json:"registers"`
	Orders            int64  `json:"orders"`
	GMVCents          int64  `json:"gmv_cents"`
	CommissionPending int64  `json:"commission_pending_cents"`
	CommissionLocked  int64  `json:"commission_locked_cents"`
	CommissionSettled int64  `json:"commission_settled_cents"`
}

// MyProfileResp 我的分销员资料。
type MyProfileResp struct {
	Distributor       *DistributorResp `json:"distributor"`
	PendingCents      int64            `json:"pending_cents"`
	LockedCents       int64            `json:"locked_cents"`
	SettledCents      int64            `json:"settled_cents"`
	TotalEarnedCents  int64            `json:"total_earned_cents"`
	WithdrawnCents    int64            `json:"withdrawn_cents"`
}

// ===== 转换 =====

func toShareLinkResp(l *ShareLink, baseURL string) ShareLinkResp {
	r := ShareLinkResp{
		ID:          types.Int64Str(l.ID),
		UserID:      types.Int64Str(l.UserID),
		Scene:       l.Scene,
		ChannelCode: l.ChannelCode,
		ShortToken:  l.ShortToken,
		ShortURL:    baseURL + "/s/" + l.ShortToken,
		ExpireAt:    l.ExpireAt,
		ClickCount:  l.ClickCount,
		OrderCount:  l.OrderCount,
		GMVCents:    l.GMVCents,
		CreatedAt:   l.CreatedAt,
	}
	if l.TargetID != nil {
		t := types.Int64Str(*l.TargetID)
		r.TargetID = &t
	}
	return r
}

func toDistributorResp(d *Distributor) DistributorResp {
	return DistributorResp{
		ID:           types.Int64Str(d.ID),
		UserID:       types.Int64Str(d.UserID),
		Level:        d.Level,
		Rate:         d.Rate,
		RateOverride: d.RateOverride,
		Status:       d.Status,
		ApplyAt:      d.ApplyAt,
		ApprovedAt:   d.ApprovedAt,
	}
}

func toCommissionResp(c *CommissionRecord) CommissionResp {
	return CommissionResp{
		ID:                types.Int64Str(c.ID),
		OrderID:           types.Int64Str(c.OrderID),
		DistributorUserID: types.Int64Str(c.DistributorUserID),
		Level:             c.Level,
		Rate:              c.Rate,
		BaseAmountCents:   c.BaseAmountCents,
		AmountCents:       c.AmountCents,
		Status:            c.Status,
		SuspectReason:     c.SuspectReason,
		FreezeUntil:       c.FreezeUntil,
		SettledAt:         c.SettledAt,
		CreatedAt:         c.CreatedAt,
	}
}

func toWithdrawResp(w *WithdrawOrder) WithdrawResp {
	return WithdrawResp{
		ID:                types.Int64Str(w.ID),
		DistributorUserID: types.Int64Str(w.DistributorUserID),
		WithdrawNo:        w.WithdrawNo,
		AmountCents:       w.AmountCents,
		Status:            w.Status,
		WxTransferNo:      w.WxTransferNo,
		WxTransferState:   w.WxTransferState,
		FailReason:        w.FailReason,
		AppliedAt:         w.AppliedAt,
		FinishedAt:        w.FinishedAt,
	}
}

func toSettlementResp(s *CommissionSettlement) SettlementResp {
	r := SettlementResp{
		ID:                 types.Int64Str(s.ID),
		DistributorUserID:  types.Int64Str(s.DistributorUserID),
		PeriodYYYYMM:       s.PeriodYYYYMM,
		RequestAmountCents: s.RequestAmountCents,
		Status:             s.Status,
		Channel:            s.Channel,
		FailReason:         s.FailReason,
		CreatedAt:          s.CreatedAt,
	}
	if s.WithdrawOrderID != nil {
		w := types.Int64Str(*s.WithdrawOrderID)
		r.WithdrawOrderID = &w
	}
	return r
}
