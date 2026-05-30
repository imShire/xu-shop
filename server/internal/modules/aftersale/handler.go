package aftersale

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 售后模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ===== C 端 =====

// UserApply C 端申请售后。
func (h *Handler) UserApply(c *gin.Context) {
	uid := c.GetInt64("user_id")
	var req ApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	resp, err := h.svc.Apply(c.Request.Context(), uid, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// UserCancel C 端撤销售后。
func (h *Handler) UserCancel(c *gin.Context) {
	uid := c.GetInt64("user_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), uid, id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// UserFillExpress C 端回填寄回运单。
func (h *Handler) UserFillExpress(c *gin.Context) {
	uid := c.GetInt64("user_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req ExpressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	if err := h.svc.FillExpress(c.Request.Context(), uid, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// UserAppendMessage C 端追加协商。
func (h *Handler) UserAppendMessage(c *gin.Context) {
	uid := c.GetInt64("user_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req MessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	if err := h.svc.AppendMessage(c.Request.Context(), RoleBuyer, uid, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// UserList C 端列表。
func (h *Handler) UserList(c *gin.Context) {
	uid := c.GetInt64("user_id")
	f := UserListFilter{
		UserID:   uid,
		Status:   c.Query("status"),
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 20),
	}
	list, total, err := h.svc.UserList(c.Request.Context(), f)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": f.Page, "page_size": f.PageSize})
}

// UserGet C 端详情。
func (h *Handler) UserGet(c *gin.Context) {
	uid := c.GetInt64("user_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	d, err := h.svc.UserGet(c.Request.Context(), uid, id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, d)
}

// ===== Admin =====

// AdminAgree 商家同意。
func (h *Handler) AdminAgree(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req AgreeReq
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.AdminAgree(c.Request.Context(), adminID, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminReject 商家拒绝。
func (h *Handler) AdminReject(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req RejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	if err := h.svc.AdminReject(c.Request.Context(), adminID, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminConfirmReceived 商家确认收货。
func (h *Handler) AdminConfirmReceived(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req ConfirmReceivedReq
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.AdminConfirmReceived(c.Request.Context(), adminID, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminClose 后台强制关闭。
func (h *Handler) AdminClose(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req CloseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	if err := h.svc.AdminClose(c.Request.Context(), adminID, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminAppendMessage 后台追加协商。
func (h *Handler) AdminAppendMessage(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req MessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg(err.Error()))
		return
	}
	if err := h.svc.AppendMessage(c.Request.Context(), RoleSeller, adminID, id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminList 后台列表。
func (h *Handler) AdminList(c *gin.Context) {
	f := AdminListFilter{
		Status:   c.Query("status"),
		Type:     c.Query("type"),
		Keyword:  c.Query("keyword"),
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 20),
	}
	if v := c.Query("applied_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.AppliedFrom = &t
		}
	}
	if v := c.Query("applied_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.AppliedTo = &t
		}
	}
	list, total, err := h.svc.AdminList(c.Request.Context(), f)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": f.Page, "page_size": f.PageSize})
}

// AdminGet 后台详情。
func (h *Handler) AdminGet(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	d, err := h.svc.AdminGet(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, d)
}

// ===== Legacy cancel-request =====

// AdminLegacyList 旧取消申请列表。
func (h *Handler) AdminLegacyList(c *gin.Context) {
	list, total, err := h.svc.ListLegacyCancelRequests(c.Request.Context(),
		queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminLegacyApprove 同意旧取消申请。
func (h *Handler) AdminLegacyApprove(c *gin.Context) {
	orderID := mustParamID(c, "order_id")
	if orderID == 0 {
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.LegacyApproveCancel(c.Request.Context(), orderID, adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminLegacyReject 拒绝旧取消申请。
func (h *Handler) AdminLegacyReject(c *gin.Context) {
	orderID := mustParamID(c, "order_id")
	if orderID == 0 {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	adminID := c.GetInt64("admin_id")
	if err := h.svc.LegacyRejectCancel(c.Request.Context(), orderID, adminID, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ===== helpers =====

func failWith(c *gin.Context, err error) {
	if ae, ok := err.(*errs.AppError); ok {
		srv.Fail(c, ae)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}

func mustParamID(c *gin.Context, name string) int64 {
	raw := c.Param(name)
	if raw == "" {
		srv.Fail(c, errs.ErrParam)
		return 0
	}
	if v, err := url.PathUnescape(raw); err == nil {
		raw = v
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam)
		return 0
	}
	return id
}

func queryInt(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
