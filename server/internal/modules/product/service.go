package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/microcosm-cc/bluemonday"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

const (
	categoryCacheKey = "product:categories"
	categoryCacheTTL = 5 * time.Minute
	maxViewHistory   = 50
)

// ImageUploader 图片上传接口（方便 mock）。
type ImageUploader interface {
	UploadProductImage(ctx context.Context, r io.Reader, filename string) (string, error)
	AllowedImageURLPrefix(ctx context.Context) (string, error)
}

// Service 商品服务。
type Service struct {
	productRepo  ProductRepo
	categoryRepo CategoryRepo
	skuRepo      SKURepo
	favRepo      FavoriteRepo
	viewRepo     ViewHistoryRepo
	rdb          *redis.Client
	ossClient    ImageUploader
	asynqClient  *asynq.Client
	policy       *bluemonday.Policy
}

// NewService 构造 Service。
func NewService(
	productRepo ProductRepo,
	categoryRepo CategoryRepo,
	skuRepo SKURepo,
	favRepo FavoriteRepo,
	viewRepo ViewHistoryRepo,
	rdb *redis.Client,
	ossClient ImageUploader,
	asynqClient *asynq.Client,
) *Service {
	policy := bluemonday.UGCPolicy()
	// 只允许安全的 img 属性，禁止 SVG
	policy.AllowElements("img")
	policy.AllowAttrs("src", "alt", "width", "height").OnElements("img")

	return &Service{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		skuRepo:      skuRepo,
		favRepo:      favRepo,
		viewRepo:     viewRepo,
		rdb:          rdb,
		ossClient:    ossClient,
		asynqClient:  asynqClient,
		policy:       policy,
	}
}

// ---- C 端服务 ----

// GetCategories 获取全部分类并构建树形结构（Redis 缓存 5min）。
func (s *Service) GetCategories(ctx context.Context) ([]CategoryTreeNode, error) {
	// 优先读缓存
	cached, err := s.rdb.Get(ctx, categoryCacheKey).Bytes()
	if err == nil {
		var tree []CategoryTreeNode
		if jsonErr := json.Unmarshal(cached, &tree); jsonErr == nil {
			return tree, nil
		}
	}

	list, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}

	tree := buildCategoryTree(list, 0)

	if data, jsonErr := json.Marshal(tree); jsonErr == nil {
		_ = s.rdb.SetEx(ctx, categoryCacheKey, data, categoryCacheTTL)
	}
	return tree, nil
}

// buildCategoryTree 平铺转树形（递归）。
func buildCategoryTree(list []Category, parentID int64) []CategoryTreeNode {
	var nodes []CategoryTreeNode
	for _, c := range list {
		if c.ParentID == parentID && c.Status == "enabled" {
			node := toCategoryResp(&c)
			node.Children = buildCategoryTree(list, c.ID)
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// ListProducts C 端商品列表（仅 onsale）。
func (s *Service) ListProducts(ctx context.Context, req ProductListReq) ([]ProductResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	filter := ProductFilter{
		CategoryID: req.CategoryID.Int64(),
		Status:     "onsale",
		Keyword:    req.Keyword,
		Sort:       req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
		InStock:    req.InStock,
	}
	products, total, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]ProductResp, len(products))
	for i, p := range products {
		resp[i] = toProductResp(&p)
	}
	return resp, total, nil
}

// GetProduct C 端商品详情（含 is_favorite）。
func (s *Service) GetProduct(ctx context.Context, id int64, userIDs ...int64) (*ProductDetailResp, error) {
	detail, err := s.productRepo.FindWithSpecs(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}

	resp := toProductDetailResp(detail)

	if len(userIDs) > 0 && userIDs[0] > 0 {
		exists, _ := s.favRepo.Exists(ctx, userIDs[0], id)
		resp.IsFavorite = exists
	}

	return &resp, nil
}

// toProductDetailResp 详情转 DTO。
func toProductDetailResp(d *ProductDetail) ProductDetailResp {
	base := toProductResp(&d.Product)
	resp := ProductDetailResp{
		ProductResp: base,
		Images:      json.RawMessage(d.Images),
		VideoURL:    d.VideoURL,
		DetailHTML:  d.DetailHTML,
		DetailNodes: json.RawMessage(d.DetailNodes),
	}

	// 构建规格树
	valuesBySpec := make(map[int64][]SpecValueResp)
	for _, v := range d.Values {
		valuesBySpec[v.SpecID] = append(valuesBySpec[v.SpecID], SpecValueResp{
			ID:    types.Int64Str(v.ID),
			Value: v.Value,
			Sort:  v.Sort,
		})
	}
	resp.Specs = make([]SpecResp, len(d.Specs))
	for i, spec := range d.Specs {
		resp.Specs[i] = SpecResp{
			ID:     types.Int64Str(spec.ID),
			Name:   spec.Name,
			Sort:   spec.Sort,
			Values: valuesBySpec[spec.ID],
		}
	}

	resp.SKUs = make([]SKUResp, len(d.SKUs))
	for i := range d.SKUs {
		resp.SKUs[i] = toSKUResp(&d.SKUs[i])
	}
	return resp
}

// AddFavorite 添加收藏。
func (s *Service) AddFavorite(ctx context.Context, userID, productID int64) error {
	if _, err := s.productRepo.FindByID(ctx, productID); err != nil {
		return errs.ErrNotFound
	}
	return s.favRepo.Add(ctx, userID, productID)
}

// RemoveFavorite 取消收藏。
func (s *Service) RemoveFavorite(ctx context.Context, userID, productID int64) error {
	return s.favRepo.Remove(ctx, userID, productID)
}

// ListFavorites 用户收藏列表。
func (s *Service) ListFavorites(ctx context.Context, userID int64, page, size int) ([]FavoriteListItemResp, int64, error) {
	items, total, err := s.favRepo.List(ctx, userID, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]FavoriteListItemResp, len(items))
	for i, item := range items {
		resp[i] = FavoriteListItemResp{
			ID:               types.Int64Str(item.ProductID),
			ProductID:        types.Int64Str(item.ProductID),
			Title:            item.Title,
			Image:            item.Image,
			PriceCents:       item.PriceCents,
			MarketPriceCents: item.MarketPriceCents,
			CreatedAt:        item.CreatedAt,
		}
	}
	return resp, total, nil
}

