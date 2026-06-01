// Package sensitive 提供请求/响应体的敏感字段脱敏。
//
// 同时服务于审计中间件、HTTP 请求日志中间件等场景，
// 避免密码 / token / 商户密钥等明文落库或落日志。
package sensitive

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SensitiveKeys 默认敏感字段名（小写匹配）。
var SensitiveKeys = []string{
	"password",
	"new_password",
	"old_password",
	"confirm_password",
	"passwd",
	"token",
	"access_token",
	"refresh_token",
	"csrf_token",
	"secret",
	"app_secret",
	"api_key",
	"api_secret",
	"private_key",
	"aes_key",
	"signature",
	"mch_id",
}

const (
	maskString      = "***"
	fallbackMaxKeep = 1024
)

// reFallback 用于非 JSON 文本（form / query / 任意 text）的兜底脱敏：
// 匹配 password=xxx&… / "password":"xxx" / password: xxx 等常见形态。
var reFallback = regexp.MustCompile(
	`(?i)(` + strings.Join(SensitiveKeys, "|") + `)` +
		`(\s*[:=]\s*"?)([^"&\s,}]*)`,
)

// Sanitize 对输入字节做敏感字段脱敏。
//
//   - 优先按 JSON 递归遍历，命中 key 时把 value 替换为 "***"
//   - 解析失败则走正则兜底，并截断到 fallbackMaxKeep
//   - 输入为空时原样返回
func Sanitize(in []byte) []byte {
	if len(in) == 0 {
		return in
	}

	var v any
	dec := json.NewDecoder(strings.NewReader(string(in)))
	dec.UseNumber()
	if err := dec.Decode(&v); err == nil {
		masked := maskJSON(v)
		if out, mErr := json.Marshal(masked); mErr == nil {
			return out
		}
	}

	src := in
	if len(src) > fallbackMaxKeep {
		src = src[:fallbackMaxKeep]
	}
	return reFallback.ReplaceAll(src, []byte(`$1$2`+maskString))
}

// maskJSON 递归遍历 JSON 解码结果，命中敏感 key 时把 value 置为 "***"。
func maskJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if isSensitive(k) {
				out[k] = maskString
				continue
			}
			out[k] = maskJSON(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = maskJSON(val)
		}
		return out
	default:
		return v
	}
}

func isSensitive(key string) bool {
	lk := strings.ToLower(key)
	for _, sk := range SensitiveKeys {
		if lk == sk {
			return true
		}
		// 前缀匹配 secret*（如 secret_key / secret_id）
		if sk == "secret" && strings.HasPrefix(lk, "secret") {
			return true
		}
	}
	return false
}
