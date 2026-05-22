// Package coupon 实现优惠券模板/用户券/兑换码/发放任务。
package coupon

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ===== 状态常量 =====

const (
	// 模板
	TplStatusDraft   = "draft"
	TplStatusOnline  = "online"
	TplStatusOffline = "offline"

	// 用户券
	UCStatusUnused  = "unused"
	UCStatusLocked  = "locked"
	UCStatusUsed    = "used"
	UCStatusExpired = "expired"

	// 兑换码
	RCStatusUnused = "unused"
	RCStatusUsed   = "used"
	RCStatusVoided = "voided"

	// 发放任务
	GTStatusPending = "pending"
	GTStatusRunning = "running"
	GTStatusDone    = "done"
	GTStatusFailed  = "failed"
)

// CouponType 优惠券类型。
const (
	TypeAmount       = "amount"        // 满减
	TypeDiscount     = "discount"      // 折扣
	TypeNoThreshold  = "no_threshold"  // 无门槛
	TypeExchange     = "exchange"      // 兑换
)

// ScopeType 适用范围。
const (
	ScopeAll      = "all"
	ScopeCategory = "category"
	ScopeProduct  = "product"
	ScopeSKU      = "sku"
	ScopeBrand    = "brand"
)

// ValidityMode 有效期模式。
const (
	ValidityAbsolute = "absolute" // 固定时段
	ValidityRelative = "relative" // 领取后 N 天
)

// ===== JSONB 辅助 =====

// JSONMap jsonb map。
type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONMap: unsupported type %T", value)
	}
	return json.Unmarshal(b, j)
}

// JSONIntArray jsonb 数组。
type JSONIntArray []int64

func (a JSONIntArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *JSONIntArray) Scan(value any) error {
	if value == nil {
		*a = make(JSONIntArray, 0)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONIntArray: unsupported type %T", value)
	}
	return json.Unmarshal(b, a)
}

// ===== GORM 模型 =====

// CouponTemplate 优惠券模板。
type CouponTemplate struct {
	ID                    int64        `gorm:"column:id;primaryKey"`
	Name                  string       `gorm:"column:name"`
	Description           string       `gorm:"column:description"`
	Type                  string       `gorm:"column:type"`
	ValueCents            int64        `gorm:"column:value_cents"`
	DiscountRate          *float64     `gorm:"column:discount_rate"`
	MaxDiscountCents      int64        `gorm:"column:max_discount_cents"`
	MinAmountCents        int64        `gorm:"column:min_amount_cents"`
	ScopeType             string       `gorm:"column:scope_type"`
	ScopeTargets          JSONIntArray `gorm:"column:scope_targets;type:jsonb"`
	ExcludePromotionItems bool         `gorm:"column:exclude_promotion_items"`
	IncludeFreight        bool         `gorm:"column:include_freight"`
	ValidityMode          string       `gorm:"column:validity_mode"`
	ValidFrom             *time.Time   `gorm:"column:valid_from"`
	ValidTo               *time.Time   `gorm:"column:valid_to"`
	ValidDays             *int         `gorm:"column:valid_days"`
	TotalQuota            int64        `gorm:"column:total_quota"`
	ClaimedCount          int64        `gorm:"column:claimed_count"`
	UsedCount             int64        `gorm:"column:used_count"`
	PerUserLimit          int          `gorm:"column:per_user_limit"`
	PerOrderLimit         int          `gorm:"column:per_order_limit"`
	StackWithPoints       bool         `gorm:"column:stack_with_points"`
	ClaimStartAt          *time.Time   `gorm:"column:claim_start_at"`
	ClaimEndAt            *time.Time   `gorm:"column:claim_end_at"`
	Status                string       `gorm:"column:status"`
	CreatedBy             int64        `gorm:"column:created_by"`
	CreatedAt             time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt             *time.Time   `gorm:"column:deleted_at"`
}

func (CouponTemplate) TableName() string { return "coupon_template" }

// UserCoupon 用户券实例。
type UserCoupon struct {
	ID               int64      `gorm:"column:id;primaryKey"`
	UserID           int64      `gorm:"column:user_id"`
	CouponTemplateID int64      `gorm:"column:coupon_template_id"`
	Source           string     `gorm:"column:source"`
	SourceRef        JSONMap    `gorm:"column:source_ref;type:jsonb"`
	Status           string     `gorm:"column:status"`
	OrderID          *int64     `gorm:"column:order_id"`
	ClaimedAt        time.Time  `gorm:"column:claimed_at;autoCreateTime"`
	LockedAt         *time.Time `gorm:"column:locked_at"`
	UsedAt           *time.Time `gorm:"column:used_at"`
	ExpireAt         time.Time  `gorm:"column:expire_at"`
	Snapshot         JSONMap    `gorm:"column:snapshot;type:jsonb"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserCoupon) TableName() string { return "user_coupon" }

// RedeemCode 兑换码。
type RedeemCode struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	TemplateID    int64      `gorm:"column:template_id"`
	BatchID       int64      `gorm:"column:batch_id"`
	Code          string     `gorm:"column:code"`
	Status        string     `gorm:"column:status"`
	UsedByUserID  *int64     `gorm:"column:used_by_user_id"`
	UsedAt        *time.Time `gorm:"column:used_at"`
	ExpireAt      *time.Time `gorm:"column:expire_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (RedeemCode) TableName() string { return "coupon_redeem_code" }

// GrantTask 发放任务。
type GrantTask struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	TemplateID      int64      `gorm:"column:template_id"`
	Filter          JSONMap    `gorm:"column:filter;type:jsonb"`
	EstimateCount   int64      `gorm:"column:estimate_count"`
	GrantedCount    int64      `gorm:"column:granted_count"`
	FailedCount     int64      `gorm:"column:failed_count"`
	FailedDetailOSS *string    `gorm:"column:failed_detail_oss"`
	Status          string     `gorm:"column:status"`
	CreatedBy       int64      `gorm:"column:created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
	StartedAt       *time.Time `gorm:"column:started_at"`
	FinishedAt      *time.Time `gorm:"column:finished_at"`
}

func (GrantTask) TableName() string { return "coupon_grant_task" }

// ===== DTO =====

// MyCouponItem 我的券列表 item。
type MyCouponItem struct {
	ID            types.Int64Str `json:"id"`
	TemplateID    types.Int64Str `json:"template_id"`
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	ValueCents    int64          `json:"value_cents"`
	DiscountRate  *float64       `json:"discount_rate,omitempty"`
	MinAmountCents int64         `json:"min_amount_cents"`
	Status        string         `json:"status"`
	ExpireAt      time.Time      `json:"expire_at"`
}

// ClaimReq C 端领券请求。
type ClaimReq struct {
	TemplateID types.Int64Str `json:"template_id" binding:"required"`
}

// QuoteReq 订单创建前预算 抵扣金额。
type QuoteReq struct {
	UserID            int64
	UserCouponID      int64
	OrderAmountCents  int64
	ItemCategoryIDs   []int64
	ItemBrandIDs      []int64
	ItemSKUIDs        []int64
	ItemProductIDs    []int64
	HasPromotionItems bool
}
