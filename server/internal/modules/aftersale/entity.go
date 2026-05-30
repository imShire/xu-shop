// Package aftersale 售后单（v1.4 完整链路）。
//
// 数据模型见 docs/arch/91-db-schema.md v1.10：aftersale_order + aftersale_negotiation。
// 业务模型 / 状态机 / 钩子矩阵见 docs/arch/09-aftersale.md v1.4。
package aftersale

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ---- 类型 / 状态枚举 ----

const (
	TypeRefundOnly   = "refund_only"
	TypeRefundReturn = "refund_return"
	TypeExchange     = "exchange"
)

const (
	StatusApplying       = "applying"
	StatusSellerAgreed   = "seller_agreed"
	StatusBuyerReturned  = "buyer_returned"
	StatusSellerReceived = "seller_received"
	StatusCompleted      = "completed"
	StatusSellerRejected = "seller_rejected"
	StatusCancelled      = "cancelled"
	StatusClosed         = "closed"
)

const (
	RoleBuyer  = "buyer"
	RoleSeller = "seller"
	RoleSystem = "system"
)

// ActionType 状态机动作。
type ActionType string

const (
	ActionApply           ActionType = "apply"
	ActionAgree           ActionType = "agree"
	ActionReject          ActionType = "reject"
	ActionFillExpress     ActionType = "fill_express"
	ActionConfirmReceived ActionType = "confirm_received"
	ActionCancel          ActionType = "cancel"
	ActionClose           ActionType = "close"
	ActionTimeoutPromote  ActionType = "timeout_promote"
	ActionRefundDone      ActionType = "refund_done"
)

// IsTerminal 判断状态是否为终态。
func IsTerminal(status string) bool {
	switch status {
	case StatusCompleted, StatusSellerRejected, StatusCancelled, StatusClosed:
		return true
	}
	return false
}

// 终态占位 auto_close_at（远未来），避免落入扫描索引。
var farFuture = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

// ---- JSON 字段 ----

// JSONStrArray jsonb 字符串数组。
type JSONStrArray []string

// Value 实现 driver.Valuer。
func (a JSONStrArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]string(a))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan 实现 sql.Scanner。
func (a *JSONStrArray) Scan(value any) error {
	if value == nil {
		*a = JSONStrArray{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("JSONStrArray: unsupported type %T", value)
	}
	if len(raw) == 0 {
		*a = JSONStrArray{}
		return nil
	}
	return json.Unmarshal(raw, (*[]string)(a))
}

// BuyerExpress 寄回运单 jsonb。
type BuyerExpress struct {
	CarrierCode string     `json:"carrier_code"`
	WaybillNo   string     `json:"waybill_no"`
	ShippedAt   *time.Time `json:"shipped_at,omitempty"`
}

// Value 实现 driver.Valuer。
func (e *BuyerExpress) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan 实现 sql.Scanner。
func (e *BuyerExpress) Scan(value any) error {
	if value == nil {
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("BuyerExpress: unsupported type %T", value)
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, e)
}

// ---- GORM 模型 ----

// AftersaleOrder 售后单主表。
type AftersaleOrder struct {
	ID                int64         `gorm:"primaryKey"`
	AftersaleNo       string        `gorm:"size:32;not null;uniqueIndex"`
	OrderID           int64         `gorm:"not null;index"`
	OrderItemID       *int64        `gorm:"column:order_item_id"`
	UserID            int64         `gorm:"not null"`
	Type              string        `gorm:"size:16;not null"`
	Reason            string        `gorm:"size:200;not null"`
	RefundAmountCents int64         `gorm:"not null;default:0"`
	Status            string        `gorm:"size:24;not null;default:'applying'"`
	BuyerEvidence     JSONStrArray  `gorm:"type:jsonb;not null;default:'[]'"`
	BuyerExpress      *BuyerExpress `gorm:"type:jsonb"`
	SellerRemark      string        `gorm:"size:500;not null;default:''"`
	RefundID          *int64
	AppliedAt         time.Time `gorm:"not null;default:now()"`
	AgreedAt          *time.Time
	ReturnedAt        *time.Time
	ReceivedAt        *time.Time
	CompletedAt       *time.Time
	ClosedAt          *time.Time
	AutoCloseAt       time.Time `gorm:"not null"`
	OperatorAdminID   *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName 表名。
func (AftersaleOrder) TableName() string { return "aftersale_order" }

// AftersaleNegotiation 售后协商记录。
type AftersaleNegotiation struct {
	ID          int64        `gorm:"primaryKey"`
	AftersaleID int64        `gorm:"not null;index"`
	Role        string       `gorm:"size:8;not null"`
	AdminID     *int64       `gorm:"column:admin_id"`
	Content     string       `gorm:"size:1000;not null;default:''"`
	Evidence    JSONStrArray `gorm:"type:jsonb;not null;default:'[]'"`
	CreatedAt   time.Time
}

// TableName 表名。
func (AftersaleNegotiation) TableName() string { return "aftersale_negotiation" }
