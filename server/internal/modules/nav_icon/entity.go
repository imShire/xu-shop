// Package nav_icon 实现金刚区快捷导航管理。
package nav_icon

import (
	"time"

	"github.com/xushop/xu-shop/internal/domain"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// NavIcon 金刚区快捷导航图标。
type NavIcon struct {
	ID        types.Int64Str `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:16;not null;default:''" json:"title"`
	IconURL   string    `gorm:"type:text;not null" json:"icon_url"`
	LinkURL    string             `gorm:"type:text;not null;default:''" json:"link_url"`
	LinkConfig *domain.LinkConfig `gorm:"column:link_config;type:jsonb"  json:"link_config,omitempty"`
	Sort       int         `gorm:"not null;default:0" json:"sort"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (NavIcon) TableName() string { return "nav_icon" }

// ---- 请求 DTO ----

// CreateNavIconReq 创建金刚区图标请求。
type CreateNavIconReq struct {
	Title      string             `json:"title" binding:"max=16"`
	IconURL    string             `json:"icon_url" binding:"required"`
	LinkURL    string             `json:"link_url"`
	LinkConfig *domain.LinkConfig `json:"link_config"`
	Sort       int                `json:"sort"`
	IsActive   *bool              `json:"is_active"`
}

// UpdateNavIconReq 更新金刚区图标请求。
type UpdateNavIconReq struct {
	Title      string             `json:"title" binding:"max=16"`
	IconURL    string             `json:"icon_url" binding:"required"`
	LinkURL    string             `json:"link_url"`
	LinkConfig *domain.LinkConfig `json:"link_config"`
	Sort       int                `json:"sort"`
	IsActive   *bool              `json:"is_active"`
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
