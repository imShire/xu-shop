package product

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	srv "github.com/xushop/xu-shop/internal/server"
)

// Handler 商品模块 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ---- C 端 ----

// ListCategories 获取分类树。
func (h *Handler) ListCategories(c *gin.Context) {
	tree, err := h.svc.GetCategories(c.Request.Context())
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, tree)
}

// ListProducts C 端商品列表。
func (h *Handler) ListProducts(c *gin.Context) {
	var req ProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	list, total, err := h.svc.ListProducts(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// GetProduct C 端商品详情。
func (h *Handler) GetProduct(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	resp, err := h.svc.GetProduct(c.Request.Context(), id, userID)
	if err != nil {
		failWith(c, err)
		return
	}
	// 异步记录浏览
	if userID > 0 {
		h.svc.RecordView(c.Request.Context(), userID, id)
	}
	srv.OK(c, resp)
}

// AddFavorite 添加收藏。
func (h *Handler) AddFavorite(c *gin.Context) {
	productID := mustParamID(c, "product_id")
	if productID == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.AddFavorite(c.Request.Context(), userID, productID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// RemoveFavorite 取消收藏。
func (h *Handler) RemoveFavorite(c *gin.Context) {
	productID := mustParamID(c, "product_id")
	if productID == 0 {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.RemoveFavorite(c.Request.Context(), userID, productID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ListFavorites 收藏列表。
func (h *Handler) ListFavorites(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.ListFavorites(c.Request.Context(), userID, page, size)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": size})
}

// GetViewHistory 浏览历史。
func (h *Handler) GetViewHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.GetViewHistory(c.Request.Context(), userID, page, size)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": size})
}

// ClearViewHistory 清空浏览历史。
func (h *Handler) ClearViewHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if err := h.svc.ClearViewHistory(c.Request.Context(), userID); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 后台分类管理 ----

// AdminListCategories 后台分类列表（平铺）。
func (h *Handler) AdminListCategories(c *gin.Context) {
	list, err := h.svc.categoryRepo.FindAll(c.Request.Context())
	if err != nil {
		failWith(c, errs.ErrInternal)
		return
	}
	resp := make([]CategoryResp, len(list))
	for i, cat := range list {
		resp[i] = toCategoryResp(&cat)
	}
	srv.OK(c, resp)
}

// AdminCreateCategory 后台创建分类。
func (h *Handler) AdminCreateCategory(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	cat, err := h.svc.AdminCreateCategory(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, toCategoryResp(cat))
}

// AdminUpdateCategory 后台更新分类。
func (h *Handler) AdminUpdateCategory(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.AdminUpdateCategory(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminDeleteCategory 后台删除分类。
func (h *Handler) AdminDeleteCategory(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminDeleteCategory(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// ---- 后台商品管理 ----

// AdminListProducts 后台商品列表。
func (h *Handler) AdminListProducts(c *gin.Context) {
	var req ProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	list, total, err := h.svc.AdminListProducts(c.Request.Context(), req)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, gin.H{"list": list, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// AdminCreateProduct 后台创建商品。
func (h *Handler) AdminCreateProduct(c *gin.Context) {
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	adminID := c.GetInt64("admin_id")
	p, err := h.svc.AdminCreateProduct(c.Request.Context(), adminID, req)
	if err != nil {
		failWith(c, err)
		return
	}
	// 返回完整详情（含规格和 SKU）
	detail, err := h.svc.AdminGetProduct(c.Request.Context(), p.ID)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, detail)
}

// AdminGetProduct 后台获取商品详情。
func (h *Handler) AdminGetProduct(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	detail, err := h.svc.AdminGetProduct(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, detail)
}

// AdminUpdateProduct 后台更新商品。
func (h *Handler) AdminUpdateProduct(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	var req UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	if err := h.svc.AdminUpdateProduct(c.Request.Context(), id, req); err != nil {
		failWith(c, err)
		return
	}
	// 返回完整详情
	detail, err := h.svc.AdminGetProduct(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, detail)
}

// AdminDeleteProduct 后台删除商品。
func (h *Handler) AdminDeleteProduct(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminDeleteProduct(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminCopyProduct 复制商品。
func (h *Handler) AdminCopyProduct(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	p, err := h.svc.AdminCopyProduct(c.Request.Context(), id)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, toProductResp(p))
}

// AdminOnSale 上架商品。
func (h *Handler) AdminOnSale(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminOnSale(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminOffSale 下架商品。
func (h *Handler) AdminOffSale(c *gin.Context) {
	id := mustParamID(c, "id")
	if id == 0 {
		return
	}
	if err := h.svc.AdminOffSale(c.Request.Context(), id); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminBatchStatus 批量修改状态。
func (h *Handler) AdminBatchStatus(c *gin.Context) {
	var req BatchStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	ids := make([]int64, len(req.IDs))
	for i, v := range req.IDs {
		ids[i] = v.Int64()
	}
	if err := h.svc.AdminBatchStatus(c.Request.Context(), ids, req.Status); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminBatchPrice 批量修改价格。
func (h *Handler) AdminBatchPrice(c *gin.Context) {
	var req BatchPriceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		srv.FailParam(c, err)
		return
	}
	ids := make([]int64, len(req.IDs))
	for i, v := range req.IDs {
		ids[i] = v.Int64()
	}
	if err := h.svc.AdminBatchPrice(c.Request.Context(), ids, req.PriceCents); err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, nil)
}

// AdminUploadImage 上传商品图片。
func (h *Handler) AdminUploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		srv.Fail(c, errs.ErrParam.WithMsg("请上传文件"))
		return
	}
	defer file.Close()

	adminID := c.GetInt64("admin_id")
	url, err := h.svc.UploadImage(c.Request.Context(), adminID, file, header.Filename)
	if err != nil {
		failWith(c, err)
		return
	}
	srv.OK(c, UploadImageResp{URL: url})
}

// ---- 工具函数 ----

func mustParamID(c *gin.Context, key string) int64 {
	s := c.Param(key)
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		srv.Fail(c, errs.ErrParam.WithMsg("无效的 "+key))
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
