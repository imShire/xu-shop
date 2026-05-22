package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CtxKeyShareTraceID 是写入 gin.Context 的分享 trace_id key。
const CtxKeyShareTraceID = "share_trace_id"

// ShareTrace 中间件：从 query / cookie / header 读取 share trace_id（参数名 `st`），
// 注入 gin.Context。由 distribution 模块在订单创建/分享点击时消费。
//
// 优先级：query.st > cookie.st > header X-Share-Trace。
// 命中后写 cookie（90 天 HttpOnly），便于跨页面延续。
func ShareTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := strings.TrimSpace(c.Query("st"))
		if st == "" {
			if v, err := c.Cookie("st"); err == nil {
				st = strings.TrimSpace(v)
			}
		}
		if st == "" {
			st = strings.TrimSpace(c.GetHeader("X-Share-Trace"))
		}

		if st != "" && len(st) <= 64 {
			c.Set(CtxKeyShareTraceID, st)
			// 续写 cookie（90 天）
			c.SetCookie("st", st, 90*86400, "/", "", false, true)
		}
		c.Next()
	}
}

// GetShareTraceID 从 gin.Context 读取 share trace_id（可能为空字符串）。
func GetShareTraceID(c *gin.Context) string {
	v, ok := c.Get(CtxKeyShareTraceID)
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
