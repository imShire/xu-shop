package address

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 地址模块 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List 查询当前用户地址列表。
func (h *Handler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// Get 查询单条地址。
func (h *Handler) Get(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.Get(c.Request.Context(), id, userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// Create 新增地址。
func (h *Handler) Create(c *gin.Context) {
	var req CreateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// Update 更新地址。
func (h *Handler) Update(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Update(c.Request.Context(), id, userID, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// Delete 删除地址。
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

// SetDefault 设置默认地址。
func (h *Handler) SetDefault(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.SetDefault(c.Request.Context(), id, userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// DecryptWxAddress 解密微信地址数据。
func (h *Handler) DecryptWxAddress(c *gin.Context) {
	var req DecryptWxAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	data, err := h.svc.DecryptWxAddress(c.Request.Context(), userID, req.EncryptedData, req.IV)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, data)
}

// ListRegions 查询行政区划列表（open 接口）。
func (h *Handler) ListRegions(c *gin.Context) {
	parentCode := c.Query("parent_code")
	resp, err := h.svc.ListRegions(c.Request.Context(), parentCode)
	if err != nil {
		failWith(c, err)
		return
	}
	// 公共接口设置缓存头
	c.Header("Cache-Control", "public, max-age=86400")
	srv.OK(c, resp)
}

// ---- 工具函数 ----

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
