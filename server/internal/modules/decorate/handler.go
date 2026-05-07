package decorate

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 页面装修模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GetActivePage C 端获取激活页面配置。
func (h *Handler) GetActivePage(c *gin.Context) {
	pageKey := c.DefaultQuery("page_key", "home")
	cfg, err := h.svc.GetActivePage(c.Request.Context(), pageKey)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, cfg)
}

// AdminListVersions Admin 获取历史版本列表。
func (h *Handler) AdminListVersions(c *gin.Context) {
	pageKey := c.DefaultQuery("page_key", "home")
	list, err := h.svc.ListVersions(c.Request.Context(), pageKey)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, list)
}

// AdminSave Admin 保存页面装修配置新版本。
func (h *Handler) AdminSave(c *gin.Context) {
	var req SaveConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	cfg, err := h.svc.Save(c.Request.Context(), adminID, req)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, cfg)
}

// AdminActivate Admin 激活指定版本。
func (h *Handler) AdminActivate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 id"))
		return
	}
	pageKey := c.DefaultQuery("page_key", "home")
	if err := h.svc.Activate(c.Request.Context(), id, pageKey); err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, nil)
}
