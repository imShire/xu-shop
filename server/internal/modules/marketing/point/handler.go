package point

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/types"
	xserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 积分 HTTP handler。
type Handler struct{ svc *Service }

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// CBalance GET /c/points/balance
func (h *Handler) CBalance(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	bal, err := h.svc.Balance(c.Request.Context(), uid)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"balance": bal})
}

// CHistory GET /c/points/history
func (h *Handler) CHistory(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.History(c.Request.Context(), uid, page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	type item struct {
		ID           types.Int64Str `json:"id"`
		Change       int64          `json:"change"`
		Type         string         `json:"type"`
		BalanceAfter int64          `json:"balance_after"`
		Reason       string         `json:"reason"`
		CreatedAt    string         `json:"created_at"`
	}
	items := make([]item, 0, len(list))
	for _, t := range list {
		items = append(items, item{
			ID:           types.Int64Str(t.ID),
			Change:       t.Change,
			Type:         t.Type,
			BalanceAfter: t.BalanceAfter,
			Reason:       t.Reason,
			CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	xserver.OK(c, gin.H{"items": items, "total": total, "page": page, "size": size})
}

// AdminCreateTicket POST /admin/points/tickets
func (h *Handler) AdminCreateTicket(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	if adminID == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req struct {
		UserID types.Int64Str `json:"user_id" binding:"required"`
		Change int64          `json:"change" binding:"required"`
		Reason string         `json:"reason" binding:"required,min=2,max=200"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	t, err := h.svc.AdjustTicketCreate(c.Request.Context(), req.UserID.Int64(), req.Change, req.Reason, adminID)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"id": types.Int64Str(t.ID), "status": t.Status})
}

// AdminApproveTicket POST /admin/points/tickets/:id/approve
func (h *Handler) AdminApproveTicket(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	if err := h.svc.AdjustTicketApprove(c.Request.Context(), id, adminID); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminRejectTicket POST /admin/points/tickets/:id/reject
func (h *Handler) AdminRejectTicket(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	if err := h.svc.AdjustTicketReject(c.Request.Context(), id, adminID); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}