// GetViewHistory 用户浏览历史。
func (s *Service) GetViewHistory(ctx context.Context, userID int64, page, size int) ([]ViewHistoryItemResp, int64, error) {
	items, total, err := s.viewRepo.List(ctx, userID, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]ViewHistoryItemResp, len(items))
	for i, item := range items {
		resp[i] = ViewHistoryItemResp{
			ID:               types.Int64Str(item.ProductID),
			ProductID:        types.Int64Str(item.ProductID),
			Title:            item.Title,
			Image:            item.Image,
			PriceCents:       item.PriceCents,
			MarketPriceCents: item.MarketPriceCents,
			ViewedAt:         item.ViewedAt,
		}
	}
	return resp, total, nil
}

// ClearViewHistory 清空浏览历史。
func (s *Service) ClearViewHistory(ctx context.Context, userID int64) error {
	return s.viewRepo.Clear(ctx, userID)
}

// RecordView 异步记录浏览（goroutine，保留最近 50 条）。
func (s *Service) RecordView(ctx context.Context, userID, productID int64) {
	go func() {
		bgCtx := context.Background()
		if err := s.viewRepo.Upsert(bgCtx, userID, productID); err != nil {
			logger.L().Warn("record view upsert", zap.Error(err))
			return
		}
		count, err := s.viewRepo.CountByUser(bgCtx, userID)
		if err != nil {
			return
		}
		if count > maxViewHistory {
			_ = s.viewRepo.DeleteOldest(bgCtx, userID, maxViewHistory)
		}
	}()
}

// ---- 后台服务 ----

// sanitizeDetailHTML 净化 detail_html：bluemonday + img src 白名单校验。
func (s *Service) sanitizeDetailHTML(ctx context.Context, html string) (string, error) {
	if html == "" {
		return "", nil
	}
	// 禁止 SVG
	lowerHTML := strings.ToLower(html)
	if strings.Contains(lowerHTML, "<svg") || strings.Contains(lowerHTML, "data:image/svg") {
		return "", errs.ErrParam.WithMsg("detail_html 禁止包含 SVG")
	}

	cleaned := s.policy.Sanitize(html)

	// 校验 img src 白名单
	imgSrcRe := regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
	matches := imgSrcRe.FindAllStringSubmatch(cleaned, -1)
	allowedPrefix, err := s.ossClient.AllowedImageURLPrefix(ctx)
	if err != nil {
		return "", errs.ErrInternal
	}
	allowedPattern := regexp.MustCompile(`(?i)^https://`)
	if allowedPrefix != "" {
		allowedPattern = regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(strings.TrimRight(allowedPrefix, "/")) + `/`)
	}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		src := m[1]
		if !allowedPattern.MatchString(src) {
			return "", errs.ErrParam.WithMsg(fmt.Sprintf("img src 不在允许的域名: %s", src))
		}
	}
	return cleaned, nil
}

