package product_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

func init() {
	snowflake.Init(1)
}

// ---- mock repos ----

type mockCategoryRepo struct {
	categories map[int64]*product.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[int64]*product.Category)}
}

func (m *mockCategoryRepo) FindAll(_ context.Context) ([]product.Category, error) {
	list := make([]product.Category, 0, len(m.categories))
	for _, c := range m.categories {
		list = append(list, *c)
	}
	return list, nil
}

func (m *mockCategoryRepo) FindByID(_ context.Context, id int64) (*product.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return c, nil
}

func (m *mockCategoryRepo) Create(_ context.Context, c *product.Category) error {
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) Update(_ context.Context, c *product.Category) error {
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) SoftDelete(_ context.Context, id int64) error {
	delete(m.categories, id)
	return nil
}

type mockProductRepo struct {
	products   map[int64]*product.ProductDetail
	listResp   []product.Product
	listTotal  int64
	lastFilter product.ProductFilter
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{products: make(map[int64]*product.ProductDetail)}
}

func (m *mockProductRepo) List(_ context.Context, filter product.ProductFilter) ([]product.Product, int64, error) {
	m.lastFilter = filter
	return m.listResp, m.listTotal, nil
}

func (m *mockProductRepo) ListAdmin(_ context.Context, _ product.ProductFilter) ([]product.ProductAdminRow, int64, error) {
	return nil, 0, nil
}

