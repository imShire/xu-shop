// Package errs 定义业务错误码和错误结构体。
package errs

import "fmt"

// AppError 是业务错误的统一结构。
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

// New 构造一个 AppError。
func New(code int, msg string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: msg, HTTPStatus: httpStatus}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d msg=%s", e.Code, e.Message)
}

// 预定义错误码常量。
var (
	ErrParam           = New(10001, "参数错误", 400)
	ErrUnauth          = New(10002, "未登录", 401)
	ErrForbidden       = New(10003, "无权限", 403)
	ErrNotFound        = New(10004, "不存在", 404)
	ErrConflict        = New(10005, "冲突", 409)
	ErrRateLimit       = New(10006, "请求过于频繁", 429)
	ErrInternal        = New(10500, "内部错误", 500)
	ErrServiceDegraded = New(10503, "系统繁忙，请稍后重试", 503)
	ErrPhoneFormat     = New(20001, "手机号格式错误", 400)
	ErrAccountLocked   = New(20003, "账号已锁定", 401)
	ErrPhoneBound      = New(20004, "手机号已被绑定", 409)
	ErrSessionExpired  = New(20005, "微信会话已过期，请重新登录", 401)

	// 支付相关错误
	ErrOpenIDMissing = New(50001, "该支付场景缺少 openid，请重新授权", 400)
)

// WithMsg 返回携带自定义消息的新 AppError（不修改原对象）。
func (e *AppError) WithMsg(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg, HTTPStatus: e.HTTPStatus}
}
