package clog

import (
	"net"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkgserver "github.com/xushop/xu-shop/internal/server"
)

// Handler 前端日志上报 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Submit 单条上报。
func (h *Handler) Submit(c *gin.Context) {
	var req SubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgserver.FailParam(c, err)
		return
	}
	if err := h.svc.SubmitBatch(c.Request.Context(), []SubmitReq{req}, h.metaFromCtx(c)); err != nil {
		pkgserver.Fail(c, asAppError(err))
		return
	}
	pkgserver.OK(c, gin.H{"accepted": 1})
}

// SubmitBatch 批量上报。
func (h *Handler) SubmitBatch(c *gin.Context) {
	var req BatchSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgserver.FailParam(c, err)
		return
	}
	if err := h.svc.SubmitBatch(c.Request.Context(), req.Logs, h.metaFromCtx(c)); err != nil {
		pkgserver.Fail(c, asAppError(err))
		return
	}
	pkgserver.OK(c, gin.H{"accepted": len(req.Logs)})
}

// metaFromCtx 从 gin.Context 提取服务端可信元信息。
//
//   - user_id / admin_id：来自 UserOptionalAuth / AdminAuth 注入；未登录则为 nil
//   - trace_id：来自 OTel span（A1 链路）
//   - client_ip：c.ClientIP()（已经过 SetTrustedProxies）
func (h *Handler) metaFromCtx(c *gin.Context) Meta {
	m := Meta{
		ClientIP: c.ClientIP(),
	}
	if net.ParseIP(m.ClientIP) == nil {
		m.ClientIP = ""
	}
	if sc := trace.SpanContextFromContext(c.Request.Context()); sc.IsValid() {
		m.TraceID = sc.TraceID().String()
	}
	if v, ok := c.Get("user_id"); ok {
		if id, ok := toInt64(v); ok {
			m.UserID = &id
		}
	}
	if v, ok := c.Get("admin_id"); ok {
		if id, ok := toInt64(v); ok {
			m.AdminID = &id
		}
	}
	return m
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	}
	return 0, false
}

// asAppError 把 service 错误归一为 *errs.AppError。
func asAppError(err error) *errs.AppError {
	if ae, ok := err.(*errs.AppError); ok {
		return ae
	}
	return errs.ErrInternal
}
