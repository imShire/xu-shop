package aftersale

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 请求 DTO ----

// ApplyReq 申请售后请求。
type ApplyReq struct {
	OrderID           types.Int64Str  `json:"order_id"           binding:"required"`
	OrderItemID       *types.Int64Str `json:"order_item_id,omitempty"`
	Type              string          `json:"type"               binding:"required,oneof=refund_only refund_return exchange"`
	Reason            string          `json:"reason"             binding:"required,min=2,max=200"`
	RefundAmountCents int64           `json:"refund_amount_cents" binding:"gte=0"`
	Evidence          []string        `json:"evidence,omitempty"`
}

// ExpressReq 回填寄回运单。
type ExpressReq struct {
	CarrierCode string `json:"carrier_code" binding:"required,max=32"`
	WaybillNo   string `json:"waybill_no"   binding:"required,max=64"`
}

// MessageReq 追加协商。
type MessageReq struct {
	Content  string   `json:"content"  binding:"omitempty,max=1000"`
	Evidence []string `json:"evidence,omitempty"`
}

// AgreeReq 商家同意。
type AgreeReq struct {
	SellerRemark    string          `json:"seller_remark,omitempty" binding:"omitempty,max=500"`
	ReturnAddressID *types.Int64Str `json:"return_address_id,omitempty"`
}

// RejectReq 商家拒绝。
type RejectReq struct {
	Reason string `json:"reason" binding:"required,min=2,max=200"`
}

// ConfirmReceivedReq 商家确认收货。
type ConfirmReceivedReq struct {
	SellerRemark string `json:"seller_remark,omitempty" binding:"omitempty,max=500"`
}

// CloseReq 手动关闭。
type CloseReq struct {
	Reason string `json:"reason" binding:"required,min=2,max=200"`
}

// AdminListFilter 后台列表过滤。
type AdminListFilter struct {
	Status      string
	Type        string
	Keyword     string
	AppliedFrom *time.Time
	AppliedTo   *time.Time
	Page        int
	PageSize    int
}

// UserListFilter C 端列表过滤。
type UserListFilter struct {
	UserID   int64
	Status   string
	Page     int
	PageSize int
}

// ---- 响应 DTO ----

// AftersaleResp 售后单响应。
type AftersaleResp struct {
	ID                types.Int64Str  `json:"id"`
	AftersaleNo       string          `json:"aftersale_no"`
	OrderID           types.Int64Str  `json:"order_id"`
	OrderItemID       *types.Int64Str `json:"order_item_id,omitempty"`
	UserID            types.Int64Str  `json:"user_id"`
	Type              string          `json:"type"`
	Status            string          `json:"status"`
	Reason            string          `json:"reason"`
	RefundAmountCents int64           `json:"refund_amount_cents"`
	BuyerEvidence     []string        `json:"buyer_evidence"`
	BuyerExpress      *BuyerExpress   `json:"buyer_express,omitempty"`
	SellerRemark      string          `json:"seller_remark"`
	RefundID          *types.Int64Str `json:"refund_id,omitempty"`
	AppliedAt         time.Time       `json:"applied_at"`
	AgreedAt          *time.Time      `json:"agreed_at,omitempty"`
	ReturnedAt        *time.Time      `json:"returned_at,omitempty"`
	ReceivedAt        *time.Time      `json:"received_at,omitempty"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty"`
	ClosedAt          *time.Time      `json:"closed_at,omitempty"`
	AutoCloseAt       time.Time       `json:"auto_close_at"`
}

// NegotiationResp 协商记录响应。
type NegotiationResp struct {
	ID        types.Int64Str  `json:"id"`
	Role      string          `json:"role"`
	AdminID   *types.Int64Str `json:"admin_id,omitempty"`
	Content   string          `json:"content"`
	Evidence  []string        `json:"evidence"`
	CreatedAt time.Time       `json:"created_at"`
}

// DetailResp 详情响应（含协商记录）。
type DetailResp struct {
	AftersaleResp
	Negotiations []NegotiationResp `json:"negotiations"`
}

// ApplyResp 申请响应。
type ApplyResp struct {
	ID          types.Int64Str `json:"id"`
	AftersaleNo string         `json:"aftersale_no"`
}

func toAftersaleResp(a *AftersaleOrder) AftersaleResp {
	r := AftersaleResp{
		ID:                types.Int64Str(a.ID),
		AftersaleNo:       a.AftersaleNo,
		OrderID:           types.Int64Str(a.OrderID),
		UserID:            types.Int64Str(a.UserID),
		Type:              a.Type,
		Status:            a.Status,
		Reason:            a.Reason,
		RefundAmountCents: a.RefundAmountCents,
		BuyerEvidence:     []string(a.BuyerEvidence),
		BuyerExpress:      a.BuyerExpress,
		SellerRemark:      a.SellerRemark,
		AppliedAt:         a.AppliedAt,
		AgreedAt:          a.AgreedAt,
		ReturnedAt:        a.ReturnedAt,
		ReceivedAt:        a.ReceivedAt,
		CompletedAt:       a.CompletedAt,
		ClosedAt:          a.ClosedAt,
		AutoCloseAt:       a.AutoCloseAt,
	}
	if a.OrderItemID != nil {
		v := types.Int64Str(*a.OrderItemID)
		r.OrderItemID = &v
	}
	if a.RefundID != nil {
		v := types.Int64Str(*a.RefundID)
		r.RefundID = &v
	}
	if r.BuyerEvidence == nil {
		r.BuyerEvidence = []string{}
	}
	return r
}

func toNegotiationResp(n *AftersaleNegotiation) NegotiationResp {
	r := NegotiationResp{
		ID:        types.Int64Str(n.ID),
		Role:      n.Role,
		Content:   n.Content,
		Evidence:  []string(n.Evidence),
		CreatedAt: n.CreatedAt,
	}
	if n.AdminID != nil {
		v := types.Int64Str(*n.AdminID)
		r.AdminID = &v
	}
	if r.Evidence == nil {
		r.Evidence = []string{}
	}
	return r
}
