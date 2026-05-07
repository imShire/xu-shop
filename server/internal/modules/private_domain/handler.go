package private_domain

import (
	"io"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// Handler 私域模块 HTTP 处理器。
type Handler struct{ svc *Service }

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ---- 渠道码 ----

func (h *Handler) AdminListChannelCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.ListChannelCodes(c.Request.Context(), page, size)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total})
}

func (h *Handler) AdminCreateChannelCode(c *gin.Context) {
	var req CreateChannelCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	cc, err := h.svc.CreateChannelCode(c.Request.Context(), adminID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, cc)
}

func (h *Handler) AdminUpdateChannelCode(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateChannelCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateChannelCode(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) AdminDeleteChannelCode(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.DeleteChannelCode(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) AdminGetChannelCodeStats(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	stats, err := h.svc.GetChannelCodeStats(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, stats)
}

// ---- 标签 ----

func (h *Handler) AdminListTags(c *gin.Context) {
	list, err := h.svc.ListTags(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, list)
}

func (h *Handler) AdminCreateTag(c *gin.Context) {
	var req CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.CreateTag(c.Request.Context(), req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) AdminUpdateTag(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.UpdateTag(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) AdminDeleteTag(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.DeleteTag(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 用户标签 ----

func (h *Handler) AdminGetUserTags(c *gin.Context) {
	uid := mustParamID(c, "id")
	if uid == 0 {
		return
	}
	tags, err := h.svc.ListUserTags(c.Request.Context(), uid)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, tags)
}

func (h *Handler) AdminAddUserTag(c *gin.Context) {
	uid := mustParamID(c, "id")
	if uid == 0 {
		return
	}
	var req struct {
		TagID  types.Int64Str `json:"tag_id" binding:"required"`
		Source string         `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.AddUserTag(c.Request.Context(), uid, req.TagID.Int64(), req.Source); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) AdminRemoveUserTag(c *gin.Context) {
	uid := mustParamID(c, "id")
	tagID := mustParamID(c, "tag_id")
	if uid == 0 || tagID == 0 {
		return
	}
	if err := h.svc.RemoveUserTag(c.Request.Context(), uid, tagID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 分享 ----

func (h *Handler) RecordShareVisit(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		ShareUserID types.Int64Str `json:"share_user_id" binding:"required"`
		ProductID   types.Int64Str `json:"product_id"`
		Channel     string         `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.RecordShareVisit(c.Request.Context(), req.ShareUserID.Int64(), userID, req.ProductID.Int64(), req.Channel); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

func (h *Handler) GeneratePoster(c *gin.Context) {
	userID := c.GetInt64("user_id")
	productID := mustParamID(c, "id")
	if productID == 0 {
		return
	}
	posterURL, err := h.svc.GeneratePoster(c.Request.Context(), userID, productID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"url": posterURL})
}

func (h *Handler) ResolveShareScene(c *gin.Context) {
	scene, _ := url.QueryUnescape(c.Query("scene"))
	if scene == "" {
		srv.Fail(c, errs.ErrParam.WithMsg("scene 不能为空"))
		return
	}
	resp, err := h.svc.ResolveShareScene(c.Request.Context(), scene)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, resp)
}

// HandleQYWXCallback 企微回调。
func (h *Handler) HandleQYWXCallback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		srv.Fail(c, errs.ErrParam)
		return
	}
	params := map[string]string{
		"msg_signature": c.Query("msg_signature"),
		"timestamp":     c.Query("timestamp"),
		"nonce":         c.Query("nonce"),
	}
	if err := h.svc.HandleQYWXCallback(c.Request.Context(), body, params); err != nil {
		failWith(c, err)
		return
	}
	c.String(200, "success")
}

// ---- helper ----

func mustParamID(c *gin.Context, key string) int64 {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 ID"))
		return 0
	}
	return id
}

func failWith(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		srv.Fail(c, appErr)
		return
	}
	srv.Fail(c, errs.ErrInternal)
}
