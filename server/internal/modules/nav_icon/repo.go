package nav_icon

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// NavIconRepo 金刚区图标数据访问接口。
type NavIconRepo interface {
	FindAll(ctx context.Context) ([]NavIcon, error)
	FindActive(ctx context.Context) ([]NavIcon, error)
	FindByID(ctx context.Context, id int64) (*NavIcon, error)
	Create(ctx context.Context, n *NavIcon) error
	Update(ctx context.Context, n *NavIcon) error
	Delete(ctx context.Context, id int64) error
	BulkUpdateSort(ctx context.Context, items []SortItem) error
}

type navIconRepoImpl struct{ db *gorm.DB }

// NewNavIconRepo 构造 NavIconRepo。
func NewNavIconRepo(db *gorm.DB) NavIconRepo {
	return &navIconRepoImpl{db: db}
}

func (r *navIconRepoImpl) FindAll(ctx context.Context) ([]NavIcon, error) {
	var list []NavIcon
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *navIconRepoImpl) FindActive(ctx context.Context) ([]NavIcon, error) {
	var list []NavIcon
	err := r.db.WithContext(ctx).Where("is_active = true").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *navIconRepoImpl) FindByID(ctx context.Context, id int64) (*NavIcon, error) {
	var n NavIcon
	err := r.db.WithContext(ctx).First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *navIconRepoImpl) Create(ctx context.Context, n *NavIcon) error {
	n.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *navIconRepoImpl) Update(ctx context.Context, n *NavIcon) error {
	n.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *navIconRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&NavIcon{}, id).Error
}

func (r *navIconRepoImpl) BulkUpdateSort(ctx context.Context, items []SortItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, item := range items {
			if err := tx.Model(&NavIcon{}).Where("id = ?", item.ID).
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
