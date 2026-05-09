package decorate

import (
	"encoding/json"
	"time"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// PageModuleResp 单个页面模块响应（保持 json.RawMessage 以便透传）。
type PageModuleResp struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// PageConfigResp 页面装修配置响应 DTO（ID 序列化为字符串）。
type PageConfigResp struct {
	ID        types.Int64Str   `json:"id"`
	PageKey   string           `json:"page_key"`
	Version   int              `json:"version"`
	Modules   []PageModuleResp `json:"modules"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
}

// newPageConfigResp 将 entity 转换为响应 DTO。
func newPageConfigResp(c *PageConfig) PageConfigResp {
	mods := make([]PageModuleResp, len(c.Modules))
	for i, m := range c.Modules {
		mods[i] = PageModuleResp{Type: m.Type, Data: m.Data}
	}
	return PageConfigResp{
		ID:        types.Int64Str(c.ID),
		PageKey:   c.PageKey,
		Version:   c.Version,
		Modules:   mods,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
	}
}

// newPageConfigRespList 批量转换。
func newPageConfigRespList(list []PageConfig) []PageConfigResp {
	result := make([]PageConfigResp, len(list))
	for i := range list {
		result[i] = newPageConfigResp(&list[i])
	}
	return result
}
