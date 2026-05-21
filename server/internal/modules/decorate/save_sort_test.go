package decorate

import (
	"context"
	"strings"
	"testing"
)

// TestSave_ProductList_InvalidSort 验证 product_list.data.sort 白名单拦截非法值。
// Given: modules=[{type:product_list, data:{sort:"invalid"}}]
// When:  调用 Save
// Then:  返回 400 错误，含 "不支持的商品排序"，repo.Save 未被调用
func TestSave_ProductList_InvalidSort(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	req := SaveConfigReq{
		PageKey: "home",
		Modules: []PageModule{
			{Type: "product_list", Data: rawJSON(map[string]any{"sort": "invalid"})},
		},
	}

	_, err := svc.Save(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for invalid sort, got nil")
	}
	if !strings.Contains(err.Error(), "不支持的商品排序") {
		t.Errorf("expected sort whitelist error, got: %v", err)
	}
	if repo.savedCfg != nil {
		t.Error("repo.Save 不应被调用")
	}
}

// TestSave_ProductList_ValidHotSort 验证 sort=hot 通过白名单。
// Given: modules=[{type:product_list, data:{sort:"hot"}}]
// When:  调用 Save
// Then:  返回成功，repo.Save 被调用
func TestSave_ProductList_ValidHotSort(t *testing.T) {
	for _, sort := range []string{"hot", "popular", "latest", "price_asc", "price_desc"} {
		t.Run(sort, func(t *testing.T) {
			repo := &mockRepo{}
			svc := NewService(repo)
			req := SaveConfigReq{
				PageKey: "home",
				Modules: []PageModule{
					{Type: "product_list", Data: rawJSON(map[string]any{"sort": sort, "limit": 10})},
				},
			}
			cfg, err := svc.Save(context.Background(), 1, req)
			if err != nil {
				t.Fatalf("sort=%s expected ok, got err=%v", sort, err)
			}
			if cfg == nil || repo.savedCfg == nil {
				t.Errorf("sort=%s repo.Save should be called", sort)
			}
		})
	}
}

// TestSave_ProductList_NoSortField 验证 data 中无 sort 字段时不报错。
// Given: modules=[{type:product_list, data:{limit:10}}] (无 sort)
// When:  调用 Save
// Then:  返回成功
func TestSave_ProductList_NoSortField(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	req := SaveConfigReq{
		PageKey: "home",
		Modules: []PageModule{
			{Type: "product_list", Data: rawJSON(map[string]any{"limit": 10})},
		},
	}
	cfg, err := svc.Save(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("expected ok when sort missing, got err=%v", err)
	}
	if cfg == nil || repo.savedCfg == nil {
		t.Error("repo.Save should be called")
	}
}
