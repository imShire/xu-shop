package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/modules/admin/audit"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/sensitive"
)

// auditAdminPathPrefix 审计仅覆盖该前缀下的写操作。
const auditAdminPathPrefix = "/api/v1/admin/"

const (
	// auditMaxBodyBytes 请求体最大记录字节数（64KB），超出截断。
	auditMaxBodyBytes = 64 * 1024
	// auditMaxRespExcerpt 响应体记录字节数上限（512B）。
	auditMaxRespExcerpt = 512
	// auditAsyncTimeout 异步写库超时。
	auditAsyncTimeout = 5 * time.Second

	// CtxKeyAuditAction handler 可通过 c.Set 注入自定义 action（如 "product.update"）。
	CtxKeyAuditAction     = "audit.action"
	CtxKeyAuditTargetType = "audit.target_type"
	CtxKeyAuditTargetID   = "audit.target_id"
	// CtxKeyAuditAdminID 登录类 handler 可在 c.Next 之后注入操作人 id。
	CtxKeyAuditAdminID = "audit.admin_id"
)

// auditRespCapture 捕获响应状态与前 N 字节。
type auditRespCapture struct {
	gin.ResponseWriter
	excerpt *bytes.Buffer
	limit   int
}

func (w *auditRespCapture) Write(b []byte) (int, error) {
	if remain := w.limit - w.excerpt.Len(); remain > 0 {
		if remain >= len(b) {
			w.excerpt.Write(b)
		} else {
			w.excerpt.Write(b[:remain])
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *auditRespCapture) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// AuditLog 后台审计日志中间件。
//
//   - 仅对方法 POST/PUT/PATCH/DELETE 生效
//   - 仅对 URL 路径中包含 "/admin/" 的请求生效（含登录 / 登出 / 内部接口）
//   - 异步写库，repo 报错只打日志不阻断主流程
//   - 敏感字段统一通过 audit.Sanitize 脱敏
//
// 中间件应挂载在 Logging 之后（依赖 request_id 与 trace span 已就绪），
// 可挂在 AdminAuth 之前；admin_id 优先级：
//
//	c.Get("audit.admin_id") > c.Get("admin_id") > 0
func AuditLog(repo audit.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if !isAuditedMethod(method) || !isAuditedPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		// 读取请求体（截断到 auditMaxBodyBytes），并 NopCloser 还原。
		var bodyRaw []byte
		var bodyTruncated bool
		if c.Request.Body != nil {
			bodyRaw, _ = io.ReadAll(io.LimitReader(c.Request.Body, auditMaxBodyBytes+1))
			if len(bodyRaw) > auditMaxBodyBytes {
				bodyRaw = bodyRaw[:auditMaxBodyBytes]
				bodyTruncated = true
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyRaw))
		}

		respCap := &auditRespCapture{
			ResponseWriter: c.Writer,
			excerpt:        &bytes.Buffer{},
			limit:          auditMaxRespExcerpt,
		}
		c.Writer = respCap

		c.Next()

		// ---- 组装审计记录 ----
		adminID := resolveAuditAdminID(c)
		action := resolveAuditAction(c)

		entry := &audit.AuditLog{
			AdminID:        adminID,
			Action:         action,
			Method:         method,
			Path:           c.Request.URL.Path,
			ResponseStatus: c.Writer.Status(),
			DurationMs:     int(time.Since(start).Milliseconds()),
		}

		if tt, ok := c.Get(CtxKeyAuditTargetType); ok {
			if s, ok := tt.(string); ok && s != "" {
				entry.TargetType = &s
			}
		}
		if ti, ok := c.Get(CtxKeyAuditTargetID); ok {
			if s := anyToStr(ti); s != "" {
				entry.TargetID = &s
			}
		}
		if q := c.Request.URL.RawQuery; q != "" {
			sq := sensitive.Sanitize([]byte(q))
			s := string(sq)
			entry.Query = &s
		}
		if len(bodyRaw) > 0 {
			masked := sensitive.Sanitize(bodyRaw)
			// 截断标记：JSON 体包裹成 {"_truncated":true,"raw":"..."}；非 JSON 仍 /*truncated*/
			if bodyTruncated {
				if json.Valid(masked) {
					wrapped, wrapErr := json.Marshal(map[string]any{
						"_truncated": true,
						"raw":        json.RawMessage(masked),
					})
					if wrapErr == nil {
						masked = wrapped
					} else {
						masked = append(masked, []byte(`/*truncated*/`)...)
					}
				} else {
					masked = append(masked, []byte(`/*truncated*/`)...)
				}
			}
			// jsonb 列要求合法 JSON：用 json.Valid 严格校验，不合法则包成 JSON 字符串
			if !json.Valid(masked) {
				if quoted, err := jsonQuote(string(masked)); err == nil {
					masked = quoted
				} else {
					masked = []byte(`""`)
				}
			}
			entry.RequestBody = masked
		}
		if respCap.excerpt.Len() > 0 {
			masked := sensitive.Sanitize(respCap.excerpt.Bytes())
			s := string(masked)
			entry.ResponseExcerpt = &s
		}
		if ip := c.ClientIP(); ip != "" {
			if net.ParseIP(ip) != nil {
				entry.ClientIP = &ip
			}
		}
		if ua := c.Request.UserAgent(); ua != "" {
			if len(ua) > 255 {
				ua = ua[:255]
			}
			entry.UserAgent = &ua
		}
		if sc := trace.SpanContextFromContext(c.Request.Context()); sc.IsValid() {
			tid := sc.TraceID().String()
			entry.TraceID = &tid
		}

		// 异步写库，避免阻塞响应。透传当前 span context，便于跨 goroutine trace 串联。
		spanCtx := trace.SpanContextFromContext(c.Request.Context())
		detached := trace.ContextWithSpanContext(context.Background(), spanCtx)
		go func(e *audit.AuditLog) {
			ctx, cancel := context.WithTimeout(detached, auditAsyncTimeout)
			defer cancel()
			if err := repo.Insert(ctx, e); err != nil {
				pkglogger.L().Warn("audit log insert failed",
					zap.Error(err),
					zap.String("action", e.Action),
					zap.Int64("admin_id", e.AdminID),
				)
			}
		}(entry)
	}
}

func isAuditedMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// isAuditedPath 仅对真正的 admin 写接口审计：要求路径以 /api/v1/admin/ 开头，
// 避免误匹配业务字段含 "admin" 子串的非 admin 路由。
func isAuditedPath(p string) bool {
	return strings.HasPrefix(p, auditAdminPathPrefix)
}

func resolveAuditAdminID(c *gin.Context) int64 {
	if v, ok := c.Get(CtxKeyAuditAdminID); ok {
		if id, ok := toInt64(v); ok {
			return id
		}
	}
	if v, ok := c.Get(ctxKeyAdminID); ok {
		if id, ok := toInt64(v); ok {
			return id
		}
	}
	return 0
}

func resolveAuditAction(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyAuditAction); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	// 兜底：method:path（如 POST:/api/v1/admin/banners）
	a := c.Request.Method + ":" + c.Request.URL.Path
	if len(a) > 64 {
		a = a[:64]
	}
	return a
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case uint64:
		return int64(x), true
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func anyToStr(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	default:
		return ""
	}
}

// jsonQuote 把任意字节串包成合法 JSON 字符串。
func jsonQuote(s string) ([]byte, error) {
	return json.Marshal(s)
}
