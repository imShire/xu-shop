package aftersale

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 售后模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// AdminListAftersales 后台售后订单列表。
func (h *Handler) AdminListAftersales(c *gin.Context) {
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 20)
	list, total, err := h.svc.ListAftersales(c.Request.Context(), page, size)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminApproveCancel 同意取消申请。
func (h *Handler) AdminApproveCancel(c *gin.Context) {
	orderID := mustParamID(c, "order_id")
	if orderID == 0 {
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.ApproveCancel(c.Request.Context(), orderID, adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminRejectCancel 拒绝取消申请。
func (h *Handler) AdminRejectCancel(c *gin.Context) {
	orderID := mustParamID(c, "order_id")
	if orderID == 0 {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	adminID := c.GetInt64("admin_id")
	if err := h.svc.RejectCancel(c.Request.Context(), orderID, adminID, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
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

func queryInt(c *gin.Context, key string, defaultVal int) int {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}