// AdminCreateProduct 后台创建商品。
func (s *Service) AdminCreateProduct(ctx context.Context, adminID int64, req CreateProductReq) (*Product, error) {
	// 校验规格数量
	if len(req.Specs) > 3 {
		return nil, errs.ErrParam.WithMsg("spec_limit_exceeded: 规格最多 3 个")
	}

	cleanHTML, err := s.sanitizeDetailHTML(ctx, req.DetailHTML)
	if err != nil {
		return nil, err
	}

	images := req.Images
	if len(images) == 0 {
		images = json.RawMessage("[]")
	}
	tags := req.Tags
	if len(tags) == 0 {
		tags = json.RawMessage("[]")
	}

	unit := req.Unit
	if unit == "" {
		unit = "件"
	}

	// 虚拟商品强制清空运费模板
	var freightTemplateID *int64
	if req.FreightTemplateID != nil && !req.IsVirtual {
		v := req.FreightTemplateID.Int64()
		freightTemplateID = &v
	}

	p := &Product{
		CategoryID:        req.CategoryID.Int64(),
		Title:             req.Title,
		Subtitle:          req.Subtitle,
		MainImage:         req.MainImage,
		Images:            JSON(images),
		VideoURL:          req.VideoURL,
		DetailHTML:        cleanHTML,
		DetailNodes:       JSON(req.DetailNodes),
		Status:            req.Status,
		Sort:              req.Sort,
		Tags:              JSON(tags),
		Unit:              unit,
		IsVirtual:         req.IsVirtual,
		FreightTemplateID: freightTemplateID,
		VirtualSales:      req.VirtualSales,
	}
	if p.Status == "" {
		p.Status = "draft"
	}

	// 构建 specs 和 values
	var specs []ProductSpec
	var values []ProductSpecValue
	for _, sr := range req.Specs {
		spec := ProductSpec{Name: sr.Name, Sort: sr.Sort}
		specs = append(specs, spec)
		for j, val := range sr.Values {
			values = append(values, ProductSpecValue{
				Value: val,
				Sort:  j,
			})
		}
	}

	// 构建 SKUs
	skus := make([]SKU, len(req.SKUs))
	for i, sr := range req.SKUs {
		skus[i] = SKU{
			Attrs:              JSON(sr.Attrs),
			PriceCents:         sr.PriceCents,
			OriginalPriceCents: sr.OriginalPriceCents,
			Stock:              sr.Stock,
			WeightG:            sr.WeightG,
			SkuCode:            sr.SkuCode,
			Barcode:            sr.Barcode,
			Image:              sr.Image,
			Status:             sr.Status,
			LowStockThreshold:  sr.LowStockThreshold,
		}
		if skus[i].Status == "" {
			skus[i].Status = "active"
		}
	}

	// 规格 ID 在 Create 时由 repo 分配，这里需要先建立 spec→values 映射
	// 简化处理：spec 和 values 在 repo.Create 按顺序分配 ID
	// 需要预先分配 spec ID 才能关联 values
	specIdx := 0
	valIdx := 0
	for _, sr := range req.Specs {
		for range sr.Values {
			// values[valIdx].SpecID 会在 repo 中设置（spec ID 是 repo 分配的）
			// 这里用 specIdx 作为临时占位符
			values[valIdx].SpecID = int64(specIdx) // 临时，repo 会覆盖
			valIdx++
		}
		specIdx++
	}

	// 重新按正确顺序关联（需要预分配 spec ID）
	// 更好的做法：在 repo 层处理
	if err := s.productRepo.Create(ctx, p, specs, values, skus); err != nil {
		logger.Ctx(ctx).Error("admin create product", zap.Error(err))
		return nil, errs.ErrInternal
	}

	// 更新价格区间
	if len(skus) > 0 {
		minPrice, maxPrice := skus[0].PriceCents, skus[0].PriceCents
		for _, sku := range skus {
			if sku.PriceCents < minPrice {
				minPrice = sku.PriceCents
			}
			if sku.PriceCents > maxPrice {
				maxPrice = sku.PriceCents
			}
		}
		_ = s.productRepo.UpdatePriceRange(ctx, p.ID, minPrice, maxPrice)
	}

	_ = s.invalidateCategoryCache(ctx)
	return p, nil
}

