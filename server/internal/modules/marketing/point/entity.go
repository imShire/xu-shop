// Package point 实现积分账户/流水/规则/调整工单。
package point

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ===== 状态/类型常量 =====

const (
	TxnTypeEarn         = "earn"
	TxnTypeSpend        = "spend"
	TxnTypeExpire       = "expire"
	TxnTypeRefund       = "refund"
	TxnTypeAdminAdjust  = "admin_adjust"
	TxnTypeFreeze       = "freeze"
	TxnTypeUnfreeze     = "unfreeze"

	TicketStatusPending  = "pending"
	TicketStatusApproved = "approved"
	TicketStatusRejected = "rejected"
)

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

// ===== 模型 =====

// Account 积分账户。
type Account struct {
	UserID      int64     `gorm:"column:user_id;primaryKey"`
	Balance     int64     `gorm:"column:balance"`
	Locked      int64     `gorm:"column:locked"`
	TotalEarned int64     `gorm:"column:total_earned"`
	TotalSpent  int64     `gorm:"column:total_spent"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Account) TableName() string { return "point_account" }

// Transaction 积分流水（不可变）。
type Transaction struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	UserID       int64      `gorm:"column:user_id"`
	Change       int64      `gorm:"column:change"`
	Type         string     `gorm:"column:type"`
	RefType      *string    `gorm:"column:ref_type"`
	RefID        *int64     `gorm:"column:ref_id"`
	BalanceAfter int64      `gorm:"column:balance_after"`
	ExpireAt     *time.Time `gorm:"column:expire_at"`
	Consumed     bool       `gorm:"column:consumed"`
	Reason       string     `gorm:"column:reason"`
	CreatedBy    *int64     `gorm:"column:created_by"`
	IdemKey      *string    `gorm:"column:idem_key"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (Transaction) TableName() string { return "point_transaction" }

// Rule 积分规则。
type Rule struct {
	Code      string    `gorm:"column:code;primaryKey"`
	Enabled   bool      `gorm:"column:enabled"`
	Config    JSONMap   `gorm:"column:config;type:jsonb"`
	UpdatedBy *int64    `gorm:"column:updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Rule) TableName() string { return "point_rule" }

// AdjustTicket 积分调整工单。
type AdjustTicket struct {
	ID                int64      `gorm:"column:id;primaryKey"`
	UserID            int64      `gorm:"column:user_id"`
	Change            int64      `gorm:"column:change"`
	Reason            string     `gorm:"column:reason"`
	Status            string     `gorm:"column:status"`
	ApplicantAdminID  int64      `gorm:"column:applicant_admin_id"`
	ApproverAdminID   *int64     `gorm:"column:approver_admin_id"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	ApprovedAt        *time.Time `gorm:"column:approved_at"`
}

func (AdjustTicket) TableName() string { return "point_adjust_ticket" }
