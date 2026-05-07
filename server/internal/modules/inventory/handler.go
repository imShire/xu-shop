package inventory

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 库存模块 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// AdjustReq 手动调整库存请求。
type AdjustReq struct {
	ChangeType string `json:"change_type" binding:"required,oneof=in out set"`
	Change     int    `json:"change"      binding:"required,min=0"`
	Reason     string `json:"reason"      binding:"omitempty,max=200"`
}

// Adjust 手动调整单个 SKU 库存。
func (h *Handler) Adjust(c *gin.Context) {
	skuIDStr := c.Param("id")
	skuID, err := strconv.ParseInt(skuIDStr, 10, 64)
	if err != nil || skuID <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 sku_id"))
		return
	}

	var req AdjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}

	adminID := c.GetInt64("admin_id")
	if err := h.svc.Adjust(c.Request.Context(), adminID, skuID, req.ChangeType, req.Change, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ListLogsReq 日志查询参数。
type ListLogsReq struct {
	SkuCode    string `form:"sku_code"`
	ChangeType string `form:"change_type"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page"      binding:"min=0"`
	PageSize   int    `form:"page_size" binding:"min=0,max=100"`
}

// ListLogs 库存日志列表。
func (h *Handler) ListLogs(c *gin.Context) {
	var req ListLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	filter := LogFilter{
		SkuCode:    req.SkuCode,
		ChangeType: req.ChangeType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}
	list, total, err := h.svc.ListLogs(c.Request.Context(), filter, req.Page, req.PageSize)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// ListAlertsReq 预警查询参数。
type ListAlertsReq struct {
	Status   string `form:"status"`
	Page     int    `form:"page"      binding:"min=0"`
	PageSize int    `form:"page_size" binding:"min=0,max=100"`
}

// ListAlerts 低库存预警列表。
func (h *Handler) ListAlerts(c *gin.Context) {
	var req ListAlertsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	list, total, err := h.svc.ListAlerts(c.Request.Context(), req.Status, req.Page, req.PageSize)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// MarkAlertRead 标记预警已读。
func (h *Handler) MarkAlertRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 alert_id"))
		return
	}

	adminID := c.GetInt64("admin_id")
	if err := h.svc.MarkAlertRead(c.Request.Context(), id, adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// MarkAllAlertsRead 批量标记所有未读预警为已读。
func (h *Handler) MarkAllAlertsRead(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	if err := h.svc.MarkAllAlertsRead(c.Request.Context(), adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func failWith(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		srv.Fail(c, appErr)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}
