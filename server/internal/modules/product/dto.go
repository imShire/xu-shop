package product

import (
	"encoding/json"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 请求 DTO ----

// CreateCategoryReq 创建分类请求。
type CreateCategoryReq struct {
	ParentID types.Int64Str `json:"parent_id"`
	Name     string         `json:"name"   binding:"required,max=64"`
	Icon     string         `json:"icon"   binding:"omitempty,max=512"`
	Sort     int            `json:"sort"`
	Status   string         `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// UpdateCategoryReq 更新分类请求。
type UpdateCategoryReq struct {
	ParentID *types.Int64Str `json:"parent_id"`
	Name     *string         `json:"name"   binding:"omitempty,max=64"`
	Icon     *string         `json:"icon"   binding:"omitempty,max=512"`
	Sort     *int            `json:"sort"`
	Status   *string         `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// SKUReq 创建/更新时的 SKU 请求体。
type SKUReq struct {
	ID                 *string         `json:"id"`
	Attrs              json.RawMessage `json:"attrs"`
	PriceCents         int64           `json:"price_cents"           binding:"required,min=1"`
	OriginalPriceCents *int64          `json:"original_price_cents"`
	Stock              int             `json:"stock"                 binding:"min=0"`
	WeightG            int             `json:"weight_g"`
	SkuCode            *string         `json:"sku_code"`
	Barcode            *string         `json:"barcode"`
	Image              string          `json:"image"`
	Status             string          `json:"status"                binding:"omitempty,oneof=active inactive"`
	LowStockThreshold  int             `json:"low_stock_threshold"`
}

// SpecReq 规格请求体。
type SpecReq struct {
	Name   string   `json:"name"   binding:"required,max=32"`
	Sort   int      `json:"sort"`
	Values []string `json:"values" binding:"required,min=1"`
}

// CreateProductReq 创建商品请求。
type CreateProductReq struct {
	CategoryID        types.Int64Str  `json:"category_id"          binding:"required"`
	Title             string          `json:"title"                binding:"required,max=60"`
	Subtitle          string          `json:"subtitle"             binding:"omitempty,max=120"`
	MainImage         string          `json:"main_image"           binding:"required,max=512"`
	Images            json.RawMessage `json:"images"`
	VideoURL          string          `json:"video_url"            binding:"omitempty,max=512"`
	DetailHTML        string          `json:"detail_html"`
	DetailNodes       json.RawMessage `json:"detail_nodes"`
	Status            string          `json:"status"               binding:"omitempty,oneof=draft onsale offsale"`
	Sort              int             `json:"sort"`
	Tags              json.RawMessage `json:"tags"`
	Unit              string          `json:"unit"                 binding:"omitempty,max=16"`
	IsVirtual         bool            `json:"is_virtual"`
	FreightTemplateID *types.Int64Str `json:"freight_template_id"`
	VirtualSales      int             `json:"virtual_sales"        binding:"min=0"`
	Specs             []SpecReq       `json:"specs"                binding:"omitempty,max=3"`
	SKUs              []SKUReq        `json:"skus"                 binding:"required,min=1"`
}

// UpdateProductReq 更新商品请求（所有字段可选）。
type UpdateProductReq struct {
	CategoryID        *types.Int64Str `json:"category_id"`
	Title             *string         `json:"title"                binding:"omitempty,max=60"`
	Subtitle          *string         `json:"subtitle"             binding:"omitempty,max=120"`
	MainImage         *string         `json:"main_image"           binding:"omitempty,max=512"`
	Images            json.RawMessage `json:"images"`
	VideoURL          *string         `json:"video_url"            binding:"omitempty,max=512"`
	DetailHTML        *string         `json:"detail_html"`
	DetailNodes       json.RawMessage `json:"detail_nodes"`
	Sort              *int            `json:"sort"`
	Tags              json.RawMessage `json:"tags"`
	Unit              *string         `json:"unit"                 binding:"omitempty,max=16"`
	IsVirtual         *bool           `json:"is_virtual"`
	FreightTemplateID *types.Int64Str `json:"freight_template_id"`
	VirtualSales      *int            `json:"virtual_sales"        binding:"omitempty,min=0"`
	Status            *string         `json:"status"               binding:"omitempty,oneof=draft onsale offsale"`
	Specs             *[]SpecReq      `json:"specs"                binding:"omitempty,max=3"`
	SKUs              *[]SKUReq       `json:"skus"                 binding:"omitempty"`
}

// BatchStatusReq 批量修改状态请求。
type BatchStatusReq struct {
	IDs    []types.Int64Str `json:"ids"    binding:"required,min=1"`
	Status string           `json:"status" binding:"required,oneof=draft onsale offsale"`
}

// BatchPriceReq 批量修改价格请求。
type BatchPriceReq struct {
	IDs        []types.Int64Str `json:"ids"         binding:"required,min=1"`
	PriceCents int64            `json:"price_cents" binding:"required,min=1"`
}

// ProductListReq 商品列表查询参数。
type ProductListReq struct {
	CategoryID types.Int64Str `form:"category_id"`
	Status     string         `form:"status"`
	Keyword    string         `form:"keyword"`
	Sort       string         `form:"sort"      binding:"omitempty,oneof=latest popular price_asc price_desc"`
	Page       int            `form:"page"      binding:"min=0"`
	PageSize   int            `form:"page_size" binding:"min=0,max=100"`
	InStock    bool           `form:"in_stock"`
}

// ---- 响应 DTO ----

// CategoryResp 分类响应。
type CategoryResp struct {
	ID        types.Int64Str `json:"id"`
	ParentID  types.Int64Str `json:"parent_id"`
	Name      string         `json:"name"`
	Icon      string         `json:"icon,omitempty"`
	Sort      int            `json:"sort"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	Children  []CategoryResp `json:"children,omitempty"`
}

// CategoryTreeNode 分类树节点。
type CategoryTreeNode = CategoryResp

// ProductResp 商品列表响应 DTO。
type ProductResp struct {
	ID                  types.Int64Str  `json:"id"`
	CategoryID          types.Int64Str  `json:"category_id"`
	CategoryName        *string         `json:"category_name,omitempty"`
	Title               string          `json:"title"`
	Subtitle            string          `json:"subtitle,omitempty"`
	MainImage           string          `json:"main_image"`
	Status              string          `json:"status"`
	Unit                string          `json:"unit"`
	IsVirtual           bool            `json:"is_virtual"`
	FreightTemplateID   *types.Int64Str `json:"freight_template_id,omitempty"`
	FreightTemplateName *string         `json:"freight_template_name,omitempty"`
	Sales               int             `json:"sales"`
	VirtualSales        int             `json:"virtual_sales"`
	TotalStock          int             `json:"total_stock"`
	Sort                int             `json:"sort"`
	Tags                json.RawMessage `json:"tags"`
	PriceMinCents       int64           `json:"price_min_cents"`
	PriceMaxCents       int64           `json:"price_max_cents"`
	OnSaleAt            *time.Time      `json:"on_sale_at,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// SpecResp 规格响应。
type SpecResp struct {
	ID     types.Int64Str  `json:"id"`
	Name   string          `json:"name"`
	Sort   int             `json:"sort"`
	Values []SpecValueResp `json:"values"`
}

// SpecValueResp 规格值响应。
type SpecValueResp struct {
	ID    types.Int64Str `json:"id"`
	Value string         `json:"value"`
	Sort  int            `json:"sort"`
}

// SKUResp SKU 响应。
type SKUResp struct {
	ID                 types.Int64Str  `json:"id"`
	ProductID          types.Int64Str  `json:"product_id"`
	Attrs              json.RawMessage `json:"attrs"`
	PriceCents         int64           `json:"price_cents"`
	OriginalPriceCents *int64          `json:"original_price_cents,omitempty"`
	Stock              int             `json:"stock"`
	LockedStock        int             `json:"locked_stock"`
	WeightG            int             `json:"weight_g"`
	SkuCode            *string         `json:"sku_code,omitempty"`
	Barcode            *string         `json:"barcode,omitempty"`
	Image              string          `json:"image,omitempty"`
	Status             string          `json:"status"`
	LowStockThreshold  int             `json:"low_stock_threshold"`
}

// ProductDetailResp 商品详情响应 DTO。
type ProductDetailResp struct {
	ProductResp
	Images      json.RawMessage `json:"images"`
	VideoURL    string          `json:"video_url,omitempty"`
	DetailHTML  string          `json:"detail_html,omitempty"`
	DetailNodes json.RawMessage `json:"detail_nodes,omitempty"`
	Specs       []SpecResp      `json:"specs"`
	SKUs        []SKUResp       `json:"skus"`
	IsFavorite  bool            `json:"is_favorite"`
}

// FavoriteListItemResp 收藏列表项。
type FavoriteListItemResp struct {
	ID               types.Int64Str `json:"id"`
	ProductID        types.Int64Str `json:"product_id"`
	Title            string         `json:"title"`
	Image            string         `json:"image"`
	PriceCents       int64          `json:"price_cents"`
	MarketPriceCents *int64         `json:"market_price_cents,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

// ViewHistoryItemResp 浏览历史列表项。
type ViewHistoryItemResp struct {
	ID               types.Int64Str `json:"id"`
	ProductID        types.Int64Str `json:"product_id"`
	Title            string         `json:"title"`
	Image            string         `json:"image"`
	PriceCents       int64          `json:"price_cents"`
	MarketPriceCents *int64         `json:"market_price_cents,omitempty"`
	ViewedAt         time.Time      `json:"viewed_at"`
}

// UploadImageResp 上传图片响应。
type UploadImageResp struct {
	URL string `json:"url"`
}

// toProductResp entity → ProductResp。
func toProductResp(p *Product) ProductResp {
	var ftID *types.Int64Str
	if p.FreightTemplateID != nil {
		v := types.Int64Str(*p.FreightTemplateID)
		ftID = &v
	}
	return ProductResp{
		ID:                types.Int64Str(p.ID),
		CategoryID:        types.Int64Str(p.CategoryID),
		Title:             p.Title,
		Subtitle:          p.Subtitle,
		MainImage:         p.MainImage,
		Status:            p.Status,
		Unit:              p.Unit,
		IsVirtual:         p.IsVirtual,
		FreightTemplateID: ftID,
		Sales:             p.Sales,
		VirtualSales:      p.VirtualSales,
		Sort:              p.Sort,
		Tags:              json.RawMessage(p.Tags),
		PriceMinCents:     p.PriceMinCents,
		PriceMaxCents:     p.PriceMaxCents,
		OnSaleAt:          p.OnSaleAt,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// toSKUResp entity → SKUResp。
func toSKUResp(s *SKU) SKUResp {
	return SKUResp{
		ID:                 types.Int64Str(s.ID),
		ProductID:          types.Int64Str(s.ProductID),
		Attrs:              json.RawMessage(s.Attrs),
		PriceCents:         s.PriceCents,
		OriginalPriceCents: s.OriginalPriceCents,
		Stock:              s.Stock,
		LockedStock:        s.LockedStock,
		WeightG:            s.WeightG,
		SkuCode:            s.SkuCode,
		Barcode:            s.Barcode,
		Image:              s.Image,
		Status:             s.Status,
		LowStockThreshold:  s.LowStockThreshold,
	}
}

// toCategoryResp entity → CategoryResp。
func toCategoryResp(c *Category) CategoryResp {
	return CategoryResp{
		ID:        types.Int64Str(c.ID),
		ParentID:  types.Int64Str(c.ParentID),
		Name:      c.Name,
		Icon:      c.Icon,
		Sort:      c.Sort,
		Status:    c.Status,
		CreatedAt: c.CreatedAt,
	}
}
