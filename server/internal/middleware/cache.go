package middleware

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PublicCache 为只读接口写 Cache-Control 响应头。
// maxAge: max-age 秒；swr: stale-while-revalidate 秒（0 则不加）。
func PublicCache(maxAge, swr int) gin.HandlerFunc {
	var val string
	if swr > 0 {
		val = fmt.Sprintf("public, max-age=%d, stale-while-revalidate=%d", maxAge, swr)
	} else {
		val = fmt.Sprintf("public, max-age=%d", maxAge)
	}
	return func(c *gin.Context) {
		// 在 c.Next() 之前设置，确保 ETagMiddleware flush 时能一并写出头部。
		c.Header("Cache-Control", val)
		c.Next()
	}
}

// captureWriter 拦截响应体用于 ETag 计算；实际写出延迟到 ETagMiddleware flush 阶段。
type captureWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *captureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func (w *captureWriter) WriteHeader(code int) {
	w.status = code
}

// WriteHeaderNow 延迟到 flush，此处空操作。
func (w *captureWriter) WriteHeaderNow() {}

// Written 始终返回 false，阻止 gin 提前 flush。
func (w *captureWriter) Written() bool { return false }

func (w *captureWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *captureWriter) Size() int { return w.body.Len() }

// ETagMiddleware 对 GET 响应体计算弱 ETag（SHA1 前 8 字节）；
// 若请求携带匹配的 If-None-Match，则返回 304，不写响应体。
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		cw := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = cw

		c.Next()

		body := cw.body.Bytes()
		if len(body) == 0 {
			cw.ResponseWriter.WriteHeader(cw.Status())
			return
		}

		// 计算 SHA1 前 8 字节（16 hex chars）构成弱 ETag。
		h := sha1.Sum(body)
		etag := `W/"` + hex.EncodeToString(h[:8]) + `"`
		cw.ResponseWriter.Header().Set("ETag", etag)

		if match := c.GetHeader("If-None-Match"); match == etag {
			cw.ResponseWriter.WriteHeader(http.StatusNotModified)
			return
		}

		cw.ResponseWriter.WriteHeader(cw.Status())
		_, _ = cw.ResponseWriter.Write(body)
	}
}
