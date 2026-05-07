// Package private_domain 实现私域流量（渠道码、客户标签、分享裂变）。
package private_domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- JSONB 辅助 ----

// JSONStrings jsonb 字符串数组。
type JSONStrings []string

func (j JSONStrings) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONStrings) Scan(value any) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONStrings: unsupported type %T", value)
	}
	return json.Unmarshal(b, j)
}

// JSONInt64s jsonb int64 数组。
type JSONInt64s []int64

func (j JSONInt64s) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONInt64s) Scan(value any) error {
	if value == nil {
		*j = []int64{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONInt64s: unsupported type %T", value)
	}
	return json.Unmarshal(b, j)
}

// ---- GORM 模型 ----

// ChannelCode 渠道码（企微联系我）。
type ChannelCode struct {
	ID              int64       `gorm:"column:id;primaryKey" json:"id"`
	Name            string      `gorm:"column:name" json:"name"`
	QRImageURL      *string     `gorm:"column:qr_image_url" json:"qr_image_url,omitempty"`
	QYWXConfigID    *string     `gorm:"column:qywx_config_id" json:"qywx_config_id,omitempty"`
	CustomerServers JSONStrings `gorm:"column:customer_servers;type:jsonb" json:"customer_servers"`
	TagIDs          JSONInt64s  `gorm:"column:tag_ids;type:jsonb" json:"tag_ids"`
	WelcomeText     *string     `gorm:"column:welcome_text" json:"welcome_text,omitempty"`
	ScanCount       int         `gorm:"column:scan_count" json:"scan_count"`
	AddCount        int         `gorm:"column:add_count" json:"add_count"`
	CreatedAt       time.Time   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ChannelCode) TableName() string { return "channel_code" }

// CustomerTag 客户标签。
type CustomerTag struct {
	ID        int64     `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name;uniqueIndex" json:"name"`
	Source    string    `gorm:"column:source" json:"source"`
	QYWXTagID *string   `gorm:"column:qywx_tag_id" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (CustomerTag) TableName() string { return "customer_tag" }

// UserTag 用户标签关联。
type UserTag struct {
	UserID    int64     `gorm:"column:user_id;primaryKey" json:"user_id"`
	TagID     int64     `gorm:"column:tag_id;primaryKey" json:"tag_id"`
	Source    string    `gorm:"column:source" json:"source"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (UserTag) TableName() string { return "user_tag" }

// ShareAttribution 分享归因。
type ShareAttribution struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	ShareUserID  int64     `gorm:"column:share_user_id" json:"share_user_id"`
	ViewerUserID *int64    `gorm:"column:viewer_user_id" json:"viewer_user_id,omitempty"`
	ProductID    *int64    `gorm:"column:product_id" json:"product_id,omitempty"`
	Channel      string    `gorm:"column:channel" json:"channel"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ShareAttribution) TableName() string { return "share_attribution" }

// ShareShortCode 分享短码。
type ShareShortCode struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	ShareUserID   int64      `gorm:"column:share_user_id" json:"share_user_id"`
	ProductID     *int64     `gorm:"column:product_id" json:"product_id,omitempty"`
	ChannelCodeID *int64     `gorm:"column:channel_code_id" json:"channel_code_id,omitempty"`
	ExpireAt      *time.Time `gorm:"column:expire_at" json:"expire_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ShareShortCode) TableName() string { return "share_short_code" }

// ---- DTO ----

// ChannelCodeResp 渠道码响应 DTO。
type ChannelCodeResp struct {
	ID              types.Int64Str `json:"id"`
	Name            string         `json:"name"`
	QRImageURL      *string        `json:"qr_image_url,omitempty"`
	QYWXConfigID    *string        `json:"qywx_config_id,omitempty"`
	CustomerServers JSONStrings    `json:"customer_servers"`
	TagIDs          JSONInt64s     `json:"tag_ids"`
	WelcomeText     *string        `json:"welcome_text,omitempty"`
	ScanCount       int            `json:"scan_count"`
	AddCount        int            `json:"add_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// TagResp 标签响应 DTO。
type TagResp struct {
	ID        types.Int64Str `json:"id"`
	Name      string         `json:"name"`
	Source    string         `json:"source"`
	CreatedAt time.Time      `json:"created_at"`
}

// toChannelCodeResp entity → ChannelCodeResp。
func toChannelCodeResp(c *ChannelCode) ChannelCodeResp {
	return ChannelCodeResp{
		ID:              types.Int64Str(c.ID),
		Name:            c.Name,
		QRImageURL:      c.QRImageURL,
		QYWXConfigID:    c.QYWXConfigID,
		CustomerServers: c.CustomerServers,
		TagIDs:          c.TagIDs,
		WelcomeText:     c.WelcomeText,
		ScanCount:       c.ScanCount,
		AddCount:        c.AddCount,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

// toTagResp entity → TagResp。
func toTagResp(t *CustomerTag) TagResp {
	return TagResp{
		ID:        types.Int64Str(t.ID),
		Name:      t.Name,
		Source:    t.Source,
		CreatedAt: t.CreatedAt,
	}
}

// ChannelCodeStats 渠道码统计。
type ChannelCodeStats struct {
	ChannelCode
	OrderCount  int   `json:"order_count"`
	AmountCents int64 `json:"amount_cents"`
}

// ChannelCodeStatsResp 渠道码统计响应 DTO。
type ChannelCodeStatsResp struct {
	ChannelCodeResp
	OrderCount  int   `json:"order_count"`
	AmountCents int64 `json:"amount_cents"`
}

// ResolveResp 解析分享场景响应。
type ResolveResp struct {
	ShareUserID   types.Int64Str  `json:"share_user_id"`
	ProductID     *types.Int64Str `json:"product_id,omitempty"`
	ChannelCodeID *types.Int64Str `json:"channel_code_id,omitempty"`
}
