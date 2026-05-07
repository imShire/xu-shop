package nav_icon

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 金刚区图标模块 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ---- C 端 ----

// ListActiveNavIcons 客户端获取激活金刚区图标列表。
func (h *Handler) ListActiveNavIcons(c *gin.Context) {
	list, err := h.svc.ListActive(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// ---- 后台 ----

// AdminListNavIcons 后台查询所有金刚区图标。
func (h *Handler) AdminListNavIcons(c *gin.Context) {
	list, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// AdminCreateNavIcon 后台创建金刚区图标。
func (h *Handler) AdminCreateNavIcon(c *gin.Context) {
	var req CreateNavIconReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	n, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, n)
}

// AdminUpdateNavIcon 后台更新金刚区图标。
func (h *Handler) AdminUpdateNavIcon(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateNavIconReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	n, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, n)
}

// AdminDeleteNavIcon 后台删除金刚区图标。
func (h *Handler) AdminDeleteNavIcon(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminToggleNavIcon 后台切换金刚区图标激活状态。
func (h *Handler) AdminToggleNavIcon(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	n, err := h.svc.Toggle(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, n)
}

// AdminBulkSortNavIcons 后台批量更新金刚区图标排序。
func (h *Handler) AdminBulkSortNavIcons(c *gin.Context) {
	var req BulkSortReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.BulkSort(c.Request.Context(), req.Items); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 工具函数 ----

func mustParamID(c *gin.Context, key string) int64 {
	s := c.Param(key)
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 "+key))
		return 0
	}
	return id
}

func failWith(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		srv.Fail(c, appErr)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}
