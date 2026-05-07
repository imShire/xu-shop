package cart

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/stock"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

const maxCartRows = 100

// CartListResp 购物车列表响应。
type CartListResp struct {
	Items []CartItemResp `json:"items"`
	Total int64          `json:"total"`
}

type CartProductResp struct {
	ID             types.Int64Str  `json:"id"`
	CategoryID     types.Int64Str  `json:"category_id"`
	Title          string          `json:"title"`
	Subtitle       string          `json:"subtitle,omitempty"`
	MainImage      string          `json:"main_image"`
	Status         string          `json:"status"`
	Sales          int             `json:"sales"`
	Tags           json.RawMessage `json:"tags"`
	PriceMinCents  int64           `json:"price_min_cents"`
	PriceMaxCents  int64           `json:"price_max_cents"`
}

// CartItemResp 单条购物车响应。
type CartItemResp struct {
	ID                 types.Int64Str  `json:"id"`
	SkuID              types.Int64Str  `json:"sku_id"`
	ProductID          types.Int64Str  `json:"product_id"`
	Product            CartProductResp `json:"product"`
	ProductTitle       string          `json:"product_title"`
	SkuImage           string          `json:"sku_image"`
	SkuAttrs           json.RawMessage `json:"sku_attrs"`
	Qty                int             `json:"qty"`
	SnapshotPriceCents int64           `json:"snapshot_price_cents"`
	CurrentPriceCents  int64           `json:"current_price_cents"`
	AvailableStock     int             `json:"available_stock"`
	IsAvailable        bool            `json:"is_available"`
	UnavailableReason  string          `json:"unavailable_reason,omitempty"`
}

// PrecheckResp 下单前检查响应。
type PrecheckResp struct {
	Conflicts []PrecheckConflict `json:"conflicts"`
	OK        bool               `json:"ok"`
}

// PrecheckConflict 冲突明细。
type PrecheckConflict struct {
	CartItemID types.Int64Str `json:"cart_item_id"`
	SkuID      types.Int64Str `json:"sku_id"`
	Reason     string         `json:"reason"`
}

// Service 购物车服务。
type Service struct {
	repo        CartRepo
	skuRepo     product.SKURepo
	productRepo product.ProductRepo
	stockClient *stock.Client
}

// NewService 构造 Service。
func NewService(
	repo CartRepo,
	skuRepo product.SKURepo,
	productRepo product.ProductRepo,
	stockClient *stock.Client,
) *Service {
	return &Service{
		repo:        repo,
		skuRepo:     skuRepo,
		productRepo: productRepo,
		stockClient: stockClient,
	}
}

// List 购物车列表（含可用性）。
func (s *Service) List(ctx context.Context, userID int64) (*CartListResp, error) {
	rows, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	items := make([]CartItemResp, len(rows))
	for i, r := range rows {
		product := CartProductResp{
			ID:            types.Int64Str(r.ProductID),
			CategoryID:    types.Int64Str(r.ProductCategoryID),
			Title:         r.ProductTitle,
			Subtitle:      r.ProductSubtitle,
			MainImage:     r.ProductMainImage,
			Status:        r.ProductStatus,
			Sales:         r.ProductSales,
			Tags:          json.RawMessage(r.ProductTags),
			PriceMinCents: r.ProductPriceMinCents,
			PriceMaxCents: r.ProductPriceMaxCents,
		}
		skuImage := r.SkuImage
		if skuImage == "" {
			skuImage = r.ProductMainImage
		}

		item := CartItemResp{
			ID:                 types.Int64Str(r.ID),
			SkuID:              types.Int64Str(r.SkuID),
			ProductID:          types.Int64Str(r.ProductID),
			Product:            product,
			ProductTitle:       r.ProductTitle,
			SkuImage:           skuImage,
			SkuAttrs:           r.SkuAttrs,
			Qty:                r.Qty,
			SnapshotPriceCents: r.SnapshotPriceCents,
			CurrentPriceCents:  r.SkuPriceCents,
			AvailableStock:     s.resolveAvailableStock(ctx, r.SkuID, r.SkuStock-r.SkuLockedStock),
			IsAvailable:        true,
		}
		// 可用性检查
		reason := checkAvailability(r, item.AvailableStock)
		if reason != "" {
			item.IsAvailable = false
			item.UnavailableReason = reason
		}
		items[i] = item
	}
	return &CartListResp{Items: items, Total: int64(len(items))}, nil
}

// checkAvailability 检查购物车条目可用性，返回不可用原因（空表示可用）。
func checkAvailability(r CartItemDetail, availableStock int) string {
	if r.ProductID == 0 || r.SkuStatus == "" {
		return "sku_not_found"
	}
	if r.ProductDeleted {
		return "product_deleted"
	}
	if r.ProductStatus != "onsale" {
		return "product_offsale"
	}
	if r.SkuStatus != "active" {
		return "sku_disabled"
	}
	if availableStock < r.Qty {
		return "sku_oos"
	}
	return ""
}

