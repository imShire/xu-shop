package recall

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	xserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 召回 HTTP handler。
type Handler struct{ svc *Service }

// NewHandler 构造。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// AdminListCampaigns GET /admin/recall/campaigns?status=&page=&page_size=
func (h *Handler) AdminListCampaigns(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.ListCampaigns(c.Request.Context(), status, page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"items": list, "total": total})
}

// AdminGetCampaign GET /admin/recall/campaigns/:id
func (h *Handler) AdminGetCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	resp, err := h.svc.GetCampaign(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, resp)
}

// AdminCreateCampaign POST /admin/recall/campaigns
func (h *Handler) AdminCreateCampaign(c *gin.Context) {
	var form CampaignForm
	if err := c.ShouldBindJSON(&form); err != nil {
		xserver.FailParam(c, err)
		return
	}
	aid := c.GetInt64("admin_id")
	resp, err := h.svc.CreateCampaign(c.Request.Context(), form, aid)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, resp)
}

// AdminUpdateCampaign PUT /admin/recall/campaigns/:id
func (h *Handler) AdminUpdateCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	var form CampaignForm
	if err := c.ShouldBindJSON(&form); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateCampaign(c.Request.Context(), id, form); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminTransition POST /admin/recall/campaigns/:id/transition  body: {"to":"online"}
func (h *Handler) AdminTransition(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	var body struct {
		To string `json:"to" binding:"required,oneof=online paused closed"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		xserver.FailParam(c, err)
		return
	}
	aid := c.GetInt64("admin_id")
	if err := h.svc.Transition(c.Request.Context(), id, body.To, aid); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminFunnel GET /admin/recall/campaigns/:id/funnel
func (h *Handler) AdminFunnel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	resp, err := h.svc.FunnelReport(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, resp)
}

// AdminListLogs GET /admin/recall/campaigns/:id/logs?page=&page_size=
func (h *Handler) AdminListLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	logs, total, err := h.svc.ListLogs(c.Request.Context(), id, page, size)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"items": logs, "total": total})
}
