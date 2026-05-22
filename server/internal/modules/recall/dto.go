package recall

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// CampaignForm 活动新建/修改表单。
//
// audience_filter / actions / trigger_config 直接透传 jsonb 对象。
type CampaignForm struct {
	Name                  string    `json:"name" binding:"required,max=128"`
	Goal                  string    `json:"goal" binding:"max=500"`
	AudienceFilter        JSONMap   `json:"audience_filter"`
	Actions               JSONArray `json:"actions"`
	TriggerType           string    `json:"trigger_type" binding:"required,oneof=cron event immediate"`
	TriggerConfig         JSONMap   `json:"trigger_config"`
	EffectiveFrom         *time.Time `json:"effective_from"`
	EffectiveTo           *time.Time `json:"effective_to"`
	ThrottlePerUserDays   int       `json:"throttle_per_user_days"`
	DailyQuota            int64     `json:"daily_quota"`
	TotalQuota            int64     `json:"total_quota"`
	AttributionWindowDays int       `json:"attribution_window_days"`
}

// CampaignResp 活动响应。
type CampaignResp struct {
	ID                    types.Int64Str `json:"id"`
	Name                  string         `json:"name"`
	Goal                  string         `json:"goal"`
	AudienceFilter        JSONMap        `json:"audience_filter"`
	Actions               JSONArray      `json:"actions"`
	TriggerType           string         `json:"trigger_type"`
	TriggerConfig         JSONMap        `json:"trigger_config"`
	EffectiveFrom         *time.Time     `json:"effective_from,omitempty"`
	EffectiveTo           *time.Time     `json:"effective_to,omitempty"`
	ThrottlePerUserDays   int            `json:"throttle_per_user_days"`
	DailyQuota            int64          `json:"daily_quota"`
	TotalQuota            int64          `json:"total_quota"`
	AttributionWindowDays int            `json:"attribution_window_days"`
	Status                string         `json:"status"`
	CreatedBy             types.Int64Str `json:"created_by"`
	ApproverAdminID       *types.Int64Str `json:"approver_admin_id,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func toCampaignResp(c *RecallCampaign) CampaignResp {
	resp := CampaignResp{
		ID:                    types.Int64Str(c.ID),
		Name:                  c.Name,
		Goal:                  c.Goal,
		AudienceFilter:        c.AudienceFilter,
		Actions:               c.Actions,
		TriggerType:           c.TriggerType,
		TriggerConfig:         c.TriggerConfig,
		EffectiveFrom:         c.EffectiveFrom,
		EffectiveTo:           c.EffectiveTo,
		ThrottlePerUserDays:   c.ThrottlePerUserDays,
		DailyQuota:            c.DailyQuota,
		TotalQuota:            c.TotalQuota,
		AttributionWindowDays: c.AttributionWindowDays,
		Status:                c.Status,
		CreatedBy:             types.Int64Str(c.CreatedBy),
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
	}
	if c.ApproverAdminID != nil {
		v := types.Int64Str(*c.ApproverAdminID)
		resp.ApproverAdminID = &v
	}
	return resp
}

// FunnelResp 漏斗统计响应。
type FunnelResp struct {
	CampaignID  types.Int64Str `json:"campaign_id"`
	Triggered   int64          `json:"triggered"`
	Opened      int64          `json:"opened"`
	Converted   int64          `json:"converted"`
	GMVCents    int64          `json:"gmv_cents"`
	OpenRate    float64        `json:"open_rate"`
	ConvertRate float64        `json:"convert_rate"`
}

// TestSendReq 测试触达。
type TestSendReq struct {
	UserID types.Int64Str `json:"user_id" binding:"required"`
}
