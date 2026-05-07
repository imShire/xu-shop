package account

import (
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 请求 DTO ----

// MpLoginReq 小程序登录请求。
type MpLoginReq struct {
	Code string `json:"code" binding:"required"`
}

// H5OAuthCallbackQuery 公众号 OAuth2 回调 query 参数。
type H5OAuthCallbackQuery struct {
	Code  string `form:"code"  binding:"required"`
	State string `form:"state"`
}

// BindPhoneReq 绑定手机号请求。
type BindPhoneReq struct {
	EncryptedData string `json:"encrypted_data" binding:"required"`
	IV            string `json:"iv"             binding:"required"`
}

// RefreshTokenReq 刷新 token 请求。
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutReq 退出登录请求（JTI 由服务端从 token 中解析，无需客户端传）。
type LogoutReq struct{}

// SmsReq 发送短信验证码请求。
type SmsReq struct {
	Phone   string `json:"phone"   binding:"required,mobile"`
	Purpose string `json:"purpose" binding:"required,oneof=register reset"`
}

// PhoneRegisterReq 手机号注册请求。
type PhoneRegisterReq struct {
	Phone    string `json:"phone"    binding:"required,mobile"`
	Code     string `json:"code"     binding:"required,len=6"`
	Password string `json:"password" binding:"required"`
}

// PhoneLoginReq 手机号密码登录请求。
type PhoneLoginReq struct {
	Phone    string `json:"phone"    binding:"required,mobile"`
	Password string `json:"password" binding:"required"`
}

// ResetPasswordReq 重置密码请求。
type ResetPasswordReq struct {
	Phone       string `json:"phone"        binding:"required,mobile"`
	Code        string `json:"code"         binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ClientLoginResp C 端手机号登录/注册响应。
type ClientLoginResp struct {
	Token string   `json:"token"`
	User  UserResp `json:"user"`
}

// UpdateUserReq 更新用户资料请求。
type UpdateUserReq struct {
	Nickname *string    `json:"nickname"`
	Avatar   *string    `json:"avatar"`
	Gender   *int       `json:"gender"`
	Birthday *time.Time `json:"birthday"`
}

// RequestDeactivateReq 申请注销请求。
type RequestDeactivateReq struct {
	Reason string `json:"reason"`
}

// AdminLoginReq 管理员登录请求。
type AdminLoginReq struct {
	Username    string `json:"username"     binding:"required"`
	Password    string `json:"password"     binding:"required"`
	CaptchaID   string `json:"captcha_id"   binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// CreateAdminReq 创建管理员请求。
type CreateAdminReq struct {
	Username string           `json:"username"  binding:"required,min=3,max=64"`
	Password string           `json:"password"  binding:"required,strongpwd"`
	RealName string           `json:"real_name"`
	Phone    *string          `json:"phone"     binding:"omitempty,mobile"`
	RoleIDs  []types.Int64Str `json:"role_ids"`
}

// UpdateAdminReq 更新管理员请求。
type UpdateAdminReq struct {
	RealName *string          `json:"real_name"`
	Phone    *string          `json:"phone" binding:"omitempty,mobile"`
	RoleIDs  []types.Int64Str `json:"role_ids"`
}

// ResetAdminPwdReq 重置密码请求。
type ResetAdminPwdReq struct {
	NewPassword string `json:"new_password" binding:"required,strongpwd"`
}

// CreateRoleReq 创建角色请求。
type CreateRoleReq struct {
	Code        string   `json:"code"         binding:"required,max=32"`
	Name        string   `json:"name"         binding:"required,max=64"`
	PermCodes   []string `json:"perm_codes"`
}

// UpdateRoleReq 更新角色请求。
type UpdateRoleReq struct {
	Name      *string  `json:"name"`
	PermCodes []string `json:"perm_codes"`
}

// ---- 响应 DTO ----

// TokenResp 登录 token 响应（C 端通过 cookie，此结构仅用于 admin）。
type TokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"` // 秒
}

// CaptchaResp 验证码响应。
type CaptchaResp struct {
	CaptchaID string `json:"captcha_id"`
	ImageB64  string `json:"captcha_b64"`
}

// UserResp 用户信息响应。
type UserResp struct {
	ID           types.Int64Str `json:"id"`
	Phone        *string        `json:"phone,omitempty"`
	Nickname     *string        `json:"nickname,omitempty"`
	Avatar       *string        `json:"avatar,omitempty"`
	Gender       int            `json:"gender"`
	Birthday     *time.Time     `json:"birthday,omitempty"`
	Status       string         `json:"status"`
	BalanceCents int64          `json:"balance_cents"`
	CreatedAt    time.Time      `json:"created_at"`
}

// AdminResp 管理员信息响应。
type AdminResp struct {
	ID          types.Int64Str `json:"id"`
	Username    string         `json:"username"`
	RealName    *string        `json:"real_name,omitempty"`
	Phone       *string        `json:"phone,omitempty"`
	Status      string         `json:"status"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	Roles       []string       `json:"roles"`
	Perms       []string       `json:"perms"`
}

// RoleResp 角色信息响应。
type RoleResp struct {
	ID          types.Int64Str   `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	IsSystem    bool             `json:"is_system"`
	Permissions []PermissionResp `json:"permissions,omitempty"`
}

// PermissionResp 权限点响应。
type PermissionResp struct {
	Code   string `json:"code"`
	Module string `json:"module"`
	Action string `json:"action"`
	Name   string `json:"name"`
	Group  string `json:"group"`
}

// toUserResp 将 entity 转为响应 DTO。
func toUserResp(u *User) UserResp {
	return UserResp{
		ID:           types.Int64Str(u.ID),
		Phone:        u.Phone,
		Nickname:     u.Nickname,
		Avatar:       u.Avatar,
		Gender:       u.Gender,
		Birthday:     u.Birthday,
		Status:       u.Status,
		BalanceCents: u.BalanceCents,
		CreatedAt:    u.CreatedAt,
	}
}

// AdminUserDetailResp 管理员查询用户详情响应（balance_cents 已内嵌于 UserResp）。
type AdminUserDetailResp = UserResp

// toAdminResp 将 entity 转为响应 DTO。
func toAdminResp(a *Admin) AdminResp {
	return AdminResp{
		ID:          types.Int64Str(a.ID),
		Username:    a.Username,
		RealName:    a.RealName,
		Phone:       a.Phone,
		Status:      a.Status,
		LastLoginAt: a.LastLoginAt,
		CreatedAt:   a.CreatedAt,
	}
}

func toRoleResp(r *Role) RoleResp {
	resp := RoleResp{
		ID:       types.Int64Str(r.ID),
		Code:     r.Code,
		Name:     r.Name,
		IsSystem: r.IsSystem,
	}
	for _, p := range r.Permissions {
		resp.Permissions = append(resp.Permissions, PermissionResp{
			Code: p.Code, Module: p.Module, Action: p.Action, Name: p.Name, Group: p.Module,
		})
	}
	return resp
}

// AdminCreateUserReq 管理员创建 C 端用户请求。
type AdminCreateUserReq struct {
	Phone    string  `json:"phone"    binding:"required"`
	Password string  `json:"password" binding:"required,min=6"`
	Nickname *string `json:"nickname"`
}

// AdminListUsersQuery 管理员查询用户列表参数。
type AdminListUsersQuery struct {
	Phone    string `form:"phone"`
	Nickname string `form:"nickname"`
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}
