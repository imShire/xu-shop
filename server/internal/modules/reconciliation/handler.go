package reconciliation

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 对账差异 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListDiffs GET /admin/reconciliation/diff
func (h *Handler) ListDiffs(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	f := Filter{
		Job:      req.Job,
		Status:   req.Status,
		Severity: req.Severity,
		Page:     req.Page,
		Size:     req.Size,
	}
	if req.BizDate != "" {
		d, err := time.Parse("2006-01-02", req.BizDate)
		if err != nil {
			srv.Fail(c, errs.ErrParam.WithMsg("biz_date 格式应为 YYYY-MM-DD"))
			return
		}
		f.BizDate = &d
	}
	list, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		failWith(c, err)
		return
	}
	items := make([]DiffResp, len(list))
	for i := range list {
		items[i] = ToDiffResp(&list[i])
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 {
		f.Size = 20
	}
	srv.OK(c, ListResp{Items: items, Total: total, Page: f.Page, Size: f.Size})
}

// AckDiff POST /admin/reconciliation/diff/:id/ack
func (h *Handler) AckDiff(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("id 非法"))
		return
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.Acknowledge(c.Request.Context(), id, adminID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ResolveDiff POST /admin/reconciliation/diff/:id/resolve
func (h *Handler) ResolveDiff(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("id 非法"))
		return
	}
	var req ResolveReq
	// body 可空
	_ = c.ShouldBindJSON(&req)
	var note *string
	if req.Note != "" {
		note = &req.Note
	}
	adminID := c.GetInt64("admin_id")
	if err := h.svc.Resolve(c.Request.Context(), id, adminID, note); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func failWith(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		srv.Fail(c, appErr)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}
