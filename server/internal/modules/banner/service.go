package banner

import (
	"context"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// Service 横幅业务服务。
type Service struct {
	repo BannerRepo
}

// NewService 构造 Service。
func NewService(repo BannerRepo) *Service {
	return &Service{repo: repo}
}

// ListAll 后台查询所有横幅（含非激活）。
func (s *Service) ListAll(ctx context.Context) ([]Banner, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return list, nil
}

// ListActive 客户端查询激活横幅。
func (s *Service) ListActive(ctx context.Context) ([]Banner, error) {
	list, err := s.repo.FindActive(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return list, nil
}

// Create 创建横幅。
func (s *Service) Create(ctx context.Context, req CreateBannerReq) (*Banner, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	linkURL := req.LinkURL
	if req.LinkConfig != nil && req.LinkConfig.URL != "" {
		linkURL = req.LinkConfig.URL
	}
	b := &Banner{
		Title:      req.Title,
		ImageURL:   req.ImageURL,
		LinkURL:    linkURL,
		LinkConfig: req.LinkConfig,
		Sort:       req.Sort,
		IsActive:   isActive,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, errs.ErrInternal
	}
	return b, nil
}

// Update 更新横幅。
func (s *Service) Update(ctx context.Context, id int64, req UpdateBannerReq) (*Banner, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	linkURL := req.LinkURL
	if req.LinkConfig != nil && req.LinkConfig.URL != "" {
		linkURL = req.LinkConfig.URL
	}
	b.Title = req.Title
	b.ImageURL = req.ImageURL
	b.LinkURL = linkURL
	b.LinkConfig = req.LinkConfig
	b.Sort = req.Sort
	if req.IsActive != nil {
		b.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, errs.ErrInternal
	}
	return b, nil
}

// Delete 删除横幅。
func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// Toggle 切换激活状态。
func (s *Service) Toggle(ctx context.Context, id int64) (*Banner, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	b.IsActive = !b.IsActive
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, errs.ErrInternal
	}
	return b, nil
}

// BulkSort 批量更新排序。
func (s *Service) BulkSort(ctx context.Context, items []SortItem) error {
	if err := s.repo.BulkUpdateSort(ctx, items); err != nil {
		return errs.ErrInternal
	}
	return nil
}
