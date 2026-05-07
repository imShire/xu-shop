package product

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// ProductFilter 商品列表过滤条件。
type ProductFilter struct {
	CategoryID int64
	Status     string
	Keyword    string
	Sort       string
	Page       int
	PageSize   int
	InStock    bool
}

// ProductDetail 商品详情（含规格和 SKU）。
type ProductDetail struct {
	Product
	Specs  []ProductSpec
	Values []ProductSpecValue
	SKUs   []SKU
}

// ProductAdminRow 后台列表查询行（含关联名称和汇总库存）。
type ProductAdminRow struct {
	Product
	CategoryName        *string `gorm:"column:category_name"`
	FreightTemplateName *string `gorm:"column:freight_template_name"`
	TotalStock          int     `gorm:"column:total_stock"`
}

// SKUStockRow 库存对账用。
type SKUStockRow struct {
	ID          int64
	Stock       int
	LockedStock int
}

// ---- CategoryRepo ----

// CategoryRepo 分类数据访问接口。
type CategoryRepo interface {
	FindAll(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id int64) (*Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, id int64) error
}

type categoryRepoImpl struct{ db *gorm.DB }

// NewCategoryRepo 构造 CategoryRepo。
func NewCategoryRepo(db *gorm.DB) CategoryRepo {
	return &categoryRepoImpl{db: db}
}

func (r *categoryRepoImpl) FindAll(ctx context.Context) ([]Category, error) {
	var list []Category
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *categoryRepoImpl) FindByID(ctx context.Context, id int64) (*Category, error) {
	var c Category
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepoImpl) Create(ctx context.Context, c *Category) error {
	c.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *categoryRepoImpl) Update(ctx context.Context, c *Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *categoryRepoImpl) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Category{}, id).Error
}

// ---- ProductRepo ----

