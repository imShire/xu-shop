package decorate

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/microcosm-cc/bluemonday"
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if len(req.Modules) > 20 {
		return nil, errs.ErrParam.WithMsg("模块数量不能超过 20")
	}
	// 校验模块类型（选项 B：banner / nav_icon 由独立 CRUD 管理，不允许进入 page_config）
	allowedTypes := map[string]bool{
		"product_list":   true,
		"category_entry": true,
		"rich_text":      true,
	}
	for _, m := range req.Modules {
		if !allowedTypes[m.Type] {
			return nil, errs.ErrParam.WithMsg("不支持的模块类型：" + m.Type)
		}
	}

	// 对 rich_text 内容做 XSS 净化
	sanitizer := bluemonday.UGCPolicy()
	sanitizer.AllowNoAttrs().OnElements("p", "br", "b", "strong", "i", "em", "u", "ul", "ol", "li", "h1", "h2", "h3", "a", "img")
	sanitizer.AllowAttrs("href").OnElements("a")
	sanitizer.AllowAttrs("src", "alt", "width", "height").OnElements("img")

	for i, m := range req.Modules {
		if m.Type == "rich_text" {
			var d struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(m.Data, &d); err == nil {
				d.Content = sanitizer.Sanitize(d.Content)
				cleaned, _ := json.Marshal(d)
				req.Modules[i].Data = json.RawMessage(cleaned)
			}
		}
	}

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
