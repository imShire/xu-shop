package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 系统设置 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GetUploadSettings 获取上传设置。
func (h *Handler) GetUploadSettings(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	resp, err := h.svc.GetUploadSettings(c.Request.Context(), adminID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// UpdateUploadSettings 更新上传设置。
func (h *Handler) UpdateUploadSettings(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	var req UpdateUploadSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateUploadSettings(c.Request.Context(), adminID, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// TestUploadSettings 测试上传设置。
func (h *Handler) TestUploadSettings(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	var req UpdateUploadSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.TestUploadSettings(c.Request.Context(), adminID, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ProbeUploadSettings 实际上传文件验证当前配置。
func (h *Handler) ProbeUploadSettings(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg("请上传测试文件"))
		return
	}
	defer file.Close()

	url, err := h.svc.ProbeUploadSettings(c.Request.Context(), adminID, file, header.Filename)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"url": url})
}

// ServeLocalFile 对外暴露本地上传文件。
func (h *Handler) ServeLocalFile(c *gin.Context) {
	fullPath, err := h.svc.upload.ResolveLocalFile(c.Request.Context(), c.Param("filepath"))
	if err != nil {
		if appErr, ok := err.(*errs.AppError); ok && appErr.Code == errs.ErrNotFound.Code {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.File(fullPath)
}

// GetSettings 获取指定分组的系统设置。
func (h *Handler) GetSettings(c *gin.Context) {
	group := c.Param("group")
	data, err := h.svc.GetSettings(c.Request.Context(), group)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, data)
}

// UpdateSettings 批量更新指定分组的系统设置。
func (h *Handler) UpdateSettings(c *gin.Context) {
	group := c.Param("group")
	adminID := c.GetInt64("admin_id")
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateSettings(c.Request.Context(), group, req, adminID); err != nil {
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
