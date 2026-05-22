package tag

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	xserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 标签 HTTP handler。
type Handler struct{ svc *Service }

// NewHandler 构造。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// AdminListTags GET /admin/user-tags?category=&source=
func (h *Handler) AdminListTags(c *gin.Context) {
	list, err := h.svc.ListTags(c.Request.Context(), c.Query("category"), c.Query("source"))
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	resp := make([]TagResp, 0, len(list))
	for i := range list {
		resp = append(resp, toTagResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": resp})
}

// AdminCreateTag POST /admin/user-tags
func (h *Handler) AdminCreateTag(c *gin.Context) {
	var req CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	t, err := h.svc.CreateTag(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, toTagResp(t))
}

// AdminUpdateTag PUT /admin/user-tags/:code
func (h *Handler) AdminUpdateTag(c *gin.Context) {
	code := c.Param("code")
	var req UpdateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateTag(c.Request.Context(), code, req); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminDeleteTag DELETE /admin/user-tags/:code
func (h *Handler) AdminDeleteTag(c *gin.Context) {
	code := c.Param("code")
	if err := h.svc.DeleteTag(c.Request.Context(), code); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminGetUserTags GET /admin/users/:user_id/tags
func (h *Handler) AdminGetUserTags(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || uid <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	list, err := h.svc.GetUserTags(c.Request.Context(), uid)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"items": list})
}

// AdminAddUserTag POST /admin/users/:user_id/tags
func (h *Handler) AdminAddUserTag(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || uid <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	var req AddManualTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.AddManualTag(c.Request.Context(), uid, req); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			xserver.Fail(c, appErr)
			return
		}
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminRemoveUserTag DELETE /admin/users/:user_id/tags/:tag_code
func (h *Handler) AdminRemoveUserTag(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || uid <= 0 {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	code := c.Param("tag_code")
	if code == "" {
		xserver.Fail(c, errs.ErrParam)
		return
	}
	if err := h.svc.RemoveManualTag(c.Request.Context(), uid, code); err != nil {
		xserver.Fail(c, errs.ErrInternal)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminPreviewAudience POST /admin/audience/preview
func (h *Handler) AdminPreviewAudience(c *gin.Context) {
	var req PreviewAudienceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	resp, err := h.svc.PreviewAudience(c.Request.Context(), req.Filter)
	if err != nil {
		xserver.Fail(c, errs.ErrInternal.WithMsg("audience preview: "+err.Error()))
		return
	}
	xserver.OK(c, resp)
}