// AdminUpdateProduct 后台更新商品。
func (s *Service) AdminUpdateProduct(ctx context.Context, id int64, req UpdateProductReq) error {
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}

	if req.CategoryID != nil {
		p.CategoryID = req.CategoryID.Int64()
	}
	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.Subtitle != nil {
		p.Subtitle = *req.Subtitle
	}
	if req.MainImage != nil {
		p.MainImage = *req.MainImage
	}
	if len(req.Images) > 0 {
		p.Images = JSON(req.Images)
	}
	if req.VideoURL != nil {
		p.VideoURL = *req.VideoURL
	}
	if req.DetailHTML != nil {
		cleanHTML, sanitizeErr := s.sanitizeDetailHTML(ctx, *req.DetailHTML)
		if sanitizeErr != nil {
			return sanitizeErr
		}
		p.DetailHTML = cleanHTML
	}
	if len(req.DetailNodes) > 0 {
		p.DetailNodes = JSON(req.DetailNodes)
	}
	if req.Sort != nil {
		p.Sort = *req.Sort
	}
	if len(req.Tags) > 0 {
		p.Tags = JSON(req.Tags)
	}
	if req.Unit != nil {
		p.Unit = *req.Unit
	}
	if req.IsVirtual != nil {
		p.IsVirtual = *req.IsVirtual
	}
	if req.VirtualSales != nil {
		p.VirtualSales = *req.VirtualSales
	}
	// 虚拟商品强制清空运费模板；非虚拟商品允许更新
	if p.IsVirtual {
		p.FreightTemplateID = nil
	} else if req.FreightTemplateID != nil {
		v := req.FreightTemplateID.Int64()
		p.FreightTemplateID = &v
	}
	if req.Status != nil {
		p.Status = *req.Status
		if *req.Status == "onsale" && p.OnSaleAt == nil {
			now := time.Now()
			p.OnSaleAt = &now
		}
	}

	if err := s.productRepo.Update(ctx, p); err != nil {
		return errs.ErrInternal
	}

	// 若请求携带 SKU，则替换规格和 SKU
	if req.SKUs != nil {
		if req.Specs != nil && len(*req.Specs) > 3 {
			return errs.ErrParam.WithMsg("spec_limit_exceeded: 规格最多 3 个")
		}

		// 构建规格及规格值（在 service 层分配 ID 确保关联正确）
		var specs []ProductSpec
		var values []ProductSpecValue
		if req.Specs != nil {
			for _, sr := range *req.Specs {
				specID := snowflake.NextID()
				specs = append(specs, ProductSpec{
					ID:        specID,
					ProductID: id,
					Name:      sr.Name,
					Sort:      sr.Sort,
				})
				for j, val := range sr.Values {
					values = append(values, ProductSpecValue{
						ID:     snowflake.NextID(),
						SpecID: specID,
						Value:  val,
						Sort:   j,
					})
				}
			}
		}

		// 构建 SKU（保留已有 ID 以便 upsert）
		skus := make([]SKU, len(*req.SKUs))
		for i, sr := range *req.SKUs {
			var skuID int64
			if sr.ID != nil {
				parsed, _ := strconv.ParseInt(*sr.ID, 10, 64)
				skuID = parsed
			}
			if skuID == 0 {
				skuID = snowflake.NextID()
			}
			skus[i] = SKU{
				ID:                 skuID,
				ProductID:          id,
				Attrs:              JSON(sr.Attrs),
				PriceCents:         sr.PriceCents,
				OriginalPriceCents: sr.OriginalPriceCents,
				Stock:              sr.Stock,
				WeightG:            sr.WeightG,
				SkuCode:            sr.SkuCode,
				Barcode:            sr.Barcode,
				Image:              sr.Image,
				Status:             sr.Status,
				LowStockThreshold:  sr.LowStockThreshold,
			}
			if skus[i].Status == "" {
				skus[i].Status = "active"
			}
		}

		if err := s.productRepo.ReplaceSpecsAndSKUs(ctx, id, specs, values, skus); err != nil {
			logger.Ctx(ctx).Error("replace specs and skus", zap.Error(err))
			return errs.ErrInternal
		}

		// 更新价格区间
		if len(skus) > 0 {
			minPrice, maxPrice := skus[0].PriceCents, skus[0].PriceCents
			for _, sku := range skus {
				if sku.PriceCents < minPrice {
					minPrice = sku.PriceCents
				}
				if sku.PriceCents > maxPrice {
					maxPrice = sku.PriceCents
				}
			}
			_ = s.productRepo.UpdatePriceRange(ctx, id, minPrice, maxPrice)
		}
	}

	return nil
}

