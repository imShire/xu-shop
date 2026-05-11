// Package domain 定义跨模块共享的值对象。
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// LinkConfig 结构化链接配置，用于 Banner/NavIcon 等模块。存储为 JSONB。
type LinkConfig struct {
	Type       string `json:"type"`        // product|category|article|product_list|custom
	TargetID   string `json:"target_id"`   // 资源 ID（可为空）
	TargetName string `json:"target_name"` // 回显名称（可为空）
	URL        string `json:"url"`         // 最终跳转路径（必填）
}

// Value 实现 driver.Valuer。
func (l LinkConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(l)
	return string(b), err
}

// Scan 实现 sql.Scanner。
func (l *LinkConfig) Scan(value any) error {
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
		return fmt.Errorf("LinkConfig: unsupported type %T", value)
	}
	return json.Unmarshal(b, l)
}
