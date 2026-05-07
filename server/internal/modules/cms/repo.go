package cms

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// ArticleFilter 文章列表过滤条件。
type ArticleFilter struct {
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

// ArticleRepo 文章仓储接口。
type ArticleRepo interface {
	List(ctx context.Context, f ArticleFilter) ([]Article, int64, error)
	FindByID(ctx context.Context, id int64) (*Article, error)
	Create(ctx context.Context, a *Article) error
	Update(ctx context.Context, a *Article) error
	SoftDelete(ctx context.Context, id int64) error
}

type articleRepoImpl struct{ db *gorm.DB }

// NewArticleRepo 创建 ArticleRepo 实现。
func NewArticleRepo(db *gorm.DB) ArticleRepo { return &articleRepoImpl{db: db} }

func (r *articleRepoImpl) List(ctx context.Context, f ArticleFilter) ([]Article, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	q := r.db.WithContext(ctx).Model(&Article{}).Where("deleted_at IS NULL")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Keyword != "" {
		q = q.Where("title ILIKE ?", "%"+f.Keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Article
	err := q.Order("sort DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *articleRepoImpl) FindByID(ctx context.Context, id int64) (*Article, error) {
	var a Article
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&a).Error
	return &a, err
}

func (r *articleRepoImpl) Create(ctx context.Context, a *Article) error {
	a.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *articleRepoImpl) Update(ctx context.Context, a *Article) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *articleRepoImpl) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&Article{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}
