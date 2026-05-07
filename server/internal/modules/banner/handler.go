package banner

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 横幅模块 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ---- C 端 ----

// ListActiveBanners 客户端获取激活横幅列表。
func (h *Handler) ListActiveBanners(c *gin.Context) {
	list, err := h.svc.ListActive(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// ---- 后台 ----

// AdminListBanners 后台查询所有横幅。
func (h *Handler) AdminListBanners(c *gin.Context) {
	list, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// AdminCreateBanner 后台创建横幅。
func (h *Handler) AdminCreateBanner(c *gin.Context) {
	var req CreateBannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	b, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, b)
}

// AdminUpdateBanner 后台更新横幅。
func (h *Handler) AdminUpdateBanner(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateBannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	b, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, b)
}

// AdminDeleteBanner 后台删除横幅。
func (h *Handler) AdminDeleteBanner(c *gin.Context) {
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

// AdminToggleBanner 后台切换横幅激活状态。
func (h *Handler) AdminToggleBanner(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	b, err := h.svc.Toggle(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, b)
}

// AdminBulkSortBanners 后台批量更新横幅排序。
func (h *Handler) AdminBulkSortBanners(c *gin.Context) {
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
