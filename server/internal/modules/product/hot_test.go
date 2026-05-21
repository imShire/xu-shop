package product_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// errProductRepo 是返回 List 错误的极简 ProductRepo 实现，仅用于 RepoError 用例。
type errProductRepo struct {
	mockProductRepo
	listErr error
}

func (e *errProductRepo) List(_ context.Context, filter product.ProductFilter) ([]product.Product, int64, error) {
	e.lastFilter = filter
	return nil, 0, e.listErr
}

// newHotTestService 构造带可控 redis 的 Service。返回 svc / miniredis 实例 / rdb。
func newHotTestService(t *testing.T, repo product.ProductRepo) (*product.Service, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svc := product.NewService(
		repo,
		newMockCategoryRepo(),
		&mockSKURepo{},
		&mockFavoriteRepo{},
		&mockViewHistoryRepo{},
		rdb,
		&mockOSSClient{baseURL: "https://cdn.example.com"},
		nil,
	)
	return svc, mr, rdb
}

// TestListHotProducts_LimitClamp 验证 limit 边界 clamp 行为。
// Given: 不同 limit 入参 (0 / 100 / -5 / 30)
// When:  调用 ListHotProducts
// Then:  传入 repo 的 PageSize 分别为 20 / 50 / 20 / 30
func TestListHotProducts_LimitClamp(t *testing.T) {
	cases := []struct {
		name     string
		limit    int
		expected int
	}{
		{"zero_default_20", 0, 20},
		{"over_max_clamp_50", 100, 50},
		{"negative_default_20", -5, 20},
		{"valid_keep_30", 30, 30},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newMockProductRepo()
			repo.listResp = []product.Product{}
			repo.listTotal = 0
			svc, _, _ := newHotTestService(t, repo)

			_, err := svc.ListHotProducts(context.Background(), c.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.lastFilter.PageSize != c.expected {
				t.Errorf("limit=%d → expect PageSize=%d, got %d", c.limit, c.expected, repo.lastFilter.PageSize)
			}
		})
	}
}

// TestListHotProducts_CacheHit 命中缓存时直接返回，不触发 repo.List。
// Given: Redis 中已存在合法 JSON 缓存
// When:  调用 ListHotProducts(limit=20)
// Then:  返回缓存内容，repo.List 未被调用（lastFilter 保持零值）
func TestListHotProducts_CacheHit(t *testing.T) {
	repo := newMockProductRepo()
	svc, mr, _ := newHotTestService(t, repo)

	cached := []product.ProductResp{{Title: "cached-product"}}
	data, _ := json.Marshal(cached)
	mr.Set(fmt.Sprintf("hot:products:limit:%d", 20), string(data))

	resp, err := svc.ListHotProducts(context.Background(), 0) // 0 → 20
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 || resp[0].Title != "cached-product" {
		t.Fatalf("expected cache hit content, got %+v", resp)
	}
	// repo.List 未被调用：lastFilter 应该为零值
	if repo.lastFilter.PageSize != 0 || repo.lastFilter.Status != "" {
		t.Errorf("repo.List should not be called on cache hit, got filter=%+v", repo.lastFilter)
	}
}

// TestListHotProducts_CacheMiss_FallbackToDB 缓存未命中走 DB 并使用 popular 排序。
// Given: Redis 中无缓存
// When:  调用 ListHotProducts(limit=15)
// Then:  repo.List 被调用，filter.Status=onsale / Sort=popular / PageSize=15
func TestListHotProducts_CacheMiss_FallbackToDB(t *testing.T) {
	repo := newMockProductRepo()
	repo.listResp = []product.Product{{ID: 1, Title: "from-db", Status: "onsale"}}
	repo.listTotal = 1
	svc, _, _ := newHotTestService(t, repo)

	resp, err := svc.ListHotProducts(context.Background(), 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 || resp[0].Title != "from-db" {
		t.Fatalf("expected db content, got %+v", resp)
	}
	if repo.lastFilter.Status != "onsale" {
		t.Errorf("expected status=onsale, got %q", repo.lastFilter.Status)
	}
	if repo.lastFilter.Sort != "popular" {
		t.Errorf("expected sort=popular, got %q", repo.lastFilter.Sort)
	}
	if repo.lastFilter.PageSize != 15 {
		t.Errorf("expected PageSize=15, got %d", repo.lastFilter.PageSize)
	}
	if repo.lastFilter.Page != 1 {
		t.Errorf("expected Page=1, got %d", repo.lastFilter.Page)
	}
}

// TestListHotProducts_RedisError_FallbackToDB Redis 故障时降级到 DB，不阻塞请求。
// Given: 提前关闭 miniredis，所有 Redis 操作返回非 Nil 错误
// When:  调用 ListHotProducts
// Then:  仍返回成功，repo.List 被调用
func TestListHotProducts_RedisError_FallbackToDB(t *testing.T) {
	repo := newMockProductRepo()
	repo.listResp = []product.Product{{ID: 2, Title: "fallback", Status: "onsale"}}
	repo.listTotal = 1
	svc, mr, _ := newHotTestService(t, repo)

	// 关闭 miniredis 模拟 Redis 故障
	mr.Close()

	resp, err := svc.ListHotProducts(context.Background(), 10)
	if err != nil {
		t.Fatalf("redis 故障不应阻塞，err=%v", err)
	}
	if len(resp) != 1 || resp[0].Title != "fallback" {
		t.Fatalf("expected db fallback content, got %+v", resp)
	}
	if repo.lastFilter.Sort != "popular" || repo.lastFilter.Status != "onsale" {
		t.Errorf("filter not applied on fallback: %+v", repo.lastFilter)
	}
}

// TestListHotProducts_RepoError repo 出错时返回 ErrInternal。
// Given: repo.List 返回 error
// When:  调用 ListHotProducts
// Then:  返回 errs.ErrInternal
func TestListHotProducts_RepoError(t *testing.T) {
	repo := &errProductRepo{listErr: errors.New("db down")}
	svc, _, _ := newHotTestService(t, repo)

	_, err := svc.ListHotProducts(context.Background(), 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errs.ErrInternal) && err != errs.ErrInternal {
		t.Errorf("expected errs.ErrInternal, got %v", err)
	}
}
