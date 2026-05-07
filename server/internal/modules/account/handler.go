package account

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
	"github.com/xushop/xu-shop/internal/pkg/pagination"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 账号模块 HTTP 处理器。
type Handler struct {
	svc    *Service
	jwtCfg pkgjwt.Config
	isProd bool
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service, jwtCfg pkgjwt.Config, isProd bool) *Handler {
	return &Handler{svc: svc, jwtCfg: jwtCfg, isProd: isProd}
}

// ---- C 端 handler ----

// MpLogin 小程序登录，返回 access_token + refresh_token（JSON）。
// C 端统一以 JSON body 返回，前端自行存储（或配合 Cookie）。
func (h *Handler) MpLogin(c *gin.Context) {
	var req MpLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	result, err := h.svc.MpLogin(c.Request.Context(), req.Code)
	if err != nil {
		if ae, ok := err.(*errs.AppError); ok {
			srv.Fail(c, ae)
		} else {
			srv.Fail(c, errs.ErrInternal)
		}
		return
	}
	srv.OK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
		"user_id":       result.UserID,
	})
}

// H5GetOAuthURL 获取公众号 OAuth2 授权 URL，返回 redirect_url。
func (h *Handler) H5GetOAuthURL(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")
	if redirectURI == "" {
		srv.Fail(c, errs.ErrParam)
		return
	}
	authURL := h.svc.H5GetOAuthURL(c.Request.Context(), redirectURI, state)
	srv.OK(c, gin.H{"url": authURL})
}

// H5Callback 公众号 OAuth2 回调，将 access_token 设为 HttpOnly cookie 后重定向。
func (h *Handler) H5Callback(c *gin.Context) {
	var query H5OAuthCallbackQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		srv.FailParam(c, err)
		return
	}
	result, err := h.svc.H5Callback(c.Request.Context(), query.Code)
	if err != nil {
		if ae, ok := err.(*errs.AppError); ok {
			srv.Fail(c, ae)
		} else {
			srv.Fail(c, errs.ErrInternal)
		}
		return
	}

	// access_token 设为 HttpOnly + Secure cookie，禁止 JS 读取
	c.SetCookie(
		"access_token",
		result.AccessToken,
		int(result.ExpiresIn),
		"/",
		"",   // domain 由 nginx 注入，此处留空
		true, // secure（HTTPS only）
		true, // httpOnly
	)
	// refresh_token 同样走 cookie
	c.SetCookie(
		"refresh_token",
		result.RefreshToken,
		int(result.ExpiresIn*3),
		"/api/v1/c/auth/refresh",
		"",
		true,
		true,
	)

	// 302 跳转到前端首页（state 中可携带 returnURL）
	target := "/"
	if query.State != "" {
		target = query.State
	}
	c.Redirect(http.StatusFound, target)
}

// BindPhone 绑定手机号。
func (h *Handler) BindPhone(c *gin.Context) {
	var req BindPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.BindPhone(c.Request.Context(), userID, req.EncryptedData, req.IV); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// RefreshToken 刷新 token（从 body 或 cookie 取 refresh_token）。
func (h *Handler) RefreshToken(c *gin.Context) {
	var refreshToken string
	var req RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err == nil {
		refreshToken = req.RefreshToken
	} else {
		// 尝试从 cookie 取
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			refreshToken = cookie
		}
	}
	if refreshToken == "" {
		srv.Fail(c, errs.ErrUnauth)
		return
	}

	result, err := h.svc.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		failWith(c, err)
		return
	}

	// 更新 cookie
	c.SetCookie("access_token", result.AccessToken, int(result.ExpiresIn), "/", "", true, true)
	srv.OK(c, gin.H{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
	})
}

// Logout C 端退出登录。
func (h *Handler) Logout(c *gin.Context) {
	// 解析 token 获取 jti
	claims := extractClaims(c, h.jwtCfg)
	if claims != nil {
		var exp time.Duration
		if claims.ExpiresAt != nil {
			exp = time.Until(claims.ExpiresAt.Time)
		}
		_ = h.svc.Logout(c.Request.Context(), claims.JTI, exp)
	}
	// 清除 cookie
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/api/v1/c/auth/refresh", "", true, true)
	srv.OK(c, nil)
}

