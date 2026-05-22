package coupon

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/types"
	xserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 优惠券 HTTP handler。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// CClaim C 端领券 POST /c/coupons/claim
func (h *Handler) CClaim(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req ClaimReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	uc, err := h.svc.Claim(c.Request.Context(), userID, req.TemplateID.Int64(), "claim", nil)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"id": types.Int64Str(uc.ID), "expire_at": uc.ExpireAt})
}

// CClaimByCode C 端兑换码领券 POST /c/coupons/redeem
func (h *Handler) CClaimByCode(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req struct {
		Code string `json:"code" binding:"required,min=4,max=32"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	uc, err := h.svc.ClaimByCode(c.Request.Context(), userID, req.Code)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"id": types.Int64Str(uc.ID)})
}

// CMyList GET /c/coupons/my?status=unused|used|expired&page=&size=
func (h *Handler) CMyList(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	list, total, err := h.svc.MyList(c.Request.Context(), userID, status, page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	items := make([]MyCouponItem, 0, len(list))
	for _, uc := range list {
		item := MyCouponItem{
			ID:         types.Int64Str(uc.ID),
			TemplateID: types.Int64Str(uc.CouponTemplateID),
			Status:     uc.Status,
			ExpireAt:   uc.ExpireAt,
		}
		if uc.Snapshot != nil {
			item.Name, _ = uc.Snapshot["name"].(string)
			item.Type, _ = uc.Snapshot["type"].(string)
			if v, ok := uc.Snapshot["value_cents"].(float64); ok {
				item.ValueCents = int64(v)
			}
			if v, ok := uc.Snapshot["min_amount_cents"].(float64); ok {
				item.MinAmountCents = int64(v)
			}
			if v, ok := uc.Snapshot["discount_rate"].(float64); ok {
				rate := v
				item.DiscountRate = &rate
			}
		}
		items = append(items, item)
	}
	xserver.OK(c, gin.H{"items": items, "total": total, "page": page, "size": size})
}

// CPublicList GET /c/coupons/list
func (h *Handler) CPublicList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.PublicList(c.Request.Context(), page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	type item struct {
		ID             types.Int64Str `json:"id"`
		Name           string         `json:"name"`
		Description    string         `json:"description"`
		Type           string         `json:"type"`
		ValueCents     int64          `json:"value_cents"`
		MinAmountCents int64          `json:"min_amount_cents"`
		PerUserLimit   int            `json:"per_user_limit"`
		Remaining      int64          `json:"remaining"`
	}
	items := make([]item, 0, len(list))
	for _, t := range list {
		remaining := int64(-1)
		if t.TotalQuota > 0 {
			remaining = t.TotalQuota - t.ClaimedCount
		}
		items = append(items, item{
			ID:             types.Int64Str(t.ID),
			Name:           t.Name,
			Description:    t.Description,
			Type:           t.Type,
			ValueCents:     t.ValueCents,
			MinAmountCents: t.MinAmountCents,
			PerUserLimit:   t.PerUserLimit,
			Remaining:      remaining,
		})
	}
	xserver.OK(c, gin.H{"items": items, "total": total, "page": page, "size": size})
}

// AdminListTemplates 后台模板列表（占位实现：批次 2 由后端 B 完善过滤+分页）。
func (h *Handler) AdminListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.PublicList(c.Request.Context(), page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"items": list, "total": total, "page": page, "size": size})
}

// AdminGetTemplate 后台获取模板详情。
func (h *Handler) AdminGetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	t, err := h.svc.repo.FindTemplate(c.Request.Context(), id)
	if err != nil {
		xserver.Fail(c, errs.ErrNotFound)
		return
	}
	xserver.OK(c, t)
}
