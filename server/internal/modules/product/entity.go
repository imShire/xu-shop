// Package product 实现商品、分类、SKU 相关逻辑。
package product

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JSON 自定义 JSONB 类型，兼容 GORM + PostgreSQL。
type JSON json.RawMessage

// Value 实现 driver.Valuer。
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

// Scan 实现 sql.Scanner。
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = JSON(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}

// MarshalJSON 透传底层 JSON。
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON 解析 JSON。
func (j *JSON) UnmarshalJSON(data []byte) error {
	*j = append((*j)[0:0], data...)
	return nil
}

// Category 商品分类。
type Category struct {
	ID        int64          `gorm:"primaryKey"`
	ParentID  int64          `gorm:"not null;default:0"`
	Name      string         `gorm:"size:64;not null"`
	Icon      string         `gorm:"size:512"`
	Sort      int            `gorm:"not null;default:0"`
	Status    string         `gorm:"size:16;not null;default:'enabled'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Category) TableName() string { return "category" }

// Product 商品主表。
type Product struct {
	ID                 int64          `gorm:"primaryKey"`
	CategoryID         int64          `gorm:"not null"`
	Title              string         `gorm:"size:60;not null"`
	Subtitle           string         `gorm:"size:120"`
	MainImage          string         `gorm:"size:512;not null"`
	Images             JSON           `gorm:"type:jsonb;not null;default:'[]'"`
	VideoURL           string         `gorm:"size:512"`
	DetailHTML         string         `gorm:"type:text"`
	DetailNodes        JSON           `gorm:"type:jsonb"`
	Status             string         `gorm:"size:16;not null;default:'draft'"`
	Sales              int            `gorm:"not null;default:0"`
	Sort               int            `gorm:"not null;default:0"`
	Tags               JSON           `gorm:"type:jsonb;not null;default:'[]'"`
	PriceMinCents      int64          `gorm:"not null;default:0"`
	PriceMaxCents      int64          `gorm:"not null;default:0"`
	Unit               string         `gorm:"column:unit;size:16;not null;default:'件'"`
	IsVirtual          bool           `gorm:"column:is_virtual;not null;default:false"`
	FreightTemplateID  *int64         `gorm:"column:freight_template_id"`
	VirtualSales       int            `gorm:"column:virtual_sales;not null;default:0"`
	OnSaleAt           *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

func (Product) TableName() string { return "product" }

// ProductSpec 规格名（如：颜色、尺寸）。
type ProductSpec struct {
	ID        int64  `gorm:"primaryKey"`
	ProductID int64  `gorm:"not null"`
	Name      string `gorm:"size:32;not null"`
	Sort      int    `gorm:"not null;default:0"`
}

func (ProductSpec) TableName() string { return "product_spec" }

// ProductSpecValue 规格值（如：红色、XL）。
type ProductSpecValue struct {
	ID     int64  `gorm:"primaryKey"`
	SpecID int64  `gorm:"not null"`
	Value  string `gorm:"size:32;not null"`
	Sort   int    `gorm:"not null;default:0"`
}

func (ProductSpecValue) TableName() string { return "product_spec_value" }

// SKU 库存单元。
type SKU struct {
	ID                  int64   `gorm:"primaryKey"`
	ProductID           int64   `gorm:"not null"`
	Attrs               JSON    `gorm:"type:jsonb;not null;default:'{}'"`
	PriceCents          int64   `gorm:"not null"`
	OriginalPriceCents  *int64
	Stock               int     `gorm:"not null;default:0"`
	LockedStock         int     `gorm:"not null;default:0"`
	WeightG             int     `gorm:"not null;default:0"`
	SkuCode             *string `gorm:"size:64"`
	Barcode             *string `gorm:"column:barcode;size:64"`
	Image               string  `gorm:"size:512"`
	Status              string  `gorm:"size:16;not null;default:'active'"`
	LowStockThreshold   int     `gorm:"not null;default:0"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (SKU) TableName() string { return "sku" }

// UserViewHistory 用户浏览历史。
type UserViewHistory struct {
	UserID    int64     `gorm:"primaryKey"`
	ProductID int64     `gorm:"primaryKey"`
	ViewedAt  time.Time `gorm:"not null;default:now()"`
}

func (UserViewHistory) TableName() string { return "user_view_history" }

// UserFavorite 用户收藏。
type UserFavorite struct {
	UserID    int64     `gorm:"primaryKey"`
	ProductID int64     `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (UserFavorite) TableName() string { return "user_favorite" }
