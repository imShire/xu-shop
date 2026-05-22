// Package distribution 实现分享溯源 + 一级分销 + 微信商家转账提现。
//
// 对应 docs/arch/18-distribution.md / docs/prd/18-distribution.md。
//
// 文件分布：
//   - entity.go   GORM 模型（8 张表）
//   - dto.go      请求/响应 DTO
//   - repo.go     仓储实现
//   - service.go  业务服务（含状态机 Transition）
//   - handler.go  HTTP handler（C 端 + admin）
//   - router.go   路由注册
package distribution

import (
	"time"
)

// ===== 状态常量 =====

const (
	// 分销员状态
	DistStatusPending  = "pending"
	DistStatusActive   = "active"
	DistStatusDisabled = "disabled"

	DistLevelNormal = "normal"
	DistLevelSenior = "senior"

	// 佣金状态
	CommissionStatusPending  = "pending"  // 冻结期
	CommissionStatusLocked   = "locked"   // 已冻结期满，可提现
	CommissionStatusSettled  = "settled"  // 已结算（出账）
	CommissionStatusCanceled = "canceled" // 作废
	CommissionStatusSuspect  = "suspect"  // 防刷标记

	// 提现工单状态
	WithdrawStatusPending    = "pending"
	WithdrawStatusProcessing = "processing"
	WithdrawStatusSuccess    = "success"
	WithdrawStatusFailed     = "failed"
	WithdrawStatusCanceled   = "canceled"

	// 结算单状态
	SettlementStatusPending    = "pending"
	SettlementStatusProcessing = "processing"
	SettlementStatusSuccess    = "success"
	SettlementStatusFailed     = "failed"

	// 分享 Scene
	ShareSceneProduct        = "product"
	ShareSceneActivity       = "activity"
	ShareSceneBrand          = "brand"
	ShareSceneInviteRegister = "invite_register"
)

// ===== 分享 =====

// ShareLink 分享链接。
type ShareLink struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64     `gorm:"column:user_id" json:"user_id"`
	Scene         string    `gorm:"column:scene" json:"scene"`
	TargetID      *int64    `gorm:"column:target_id" json:"target_id,omitempty"`
	ChannelCode   string    `gorm:"column:channel_code" json:"channel_code"`
	ShortToken    string    `gorm:"column:short_token" json:"short_token"`
	ExpireAt      time.Time `gorm:"column:expire_at" json:"expire_at"`
	ClickCount    int64     `gorm:"column:click_count" json:"click_count"`
	RegisterCount int64     `gorm:"column:register_count" json:"register_count"`
	OrderCount    int64     `gorm:"column:order_count" json:"order_count"`
	GMVCents      int64     `gorm:"column:gmv_cents" json:"gmv_cents"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名。
func (ShareLink) TableName() string { return "share_link" }

// ShareClick 分享点击。
type ShareClick struct {
	ID                 int64     `gorm:"column:id;primaryKey" json:"id"`
	TraceID            string    `gorm:"column:trace_id" json:"trace_id"`
	ShareLinkID        int64     `gorm:"column:share_link_id" json:"share_link_id"`
	VisitorFingerprint *string   `gorm:"column:visitor_fingerprint" json:"visitor_fingerprint,omitempty"`
	TS                 time.Time `gorm:"column:ts" json:"ts"`
	UA                 *string   `gorm:"column:ua" json:"ua,omitempty"`
	IP                 *string   `gorm:"column:ip;type:inet" json:"ip,omitempty"`
	Device             *string   `gorm:"column:device" json:"device,omitempty"`
	Referer            *string   `gorm:"column:referer" json:"referer,omitempty"`
}

// TableName 指定表名。
func (ShareClick) TableName() string { return "share_click" }

// ShareAttribution 分享归因（trace_id 唯一）。
type ShareAttribution struct {
	ID                    int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID                *int64    `gorm:"column:user_id" json:"user_id,omitempty"`
	ShareLinkID           int64     `gorm:"column:share_link_id" json:"share_link_id"`
	TraceID               string    `gorm:"column:trace_id" json:"trace_id"`
	FirstTouchTS          time.Time `gorm:"column:first_touch_ts" json:"first_touch_ts"`
	LastTouchTS           time.Time `gorm:"column:last_touch_ts" json:"last_touch_ts"`
	AttributionWindowDays int       `gorm:"column:attribution_window_days" json:"attribution_window_days"`
}

// TableName 指定表名。
func (ShareAttribution) TableName() string { return "share_attribution" }

// ===== 分销 =====

// Distributor 分销员。
type Distributor struct {
	ID               int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID           int64      `gorm:"column:user_id" json:"user_id"`
	Level            string     `gorm:"column:level" json:"level"`
	Rate             float64    `gorm:"column:rate" json:"rate"`
	RateOverride     *float64   `gorm:"column:rate_override" json:"rate_override,omitempty"`
	Status           string     `gorm:"column:status" json:"status"`
	ApplyAt          time.Time  `gorm:"column:apply_at" json:"apply_at"`
	ApprovedAt       *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	ApproverAdminID  *int64     `gorm:"column:approver_admin_id" json:"approver_admin_id,omitempty"`
	SuspendedAt      *time.Time `gorm:"column:suspended_at" json:"suspended_at,omitempty"`
	SuspendedReason  *string    `gorm:"column:suspended_reason" json:"suspended_reason,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名。