// ProductRepo 商品数据访问接口。
type ProductRepo interface {
	List(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	ListAdmin(ctx context.Context, filter ProductFilter) ([]ProductAdminRow, int64, error)
	FindByID(ctx context.Context, id int64) (*Product, error)
	FindByIDs(ctx context.Context, ids []int64) ([]Product, error)
	FindWithSpecs(ctx context.Context, id int64) (*ProductDetail, error)
	Create(ctx context.Context, p *Product, specs []ProductSpec, values []ProductSpecValue, skus []SKU) error
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, id int64) error
	Copy(ctx context.Context, productID int64) (*Product, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	BatchUpdateStatus(ctx context.Context, ids []int64, status string) error
	UpdatePriceRange(ctx context.Context, productID int64, min, max int64) error
	// ReplaceSpecsAndSKUs 在事务中替换商品的规格、规格值和 SKU。
	// specs/values 已含 ID；skus 中 ID>0 的执行 upsert，ID=0 的生成新 ID；
	// 未出现在 skus 中的旧 SKU 被标记为 inactive（保留行，避免破坏订单引用）。
	ReplaceSpecsAndSKUs(ctx context.Context, productID int64, specs []ProductSpec, values []ProductSpecValue, skus []SKU) error
}

type productRepoImpl struct{ db *gorm.DB }

// NewProductRepo 构造 ProductRepo。
func NewProductRepo(db *gorm.DB) ProductRepo {
	return &productRepoImpl{db: db}
}

func (r *productRepoImpl) List(ctx context.Context, f ProductFilter) ([]Product, int64, error) {
	q := r.db.WithContext(ctx).Model(&Product{})
	if f.CategoryID > 0 {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.InStock {
		q = q.Where("EXISTS (SELECT 1 FROM sku WHERE sku.product_id = product.id AND sku.stock > sku.locked_stock AND sku.status = 'active' AND sku.deleted_at IS NULL)")
	}
	if f.Keyword != "" {
		q = q.Where("title ILIKE ? OR subtitle ILIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 20
	}

	var list []Product
	orderBy := "sort DESC, id DESC"
	switch f.Sort {
	case "latest":
		orderBy = "on_sale_at DESC NULLS LAST, created_at DESC, id DESC"
	case "popular":
		orderBy = "(COALESCE(sales, 0) + COALESCE(virtual_sales, 0)) DESC, sort DESC, id DESC"
	case "price_asc":
		orderBy = "price_cents ASC, id DESC"
	case "price_desc":
		orderBy = "price_cents DESC, id DESC"
	}

	err := q.Order(orderBy).
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	return list, total, err
}

// ListAdmin 后台商品列表，JOIN 分类名和运费模板名，汇总库存。
func (r *productRepoImpl) ListAdmin(ctx context.Context, f ProductFilter) ([]ProductAdminRow, int64, error) {
	baseQ := r.db.WithContext(ctx).Table("product p").
		Select(`p.*,
			c.name AS category_name,
			ft.name AS freight_template_name,
			COALESCE(SUM(s.stock), 0) AS total_stock`).
		Joins("LEFT JOIN category c ON c.id = p.category_id AND c.deleted_at IS NULL").
		Joins("LEFT JOIN freight_template ft ON ft.id = p.freight_template_id").
		Joins("LEFT JOIN sku s ON s.product_id = p.id").
		Where("p.deleted_at IS NULL").
		Group("p.id, c.name, ft.name")

	if f.CategoryID > 0 {
		baseQ = baseQ.Where("p.category_id = ?", f.CategoryID)
	}
	if f.Status != "" {
		baseQ = baseQ.Where("p.status = ?", f.Status)
	}
	if f.Keyword != "" {
		baseQ = baseQ.Where("p.title ILIKE ? OR p.subtitle ILIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}

	// 计数查询（不带 GROUP BY / ORDER BY）
	cq := r.db.WithContext(ctx).Model(&Product{}).Where("deleted_at IS NULL")
	if f.CategoryID > 0 {
		cq = cq.Where("category_id = ?", f.CategoryID)
	}
	if f.Status != "" {
		cq = cq.Where("status = ?", f.Status)
	}
	if f.Keyword != "" {
		cq = cq.Where("title ILIKE ? OR subtitle ILIKE ?", "%"+f.Keyword+"%", "%"+f.Keyword+"%")
	}
	var total int64
	if err := cq.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.PageSize
	if size < 1 {
		size = 20
	}

	var list []ProductAdminRow
	err := baseQ.Order("p.sort DESC, p.id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	return list, total, err
}

func (r *productRepoImpl) FindByID(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepoImpl) FindByIDs(ctx context.Context, ids []int64) ([]Product, error) {
	var list []Product
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *productRepoImpl) FindWithSpecs(ctx context.Context, id int64) (*ProductDetail, error) {
	p, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	detail := &ProductDetail{Product: *p}

	if err := r.db.WithContext(ctx).
		Where("product_id = ?", id).Order("sort ASC").
		Find(&detail.Specs).Error; err != nil {
		return nil, err
	}

	if len(detail.Specs) > 0 {
		specIDs := make([]int64, len(detail.Specs))
		for i, s := range detail.Specs {
			specIDs[i] = s.ID
		}
		if err := r.db.WithContext(ctx).
			Where("spec_id IN ?", specIDs).Order("sort ASC").
			Find(&detail.Values).Error; err != nil {
			return nil, err
		}
	}

	if err := r.db.WithContext(ctx).
		Where("product_id = ?", id).
		Find(&detail.SKUs).Error; err != nil {
		return nil, err
	}

	return detail, nil
}

func (r *productRepoImpl) Create(ctx context.Context, p *Product, specs []ProductSpec, values []ProductSpecValue, skus []SKU) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p.ID = snowflake.NextID()
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for i := range specs {
			specs[i].ID = snowflake.NextID()
			specs[i].ProductID = p.ID
		}
		if len(specs) > 0 {
			if err := tx.Create(&specs).Error; err != nil {
				return err
			}
		}
		for i := range values {
			values[i].ID = snowflake.NextID()
		}
		if len(values) > 0 {
			if err := tx.Create(&values).Error; err != nil {
				return err
			}
		}
		for i := range skus {
			skus[i].ID = snowflake.NextID()
			skus[i].ProductID = p.ID
		}
		if len(skus) > 0 {
			if err := tx.Create(&skus).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *productRepoImpl) Update(ctx context.Context, p *Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productRepoImpl) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Product{}, id).Error
}

func (r *productRepoImpl) Copy(ctx context.Context, productID int64) (*Product, error) {
	detail, err := r.FindWithSpecs(ctx, productID)
	if err != nil {
		return nil, err
	}

	newProduct := detail.Product
	newProduct.ID = snowflake.NextID()
	newProduct.Title = newProduct.Title + "_copy"
	newProduct.Status = "draft"
	newProduct.Sales = 0
	newProduct.OnSaleAt = nil
	now := time.Now()
	newProduct.CreatedAt = now
	newProduct.UpdatedAt = now
	newProduct.DeletedAt = gorm.DeletedAt{}

	// 构建规格 ID 映射
	oldToNewSpecID := make(map[int64]int64, len(detail.Specs))
	newSpecs := make([]ProductSpec, len(detail.Specs))
	for i, s := range detail.Specs {
		newID := snowflake.NextID()
		oldToNewSpecID[s.ID] = newID
		newSpecs[i] = ProductSpec{
			ID:        newID,
			ProductID: newProduct.ID,
			Name:      s.Name,
			Sort:      s.Sort,
		}
	}

	newValues := make([]ProductSpecValue, len(detail.Values))
	for i, v := range detail.Values {
		newValues[i] = ProductSpecValue{
			ID:     snowflake.NextID(),
			SpecID: oldToNewSpecID[v.SpecID],
			Value:  v.Value,
			Sort:   v.Sort,
		}
	}

	newSKUs := make([]SKU, len(detail.SKUs))
	for i, s := range detail.SKUs {
		newSKUs[i] = s
		newSKUs[i].ID = snowflake.NextID()
		newSKUs[i].ProductID = newProduct.ID
		newSKUs[i].Stock = 0
		newSKUs[i].LockedStock = 0
		newSKUs[i].SkuCode = nil
	}

	if err := r.Create(ctx, &newProduct, newSpecs, newValues, newSKUs); err != nil {
		return nil, err
	}
	return &newProduct, nil
}

func (r *productRepoImpl) UpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]any{"status": status}
	if status == "onsale" {
		now := time.Now()
		updates["on_sale_at"] = now
	}
	return r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", id).Updates(updates).Error
}

func (r *productRepoImpl) BatchUpdateStatus(ctx context.Context, ids []int64, status string) error {
	return r.db.WithContext(ctx).Model(&Product{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

func (r *productRepoImpl) UpdatePriceRange(ctx context.Context, productID int64, min, max int64) error {
	return r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", productID).
		Updates(map[string]any{
			"price_min_cents": min,
			"price_max_cents": max,
		}).Error
}

func (r *productRepoImpl) ReplaceSpecsAndSKUs(ctx context.Context, productID int64, specs []ProductSpec, values []ProductSpecValue, skus []SKU) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除旧规格值（无 FK cascade，需手动清理）
		if err := tx.Exec(
			"DELETE FROM product_spec_value WHERE spec_id IN (SELECT id FROM product_spec WHERE product_id = ?)",
			productID,
		).Error; err != nil {
			return err
		}
		// 2. 删除旧规格
		if err := tx.Where("product_id = ?", productID).Delete(&ProductSpec{}).Error; err != nil {
			return err
		}
		// 3. 插入新规格和规格值
		if len(specs) > 0 {
			if err := tx.Create(&specs).Error; err != nil {
				return err
			}
		}
		if len(values) > 0 {
			if err := tx.Create(&values).Error; err != nil {
				return err
			}
		}
		// 4. 对于不在新 SKU 列表中的旧 SKU，标记为 inactive（保留行，保护订单引用）
		newIDSet := make([]int64, 0, len(skus))
		for _, s := range skus {
			if s.ID > 0 {
				newIDSet = append(newIDSet, s.ID)
			}
		}
		deactivateQ := tx.Model(&SKU{}).Where("product_id = ?", productID)
		if len(newIDSet) > 0 {
			deactivateQ = deactivateQ.Where("id NOT IN ?", newIDSet)
		}
		if err := deactivateQ.Update("status", "inactive").Error; err != nil {
			return err
		}
		// 5. Upsert 新 SKU（ID 已由 service 分配）
		if len(skus) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"product_id", "attrs", "price_cents", "original_price_cents",
					"stock", "weight_g", "sku_code", "barcode", "image",
					"status", "low_stock_threshold",
				}),
			}).Create(&skus).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- SKURepo ----

// SKURepo SKU 数据访问接口。
type SKURepo interface {
	FindByProduct(ctx context.Context, productID int64) ([]SKU, error)
	FindByID(ctx context.Context, id int64) (*SKU, error)
	BatchUpdatePrice(ctx context.Context, ids []int64, priceCents int64) error
	FindByIDs(ctx context.Context, ids []int64) ([]SKU, error)
}

type skuRepoImpl struct{ db *gorm.DB }

// NewSKURepo 构造 SKURepo。
func NewSKURepo(db *gorm.DB) SKURepo {
	return &skuRepoImpl{db: db}
}

func (r *skuRepoImpl) FindByProduct(ctx context.Context, productID int64) ([]SKU, error) {
	var list []SKU
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&list).Error
	return list, err
}

