// Package decorate 实现首页/页面装修配置。
package decorate

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// PageModule 页面组件配置（JSON 结构）。
// 支持 type: product_list / category_entry / rich_text
type PageModule struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Modules 页面组件列表（JSONB 字段）。
type Modules []PageModule

// Value 实现 driver.Valuer。
func (m Modules) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	return string(b), err
}

// Scan 实现 sql.Scanner。
func (m *Modules) Scan(value any) error {
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("Modules: unsupported type %T", value)
	}
	return json.Unmarshal(b, m)
}

// PageConfig 页面装修配置。
type PageConfig struct {
	ID        int64      `gorm:"primaryKey"                       json:"id"`
	PageKey   string     `gorm:"size:32;not null"                 json:"page_key"` // home / category
	Version   int        `gorm:"not null;default:1"               json:"version"`
	Modules   Modules    `gorm:"type:jsonb;not null;default:'[]'" json:"modules"`
	IsActive  bool       `gorm:"not null;default:false"           json:"is_active"`
	CreatedBy *int64     `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName 指定数据库表名。
func (PageConfig) TableName() string { return "page_config" }
