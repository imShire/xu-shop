package cart

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// CartRepo 购物车数据访问接口。
type CartRepo interface {
	FindByUserID(ctx context.Context, userID int64) ([]CartItemDetail, error)
	FindByID(ctx context.Context, id int64) (*CartItem, error)
	Upsert(ctx context.Context, userID, skuID int64, qty int, snapshotPrice int64) error
	Update(ctx context.Context, id int64, qty int, snapshotPrice int64) error
	Delete(ctx context.Context, id int64) error
	BatchDelete(ctx context.Context, ids []int64) error
	DeleteByUserAndIDs(ctx context.Context, userID int64, ids []int64) error
	CountByUser(ctx context.Context, userID int64) (int64, error)
	FindByIDs(ctx context.Context, ids []int64) ([]CartItem, error)
}

type cartRepoImpl struct{ db *gorm.DB }

// NewCartRepo 构造 CartRepo。
func NewCartRepo(db *gorm.DB) CartRepo { return &cartRepoImpl{db: db} }

func (r *cartRepoImpl) FindByUserID(ctx context.Context, userID int64) ([]CartItemDetail, error) {
	var rows []CartItemDetail
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			ci.id, ci.user_id, ci.sku_id, ci.qty, ci.snapshot_price_cents,
			ci.created_at, ci.updated_at,
			s.price_cents  AS sku_price_cents,
			s.status       AS sku_status,
			s.stock        AS sku_stock,
			s.locked_stock AS sku_locked_stock,
			s.image        AS sku_image,
			s.attrs        AS sku_attrs,
			p.id           AS product_id,
			p.category_id  AS product_category_id,
			p.title        AS product_title,
			p.subtitle     AS product_subtitle,
			p.main_image   AS product_main_image,
			p.tags         AS product_tags,
			p.status       AS product_status,
			p.sales        AS product_sales,
			p.price_min_cents AS product_price_min_cents,
			p.price_max_cents AS product_price_max_cents,
			(p.deleted_at IS NOT NULL) AS product_deleted
		FROM cart_item ci
		LEFT JOIN sku     s ON ci.sku_id    = s.id
		LEFT JOIN product p ON s.product_id = p.id
		WHERE ci.user_id = ?
		ORDER BY ci.updated_at DESC
	`, userID).Scan(&rows).Error
	return rows, err
}

func (r *cartRepoImpl) FindByID(ctx context.Context, id int64) (*CartItem, error) {
	var item CartItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepoImpl) Upsert(ctx context.Context, userID, skuID int64, qty int, snapshotPrice int64) error {
	now := time.Now()
	item := CartItem{
		ID:                 snowflake.NextID(),
		UserID:             userID,
		SkuID:              skuID,
		Qty:                qty,
		SnapshotPriceCents: snapshotPrice,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	// ON CONFLICT (user_id, sku_id) 时累加数量并更新快照价格
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "sku_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"qty":                  gorm.Expr("LEAST(cart_item.qty + EXCLUDED.qty, 999)"),
			"snapshot_price_cents": gorm.Expr("EXCLUDED.snapshot_price_cents"),
			"updated_at":           gorm.Expr("EXCLUDED.updated_at"),
		}),
	}).Create(&item).Error
}

func (r *cartRepoImpl) Update(ctx context.Context, id int64, qty int, snapshotPrice int64) error {
	return r.db.WithContext(ctx).Model(&CartItem{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"qty":                  qty,
			"snapshot_price_cents": snapshotPrice,
			"updated_at":           time.Now(),
		}).Error
}

func (r *cartRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&CartItem{}, id).Error
}

func (r *cartRepoImpl) BatchDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&CartItem{}).Error
}

func (r *cartRepoImpl) DeleteByUserAndIDs(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND id IN ?", userID, ids).
		Delete(&CartItem{}).Error
}

func (r *cartRepoImpl) CountByUser(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&CartItem{}).
		Where("user_id = ?", userID).Count(&cnt).Error
	return cnt, err
}

func (r *cartRepoImpl) FindByIDs(ctx context.Context, ids []int64) ([]CartItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []CartItem
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}
