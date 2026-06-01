package reconciliation

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// DiffResp 差异响应 DTO（ID/ref_id 序列化为字符串，沿用红线）。
type DiffResp struct {
	ID            types.Int64Str  `json:"id"`
	Job           string          `json:"job"`
	BizDate       string          `json:"biz_date"` // YYYY-MM-DD
	RefType       string          `json:"ref_type"`
	RefID         string          `json:"ref_id"`
	Field         string          `json:"field"`
	ExpectedValue *string         `json:"expected_value,omitempty"`
	ActualValue   *string         `json:"actual_value,omitempty"`
	DiffCents     *int64          `json:"diff_cents,omitempty"`
	Severity      string          `json:"severity"`
	Status        string          `json:"status"`
	Note          *string         `json:"note,omitempty"`
	AckedBy       *types.Int64Str `json:"acked_by,omitempty"`
	AckedAt       *time.Time      `json:"acked_at,omitempty"`
	ResolvedBy    *types.Int64Str `json:"resolved_by,omitempty"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ToDiffResp 将实体转换为响应 DTO。
func ToDiffResp(d *Diff) DiffResp {
	resp := DiffResp{
		ID:            types.Int64Str(d.ID),
		Job:           d.Job,
		BizDate:       d.BizDate.Format("2006-01-02"),
		RefType:       d.RefType,
		RefID:         d.RefID,
		Field:         d.Field,
		ExpectedValue: d.ExpectedValue,
		ActualValue:   d.ActualValue,
		DiffCents:     d.DiffCents,
		Severity:      d.Severity,
		Status:        d.Status,
		Note:          d.Note,
		AckedAt:       d.AckedAt,
		ResolvedAt:    d.ResolvedAt,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	if d.AckedBy != nil {
		v := types.Int64Str(*d.AckedBy)
		resp.AckedBy = &v
	}
	if d.ResolvedBy != nil {
		v := types.Int64Str(*d.ResolvedBy)
		resp.ResolvedBy = &v
	}
	return resp
}

// ListReq 差异查询请求。
type ListReq struct {
	Job      string `form:"job"`
	BizDate  string `form:"biz_date"` // YYYY-MM-DD
	Status   string `form:"status"`
	Severity string `form:"severity"`
	Page     int    `form:"page"`
	Size     int    `form:"size"`
}

// ListResp 差异列表响应。
type ListResp struct {
	Items []DiffResp `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// ResolveReq 解决差异请求体。
type ResolveReq struct {
	Note string `json:"note" binding:"max=500"`
}
