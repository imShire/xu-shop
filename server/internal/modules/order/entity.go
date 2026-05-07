// Package order 实现订单相关逻辑。
package order

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ---- 状态枚举 ----

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
	StatusRefunding = "refunding"
	StatusRefunded  = "refunded"
)

// ---- 自定义 JSONB 类型 ----

// AddressSnapshot 地址快照（下单时固化，不随地址变更）。
type AddressSnapshot struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Province     string `json:"province"`
	ProvinceCode string `json:"province_code"`
	City         string `json:"city"`
	CityCode     string `json:"city_code"`
	District     string `json:"district"`
	DistrictCode string `json:"district_code"`
	Street       string `json:"street,omitempty"`
	StreetCode   string `json:"street_code,omitempty"`
	Detail       string `json:"detail"`
}

func (a AddressSnapshot) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *AddressSnapshot) Scan(value any) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("AddressSnapshot: unsupported type %T", value)
	}
	return json.Unmarshal(b, a)
}

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

// Order 订单主表。
type Order struct {
	ID                    int64           `gorm:"primaryKey"`
	OrderNo               string          `gorm:"size:32;not null;uniqueIndex"`
	ShopID                int64           `gorm:"not null;default:1"`
	UserID                int64           `gorm:"not null"`
	Status                string          `gorm:"size:16;not null"`
	GoodsCents            int64           `gorm:"not null"`
	FreightCents          int64           `gorm:"not null;default:0"`
	DiscountCents         int64           `gorm:"not null;default:0"`
	CouponDiscountCents   int64           `gorm:"not null;default:0"`
	TotalCents            int64           `gorm:"not null"`
	PayCents              int64           `gorm:"not null"`
	BalancePayCents       int64           `gorm:"column:balance_pay_cents;not null;default:0"`
	AddressSnapshot       AddressSnapshot `gorm:"type:jsonb;not null"`
	BuyerRemark           *string         `gorm:"size:200"`
	CancelRequestPending  bool            `gorm:"not null;default:false"`
	CancelRequestReason   *string         `gorm:"size:200"`
	CancelRequestAt       *time.Time
	CancelReason          *string `gorm:"size:200"`
	DistributorID         *int64
	DistributionPath      RawJSON `gorm:"type:jsonb"`
	GroupBuyOrderID       *int64
	CouponID              *int64
	FromShareUserID       *int64
	FromChannelCodeID     *int64
	IdempotencyKey        *string `gorm:"size:64"`
	CurrentPrepayID       *string `gorm:"size:64"`
	CurrentPrepayExpireAt *time.Time
	ExpireAt              time.Time `gorm:"not null"`
	PaidAt                *time.Time
	ShippedAt             *time.Time
	CompletedAt           *time.Time
	CancelledAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (Order) TableName() string { return `"order"` }

// OrderItem 订单商品行。
type OrderItem struct {
	ID              int64   `gorm:"primaryKey"`
	OrderID         int64   `gorm:"not null"`
	ProductID       int64   `gorm:"not null"`
	SkuID           int64   `gorm:"not null"`
	ProductSnapshot RawJSON `gorm:"type:jsonb;not null"`
	PriceCents      int64   `gorm:"not null"`
	Qty             int     `gorm:"not null"`
	WeightG         int     `gorm:"not null"`
	CreatedAt       time.Time
}

func (OrderItem) TableName() string { return "order_item" }

// OrderLog 订单状态变更日志。
type OrderLog struct {
	ID           int64     `gorm:"primaryKey"          json:"id"`
	OrderID      int64     `gorm:"not null"            json:"order_id"`
	FromStatus   *string   `gorm:"size:16"             json:"from_status,omitempty"`
	ToStatus     *string   `gorm:"size:16"             json:"to_status,omitempty"`
	Reason       *string   `gorm:"size:200"            json:"reason,omitempty"`
	OperatorType *string   `gorm:"size:8"              json:"operator_type,omitempty"`
	OperatorID   *int64    `json:"operator_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (OrderLog) TableName() string { return "order_log" }

// OrderRemark 管理员备注。
type OrderRemark struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	OrderID   int64     `gorm:"not null"   json:"order_id"`
	AdminID   int64     `gorm:"not null"   json:"admin_id"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (OrderRemark) TableName() string { return "order_remark" }

// FreightTemplate 运费模板。
type FreightTemplate struct {
	ID                   int64   `gorm:"primaryKey"`
	Name                 string  `gorm:"size:64;not null"`
	FreeThresholdCents   int64   `gorm:"not null;default:9900"`
	DefaultFeeCents      int64   `gorm:"not null;default:1000"`
	RemoteThresholdCents int64   `gorm:"not null;default:19900"`
	RemoteFeeCents       int64   `gorm:"not null;default:2000"`
	RemoteProvinces      RawJSON `gorm:"type:jsonb;not null;default:'[]'"`
	IsDefault            bool    `gorm:"not null;default:false"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (FreightTemplate) TableName() string { return "freight_template" }
