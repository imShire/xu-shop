package notification

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 通知模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// AdminListNotifications 通知任务列表。
func (h *Handler) AdminListNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filter := TaskFilter{
		Status:       c.Query("status"),
		TemplateCode: c.Query("template_code"),
		Page:         page,
		Size:         size,
	}
	list, total, err := h.svc.ListNotifications(c.Request.Context(), filter)
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

// AdminListTemplates 通知模板列表。
func (h *Handler) AdminListTemplates(c *gin.Context) {
	list, err := h.svc.ListTemplates(c.Request.Context())
	if err != nil {
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, list)
}

// AdminUpdateTemplate 更新通知模板。
func (h *Handler) AdminUpdateTemplate(c *gin.Context) {
	code := c.Param("code")
	var req UpdateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateTemplate(c.Request.Context(), code, req); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			srv.Fail(c, appErr)
			return
		}
		srv.Fail(c, errs.ErrInternal)
		return
	}
	srv.OK(c, nil)
}

// AdminTestSend 测试发送通知。
func (h *Handler) AdminTestSend(c *gin.Context) {
	code := c.Param("code")
	var req struct {
		Openid string `json:"openid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.TestSend(c.Request.Context(), code, req.Openid); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			srv.Fail(c, appErr)
			return
		}
		srv.Fail(c, errs.ErrInternal.WithMsg(err.Error()))
		return
	}
	srv.OK(c, nil)
}
