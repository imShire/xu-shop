// Package payment 实现支付、退款、对账逻辑。
package payment

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ---- 状态常量 ----

const (
	PayStatusPending = "pending"
	PayStatusSuccess = "success"
	PayStatusFailed  = "failed"
	PayStatusOrphan  = "orphan" // 金额不一致，已入队自动退款

	RefundStatusPending = "pending"
	RefundStatusSuccess = "success"
	RefundStatusFailed  = "failed"

	DiffStatusUnresolved = "unresolved"
	DiffStatusResolved   = "resolved"
)

// ---- 自定义 JSONB 类型 ----

// RawJSON 通用 JSONB 字段。
type RawJSON []byte

func (j RawJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

func (j *RawJSON) Scan(value any) error {
	if value == nil {
		*j = RawJSON("null")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = RawJSON(v)
	default:
		return fmt.Errorf("RawJSON: unsupported type %T", value)
	}
	return nil
}

func (j RawJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *RawJSON) UnmarshalJSON(data []byte) error {
	*j = append((*j)[0:0], data...)
	return nil
}

// ---- GORM 模型 ----

// Payment 支付记录。
type Payment struct {
	ID            int64      `gorm:"primaryKey"                       json:"id"`
	OrderID       int64      `gorm:"not null"                         json:"order_id"`
	Channel       string     `gorm:"size:16;not null;default:wxpay"   json:"channel"`
	TradeType     string     `gorm:"size:16;not null"                 json:"trade_type"`
	AppID         *string    `gorm:"size:64"                          json:"app_id,omitempty"`
	PrepayID      *string    `gorm:"size:64"                          json:"-"`
	TransactionID *string    `gorm:"size:64;uniqueIndex"              json:"transaction_id,omitempty"`
	AmountCents   int64      `gorm:"not null"                         json:"amount_cents"`
	Status        string     `gorm:"size:16;not null;default:pending" json:"status"`
	RawNotify     RawJSON    `gorm:"type:jsonb"                       json:"-"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Payment) TableName() string { return "payment" }

// Refund 退款记录。
type Refund struct {
	ID          int64      `gorm:"primaryKey"                       json:"id"`
	OrderID     int64      `gorm:"not null"                         json:"order_id"`
	PaymentID   int64      `gorm:"not null"                         json:"payment_id"`
	RefundNo    string     `gorm:"size:32;not null;uniqueIndex"     json:"refund_no"`
	AmountCents int64      `gorm:"not null"                         json:"amount_cents"`
	Reason      *string    `gorm:"size:200"                         json:"reason,omitempty"`
	Status      string     `gorm:"size:16;not null;default:pending" json:"status"`
	RawNotify   RawJSON    `gorm:"type:jsonb"                       json:"-"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
	OperatorID  *int64     `json:"operator_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Refund) TableName() string { return "refund" }

// ReconciliationDiff 对账差异记录。
type ReconciliationDiff struct {
	ID             int64      `gorm:"primaryKey"                          json:"id"`
	BillDate       time.Time  `gorm:"type:date;not null"                  json:"bill_date"`
	TransactionID  *string    `gorm:"size:64"                             json:"transaction_id,omitempty"`
	OrderNo        *string    `gorm:"size:32"                             json:"order_no,omitempty"`
	OurAmountCents *int64     `json:"our_amount_cents,omitempty"`
	WxAmountCents  *int64     `json:"wx_amount_cents,omitempty"`
	DiffType       *string    `gorm:"size:32"                             json:"diff_type,omitempty"`
	Status         string     `gorm:"size:16;not null;default:unresolved" json:"status"`
	ResolvedBy     *int64     `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (ReconciliationDiff) TableName() string { return "reconciliation_diff" }

// AuditLog 操作审计日志。
type AuditLog struct {
	ID            int64     `gorm:"primaryKey"        json:"id"`
	AdminID       int64     `gorm:"not null"          json:"admin_id"`
	AdminUsername *string   `gorm:"size:64"           json:"admin_username,omitempty"`
	AdminRealName *string   `gorm:"size:64"           json:"admin_real_name,omitempty"`
	Module        string    `gorm:"size:32;not null"  json:"module"`
	Action        string    `gorm:"size:32;not null"  json:"action"`
	TargetID      *string   `gorm:"size:64"           json:"target_id,omitempty"`
	Diff          RawJSON   `gorm:"type:jsonb"        json:"diff,omitempty"`
	IP            *string   `gorm:"type:inet"         json:"ip,omitempty"`
	UA            *string   `gorm:"type:text"         json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_log" }

// ---- DTO ----

// PayStatusResp C 端查询支付状态响应。
type PayStatusResp struct {
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paid_at"`
	AmtCents  int64      `json:"amount_cents"`
	TradeType string     `json:"trade_type"`
}

// PaymentFilter 后台支付列表过滤条件。
type PaymentFilter struct {
	OrderID int64
	Status  string
	Page    int
	Size    int
}

// RefundFilter 后台退款列表过滤条件。
type RefundFilter struct {
	OrderID int64
	Status  string
	Page    int
	Size    int
}

// DiffFilter 对账差异过滤条件。
type DiffFilter struct {
	Date   string
	Status string
	Page   int
	Size   int
}

// AuditLogFilter 操作审计日志过滤条件。
type AuditLogFilter struct {
	Module    string
	Operator  string
	StartDate string
	EndDate   string
	Page      int
	Size      int
}

// systemSetting 系统设置（JSON 序列化辅助）。
type systemSetting struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// rawNotifyFromBytes 将字节切片转为 RawJSON（兼容 nil）。
func rawNotifyFromBytes(b []byte) RawJSON {
	if len(b) == 0 {
		return RawJSON("null")
	}
	return RawJSON(b)
}

// marshalJSON 将任意值序列化为 RawJSON。
func marshalJSON(v any) RawJSON {
	b, err := json.Marshal(v)
	if err != nil {
		return RawJSON("null")
	}
	return RawJSON(b)
}