func (s *Service) resolveAvailableStock(ctx context.Context, skuID int64, fallback int) int {
	if fallback < 0 {
		fallback = 0
	}
	if s.stockClient == nil || skuID == 0 {
		return fallback
	}
	available, exists, err := s.stockClient.GetWithExists(ctx, skuID)
	if err != nil || !exists {
		return fallback
	}
	if available < 0 {
		return 0
	}
	return available
}

// Add 添加商品到购物车。
func (s *Service) Add(ctx context.Context, userID, skuID int64, qty int) error {
	sku, err := s.skuRepo.FindByID(ctx, skuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("SKU 不存在")
		}
		return errs.ErrInternal
	}
	if sku.Status != "active" {
		return errs.ErrParam.WithMsg("SKU 已下架")
	}

	p, err := s.productRepo.FindByID(ctx, sku.ProductID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("商品不存在")
		}
		return errs.ErrInternal
	}
	if p.Status != "onsale" {
		return errs.ErrParam.WithMsg("商品未上架")
	}

	// 检查购物车总行数 ≤ 100
	cnt, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return errs.ErrInternal
	}
	if cnt >= maxCartRows {
		return errs.ErrConflict.WithMsg("购物车已满（最多 100 种商品）")
	}

	return s.repo.Upsert(ctx, userID, skuID, qty, sku.PriceCents)
}

// Update 修改购物车条目数量。
func (s *Service) Update(ctx context.Context, id, userID int64, qty int) error {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if item.UserID != userID {
		return errs.ErrForbidden
	}
	sku, err := s.skuRepo.FindByID(ctx, item.SkuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrParam.WithMsg("商品规格已失效")
		}
		return errs.ErrInternal
	}
	if sku.Status != "active" {
		return errs.ErrParam.WithMsg("商品规格已失效")
	}
	p, err := s.productRepo.FindByID(ctx, sku.ProductID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrParam.WithMsg("商品已失效")
		}
		return errs.ErrInternal
	}
	if p.Status != "onsale" {
		return errs.ErrParam.WithMsg("商品已下架")
	}

	available := s.resolveAvailableStock(ctx, sku.ID, sku.Stock-sku.LockedStock)
	if available < qty {
		if available <= 0 {
			return errs.ErrParam.WithMsg("当前无库存")
		}
		return errs.ErrParam.WithMsg(fmt.Sprintf("库存不足，当前最多可购买 %d 件", available))
	}

	return s.repo.Update(ctx, id, qty, sku.PriceCents)
}

// Delete 删除单条购物车条目。
func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if item.UserID != userID {
		return errs.ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

// BatchDelete 批量删除（校验归属）。
func (s *Service) BatchDelete(ctx context.Context, userID int64, ids []int64) error {
	return s.repo.DeleteByUserAndIDs(ctx, userID, ids)
}

// CleanInvalid 清除不可用条目。
func (s *Service) CleanInvalid(ctx context.Context, userID int64) error {
	rows, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return errs.ErrInternal
	}
	var invalidIDs []int64
	for _, r := range rows {
		available := s.resolveAvailableStock(ctx, r.SkuID, r.SkuStock-r.SkuLockedStock)
		if checkAvailability(r, available) != "" {
			invalidIDs = append(invalidIDs, r.ID)
		}
	}
	if len(invalidIDs) == 0 {
		return nil
	}
	return s.repo.BatchDelete(ctx, invalidIDs)
}

// Count 购物车条目数。
func (s *Service) Count(ctx context.Context, userID int64) (int64, error) {
	cnt, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return 0, errs.ErrInternal
	}
	return cnt, nil
}

// Precheck 下单前检查选中条目（价格/库存冲突）。
func (s *Service) Precheck(ctx context.Context, userID int64, ids []int64) (*PrecheckResp, error) {
	items, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, errs.ErrInternal
	}

	resp := &PrecheckResp{
		Conflicts: make([]PrecheckConflict, 0),
		OK:        true,
	}
	for _, item := range items {
		if item.UserID != userID {
			continue
		}
		sku, skuErr := s.skuRepo.FindByID(ctx, item.SkuID)
		if skuErr != nil {
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     "sku_not_found",
			})
			resp.OK = false
			continue
		}
		if sku.Status != "active" {
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     "sku_disabled",
			})
			resp.OK = false
			continue
		}

		p, prodErr := s.productRepo.FindByID(ctx, sku.ProductID)
		if prodErr != nil {
			reason := "product_deleted"
			if prodErr != gorm.ErrRecordNotFound {
				return nil, errs.ErrInternal
			}
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     reason,
			})
			resp.OK = false
			continue
		}
		if p.Status != "onsale" {
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     "product_offsale",
			})
			resp.OK = false
			continue
		}

		// 价格变动
		if sku.PriceCents != item.SnapshotPriceCents {
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     "price_changed",
			})
			resp.OK = false
		}

		// 库存不足
		available := s.resolveAvailableStock(ctx, sku.ID, sku.Stock-sku.LockedStock)
		if available < item.Qty {
			resp.Conflicts = append(resp.Conflicts, PrecheckConflict{
				CartItemID: types.Int64Str(item.ID),
				SkuID:      types.Int64Str(item.SkuID),
				Reason:     "stock_insufficient",
			})
			resp.OK = false
		}
	}
	return resp, nil
}
