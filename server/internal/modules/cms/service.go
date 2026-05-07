package cms

import (
	"context"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// Service CMS 文章服务。
type Service struct {
	repo ArticleRepo
}

// NewService 创建 Service。
func NewService(repo ArticleRepo) *Service { return &Service{repo: repo} }

// ---- C 端 ----

// ListPublished C 端获取已发布文章列表。
func (s *Service) ListPublished(ctx context.Context, keyword string, page, size int) ([]Article, int64, error) {
	return s.repo.List(ctx, ArticleFilter{Status: "published", Keyword: keyword, Page: page, PageSize: size})
}

// GetPublished C 端获取已发布文章详情。
func (s *Service) GetPublished(ctx context.Context, id int64) (*Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	if a.Status != "published" {
		return nil, errs.ErrNotFound
	}
	return a, nil
}

// ---- Admin ----

// AdminList Admin 获取文章列表（含草稿）。
func (s *Service) AdminList(ctx context.Context, f ArticleFilter) ([]Article, int64, error) {
	return s.repo.List(ctx, f)
}

// AdminGet Admin 获取文章详情。
func (s *Service) AdminGet(ctx context.Context, id int64) (*Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	return a, nil
}

// UpsertArticleReq 创建/更新文章请求。
type UpsertArticleReq struct {
	Title   string `json:"title"   binding:"required,max=255"`
	Cover   string `json:"cover"   binding:"omitempty,max=512"`
	Content string `json:"content"`
	Status  string `json:"status"  binding:"omitempty,oneof=draft published"`
	Sort    int    `json:"sort"`
}

// Create Admin 创建文章。
func (s *Service) Create(ctx context.Context, adminID int64, req UpsertArticleReq) (*Article, error) {
	now := time.Now()
	a := &Article{
		Title:     req.Title,
		Cover:     req.Cover,
		Content:   req.Content,
		Status:    req.Status,
		Sort:      req.Sort,
		CreatedBy: &adminID,
	}
	if req.Status == "" {
		a.Status = "draft"
	}
	if a.Status == "published" {
		a.PublishedAt = &now
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, errs.ErrInternal
	}
	return a, nil
}

// Update Admin 更新文章。
func (s *Service) Update(ctx context.Context, id int64, req UpsertArticleReq) (*Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	a.Title = req.Title
	a.Cover = req.Cover
	a.Content = req.Content
	if req.Status != "" {
		a.Status = req.Status
	}
	a.Sort = req.Sort
	if a.Status == "published" && a.PublishedAt == nil {
		now := time.Now()
		a.PublishedAt = &now
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, errs.ErrInternal
	}
	return a, nil
}

// Delete Admin 软删除文章。
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return errs.ErrInternal
	}
	return nil
}
