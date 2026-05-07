package cms

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler CMS 模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---- C 端 ----

// ListPublished C 端获取已发布文章列表。
func (h *Handler) ListPublished(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.ListPublished(c.Request.Context(), keyword, page, size)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": size})
}

// GetPublished C 端获取已发布文章详情。
func (h *Handler) GetPublished(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 id"))
		return
	}
	a, err := h.svc.GetPublished(c.Request.Context(), id)
	if err != nil {
		srv.Fail(c, errs.ErrNotFound)
		return
	}
	srv.OK(c, a)
}

// ---- Admin ----

// AdminList Admin 获取文章列表。
func (h *Handler) AdminList(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.AdminList(c.Request.Context(), ArticleFilter{
		Status: status, Keyword: keyword, Page: page, PageSize: size,
	})
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": size})
}

// AdminGet Admin 获取文章详情。
func (h *Handler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 id"))
		return
	}
	a, err := h.svc.AdminGet(c.Request.Context(), id)
	if err != nil {
		srv.Fail(c, errs.ErrNotFound)
		return
	}
	srv.OK(c, a)
}

// AdminCreate Admin 创建文章。
func (h *Handler) AdminCreate(c *gin.Context) {
	var req UpsertArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	a, err := h.svc.Create(c.Request.Context(), adminID, req)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, a)
}

// AdminUpdate Admin 更新文章。
func (h *Handler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 id"))
		return
	}
	var req UpsertArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		srv.Fail(c, errs.ErrNotFound)
		return
	}
	srv.OK(c, a)
}

// AdminDelete Admin 软删除文章。
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, nil)
}
