// Package reconciliation 实现日终对账差异的存储、查询与处置。
// 三种对账作业（支付 / 库存 / 佣金）通过 RecordDiff 写入；
// 后台通过 /admin/reconciliation/diff/* 接口查阅、确认与解决。
package reconciliation

import "time"

// 差异 job 类型常量。
const (
	JobPayment    = "payment"
	JobInventory  = "inventory"
	JobCommission = "commission"
)

// 资源类型常量。
const (
	RefTypeOrder            = "order"
	RefTypeSKU              = "sku"
	RefTypeCommissionRecord = "commission_record"
	RefTypeAccount          = "account"
)

// 差异严重程度。
const (
	SeverityInfo     = "info"
	SeverityWarn     = "warn"
	SeverityCritical = "critical"
)

// 差异状态。
const (
	StatusOpen         = "open"
	StatusAcknowledged = "acknowledged"
	StatusResolved     = "resolved"
)

// Diff 对账差异记录，对应表 reconciliation_diff。
type Diff struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	Job           string     `gorm:"column:job;size:32;not null"`
	BizDate       time.Time  `gorm:"column:biz_date;type:date;not null"`
	RefType       string     `gorm:"column:ref_type;size:32;not null"`
	RefID         string     `gorm:"column:ref_id;size:64;not null"`
	Field         string     `gorm:"column:field;size:64;not null"`
	ExpectedValue *string    `gorm:"column:expected_value;type:text"`
	ActualValue   *string    `gorm:"column:actual_value;type:text"`
	DiffCents     *int64     `gorm:"column:diff_cents"`
	Severity      string     `gorm:"column:severity;size:8;not null;default:'warn'"`
	Status        string     `gorm:"column:status;size:16;not null;default:'open'"`
	Note          *string    `gorm:"column:note;type:text"`
	AckedBy       *int64     `gorm:"column:acked_by"`
	AckedAt       *time.Time `gorm:"column:acked_at"`
	ResolvedBy    *int64     `gorm:"column:resolved_by"`
	ResolvedAt    *time.Time `gorm:"column:resolved_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名。
func (Diff) TableName() string { return "reconciliation_diff" }
