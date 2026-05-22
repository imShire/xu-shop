// Package recall 实现召回活动引擎：触发器（cron / event / immediate）+ 节流 + 触达执行 + 漏斗统计。
package recall

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ===== 状态常量 =====

const (
	StatusDraft  = "draft"
	StatusOnline = "online"
	StatusPaused = "paused"
	StatusClosed = "closed"

	TriggerCron      = "cron"
	TriggerEvent     = "event"
	TriggerImmediate = "immediate"

	// 事件触发器名称（trigger_config.event 可选值）。
	EventOrderPaidAfter30D = "order_paid_after_30d"
	EventBirthdayToday     = "birthday_today"
	EventCartAbandoned2H   = "cart_abandoned_2h"
	EventOrderCompleted    = "order_completed"

	// 动作类型
	ActionGrantCoupon = "grant_coupon"
	ActionWxSubscribe = "wx_subscribe_msg"
	ActionInbox       = "inbox"
)

// ===== JSONB 辅助 =====

// JSONMap jsonb map。
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

// JSONArray jsonb 任意数组。
type JSONArray []any

func (a JSONArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *JSONArray) Scan(value any) error {
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
		return fmt.Errorf("JSONArray: unsupported type %T", value)
	}
	return json.Unmarshal(b, a)
}

// ===== GORM 模型 =====

// RecallCampaign 召回活动。
type RecallCampaign struct {
	ID                    int64      `gorm:"column:id;primaryKey" json:"id"`
	Name                  string     `gorm:"column:name" json:"name"`
	Goal                  string     `gorm:"column:goal" json:"goal"`
	AudienceFilter        JSONMap    `gorm:"column:audience_filter;type:jsonb" json:"audience_filter"`
	Actions               JSONArray  `gorm:"column:actions;type:jsonb" json:"actions"`
	TriggerType           string     `gorm:"column:trigger_type" json:"trigger_type"`
	TriggerConfig         JSONMap    `gorm:"column:trigger_config;type:jsonb" json:"trigger_config"`
	EffectiveFrom         *time.Time `gorm:"column:effective_from" json:"effective_from,omitempty"`
	EffectiveTo           *time.Time `gorm:"column:effective_to" json:"effective_to,omitempty"`
	ThrottlePerUserDays   int        `gorm:"column:throttle_per_user_days" json:"throttle_per_user_days"`
	DailyQuota            int64      `gorm:"column:daily_quota" json:"daily_quota"`
	TotalQuota            int64      `gorm:"column:total_quota" json:"total_quota"`
	AttributionWindowDays int        `gorm:"column:attribution_window_days" json:"attribution_window_days"`
	Status                string     `gorm:"column:status" json:"status"`
	CreatedBy             int64      `gorm:"column:created_by" json:"created_by"`
	ApproverAdminID       *int64     `gorm:"column:approver_admin_id" json:"approver_admin_id,omitempty"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 显式声明。
func (RecallCampaign) TableName() string { return "recall_campaign" }

// RecallLog 召回触达日志。
type RecallLog struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	CampaignID        int64      `gorm:"column:campaign_id" json:"campaign_id"`
	UserID            int64      `gorm:"column:user_id" json:"user_id"`
	TriggeredAt       time.Time  `gorm:"column:triggered_at" json:"triggered_at"`
	AudienceSnapshot  JSONMap    `gorm:"column:audience_snapshot;type:jsonb" json:"audience_snapshot"`
	ActionsResult     JSONArray  `gorm:"column:actions_result;type:jsonb" json:"actions_result"`
	OpenedAt          *time.Time `gorm:"column:opened_at" json:"opened_at,omitempty"`
	ConvertedOrderID  *int64     `gorm:"column:converted_order_id" json:"converted_order_id,omitempty"`
	ConvertedAt       *time.Time `gorm:"column:converted_at" json:"converted_at,omitempty"`
	ConvertedGMVCents int64      `gorm:"column:converted_gmv_cents" json:"converted_gmv_cents"`
}

// TableName 显式声明。
func (RecallLog) TableName() string { return "recall_log" }
