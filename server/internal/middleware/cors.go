package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 配置跨域策略，allowOrigins 为允许的前端域名列表。
func CORS(allowOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		originSet[strings.TrimRight(o, "/")] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := originSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-Id,X-CSRF-Token,Idempotency-Key")
			c.Header("Access-Control-Expose-Headers", "X-Request-Id")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
