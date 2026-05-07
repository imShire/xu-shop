package order

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 订单模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---- C 端 ----

// Create 下单。
func (h *Handler) Create(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	// 从 Idempotency-Key 头覆盖（优先使用 header）
	if idem := c.GetHeader("Idempotency-Key"); idem != "" {
		req.IdempotencyKey = idem
	}
	userID := c.GetInt64("user_id")
	o, err := h.svc.CreateOrder(c.Request.Context(), userID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, o)
}

// List 订单列表。
func (h *Handler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	status := c.Query("status")
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 20)
	list, total, err := h.svc.ListOrders(c.Request.Context(), userID, status, page, size)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// Get 订单详情。
func (h *Handler) Get(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.GetOrder(c.Request.Context(), id, userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// Cancel 取消订单。
func (h *Handler) Cancel(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	userID := c.GetInt64("user_id")
	if err := h.svc.CancelOrder(c.Request.Context(), id, userID, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// WithdrawCancel 撤回取消申请。
func (h *Handler) WithdrawCancel(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.WithdrawCancelRequest(c.Request.Context(), id, userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// Confirm 确认收货。
func (h *Handler) Confirm(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.ConfirmReceived(c.Request.Context(), id, userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// BuyAgain 再次购买。
func (h *Handler) BuyAgain(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	added, err := h.svc.BuyAgain(c.Request.Context(), id, userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"items": added})
}

// ---- 后台 ----

// AdminList 后台订单列表。
func (h *Handler) AdminList(c *gin.Context) {
	filter := AdminOrderFilter{
		Status:  c.Query("status"),
		OrderNo: c.Query("order_no"),
		Page:    queryInt(c, "page", 1),
		Size:    queryInt(c, "page_size", 20),
	}
	if uid := c.Query("user_id"); uid != "" {
		if v, err := strconv.ParseInt(uid, 10, 64); err == nil {
			filter.UserID = v
		}
	}
	list, total, err := h.svc.AdminListOrders(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminGet 后台订单详情。
func (h *Handler) AdminGet(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	resp, err := h.svc.AdminGetOrder(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// AdminExport 导出 CSV。
func (h *Handler) AdminExport(c *gin.Context) {
	filter := AdminOrderFilter{
		Status:  c.Query("status"),
		OrderNo: c.Query("order_no"),
	}
	data, err := h.svc.AdminExportOrders(c.Request.Context(), filter)
	if err != nil {
		failWith(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="orders.csv"`)
	c.Data(200, "text/csv", data)
}

// AdminCancel 后台取消订单。
func (h *Handler) AdminCancel(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	adminID := c.GetInt64("admin_id")
	if err := h.svc.AdminCancelOrder(c.Request.Context(), id, adminID, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminAddRemark 添加管理备注。
func (h *Handler) AdminAddRemark(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.AddRemark(c.Request.Context(), id, adminID, req.Content); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminListRemarks 管理备注列表。
func (h *Handler) AdminListRemarks(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	list, err := h.svc.ListRemarks(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// AdminListFreight 运费模板列表。
func (h *Handler) AdminListFreight(c *gin.Context) {
	list, err := h.svc.ListFreightTemplates(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

// AdminCreateFreight 创建运费模板。
func (h *Handler) AdminCreateFreight(c *gin.Context) {
	var req FreightTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	t, err := h.svc.CreateFreightTemplate(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, t)
}

// AdminUpdateFreight 更新运费模板。
func (h *Handler) AdminUpdateFreight(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req FreightTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateFreightTemplate(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminDeleteFreight 删除运费模板。
func (h *Handler) AdminDeleteFreight(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.DeleteFreightTemplate(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
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