func (r *skuRepoImpl) FindByID(ctx context.Context, id int64) (*SKU, error) {
	var sku SKU
	err := r.db.WithContext(ctx).First(&sku, id).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepoImpl) BatchUpdatePrice(ctx context.Context, ids []int64, priceCents int64) error {
	return r.db.WithContext(ctx).Model(&SKU{}).
		Where("id IN ?", ids).
		Update("price_cents", priceCents).Error
}

func (r *skuRepoImpl) FindByIDs(ctx context.Context, ids []int64) ([]SKU, error) {
	var list []SKU
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// ---- FavoriteRepo ----

// FavoriteListItem 收藏列表查询结果。
type FavoriteListItem struct {
	ProductID        int64     `gorm:"column:product_id"`
	Title            string    `gorm:"column:title"`
	Image            string    `gorm:"column:image"`
	PriceCents       int64     `gorm:"column:price_cents"`
	MarketPriceCents *int64    `gorm:"column:market_price_cents"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

// FavoriteRepo 收藏数据访问接口。
type FavoriteRepo interface {
	Add(ctx context.Context, userID, productID int64) error
	Remove(ctx context.Context, userID, productID int64) error
	List(ctx context.Context, userID int64, page, size int) ([]FavoriteListItem, int64, error)
	Exists(ctx context.Context, userID, productID int64) (bool, error)
}

type favoriteRepoImpl struct{ db *gorm.DB }

// NewFavoriteRepo 构造 FavoriteRepo。
func NewFavoriteRepo(db *gorm.DB) FavoriteRepo {
	return &favoriteRepoImpl{db: db}
}

func (r *favoriteRepoImpl) Add(ctx context.Context, userID, productID int64) error {
	fav := UserFavorite{UserID: userID, ProductID: productID}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&fav).Error
}

func (r *favoriteRepoImpl) Remove(ctx context.Context, userID, productID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&UserFavorite{}).Error
}

func (r *favoriteRepoImpl) List(ctx context.Context, userID int64, page, size int) ([]FavoriteListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Table("user_favorite AS uf").
		Joins("JOIN product AS p ON p.id = uf.product_id AND p.deleted_at IS NULL").
		Where("uf.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []FavoriteListItem
	if err := r.db.WithContext(ctx).
		Table("user_favorite AS uf").
		Select(`
			uf.product_id AS product_id,
			p.title AS title,
			p.main_image AS image,
			p.price_min_cents AS price_cents,
			NULL::bigint AS market_price_cents,
			uf.created_at AS created_at
		`).
		Joins("JOIN product AS p ON p.id = uf.product_id AND p.deleted_at IS NULL").
		Where("uf.user_id = ?", userID).
		Order("uf.created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	if len(items) == 0 {
		return nil, total, nil
	}
	return items, total, nil
}

func (r *favoriteRepoImpl) Exists(ctx context.Context, userID, productID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

// ---- ViewHistoryRepo ----

// ViewHistoryItem 浏览历史查询结果。
type ViewHistoryItem struct {
	ProductID        int64     `gorm:"column:product_id"`
	Title            string    `gorm:"column:title"`
	Image            string    `gorm:"column:image"`
	PriceCents       int64     `gorm:"column:price_cents"`
	MarketPriceCents *int64    `gorm:"column:market_price_cents"`
	ViewedAt         time.Time `gorm:"column:viewed_at"`
}

// ViewHistoryRepo 浏览历史数据访问接口。
type ViewHistoryRepo interface {
	Upsert(ctx context.Context, userID, productID int64) error
	List(ctx context.Context, userID int64, page, size int) ([]ViewHistoryItem, int64, error)
	Clear(ctx context.Context, userID int64) error
	CountByUser(ctx context.Context, userID int64) (int64, error)
	DeleteOldest(ctx context.Context, userID int64, keepCount int) error
}

type viewHistoryRepoImpl struct{ db *gorm.DB }

// NewViewHistoryRepo 构造 ViewHistoryRepo。
func NewViewHistoryRepo(db *gorm.DB) ViewHistoryRepo {
	return &viewHistoryRepoImpl{db: db}
}

func (r *viewHistoryRepoImpl) Upsert(ctx context.Context, userID, productID int64) error {
	h := UserViewHistory{UserID: userID, ProductID: productID, ViewedAt: time.Now()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&h).Error
}

func (r *viewHistoryRepoImpl) List(ctx context.Context, userID int64, page, size int) ([]ViewHistoryItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > maxViewHistory {
		size = maxViewHistory
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Table("user_view_history AS uvh").
		Joins("JOIN product AS p ON p.id = uvh.product_id AND p.deleted_at IS NULL").
		Where("uvh.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []ViewHistoryItem
	if err := r.db.WithContext(ctx).
		Table("user_view_history AS uvh").
		Select(`
			uvh.product_id AS product_id,
			p.title AS title,
			p.main_image AS image,
			p.price_min_cents AS price_cents,
			NULL::bigint AS market_price_cents,
			uvh.viewed_at AS viewed_at
		`).
		Joins("JOIN product AS p ON p.id = uvh.product_id AND p.deleted_at IS NULL").
		Where("uvh.user_id = ?", userID).
		Order("uvh.viewed_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *viewHistoryRepoImpl) Clear(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&UserViewHistory{}).Error
}

func (r *viewHistoryRepoImpl) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserViewHistory{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *viewHistoryRepoImpl) DeleteOldest(ctx context.Context, userID int64, keepCount int) error {
	// 删除超出 keepCount 条之外最旧的记录
	return r.db.WithContext(ctx).Exec(`
		DELETE FROM user_view_history
		WHERE user_id = ? AND product_id NOT IN (
			SELECT product_id FROM user_view_history
			WHERE user_id = ?
			ORDER BY viewed_at DESC
			LIMIT ?
		)
	`, userID, userID, keepCount).Error
}