func (Distributor) TableName() string { return "distributor" }

// EffectiveRate 取专属费率优先，否则 base rate。
func (d *Distributor) EffectiveRate() float64 {
	if d.RateOverride != nil && *d.RateOverride > 0 {
		return *d.RateOverride
	}
	return d.Rate
}

// DistributorRelation 分销邀请关系。
type DistributorRelation struct {
	ID             int64     `gorm:"column:id;primaryKey" json:"id"`
	InviteeUserID  int64     `gorm:"column:invitee_user_id" json:"invitee_user_id"`
	InviterUserID  int64     `gorm:"column:inviter_user_id" json:"inviter_user_id"`
	ShareLinkID    int64     `gorm:"column:share_link_id" json:"share_link_id"`
	BoundAt        time.Time `gorm:"column:bound_at" json:"bound_at"`
	ExpireAt       time.Time `gorm:"column:expire_at" json:"expire_at"`
	LastRenewedAt  time.Time `gorm:"column:last_renewed_at" json:"last_renewed_at"`
}

// TableName 指定表名。
func (DistributorRelation) TableName() string { return "distributor_relation" }

// ===== 佣金 =====

// CommissionRecord 佣金记录。
type CommissionRecord struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	OrderID           int64      `gorm:"column:order_id" json:"order_id"`
	DistributorUserID int64      `gorm:"column:distributor_user_id" json:"distributor_user_id"`
	Level             string     `gorm:"column:level" json:"level"`
	Rate              float64    `gorm:"column:rate" json:"rate"`
	BaseAmountCents   int64      `gorm:"column:base_amount_cents" json:"base_amount_cents"`
	AmountCents       int64      `gorm:"column:amount_cents" json:"amount_cents"`
	Status            string     `gorm:"column:status" json:"status"`
	SuspectReason     *string    `gorm:"column:suspect_reason" json:"suspect_reason,omitempty"`
	FreezeUntil       time.Time  `gorm:"column:freeze_until" json:"freeze_until"`
	SettlementID      *int64     `gorm:"column:settlement_id" json:"settlement_id,omitempty"`
	SettledAt         *time.Time `gorm:"column:settled_at" json:"settled_at,omitempty"`
	CanceledReason    *string    `gorm:"column:canceled_reason" json:"canceled_reason,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名。
func (CommissionRecord) TableName() string { return "commission_record" }

// CommissionSettlement 结算单。
type CommissionSettlement struct {
	ID                  int64     `gorm:"column:id;primaryKey" json:"id"`
	DistributorUserID   int64     `gorm:"column:distributor_user_id" json:"distributor_user_id"`
	PeriodYYYYMM        *string   `gorm:"column:period_yyyymm" json:"period_yyyymm,omitempty"`
	RequestAmountCents  int64     `gorm:"column:request_amount_cents" json:"request_amount_cents"`
	Records             string    `gorm:"column:records;type:jsonb" json:"records"`
	WithdrawOrderID     *int64    `gorm:"column:withdraw_order_id" json:"withdraw_order_id,omitempty"`
	Status              string    `gorm:"column:status" json:"status"`
	Channel             string    `gorm:"column:channel" json:"channel"`
	FailReason          *string   `gorm:"column:fail_reason" json:"fail_reason,omitempty"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名。
func (CommissionSettlement) TableName() string { return "commission_settlement" }

// ===== 提现 =====

// WithdrawOrder 提现工单。
type WithdrawOrder struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	DistributorUserID int64      `gorm:"column:distributor_user_id" json:"distributor_user_id"`
	WithdrawNo        string     `gorm:"column:withdraw_no" json:"withdraw_no"`
	AmountCents       int64      `gorm:"column:amount_cents" json:"amount_cents"`
	Channel           string     `gorm:"column:channel" json:"channel"`
	Status            string     `gorm:"column:status" json:"status"`
	WxTransferNo      *string    `gorm:"column:wx_transfer_no" json:"wx_transfer_no,omitempty"`
	WxTransferState   *string    `gorm:"column:wx_transfer_state" json:"wx_transfer_state,omitempty"`
	FailReason        *string    `gorm:"column:fail_reason" json:"fail_reason,omitempty"`
	AppliedAt         time.Time  `gorm:"column:applied_at" json:"applied_at"`
	ProcessedAt       *time.Time `gorm:"column:processed_at" json:"processed_at,omitempty"`
	FinishedAt        *time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
	IdemKey           string     `gorm:"column:idem_key" json:"-"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名。
func (WithdrawOrder) TableName() string { return "withdraw_order" }
