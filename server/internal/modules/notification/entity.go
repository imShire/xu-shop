// Package notification 实现通知模板管理和消息发送。
package notification

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 状态常量 ----

const (
	TaskStatusPending = "pending"
	TaskStatusSent    = "sent"
	TaskStatusFailed  = "failed"
	TaskStatusSkipped = "skipped"

	TargetTypeUser    = "user"
	TargetTypeWebhook = "webhook"
)

// ---- JSONB 辅助 ----

// JSONMap 通用 jsonb map 字段。
type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

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

// ---- GORM 模型 ----

// NotificationTemplate 通知模板。
type NotificationTemplate struct {
	ID                 int64     `gorm:"column:id;primaryKey" json:"id"`
	Code               string    `gorm:"column:code;uniqueIndex" json:"code"`
	Channel            string    `gorm:"column:channel" json:"channel"`
	TemplateIDExternal string    `gorm:"column:template_id_external" json:"template_id_external"`
	Fields             JSONMap   `gorm:"column:fields;type:jsonb" json:"fields"`
	Enabled            bool      `gorm:"column:enabled" json:"enabled"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (NotificationTemplate) TableName() string { return "notification_template" }

// NotificationTask 通知任务。
type NotificationTask struct {
	ID           int64      `gorm:"column:id;primaryKey" json:"id"`
	TemplateCode string     `gorm:"column:template_code" json:"template_code"`
	TargetType   string     `gorm:"column:target_type" json:"target_type"`
	Target       string     `gorm:"column:target" json:"target"`
	Params       JSONMap    `gorm:"column:params;type:jsonb" json:"params"`
	Status       string     `gorm:"column:status" json:"status"`
	RetryCount   int        `gorm:"column:retry_count" json:"retry_count"`
	LastError    *string    `gorm:"column:last_error" json:"last_error,omitempty"`
	DedupKey     *string    `gorm:"column:dedup_key" json:"dedup_key,omitempty"`
	SentAt       *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (NotificationTask) TableName() string { return "notification_task" }

// ---- DTO ----

// NotificationTemplateResp 通知模板响应 DTO。
type NotificationTemplateResp struct {
	ID                 types.Int64Str `json:"id"`
	Code               string         `json:"code"`
	Channel            string         `json:"channel"`
	TemplateIDExternal string         `json:"template_id_external"`
	Fields             JSONMap        `json:"fields"`
	Enabled            bool           `json:"enabled"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// NotificationTaskResp 通知任务响应 DTO。
type NotificationTaskResp struct {
	ID           types.Int64Str `json:"id"`
	TemplateCode string         `json:"template_code"`
	TargetType   string         `json:"target_type"`
	Target       string         `json:"target"`
	Params       JSONMap        `json:"params"`
	Status       string         `json:"status"`
	RetryCount   int            `json:"retry_count"`
	LastError    *string        `json:"last_error,omitempty"`
	DedupKey     *string        `json:"dedup_key,omitempty"`
	SentAt       *time.Time     `json:"sent_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

func toNotificationTemplateResp(t *NotificationTemplate) NotificationTemplateResp {
	return NotificationTemplateResp{
		ID:                 types.Int64Str(t.ID),
		Code:               t.Code,
		Channel:            t.Channel,
		TemplateIDExternal: t.TemplateIDExternal,
		Fields:             t.Fields,
		Enabled:            t.Enabled,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
	}
}

func toNotificationTaskResp(t *NotificationTask) NotificationTaskResp {
	return NotificationTaskResp{
		ID:           types.Int64Str(t.ID),
		TemplateCode: t.TemplateCode,
		TargetType:   t.TargetType,
		Target:       t.Target,
		Params:       t.Params,
		Status:       t.Status,
		RetryCount:   t.RetryCount,
		LastError:    t.LastError,
		DedupKey:     t.DedupKey,
		SentAt:       t.SentAt,
		CreatedAt:    t.CreatedAt,
	}
}
