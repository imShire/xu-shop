package distribution

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
	xserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 分销 HTTP 控制器。
type Handler struct {
	svc   *Service
	wxpay pkgwxpay.TransferClient
}

// NewHandler 构造控制器。
func NewHandler(svc *Service, wxpay pkgwxpay.TransferClient) *Handler {
	return &Handler{svc: svc, wxpay: wxpay}
}

func parsePaging(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	return limit, (page - 1) * limit
}

func parseIDParam(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.ErrParam.WithMsg("invalid id")
	}
	return id, nil
}

// ===== C 端 =====

// CreateShareLink 创建分享链接。
func (h *Handler) CreateShareLink(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req CreateShareLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	link, err := h.svc.CreateShareLink(c.Request.Context(), uid, req)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, toShareLinkResp(link, h.svc.ShareBaseURL()))
}

// ResolveShortToken 短链 302 跳转。
//
// 解析 token → ShareLink；按 scene 拼业务页 URL，下发 set-cookie(st=trace) 90d。
func (h *Handler) ResolveShortToken(c *gin.Context) {
	token := c.Param("short_token")
	traceID := middleware.GetShareTraceID(c)
	ua := c.Request.UserAgent()
	ip := c.ClientIP()
	device := c.Query("device")
	referer := c.Request.Referer()
	fingerprint := c.Query("fp")

	link, err := h.svc.ResolveShortToken(c.Request.Context(), token, traceID, ua, ip, device, referer, fingerprint)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	// 拼接前端业务页 URL（H5）
	target := h.svc.ShareBaseURL()
	switch link.Scene {
	case ShareSceneProduct:
		if link.TargetID != nil {
			target += "/pages/product/detail?id=" + strconv.FormatInt(*link.TargetID, 10)
		} else {
			target += "/pages/index/index"
		}
	case ShareSceneActivity:
		if link.TargetID != nil {
			target += "/pages/activity/index?id=" + strconv.FormatInt(*link.TargetID, 10)
		} else {
			target += "/pages/index/index"
		}
	case ShareSceneBrand:
		target += "/pages/brand/index"
	case ShareSceneInviteRegister:
		target += "/pages/account/login"
	default:
		target += "/pages/index/index"
	}
	// 写 cookie 持久化 trace_id
	if traceID == "" {
		traceID = newTraceID()
	}
	c.SetCookie("st", traceID, 90*24*3600, "/", "", false, true)
	c.Redirect(http.StatusFound, appendQuery(target, "st", traceID))
}

func appendQuery(u, k, v string) string {
	sep := "?"
	for _, r := range u {
		if r == '?' {
			sep = "&"
			break
		}
	}
	return u + sep + k + "=" + v
}

// TrackShare H5/SPA 调用：上报点击。
func (h *Handler) TrackShare(c *gin.Context) {
	var req TrackShareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	link, err := h.svc.repo.GetShareLinkByToken(c.Request.Context(), req.ShortToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			xserver.Fail(c, errs.ErrNotFound)
			return
		}
		xserver.Fail(c, errs.ErrInternal.WithMsg(err.Error()))
		return
	}
	device := ""
	if req.Device != nil {
		device = *req.Device
	}
	referer := ""
	if req.Referer != nil {
		referer = *req.Referer
	}
	fp := ""
	if req.VisitorFingerprint != nil {
		fp = *req.VisitorFingerprint
	}
	if err := h.svc.TrackClick(c.Request.Context(), link, req.TraceID, c.Request.UserAgent(), c.ClientIP(), device, referer, fp); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// ApplyDistributor 申请成为分销员。
func (h *Handler) ApplyDistributor(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req ApplyDistributorReq
	_ = c.ShouldBindJSON(&req)
	d, err := h.svc.Apply(c.Request.Context(), uid, req)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, toDistributorResp(d))
}

// GetMyProfile 我的分销资料。
func (h *Handler) GetMyProfile(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	resp, err := h.svc.GetMyProfile(c.Request.Context(), uid)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, resp)
}

