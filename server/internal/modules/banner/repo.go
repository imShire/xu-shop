package banner

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// BannerRepo 横幅数据访问接口。
type BannerRepo interface {
	FindAll(ctx context.Context) ([]Banner, error)
	FindActive(ctx context.Context) ([]Banner, error)
	FindByID(ctx context.Context, id int64) (*Banner, error)
	Create(ctx context.Context, b *Banner) error
	Update(ctx context.Context, b *Banner) error
	Delete(ctx context.Context, id int64) error
	BulkUpdateSort(ctx context.Context, items []SortItem) error
}

type bannerRepoImpl struct{ db *gorm.DB }

// NewBannerRepo 构造 BannerRepo。
func NewBannerRepo(db *gorm.DB) BannerRepo {
	return &bannerRepoImpl{db: db}
}

func (r *bannerRepoImpl) FindAll(ctx context.Context) ([]Banner, error) {
	var list []Banner
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *bannerRepoImpl) FindActive(ctx context.Context) ([]Banner, error) {
	var list []Banner
	err := r.db.WithContext(ctx).Where("is_active = true").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *bannerRepoImpl) FindByID(ctx context.Context, id int64) (*Banner, error) {
	var b Banner
	err := r.db.WithContext(ctx).First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bannerRepoImpl) Create(ctx context.Context, b *Banner) error {
	b.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *bannerRepoImpl) Update(ctx context.Context, b *Banner) error {
	b.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *bannerRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Banner{}, id).Error
}

func (r *bannerRepoImpl) BulkUpdateSort(ctx context.Context, items []SortItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, item := range items {
			if err := tx.Model(&Banner{}).Where("id = ?", item.ID.Int64()).
				Updates(map[string]any{
					"sort":       item.Sort,
					"updated_at": now,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
