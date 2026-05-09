package decorate

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gorm.io/gorm"
)

// mockRepo 实现 PageConfigRepo 接口，用于纯内存单元测试。
type mockRepo struct {
	getActiveErr error
	getActiveCfg *PageConfig
	savedCfg     *PageConfig
	saveErr      error
	listResult   []PageConfig
}

func (m *mockRepo) GetActive(_ context.Context, _ string) (*PageConfig, error) {
	return m.getActiveCfg, m.getActiveErr
}

func (m *mockRepo) ListByKey(_ context.Context, _ string) ([]PageConfig, error) {
	return m.listResult, nil
}

func (m *mockRepo) Save(_ context.Context, cfg *PageConfig) error {
	m.savedCfg = cfg
	return m.saveErr
}

func (m *mockRepo) Activate(_ context.Context, _ int64, _ string) error { return nil }

// rawJSON 辅助函数，快速构建 json.RawMessage。
func rawJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// TestService_Save_RejectBannerType 测试 Save() 拒绝不支持的 banner 模块类型。
// Given: 请求中包含 type=banner 的模块
// When:  调用 Save()
// Then:  返回含 "不支持的模块类型" 的错误，repo.Save 未被调用
func TestService_Save_RejectBannerType(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	req := SaveConfigReq{
		PageKey: "home",
		Modules: []PageModule{
			{Type: "banner", Data: rawJSON(map[string]any{"images": []string{"http://example.com/a.jpg"}})},
		},
	}

	_, err := svc.Save(context.Background(), 1, req)
	if err == nil {
		t.Fatal("期望返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "不支持的模块类型") {
		t.Errorf("错误消息不匹配，got: %v", err.Error())
	}
	if repo.savedCfg != nil {
		t.Error("repo.Save 不应被调用")
	}
}

// TestService_Save_AcceptValidTypes 测试 Save() 接受合法模块类型。
// Given: 请求中包含 product_list / category_entry / rich_text 三种类型
// When:  调用 Save()
// Then:  成功返回 PageConfig，repo.Save 被调用一次
func TestService_Save_AcceptValidTypes(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	req := SaveConfigReq{
		PageKey: "home",
		Modules: []PageModule{
			{Type: "product_list", Data: rawJSON(map[string]any{"limit": 10})},
			{Type: "category_entry", Data: rawJSON(map[string]any{"ids": []int{1, 2, 3}})},
			{Type: "rich_text", Data: rawJSON(map[string]string{"content": "<p>hello</p>"})},
		},
	}

	cfg, err := svc.Save(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("返回 PageConfig 不应为 nil")
	}
	if cfg.PageKey != "home" {
		t.Errorf("PageKey 不匹配，got: %s", cfg.PageKey)
	}
	if repo.savedCfg == nil {
		t.Error("repo.Save 应被调用一次")
	}
}

// TestService_Save_RichTextXSSSanitize 测试 Save() 对 rich_text 内容做 XSS 净化。
// Given: rich_text 模块 content 含 <script>alert(1)</script><p>hello</p>
// When:  调用 Save()
// Then:  repo.Save 被调用，传入的 rich_text content 不含 <script> 标签
func TestService_Save_RichTextXSSSanitize(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	maliciousContent := "<script>alert(1)</script><p>hello</p>"
	req := SaveConfigReq{
		PageKey: "home",
		Modules: []PageModule{
			{Type: "rich_text", Data: rawJSON(map[string]string{"content": maliciousContent})},
		},
	}

	cfg, err := svc.Save(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("期望成功，实际错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("返回 PageConfig 不应为 nil")
	}
	if repo.savedCfg == nil {
		t.Fatal("repo.Save 应被调用")
	}

	// 从保存的 cfg 中提取 rich_text 模块的 content 字段
	var found bool
	for _, m := range repo.savedCfg.Modules {
		if m.Type != "rich_text" {
			continue
		}
		found = true
		var d struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(m.Data, &d); err != nil {
			t.Fatalf("解析 rich_text data 失败: %v", err)
		}
		if strings.Contains(d.Content, "<script") {
			t.Errorf("XSS 净化失败，content 仍含 <script>，got: %s", d.Content)
		}
	}
	if !found {
		t.Error("保存的 Modules 中未找到 rich_text 类型")
	}
}

// TestService_GetActivePage_NoActiveVersion 测试无激活版本时返回空 Modules。
// Given: repo.GetActive 返回 gorm.ErrRecordNotFound
// When:  调用 GetActivePage("home")
// Then:  返回 &PageConfig{PageKey: "home", Modules: Modules{}}, err == nil
func TestService_GetActivePage_NoActiveVersion(t *testing.T) {
	repo := &mockRepo{
		getActiveErr: gorm.ErrRecordNotFound,
		getActiveCfg: nil,
	}
	svc := NewService(repo)

	cfg, err := svc.GetActivePage(context.Background(), "home")
	if err != nil {
		t.Fatalf("期望 nil 错误，实际: %v", err)
	}
	if cfg == nil {
		t.Fatal("返回 PageConfig 不应为 nil")
	}
	if cfg.PageKey != "home" {
		t.Errorf("PageKey 不匹配，got: %s", cfg.PageKey)
	}
	if cfg.Modules == nil {
		t.Error("Modules 不应为 nil，应为空切片")
	}
	if len(cfg.Modules) != 0 {
		t.Errorf("Modules 应为空，got: %v", cfg.Modules)
	}
}
