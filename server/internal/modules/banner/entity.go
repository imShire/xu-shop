// Package banner 实现首页横幅管理。
package banner

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// Banner 首页横幅。
type Banner struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:128;not null;default:''" json:"title"`
	ImageURL  string    `gorm:"type:text;not null" json:"image_url"`
	LinkURL   string    `gorm:"type:text;not null;default:''" json:"link_url"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Banner) TableName() string { return "banner" }

// ---- 请求 DTO ----

// CreateBannerReq 创建横幅请求。
type CreateBannerReq struct {
	Title    string `json:"title" binding:"max=128"`
	ImageURL string `json:"image_url" binding:"required"`
	LinkURL  string `json:"link_url"`
	Sort     int    `json:"sort"`
	IsActive *bool  `json:"is_active"`
}

// UpdateBannerReq 更新横幅请求。
type UpdateBannerReq struct {
	Title    string `json:"title" binding:"max=128"`
	ImageURL string `json:"image_url" binding:"required"`
	LinkURL  string `json:"link_url"`
	Sort     int    `json:"sort"`
	IsActive *bool  `json:"is_active"`
}

// SortItem 单个排序项。
type SortItem struct {
	ID   types.Int64Str `json:"id" binding:"required"`
	Sort int            `json:"sort"`
}

// BulkSortReq 批量排序请求。
type BulkSortReq struct {
	Items []SortItem `json:"items" binding:"required,min=1"`
}
