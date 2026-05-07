// Package inventory 实现库存管理、调整日志和低库存预警。
package inventory

import "time"

// InventoryLog 库存变动日志。
type InventoryLog struct {
	ID            int64     `gorm:"primaryKey"  json:"id"`
	SkuID         int64     `gorm:"not null"    json:"sku_id"`
	Change        int       `gorm:"not null"    json:"change"`
	Type          string    `gorm:"size:16;not null" json:"type"`
	RefType       string    `gorm:"size:16"     json:"ref_type"`
	RefID         *int64    `json:"ref_id,omitempty"`
	BalanceBefore int       `gorm:"not null"    json:"balance_before"`
	BalanceAfter  int       `gorm:"not null"    json:"balance_after"`
	LockedBefore  int       `gorm:"not null"    json:"locked_before"`
	LockedAfter   int       `gorm:"not null"    json:"locked_after"`
	OperatorType  string    `gorm:"size:8"      json:"operator_type"`
	OperatorID    *int64    `json:"operator_id,omitempty"`
	Reason        string    `gorm:"size:200"    json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

func (InventoryLog) TableName() string { return "inventory_log" }

// LowStockAlert 低库存预警。
type LowStockAlert struct {
	ID               int64      `gorm:"primaryKey"                      json:"id"`
	SkuID            int64      `gorm:"not null"                        json:"sku_id"`
	ThresholdAtAlert int        `gorm:"not null"                        json:"threshold_at_alert"`
	CurrentStock     int        `gorm:"not null"                        json:"current_stock"`
	Status           string     `gorm:"size:8;not null;default:'unread'" json:"status"`
	ReadBy           *int64     `json:"read_by,omitempty"`
	ReadAt           *time.Time `json:"read_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (LowStockAlert) TableName() string { return "low_stock_alert" }