// GetMyCommissions 我的佣金。
func (h *Handler) GetMyCommissions(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	limit, offset := parsePaging(c)
	status := c.Query("status")
	list, total, err := h.svc.MyCommissions(c.Request.Context(), uid, status, limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]CommissionResp, 0, len(list))
	for i := range list {
		out = append(out, toCommissionResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// GetMyWithdraws 我的提现。
func (h *Handler) GetMyWithdraws(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	limit, offset := parsePaging(c)
	list, total, err := h.svc.MyWithdraws(c.Request.Context(), uid, limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]WithdrawResp, 0, len(list))
	for i := range list {
		out = append(out, toWithdrawResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// SendWithdrawSms 发送提现验证码。
func (h *Handler) SendWithdrawSms(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	if _, err := h.svc.WithdrawSmsRequest(c.Request.Context(), uid); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// RequestWithdraw 申请提现。
func (h *Handler) RequestWithdraw(c *gin.Context) {
	uid := c.GetInt64("user_id")
	if uid == 0 {
		xserver.Fail(c, errs.ErrUnauth)
		return
	}
	var req WithdrawReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	idemKey := c.GetHeader("Idempotency-Key")
	w, err := h.svc.RequestWithdraw(c.Request.Context(), uid, idemKey, req)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, toWithdrawResp(w))
}

// ===== Admin =====

// AdminListDistributors 列表。
func (h *Handler) AdminListDistributors(c *gin.Context) {
	limit, offset := parsePaging(c)
	list, total, err := h.svc.ListDistributors(c.Request.Context(), c.Query("status"), c.Query("level"), limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]DistributorResp, 0, len(list))
	for i := range list {
		out = append(out, toDistributorResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// AdminApproveDistributor 通过申请。
func (h *Handler) AdminApproveDistributor(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.Approve(c.Request.Context(), id, adminID); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminRejectDistributor 拒绝申请。
func (h *Handler) AdminRejectDistributor(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req RejectReq
	_ = c.ShouldBindJSON(&req)
	if req.Reason == "" {
		req.Reason = "不符合条件"
	}
	if err := h.svc.Reject(c.Request.Context(), id, req.Reason); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminBanDistributor 停用分销员。
func (h *Handler) AdminBanDistributor(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req RejectReq
	_ = c.ShouldBindJSON(&req)
	if req.Reason == "" {
		req.Reason = "管理员停用"
	}
	if err := h.svc.Ban(c.Request.Context(), id, req.Reason); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminAdjustLevel 调整等级。
func (h *Handler) AdminAdjustLevel(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req AdjustLevelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.AdjustLevel(c.Request.Context(), id, req.Level); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminAdjustRate 调整专属费率。
func (h *Handler) AdminAdjustRate(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req AdjustRateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.AdjustRate(c.Request.Context(), id, req.RateOverride); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminListShareLinks admin 列表。
func (h *Handler) AdminListShareLinks(c *gin.Context) {
	uid, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	limit, _ := parsePaging(c)
	list, err := h.svc.ListShareLinks(c.Request.Context(), uid, c.Query("scene"), limit)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]ShareLinkResp, 0, len(list))
	for i := range list {
		out = append(out, toShareLinkResp(&list[i], h.svc.ShareBaseURL()))
	}
	xserver.OK(c, gin.H{"items": out})
}

// AdminListCommissions admin。
func (h *Handler) AdminListCommissions(c *gin.Context) {
	limit, offset := parsePaging(c)
	uid, _ := strconv.ParseInt(c.Query("distributor_user_id"), 10, 64)
	list, total, err := h.svc.AdminListCommissions(c.Request.Context(), c.Query("status"), uid, limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]CommissionResp, 0, len(list))
	for i := range list {
		out = append(out, toCommissionResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// AdminAuditCommission 通用审核入口（action: release/cancel）。
func (h *Handler) AdminAuditCommission(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req AuditCommissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		xserver.FailParam(c, err)
		return
	}
	if err := h.svc.AdminAuditCommission(c.Request.Context(), id, req); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminAuditCommissionRelease release 别名（POST /admin/commissions/:id/release）。
func (h *Handler) AdminAuditCommissionRelease(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	if err := h.svc.AdminAuditCommission(c.Request.Context(), id, AuditCommissionReq{Action: "release"}); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminAuditCommissionCancel cancel 别名。
func (h *Handler) AdminAuditCommissionCancel(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.AdminAuditCommission(c.Request.Context(), id, AuditCommissionReq{Action: "cancel", Reason: req.Reason}); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminListSettlements 结算列表。
func (h *Handler) AdminListSettlements(c *gin.Context) {
	limit, offset := parsePaging(c)
	list, total, err := h.svc.AdminListSettlements(c.Request.Context(), c.Query("status"), limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]SettlementResp, 0, len(list))
	for i := range list {
		out = append(out, toSettlementResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// AdminListWithdraws 提现列表。
func (h *Handler) AdminListWithdraws(c *gin.Context) {
	limit, offset := parsePaging(c)
	list, total, err := h.svc.AdminListWithdraws(c.Request.Context(), c.Query("status"), limit, offset)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	out := make([]WithdrawResp, 0, len(list))
	for i := range list {
		out = append(out, toWithdrawResp(&list[i]))
	}
	xserver.OK(c, gin.H{"items": out, "total": total})
}

// AdminRetryWithdraw 重试失败的提现工单。
func (h *Handler) AdminRetryWithdraw(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		xserver.Fail(c, err.(*errs.AppError))
		return
	}
	w, err := h.svc.repo.GetWithdraw(c.Request.Context(), id)
	if err != nil {
		xserver.Fail(c, errs.ErrNotFound)
		return
	}
	if err := h.svc.WithdrawTransition(c.Request.Context(), w, "retry"); err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, gin.H{"ok": true})
}

// AdminFunnelReport 漏斗报表。
func (h *Handler) AdminFunnelReport(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse("2006-01-02", startStr)
	}
	if endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			end = t.Add(24 * time.Hour)
		}
	}
	resp, err := h.svc.FunnelReport(c.Request.Context(), start, end)
	if err != nil {
		failOrInternal(c, err)
		return
	}
	xserver.OK(c, resp)
}

// NotifyWxTransfer 微信商家转账回调。
//
// 微信要求：成功返回 200 + JSON {"code":"SUCCESS","message":"成功"}；失败返回 4xx/5xx + 同结构。
func (h *Handler) NotifyWxTransfer(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "read body"})
		return
	}
	headers := map[string]string{}
	for k := range c.Request.Header {
		headers[k] = c.Request.Header.Get(k)
	}
	if h.wxpay == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "FAIL", "message": "wxpay not configured"})
		return
	}
	result, err := h.wxpay.VerifyTransferNotify(c.Request.Context(), body, headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}
	if err := h.svc.OnTransferNotify(c.Request.Context(), result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// failOrInternal 统一错误转换。
func failOrInternal(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if ae, ok := err.(*errs.AppError); ok {
		xserver.Fail(c, ae)
		return
	}
	xserver.Fail(c, errs.ErrInternal.WithMsg(err.Error()))
}
