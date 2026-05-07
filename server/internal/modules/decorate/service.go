package decorate

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// Service 页面装修服务。
type Service struct{ repo PageConfigRepo }

// NewService 创建 Service。
func NewService(repo PageConfigRepo) *Service { return &Service{repo: repo} }

// GetActivePage C 端获取激活配置（无则返回空 modules）。
func (s *Service) GetActivePage(ctx context.Context, pageKey string) (*PageConfig, error) {
	cfg, err := s.repo.GetActive(ctx, pageKey)
	if err == gorm.ErrRecordNotFound {
		return &PageConfig{PageKey: pageKey, Modules: Modules{}}, nil
	}
	if err != nil {
		return nil, errs.ErrInternal
	}
	return cfg, nil
}

// ListVersions Admin 获取历史版本列表。
func (s *Service) ListVersions(ctx context.Context, pageKey string) ([]PageConfig, error) {
	return s.repo.ListByKey(ctx, pageKey)
}

// SaveConfigReq 保存页面装修配置请求。
type SaveConfigReq struct {
	PageKey string       `json:"page_key" binding:"required,max=32"`
	Modules []PageModule `json:"modules"  binding:"required"`
}

// Save Admin 保存新版本（不自动激活）。
func (s *Service) Save(ctx context.Context, adminID int64, req SaveConfigReq) (*PageConfig, error) {
	// 计算版本号
	list, _ := s.repo.ListByKey(ctx, req.PageKey)
	version := len(list) + 1

	now := time.Now()
	cfg := &PageConfig{
		PageKey:   req.PageKey,
		Version:   version,
		Modules:   req.Modules,
		IsActive:  false,
		CreatedBy: &adminID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Save(ctx, cfg); err != nil {
		return nil, errs.ErrInternal
	}
	return cfg, nil
}

// Activate Admin 激活某版本。
func (s *Service) Activate(ctx context.Context, id int64, pageKey string) error {
	if err := s.repo.Activate(ctx, id, pageKey); err != nil {
		return errs.ErrInternal
	}
	return nil
}
