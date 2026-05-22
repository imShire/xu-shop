package tag

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ===== 请求 DTO =====

// CreateTagReq 创建标签字典请求。
//
// 仅允许创建 source=manual 的 business 类标签；其他 category 由系统内置。
type CreateTagReq struct {
	Code        string  `json:"code" binding:"required,max=64"`
	Name        string  `json:"name" binding:"required,max=64"`
	Category    string  `json:"category" binding:"required,oneof=business member"`
	ParentCode  *string `json:"parent_code"`
	Color       *string `json:"color"`
	Description string  `json:"description" binding:"max=500"`
	Sort        int     `json:"sort"`
}

// UpdateTagReq 更新标签字典请求。
type UpdateTagReq struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	Sort        *int    `json:"sort"`
}

// AddManualTagReq 给用户加 manual 标签。
type AddManualTagReq struct {
	TagCode  string     `json:"tag_code" binding:"required,max=64"`
	ExpireAt *time.Time `json:"expire_at"`
}

// AudienceFilter 人群筛选条件。
//
// 嵌套结构：
//   - Op: "and" | "or"，默认 and
//   - Children: 嵌套子条件
//   - IncludeTags / ExcludeTags: 当前节点直接命中的标签条件
//   - Behavior: 行为条件（最近 N 天下单 / 累计单数 / GMV）
//
// 仅在叶子节点写 IncludeTags / ExcludeTags / Behavior；中间节点用 Children。
type AudienceFilter struct {
	Op           string           `json:"op,omitempty"`
	Children     []AudienceFilter `json:"children,omitempty"`
	IncludeTags  []string         `json:"include_tags,omitempty"`
	ExcludeTags  []string         `json:"exclude_tags,omitempty"`
	Behavior     *AudienceBehavior `json:"behavior,omitempty"`
}

// AudienceBehavior 行为条件。
type AudienceBehavior struct {
	LastOrderDaysGTE *int   `json:"last_order_days_gte,omitempty"`
	LastOrderDaysLTE *int   `json:"last_order_days_lte,omitempty"`
	OrderCountGTE    *int   `json:"order_count_gte,omitempty"`
	OrderCountLTE    *int   `json:"order_count_lte,omitempty"`
	GMVCentsGTE      *int64 `json:"gmv_cents_gte,omitempty"`
	GMVCentsLTE      *int64 `json:"gmv_cents_lte,omitempty"`
}

// PreviewAudienceReq 人群预估请求。
type PreviewAudienceReq struct {
	Filter AudienceFilter `json:"filter"`
}

// ===== 响应 DTO =====

// TagResp 标签字典响应。
type TagResp struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	ParentCode  *string   `json:"parent_code,omitempty"`
	Color       *string   `json:"color,omitempty"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Enabled     bool      `json:"enabled"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserTagResp 用户标签关系响应。
type UserTagResp struct {
	ID        types.Int64Str `json:"id"`
	TagCode   string         `json:"tag_code"`
	TagName   string         `json:"tag_name"`
	Category  string         `json:"category"`
	Source    string         `json:"source"`
	Score     int            `json:"score"`
	ExpireAt  *time.Time     `json:"expire_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// PreviewAudienceResp 人群预估响应。
type PreviewAudienceResp struct {
	Count       int64            `json:"count"`
	SampleUsers []types.Int64Str `json:"sample_users"`
}

func toTagResp(t *UserTag) TagResp {
	return TagResp{
		Code:        t.Code,
		Name:        t.Name,
		Category:    t.Category,
		ParentCode:  t.ParentCode,
		Color:       t.Color,
		Description: t.Description,
		Source:      t.Source,
		Enabled:     t.Enabled,
		Sort:        t.Sort,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
