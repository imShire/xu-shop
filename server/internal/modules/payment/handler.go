package payment

import (
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// Handler 支付模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---- C 端 ----

// Prepay 发起预支付。
func (h *Handler) Prepay(c *gin.Context) {
	var req struct {
		OrderID  types.Int64Str `json:"order_id"  binding:"required"`
		Scene    string         `json:"scene"     binding:"required"`
		ClientIP string         `json:"client_ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	result, err := h.svc.Prepay(c.Request.Context(), userID, req.OrderID.Int64(), req.Scene, req.ClientIP)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, result)
}

// QueryPayStatus C 端查询订单支付状态。
func (h *Handler) QueryPayStatus(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.QueryPayStatus(c.Request.Context(), id, userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// ---- 微信回调 ----

// WxpayNotify 微信支付回调（无 auth）。
func (h *Handler) WxpayNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		srv.Fail(c, errs.ErrParam)
		return
	}
	headers := extractWxHeaders(c)
	if err := h.svc.HandleWxpayNotify(c.Request.Context(), body, headers); err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	c.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

// WxpayRefundNotify 微信退款回调（无 auth）。
func (h *Handler) WxpayRefundNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		srv.Fail(c, errs.ErrParam)
		return
	}
	headers := extractWxHeaders(c)
	if err := h.svc.HandleWxpayRefundNotify(c.Request.Context(), body, headers); err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	c.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
}

// ---- 后台 ----

// AdminApplyRefund 管理员发起退款。
func (h *Handler) AdminApplyRefund(c *gin.Context) {
	orderID := mustParamID(c, "id")
	if orderID == 0 {
		return
	}
	var req struct {
		AmtCents int64  `json:"amount_cents" binding:"required,min=1"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.ApplyRefund(c.Request.Context(), orderID, adminID, req.AmtCents, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminListPayments 后台支付列表。
func (h *Handler) AdminListPayments(c *gin.Context) {
	filter := PaymentFilter{
		Status: c.Query("status"),
		Page:   queryInt(c, "page", 1),
		Size:   queryInt(c, "page_size", 20),
	}
	if v := c.Query("order_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.OrderID = id
		}
	}
	list, total, err := h.svc.ListPayments(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminListRefunds 后台退款列表。
func (h *Handler) AdminListRefunds(c *gin.Context) {
	var orderID int64
	if v := c.Query("order_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			orderID = id
		}
	}
	list, err := h.svc.ListRefunds(c.Request.Context(), orderID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": int64(len(list))})
}

// AdminListDiffs 后台对账差异列表。
func (h *Handler) AdminListDiffs(c *gin.Context) {
	filter := DiffFilter{
		Date:   c.Query("date"),
		Status: c.Query("status"),
		Page:   queryInt(c, "page", 1),
		Size:   queryInt(c, "page_size", 20),
	}
	list, total, err := h.svc.ListDiffs(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminResolveDiff 标记对账差异已处理。
func (h *Handler) AdminResolveDiff(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.ResolveDiff(c.Request.Context(), id, adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminListAuditLogs 后台操作审计日志列表。
func (h *Handler) AdminListAuditLogs(c *gin.Context) {
	filter := AuditLogFilter{
		Module:    c.Query("module"),
		Operator:  c.Query("operator"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Page:      queryInt(c, "page", 1),
		Size:      queryInt(c, "page_size", 20),
	}
	list, total, err := h.svc.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// ---- 工具 ----

func extractWxHeaders(c *gin.Context) map[string]string {
	return map[string]string{
		"Wechatpay-Signature": c.GetHeader("Wechatpay-Signature"),
		"Wechatpay-Timestamp": c.GetHeader("Wechatpay-Timestamp"),
		"Wechatpay-Nonce":     c.GetHeader("Wechatpay-Nonce"),
		"Wechatpay-Serial":    c.GetHeader("Wechatpay-Serial"),
	}
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
