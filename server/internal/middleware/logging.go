package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
)

// maxLogBodyBytes 日志记录请求/响应体的最大字节数（64KB）。
const maxLogBodyBytes = 64 * 1024

// bodyLogWriter 包装 gin.ResponseWriter，将响应体同时写入 buf 用于日志记录。
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging 注入 request_id，记录请求日志（含出入参）。
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 优先取 X-Request-Id 头，否则生成 UUID
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-Id", requestID)

		// 将带 request_id 的 logger 注入 context
		l := pkglogger.L().With(zap.String("request_id", requestID))
		c.Request = c.Request.WithContext(pkglogger.WithContext(c.Request.Context(), l))

		// 读取请求体（有 Content-Length 且不超限时才记录）
		var reqBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= maxLogBodyBytes {
			reqBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, maxLogBodyBytes))
			// 将读过的 body 放回，保证 handler 仍可读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 包装 ResponseWriter 捕获响应体
		blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		userID, _ := c.Get("user_id")

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Any("user_id", userID),
		}
		if len(reqBody) > 0 {
			fields = append(fields, zap.ByteString("req_body", reqBody))
		}
		if blw.body.Len() > 0 && blw.body.Len() <= maxLogBodyBytes {
			fields = append(fields, zap.String("resp_body", blw.body.String()))
		}

		l.Info("http", fields...)
	}
}