// GetMe 查询当前用户信息；未登录时返回 200 + null data。
func (h *Handler) GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		srv.OK(c, nil)
		return
	}
	resp, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// UpdateMe 更新当前用户资料。
func (h *Handler) UpdateMe(c *gin.Context) {
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.UpdateUser(c.Request.Context(), userID, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// RequestDeactivate 申请注销。
func (h *Handler) RequestDeactivate(c *gin.Context) {
	var req RequestDeactivateReq
	_ = c.ShouldBindJSON(&req)
	userID := c.GetInt64("user_id")
	if err := h.svc.RequestDeactivate(c.Request.Context(), userID, req.Reason); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// CancelDeactivate 撤销注销申请。
func (h *Handler) CancelDeactivate(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if err := h.svc.CancelDeactivate(c.Request.Context(), userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 手机号注册/登录 handler ----

// SendSmsCode 发送短信验证码。
func (h *Handler) SendSmsCode(c *gin.Context) {
	var req SmsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.SendSmsCode(c.Request.Context(), req.Phone, req.Purpose); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"msg": "ok"})
}

// PhoneRegister 手机号注册。
func (h *Handler) PhoneRegister(c *gin.Context) {
	var req PhoneRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	resp, err := h.svc.PhoneRegister(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	c.SetCookie(
		"access_token",
		resp.Token,
		int(h.jwtCfg.UserExpiry.Seconds()),
		"/",
		"",
		h.isProd,
		true,
	)
	srv.OK(c, gin.H{"token": resp.Token, "user": resp.User})
}

// PhoneLogin 手机号密码登录。
func (h *Handler) PhoneLogin(c *gin.Context) {
	var req PhoneLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	resp, err := h.svc.PhoneLogin(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	c.SetCookie(
		"access_token",
		resp.Token,
		int(h.jwtCfg.UserExpiry.Seconds()),
		"/",
		"",
		h.isProd,
		true,
	)
	srv.OK(c, gin.H{"token": resp.Token, "user": resp.User})
}

// ResetPassword 重置密码。
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"msg": "ok"})
}

// ---- Admin handler ----

// AdminGetCaptcha 获取图形验证码。
func (h *Handler) AdminGetCaptcha(c *gin.Context) {
	resp, err := h.svc.AdminGetCaptcha(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// AdminLogin 管理员登录。
func (h *Handler) AdminLogin(c *gin.Context) {
	var req AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	result, err := h.svc.AdminLogin(c.Request.Context(), req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, result)
}

// AdminLogout 管理员退出。
func (h *Handler) AdminLogout(c *gin.Context) {
	claims := extractClaims(c, h.jwtCfg)
	if claims != nil {
		var exp time.Duration
		if claims.ExpiresAt != nil {
			exp = time.Until(claims.ExpiresAt.Time)
		}
		_ = h.svc.AdminLogout(c.Request.Context(), claims.JTI, exp)
	}
	srv.OK(c, nil)
}

// AdminGetMe 查询当前管理员信息。
func (h *Handler) AdminGetMe(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	resp, err := h.svc.GetAdminMe(c.Request.Context(), adminID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// ListAdmins 分页查询管理员列表。
func (h *Handler) ListAdmins(c *gin.Context) {
	var pager pagination.Req
	if err := c.ShouldBindQuery(&pager); err != nil {
		pager = pagination.DefaultReq()
	}
	list, total, err := h.svc.ListAdmins(c.Request.Context(), pager.Page, pager.PageSize)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, pagination.Resp[AdminResp]{
		List:     list,
		Page:     pager.Page,
		PageSize: pager.PageSize,
		Total:    total,
	})
}

// CreateAdmin 创建管理员。
func (h *Handler) CreateAdmin(c *gin.Context) {
	var req CreateAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	resp, err := h.svc.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// UpdateAdmin 更新管理员信息。
func (h *Handler) UpdateAdmin(c *gin.Context) {
	id := mustParamID(c, "id")
	var req UpdateAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateAdmin(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// DisableAdmin 禁用管理员。
func (h *Handler) DisableAdmin(c *gin.Context) {
	id := mustParamID(c, "id")
	if err := h.svc.DisableAdmin(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// EnableAdmin 启用管理员。
func (h *Handler) EnableAdmin(c *gin.Context) {
	id := mustParamID(c, "id")
	if err := h.svc.EnableAdmin(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ResetAdminPwd 重置管理员密码。
func (h *Handler) ResetAdminPwd(c *gin.Context) {
	id := mustParamID(c, "id")
	var req ResetAdminPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.ResetAdminPwd(c.Request.Context(), id, req.NewPassword); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ListRoles 查询角色列表。
func (h *Handler) ListRoles(c *gin.Context) {
	resp, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// CreateRole 创建角色。
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	resp, err := h.svc.CreateRole(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// UpdateRole 更新角色。
func (h *Handler) UpdateRole(c *gin.Context) {
	id := mustParamID(c, "id")
	var req UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateRole(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// DeleteRole 删除角色。
func (h *Handler) DeleteRole(c *gin.Context) {
	id := mustParamID(c, "id")
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ListPermissions 查询所有权限点。
func (h *Handler) ListPermissions(c *gin.Context) {
	resp, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// ---- 工具函数 ----

// failWith 根据错误类型返回对应响应。
func failWith(c *gin.Context, err error) {
	if ae, ok := err.(*errs.AppError); ok {
		srv.Fail(c, ae)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}

// extractClaims 从请求中解析 JWT Claims（不验证状态，仅取 jti）。
func extractClaims(c *gin.Context, cfg pkgjwt.Config) *pkgjwt.Claims {
	var token string
	if h := c.GetHeader("Authorization"); len(h) > 7 {
		token = h[7:]
	}
	if token == "" {
		if cookie, err := c.Cookie("access_token"); err == nil {
			token = cookie
		}
	}
	if token == "" {
		return nil
	}
	claims, _ := pkgjwt.Parse(cfg, token)
	return claims
}

// mustParamID 解析路由中的 int64 ID，失败则直接响应 400。
func mustParamID(c *gin.Context, name string) int64 {
	var id int64
	if _, err := fmt.Sscanf(c.Param(name), "%d", &id); err != nil || id == 0 {
		srv.Fail(c, errs.ErrParam)
		return 0
	}
	return id
}

// AdminListUsers 管理员查询 C 端用户列表。
func (h *Handler) AdminListUsers(c *gin.Context) {
	var q AdminListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		srv.FailParam(c, err)
		return
	}
	list, total, err := h.svc.AdminListUsers(c.Request.Context(), q)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

// AdminGetUser 管理员查询单个 C 端用户详情（含余额）。
func (h *Handler) AdminGetUser(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	resp, err := h.svc.AdminGetUser(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// AdminCreateUser 管理员创建 C 端用户。
func (h *Handler) AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	resp, err := h.svc.AdminCreateUser(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// AdminDisableUser 管理员禁用 C 端用户。
func (h *Handler) AdminDisableUser(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminDisableUser(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminEnableUser 管理员启用 C 端用户。
func (h *Handler) AdminEnableUser(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminEnableUser(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 余额 ----

// GetMyBalance C 端获取本人余额和流水。
func (h *Handler) GetMyBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	balance, err := h.svc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	logs, total, err := h.svc.ListBalanceLogs(c.Request.Context(), userID, page, size)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{
		"balance_cents": balance,
		"logs":          logs,
		"total":         total,
		"page":          page,
		"page_size":     size,
	})
}

// RechargeReq Admin 充值请求。
type RechargeReq struct {
	AmountCents int64  `json:"amount_cents" binding:"required,min=1"`
	Remark      string `json:"remark"       binding:"omitempty,max=200"`
}

// AdminRechargeBalance Admin 给用户充值。
func (h *Handler) AdminRechargeBalance(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 user_id"))
		return
	}
	var req RechargeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.RechargeBalance(c.Request.Context(), userID, req.AmountCents, adminID, req.Remark); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminListBalanceLogs Admin 查看用户余额流水。
func (h *Handler) AdminListBalanceLogs(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 user_id"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	balance, err := h.svc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	logs, total, err := h.svc.ListBalanceLogs(c.Request.Context(), userID, page, size)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{"balance_cents": balance, "list": logs, "total": total, "page": page, "page_size": size})
}