func (m *mockProductRepo) FindByID(_ context.Context, id int64) (*product.Product, error) {
	d, ok := m.products[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return &d.Product, nil
}

func (m *mockProductRepo) FindWithSpecs(_ context.Context, id int64) (*product.ProductDetail, error) {
	d, ok := m.products[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return d, nil
}

func (m *mockProductRepo) Create(_ context.Context, p *product.Product, _ []product.ProductSpec, _ []product.ProductSpecValue, skus []product.SKU) error {
	m.products[p.ID] = &product.ProductDetail{
		Product: *p,
		SKUs:    skus,
	}
	return nil
}

func (m *mockProductRepo) Update(_ context.Context, p *product.Product) error {
	if d, ok := m.products[p.ID]; ok {
		d.Product = *p
	}
	return nil
}

func (m *mockProductRepo) SoftDelete(_ context.Context, id int64) error {
	delete(m.products, id)
	return nil
}

func (m *mockProductRepo) Copy(_ context.Context, productID int64) (*product.Product, error) {
	d, ok := m.products[productID]
	if !ok {
		return nil, errors.New("record not found")
	}
	newP := d.Product
	newP.ID = snowflake.NextID()
	newP.Title += "_copy"
	newP.Status = "draft"
	m.products[newP.ID] = &product.ProductDetail{Product: newP}
	return &newP, nil
}

func (m *mockProductRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	if d, ok := m.products[id]; ok {
		d.Product.Status = status
		if status == "onsale" {
			now := time.Now()
			d.Product.OnSaleAt = &now
		}
	}
	return nil
}

func (m *mockProductRepo) BatchUpdateStatus(_ context.Context, ids []int64, status string) error {
	for _, id := range ids {
		if d, ok := m.products[id]; ok {
			d.Product.Status = status
		}
	}
	return nil
}

func (m *mockProductRepo) UpdatePriceRange(_ context.Context, productID int64, min, max int64) error {
	if d, ok := m.products[productID]; ok {
		d.Product.PriceMinCents = min
		d.Product.PriceMaxCents = max
	}
	return nil
}

func (m *mockProductRepo) ReplaceSpecsAndSKUs(_ context.Context, productID int64, specs []product.ProductSpec, values []product.ProductSpecValue, skus []product.SKU) error {
	if d, ok := m.products[productID]; ok {
		d.Specs = specs
		d.Values = values
		d.SKUs = skus
	}
	return nil
}

func (m *mockProductRepo) FindByIDs(_ context.Context, ids []int64) ([]product.Product, error) {
	var result []product.Product
	for _, id := range ids {
		if d, ok := m.products[id]; ok {
			result = append(result, d.Product)
		}
	}
	return result, nil
}

type mockSKURepo struct{}

func (m *mockSKURepo) FindByProduct(_ context.Context, _ int64) ([]product.SKU, error) {
	return nil, nil
}

func (m *mockSKURepo) FindByID(_ context.Context, _ int64) (*product.SKU, error) {
	return nil, errors.New("not found")
}

func (m *mockSKURepo) BatchUpdatePrice(_ context.Context, _ []int64, _ int64) error {
	return nil
}

func (m *mockSKURepo) FindByIDs(_ context.Context, _ []int64) ([]product.SKU, error) {
	return nil, nil
}

type mockFavoriteRepo struct {
	items []product.FavoriteListItem
}

func (m *mockFavoriteRepo) Add(_ context.Context, _, _ int64) error    { return nil }
func (m *mockFavoriteRepo) Remove(_ context.Context, _, _ int64) error { return nil }
func (m *mockFavoriteRepo) List(_ context.Context, _ int64, _, _ int) ([]product.FavoriteListItem, int64, error) {
	return m.items, int64(len(m.items)), nil
}
func (m *mockFavoriteRepo) Exists(_ context.Context, _, _ int64) (bool, error) { return false, nil }

type mockViewHistoryRepo struct {
	items []product.ViewHistoryItem
}

func (m *mockViewHistoryRepo) Upsert(_ context.Context, _, _ int64) error { return nil }
func (m *mockViewHistoryRepo) List(_ context.Context, _ int64, _, _ int) ([]product.ViewHistoryItem, int64, error) {
	return m.items, int64(len(m.items)), nil
}
func (m *mockViewHistoryRepo) Clear(_ context.Context, _ int64) error                { return nil }
func (m *mockViewHistoryRepo) CountByUser(_ context.Context, _ int64) (int64, error) { return 0, nil }
func (m *mockViewHistoryRepo) DeleteOldest(_ context.Context, _ int64, _ int) error  { return nil }

// mockOSSClient mock 上传客户端。
type mockOSSClient struct {
	baseURL string
	fail    bool
}

func (m *mockOSSClient) UploadProductImage(_ context.Context, _ io.Reader, filename string) (string, error) {
	if m.fail {
		return "", errors.New("upload failed")
	}
	return m.baseURL + "/" + filename, nil
}

func (m *mockOSSClient) AllowedImageURLPrefix(_ context.Context) (string, error) {
	return m.baseURL, nil
}

func newTestService(t *testing.T, productRepo product.ProductRepo, categoryRepo product.CategoryRepo, ossClient product.ImageUploader) (*product.Service, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	svc := product.NewService(
		productRepo,
		categoryRepo,
		&mockSKURepo{},
		&mockFavoriteRepo{},
		&mockViewHistoryRepo{},
		rdb,
		ossClient,
		nil,
	)
	return svc, rdb
}

// TestAdminCreateProduct_XSSFiltered 验证 detail_html 中的 script 被净化。
func TestAdminCreateProduct_XSSFiltered(t *testing.T) {
	svc, _ := newTestService(t, newMockProductRepo(), newMockCategoryRepo(), &mockOSSClient{baseURL: "https://cdn.example.com"})

	req := product.CreateProductReq{
		CategoryID: 1,
		Title:      "测试商品",
		MainImage:  "https://cdn.example.com/test.jpg",
		DetailHTML: `<p>正常内容</p><script>alert('xss')</script>`,
		SKUs: []product.SKUReq{
			{PriceCents: 9900, Stock: 10},
		},
	}

	p, err := svc.AdminCreateProduct(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// detail_html 经 bluemonday 净化后不含 script
	detail, detailErr := svc.AdminGetProduct(context.Background(), p.ID)
	if detailErr != nil {
		t.Fatalf("get product: %v", detailErr)
	}
	if containsScript := containsStr(detail.DetailHTML, "<script>"); containsScript {
		t.Errorf("XSS not filtered: %s", detail.DetailHTML)
	}
}

// TestAdminCreateProduct_InvalidImageURL 验证 img src 不在 OSS 域返回 400。
func TestAdminCreateProduct_InvalidImageURL(t *testing.T) {
	svc, _ := newTestService(t, newMockProductRepo(), newMockCategoryRepo(), &mockOSSClient{baseURL: "https://cdn.example.com"})

	req := product.CreateProductReq{
		CategoryID: 1,
		Title:      "测试商品",
		MainImage:  "https://cdn.example.com/main.jpg",
		DetailHTML: `<img src="https://evil.com/hack.jpg" />`,
		SKUs: []product.SKUReq{
			{PriceCents: 9900, Stock: 10},
		},
	}

	_, err := svc.AdminCreateProduct(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for invalid image URL, got nil")
	}
}

// TestGetCategories_Tree 验证平铺转树形正确。
func TestGetCategories_Tree(t *testing.T) {
	catRepo := newMockCategoryRepo()
	// 构建测试数据：根分类 + 子分类
	root1 := &product.Category{ID: 1, ParentID: 0, Name: "根分类1", Status: "enabled"}
	root2 := &product.Category{ID: 2, ParentID: 0, Name: "根分类2", Status: "enabled"}
	child1 := &product.Category{ID: 3, ParentID: 1, Name: "子分类1", Status: "enabled"}
	child2 := &product.Category{ID: 4, ParentID: 1, Name: "子分类2", Status: "enabled"}
	catRepo.categories[1] = root1
	catRepo.categories[2] = root2
	catRepo.categories[3] = child1
	catRepo.categories[4] = child2

	svc, _ := newTestService(t, newMockProductRepo(), catRepo, &mockOSSClient{baseURL: "https://cdn.example.com"})

	tree, err := svc.GetCategories(context.Background())
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}

	if len(tree) != 2 {
		t.Errorf("expected 2 root nodes, got %d", len(tree))
	}

	// 找到 root1，验证其有 2 个子节点
	var root1Node *product.CategoryTreeNode
	for i := range tree {
		if tree[i].ID == 1 {
			root1Node = &tree[i]
			break
		}
	}
	if root1Node == nil {
		t.Fatal("root1 not found in tree")
	}
	if len(root1Node.Children) != 2 {
		t.Errorf("root1 expected 2 children, got %d", len(root1Node.Children))
	}
}

func TestListProducts_UsesClientSortAndOnsaleStatus(t *testing.T) {
	productRepo := newMockProductRepo()
	productRepo.listResp = []product.Product{
		{
			ID:            101,
			CategoryID:    9,
			Title:         "热门单品",
			MainImage:     "https://cdn.example.com/popular.jpg",
			Status:        "onsale",
			PriceMinCents: 1990,
			PriceMaxCents: 1990,
		},
	}
	productRepo.listTotal = 1

	svc, _ := newTestService(
		t,
		productRepo,
		newMockCategoryRepo(),
		&mockOSSClient{baseURL: "https://cdn.example.com"},
	)

	list, total, err := svc.ListProducts(context.Background(), product.ProductListReq{
		CategoryID: 9,
		Keyword:    "单品",
		Sort:       "popular",
		Page:       2,
		PageSize:   12,
	})
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("unexpected list shape: total=%d len=%d", total, len(list))
	}
	if productRepo.lastFilter.Status != "onsale" {
		t.Fatalf("expected onsale status, got %q", productRepo.lastFilter.Status)
	}
	if productRepo.lastFilter.Sort != "popular" {
		t.Fatalf("expected popular sort, got %q", productRepo.lastFilter.Sort)
	}
	if productRepo.lastFilter.Page != 2 || productRepo.lastFilter.PageSize != 12 {
		t.Fatalf("unexpected paging: %+v", productRepo.lastFilter)
	}
	if productRepo.lastFilter.CategoryID != 9 || productRepo.lastFilter.Keyword != "单品" {
		t.Fatalf("unexpected filter: %+v", productRepo.lastFilter)
	}
}

func TestListFavorites_ReturnsClientShape(t *testing.T) {
	svc := product.NewService(
		newMockProductRepo(),
		newMockCategoryRepo(),
		&mockSKURepo{},
		&mockFavoriteRepo{
			items: []product.FavoriteListItem{
				{
					ProductID:  1001,
					Title:      "收藏商品",
					Image:      "https://cdn.example.com/favorite.jpg",
					PriceCents: 29900,
					CreatedAt:  time.Unix(1710000000, 0),
				},
			},
		},
		&mockViewHistoryRepo{},
		redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
		&mockOSSClient{baseURL: "https://cdn.example.com"},
		nil,
	)

	list, total, err := svc.ListFavorites(context.Background(), 1, 1, 20)
	if err != nil {
		t.Fatalf("ListFavorites: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	if list[0].ProductID != 1001 || list[0].ID != 1001 {
		t.Fatalf("unexpected product ids: %+v", list[0])
	}
	if list[0].Image != "https://cdn.example.com/favorite.jpg" {
		t.Fatalf("unexpected image: %s", list[0].Image)
	}
}

func TestGetViewHistory_ReturnsClientShape(t *testing.T) {
	viewedAt := time.Unix(1710003600, 0)
	svc := product.NewService(
		newMockProductRepo(),
		newMockCategoryRepo(),
		&mockSKURepo{},
		&mockFavoriteRepo{},
		&mockViewHistoryRepo{
			items: []product.ViewHistoryItem{
				{
					ProductID:  2002,
					Title:      "历史商品",
					Image:      "https://cdn.example.com/history.jpg",
					PriceCents: 18800,
					ViewedAt:   viewedAt,
				},
			},
		},
		redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
		&mockOSSClient{baseURL: "https://cdn.example.com"},
		nil,
	)

	list, total, err := svc.GetViewHistory(context.Background(), 1, 1, 20)
	if err != nil {
		t.Fatalf("GetViewHistory: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	if list[0].ProductID != 2002 || !list[0].ViewedAt.Equal(viewedAt) {
		t.Fatalf("unexpected history item: %+v", list[0])
	}
}

func containsStr(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && (s != sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
