package shipping

import (
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 发货模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---- 发货地址 ----

func (h *Handler) ListSenderAddresses(c *gin.Context) {
	list, err := h.svc.ListSenderAddresses(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

func (h *Handler) CreateSenderAddress(c *gin.Context) {
	var req SenderAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	a, err := h.svc.CreateSenderAddress(c.Request.Context(), &req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, a)
}

func (h *Handler) UpdateSenderAddress(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req SenderAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateSenderAddress(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) DeleteSenderAddress(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.DeleteSenderAddress(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) SetDefaultSenderAddress(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.SetDefaultSenderAddress(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 快递公司 ----

func (h *Handler) ListCarriers(c *gin.Context) {
	list, err := h.svc.ListCarriers(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

func (h *Handler) UpdateCarrier(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		srv.Fail(c, errs.ErrParam)
		return
	}
	var req Carrier
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateCarrier(c.Request.Context(), code, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 发货 ----

func (h *Handler) Ship(c *gin.Context) {
	orderID := mustParamID(c, "id")
	if orderID == 0 {
		return
	}
	var req ShipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	ship, err := h.svc.Ship(c.Request.Context(), orderID, adminID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, ship)
}

func (h *Handler) BatchShip(c *gin.Context) {
	var req BatchShipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	taskID, err := h.svc.BatchShip(c.Request.Context(), adminID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"task_id": taskID})
}

func (h *Handler) GetBatchShipStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	status, err := h.svc.GetBatchShipStatus(c.Request.Context(), taskID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, status)
}

func (h *Handler) GetBatchShipPDF(c *gin.Context) {
	taskID := c.Param("task_id")
	pdfURL, err := h.svc.GetBatchShipPDF(c.Request.Context(), taskID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"pdf_url": pdfURL})
}

func (h *Handler) ListShipments(c *gin.Context) {
	filter := ShipmentFilter{
		Status: c.Query("status"),
		Page:   queryInt(c, "page", 1),
		Size:   queryInt(c, "page_size", 20),
	}
	if v := c.Query("order_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.OrderID = id
		}
	}
	list, total, err := h.svc.ListShipments(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

func (h *Handler) UpdateShipment(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateShipmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateShipment(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- C 端 ----

func (h *Handler) ListTracks(c *gin.Context) {
	orderID := mustParamID(c, "id")
	if orderID == 0 {
		return
	}
	tracks, err := h.svc.ListTracks(c.Request.Context(), orderID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"tracks": tracks})
}

// ---- 快递鸟 Webhook ----

func (h *Handler) ExpressPush(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		srv.Fail(c, errs.ErrParam)
		return
	}
	if err := h.svc.HandleExpressPush(c.Request.Context(), body); err != nil {
		// 即使失败也返回 200，避免快递鸟重复推送
		srv.Fail(c, errs.ErrInternal)
		return
	}
	c.JSON(200, gin.H{"EBusinessID": "ok", "UpdateStatus": 1, "Reason": ""})
}

// ---- 工具 ----

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
