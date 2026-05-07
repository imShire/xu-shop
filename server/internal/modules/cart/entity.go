// Package cart 实现购物车相关逻辑。
package cart

import "time"

// CartItem 购物车条目。
type CartItem struct {
	ID                 int64     `gorm:"primaryKey"`
	UserID             int64     `gorm:"not null"`
	SkuID              int64     `gorm:"column:sku_id;not null"`
	Qty                int       `gorm:"not null"`
	SnapshotPriceCents int64     `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (CartItem) TableName() string { return "cart_item" }

// CartItemDetail 购物车条目详情（JOIN sku + product）。
type CartItemDetail struct {
	// cart_item 字段
	ID                 int64
	UserID             int64
	SkuID              int64
	Qty                int
	SnapshotPriceCents int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	// sku 字段
	SkuPriceCents  int64
	SkuStatus      string
	SkuStock       int
	SkuLockedStock int
	SkuImage       string
	SkuAttrs       []byte
	// product 字段
	ProductID            int64
	ProductCategoryID    int64
	ProductTitle         string
	ProductSubtitle      string
	ProductMainImage     string
	ProductTags          []byte
	ProductStatus        string
	ProductSales         int
	ProductPriceMinCents int64
	ProductPriceMaxCents int64
	ProductDeleted       bool
}
