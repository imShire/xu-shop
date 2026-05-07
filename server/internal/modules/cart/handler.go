package cart

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// Handler 购物车模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List 获取购物车列表。
func (h *Handler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// Add 添加商品到购物车。
func (h *Handler) Add(c *gin.Context) {
	var req struct {
		SkuID types.Int64Str `json:"sku_id" binding:"required"`
		Qty   int            `json:"qty"    binding:"required,min=1,max=999"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Add(c.Request.Context(), userID, req.SkuID.Int64(), req.Qty); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// Update 修改数量。
func (h *Handler) Update(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req struct {
		Qty int `json:"qty" binding:"required,min=1,max=999"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Update(c.Request.Context(), id, userID, req.Qty); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// Delete 删除单条。
func (h *Handler) Delete(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// BatchDelete 批量删除。
func (h *Handler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []types.Int64Str `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	ids := make([]int64, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = id.Int64()
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.BatchDelete(c.Request.Context(), userID, ids); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// CleanInvalid 清除不可用条目。
func (h *Handler) CleanInvalid(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if err := h.svc.CleanInvalid(c.Request.Context(), userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// Count 购物车数量。
func (h *Handler) Count(c *gin.Context) {
	userID := c.GetInt64("user_id")
	cnt, err := h.svc.Count(c.Request.Context(), userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"count": cnt})
}

// Precheck 下单前检查。
func (h *Handler) Precheck(c *gin.Context) {
	var req struct {
		IDs []types.Int64Str `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	ids := make([]int64, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = id.Int64()
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.Precheck(c.Request.Context(), userID, ids)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

func failWith(c *gin.Context, err error) {
	if ae, ok := err.(*errs.AppError); ok {
		srv.Fail(c, ae)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}

func mustParamID(c *gin.Context, name string) int64 {
	var id int64
	if _, err := fmt.Sscanf(c.Param(name), "%d", &id); err != nil || id == 0 {
		srv.Fail(c, errs.ErrParam)
		return 0
	}
	return id
}