// AdminDeleteProduct 后台软删除商品。
func (s *Service) AdminDeleteProduct(ctx context.Context, id int64) error {
	if err := s.productRepo.SoftDelete(ctx, id); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// AdminCopyProduct 复制商品（状态重置为 draft）。
func (s *Service) AdminCopyProduct(ctx context.Context, id int64) (*Product, error) {
	p, err := s.productRepo.Copy(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	return p, nil
}

// AdminOnSale 上架商品。
func (s *Service) AdminOnSale(ctx context.Context, id int64) error {
	if err := s.productRepo.UpdateStatus(ctx, id, "onsale"); err != nil {
		return errs.ErrInternal
	}
	_ = s.invalidateCategoryCache(ctx)
	return nil
}

// AdminOffSale 下架商品。
func (s *Service) AdminOffSale(ctx context.Context, id int64) error {
	if err := s.productRepo.UpdateStatus(ctx, id, "offsale"); err != nil {
		return errs.ErrInternal
	}
	_ = s.invalidateCategoryCache(ctx)
	return nil
}

// AdminBatchStatus 批量修改状态。
func (s *Service) AdminBatchStatus(ctx context.Context, ids []int64, status string) error {
	if err := s.productRepo.BatchUpdateStatus(ctx, ids, status); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// AdminBatchPrice 批量修改 SKU 价格。
func (s *Service) AdminBatchPrice(ctx context.Context, ids []int64, priceCents int64) error {
	if err := s.skuRepo.BatchUpdatePrice(ctx, ids, priceCents); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// AdminListProducts 后台商品列表（不限状态）。
func (s *Service) AdminListProducts(ctx context.Context, req ProductListReq) ([]ProductResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	filter := ProductFilter{
		CategoryID: req.CategoryID.Int64(),
		Status:     req.Status,
		Keyword:    req.Keyword,
		Sort:       req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	rows, total, err := s.productRepo.ListAdmin(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]ProductResp, len(rows))
	for i, row := range rows {
		r := toProductResp(&row.Product)
		r.CategoryName = row.CategoryName
		r.FreightTemplateName = row.FreightTemplateName
		r.TotalStock = row.TotalStock
		resp[i] = r
	}
	return resp, total, nil
}

// AdminCreateCategory 创建分类。
func (s *Service) AdminCreateCategory(ctx context.Context, req CreateCategoryReq) (*Category, error) {
	c := &Category{
		ParentID: req.ParentID.Int64(),
		Name:     req.Name,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   req.Status,
	}
	if c.Status == "" {
		c.Status = "enabled"
	}
	if err := s.categoryRepo.Create(ctx, c); err != nil {
		return nil, errs.ErrInternal
	}
	_ = s.invalidateCategoryCache(ctx)
	return c, nil
}

// AdminUpdateCategory 更新分类。
func (s *Service) AdminUpdateCategory(ctx context.Context, id int64, req UpdateCategoryReq) error {
	c, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if req.ParentID != nil {
		c.ParentID = req.ParentID.Int64()
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Icon != nil {
		c.Icon = *req.Icon
	}
	if req.Sort != nil {
		c.Sort = *req.Sort
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	c.UpdatedAt = time.Now()
	if err := s.categoryRepo.Update(ctx, c); err != nil {
		return errs.ErrInternal
	}
	_ = s.invalidateCategoryCache(ctx)
	return nil
}

// AdminDeleteCategory 软删除分类。
func (s *Service) AdminDeleteCategory(ctx context.Context, id int64) error {
	if err := s.categoryRepo.SoftDelete(ctx, id); err != nil {
		return errs.ErrInternal
	}
	_ = s.invalidateCategoryCache(ctx)
	return nil
}

// UploadImage 上传商品图片（校验 mime）。
func (s *Service) UploadImage(ctx context.Context, adminID int64, r io.Reader, filename string) (string, error) {
	url, err := s.ossClient.UploadProductImage(ctx, r, filename)
	if err != nil {
		logger.Ctx(ctx).Warn("upload image failed",
			zap.Int64("admin_id", adminID),
			zap.Error(err))
		return "", errs.ErrParam.WithMsg(err.Error())
	}
	return url, nil
}

// invalidateCategoryCache 清除分类缓存。
func (s *Service) invalidateCategoryCache(ctx context.Context) error {
	return s.rdb.Del(ctx, categoryCacheKey).Err()
}

// AdminGetProduct 后台获取商品详情。
func (s *Service) AdminGetProduct(ctx context.Context, id int64) (*ProductDetailResp, error) {
	detail, err := s.productRepo.FindWithSpecs(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	resp := toProductDetailResp(detail)
	return &resp, nil
}
