// Package tag 实现用户标签字典 / 关系 / 自动重算 / 人群预估。
//
// 与 recall 模块拆分：标签是数据资产，召回是动作引擎；
// 字典 user_tag、关系 user_tag_relation、月度快照 user_tag_snapshot。
package tag

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ===== 状态 / 枚举常量 =====

const (
	// SourceAuto 自动计算（每日 / 增量任务维护）。
	SourceAuto = "auto"
	// SourceManual 后台或运营手动维护。
	SourceManual = "manual"

	// 内置类目（CHECK 约束保持一致）。
	CategoryRFM         = "rfm"
	CategoryLifecycle   = "lifecycle"
	CategoryCategoryPref = "category_pref"
	CategoryPriceBand   = "price_band"
	CategorySource      = "source"
	CategoryBusiness    = "business"
	CategoryMember      = "member"
	CategorySystem      = "system"
)

// JSONMap jsonb 通用 map（与 marketing/notification 包内部 JSONMap 对齐）。
type JSONMap map[string]any

// Value 实现 driver.Valuer。
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

// Scan 实现 sql.Scanner。
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

// JSONStrArray jsonb 字符串数组。
type JSONStrArray []string

func (a JSONStrArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *JSONStrArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONStrArray: unsupported type %T", value)
	}
	return json.Unmarshal(b, a)
}

// ===== GORM 模型 =====

// UserTag 标签字典。
type UserTag struct {
	Code        string    `gorm:"column:code;primaryKey" json:"code"`
	Name        string    `gorm:"column:name" json:"name"`
	Category    string    `gorm:"column:category" json:"category"`
	ParentCode  *string   `gorm:"column:parent_code" json:"parent_code,omitempty"`
	Color       *string   `gorm:"column:color" json:"color,omitempty"`
	Description string    `gorm:"column:description" json:"description"`
	Source      string    `gorm:"column:source" json:"source"`
	Config      JSONMap   `gorm:"column:config;type:jsonb" json:"config"`
	Enabled     bool      `gorm:"column:enabled" json:"enabled"`
	Sort        int       `gorm:"column:sort" json:"sort"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 显式声明（避免 gorm 复数化）。
func (UserTag) TableName() string { return "user_tag" }

// UserTagRelation 用户与标签关系。
type UserTagRelation struct {
	ID        int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID    int64      `gorm:"column:user_id" json:"user_id"`
	TagCode   string     `gorm:"column:tag_code" json:"tag_code"`
	Score     int        `gorm:"column:score" json:"score"`
	Source    string     `gorm:"column:source" json:"source"`
	SourceRef *string    `gorm:"column:source_ref" json:"source_ref,omitempty"`
	ExpireAt  *time.Time `gorm:"column:expire_at" json:"expire_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 显式声明。
func (UserTagRelation) TableName() string { return "user_tag_relation" }

// UserTagSnapshot 月度全量快照。
type UserTagSnapshot struct {
	ID           int64        `gorm:"column:id;primaryKey" json:"id"`
	SnapshotDate time.Time    `gorm:"column:snapshot_date" json:"snapshot_date"`
	UserID       int64        `gorm:"column:user_id" json:"user_id"`
	Tags         JSONStrArray `gorm:"column:tags;type:jsonb" json:"tags"`
	CreatedAt    time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 显式声明。
func (UserTagSnapshot) TableName() string { return "user_tag_snapshot" }
