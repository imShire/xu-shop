package decorate

import (
	"context"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// PageConfigRepo 页面装修配置仓储接口。
type PageConfigRepo interface {
	GetActive(ctx context.Context, pageKey string) (*PageConfig, error)
	ListByKey(ctx context.Context, pageKey string) ([]PageConfig, error)
	Save(ctx context.Context, cfg *PageConfig) error
	Activate(ctx context.Context, id int64, pageKey string) error
}

type pageConfigRepoImpl struct{ db *gorm.DB }

// NewPageConfigRepo 创建 PageConfigRepo 实现。
func NewPageConfigRepo(db *gorm.DB) PageConfigRepo { return &pageConfigRepoImpl{db: db} }

func (r *pageConfigRepoImpl) GetActive(ctx context.Context, pageKey string) (*PageConfig, error) {
	var cfg PageConfig
	err := r.db.WithContext(ctx).
		Where("page_key = ? AND is_active = TRUE", pageKey).
		First(&cfg).Error
	return &cfg, err
}

func (r *pageConfigRepoImpl) ListByKey(ctx context.Context, pageKey string) ([]PageConfig, error) {
	var list []PageConfig
	err := r.db.WithContext(ctx).
		Where("page_key = ?", pageKey).
		Order("version DESC").
		Limit(20).
		Find(&list).Error
	return list, err
}

func (r *pageConfigRepoImpl) Save(ctx context.Context, cfg *PageConfig) error {
	if cfg.ID == 0 {
		cfg.ID = snowflake.NextID()
		return r.db.WithContext(ctx).Create(cfg).Error
	}
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *pageConfigRepoImpl) Activate(ctx context.Context, id int64, pageKey string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先取消同 pageKey 所有激活
		if err := tx.Model(&PageConfig{}).
			Where("page_key = ? AND is_active = TRUE", pageKey).
			Update("is_active", false).Error; err != nil {
			return err
		}
		// 激活指定 id
		return tx.Model(&PageConfig{}).
			Where("id = ?", id).
			Update("is_active", true).Error
	})
}
