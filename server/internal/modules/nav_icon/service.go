package nav_icon

import (
	"context"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// Service 金刚区图标业务服务。
type Service struct {
	repo NavIconRepo
}

// NewService 构造 Service。
func NewService(repo NavIconRepo) *Service {
	return &Service{repo: repo}
}

// ListAll 后台查询所有金刚区图标（含非激活）。
func (s *Service) ListAll(ctx context.Context) ([]NavIcon, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return list, nil
}

// ListActive 客户端查询激活金刚区图标。
func (s *Service) ListActive(ctx context.Context) ([]NavIcon, error) {
	list, err := s.repo.FindActive(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return list, nil
}

// Create 创建金刚区图标。
func (s *Service) Create(ctx context.Context, req CreateNavIconReq) (*NavIcon, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	n := &NavIcon{
		Title:    req.Title,
		IconURL:  req.IconURL,
		LinkURL:  req.LinkURL,
		Sort:     req.Sort,
		IsActive: isActive,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, errs.ErrInternal
	}
	return n, nil
}

// Update 更新金刚区图标。
func (s *Service) Update(ctx context.Context, id int64, req UpdateNavIconReq) (*NavIcon, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	n.Title = req.Title
	n.IconURL = req.IconURL
	n.LinkURL = req.LinkURL
	n.Sort = req.Sort
	if req.IsActive != nil {
		n.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, errs.ErrInternal
	}
	return n, nil
}

// Delete 删除金刚区图标。
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
func (s *Service) Toggle(ctx context.Context, id int64) (*NavIcon, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	n.IsActive = !n.IsActive
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, errs.ErrInternal
	}
	return n, nil
}

// BulkSort 批量更新排序。
func (s *Service) BulkSort(ctx context.Context, items []SortItem) error {
	if err := s.repo.BulkUpdateSort(ctx, items); err != nil {
		return errs.ErrInternal
	}
	return nil
}
