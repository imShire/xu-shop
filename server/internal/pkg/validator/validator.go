// Package validator 封装 go-playground/validator，注册自定义校验规则。
package validator

import (
	"regexp"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var mobileRegexp = regexp.MustCompile(`^(1[3-9])\d{9}$`)

// Setup 将自定义校验规则注册到 gin 的全局 binding.Validator 中。
// 必须在 gin.New() 之后、路由注册之前调用。
func Setup() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	_ = v.RegisterValidation("mobile", validateMobile)
	_ = v.RegisterValidation("strongpwd", validateStrongPwd)
}

// New 返回配置好的独立 Validate 实例（供非 HTTP 场景使用）。
func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("mobile", validateMobile)
	_ = v.RegisterValidation("strongpwd", validateStrongPwd)
	return v
}

// validateMobile 校验中国大陆手机号。
func validateMobile(fl validator.FieldLevel) bool {
	return mobileRegexp.MatchString(fl.Field().String())
}

// validateStrongPwd 校验密码强度：≥12 字符，含字母 + 数字 + 特殊符号。
func validateStrongPwd(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 12 {
		return false
	}
	var hasLetter, hasDigit, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsLetter(c):
			hasLetter = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasLetter && hasDigit && hasSpecial
}

// TranslateError 将 validator 错误翻译成可读字符串。
func TranslateError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
