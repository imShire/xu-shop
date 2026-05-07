package cart

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/product"
	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// ---- mock CartRepo ----

type mockCartRepo struct {
	items       []CartItem
	details     []CartItemDetail
	upsertFn    func(userID, skuID int64, qty int, price int64) error
	countByUser int64
}

func (m *mockCartRepo) FindByUserID(_ context.Context, _ int64) ([]CartItemDetail, error) {
	return m.details, nil
}
func (m *mockCartRepo) FindByID(_ context.Context, id int64) (*CartItem, error) {
	for i := range m.items {
		if m.items[i].ID == id {
			return &m.items[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockCartRepo) Upsert(_ context.Context, userID, skuID int64, qty int, price int64) error {
	if m.upsertFn != nil {
		return m.upsertFn(userID, skuID, qty, price)
	}
	return nil
}
func (m *mockCartRepo) Update(_ context.Context, _ int64, _ int, _ int64) error { return nil }
func (m *mockCartRepo) Delete(_ context.Context, _ int64) error        { return nil }
func (m *mockCartRepo) BatchDelete(_ context.Context, _ []int64) error { return nil }
func (m *mockCartRepo) DeleteByUserAndIDs(_ context.Context, _ int64, _ []int64) error {
	return nil
}
func (m *mockCartRepo) CountByUser(_ context.Context, _ int64) (int64, error) {
	return m.countByUser, nil
}
func (m *mockCartRepo) FindByIDs(_ context.Context, ids []int64) ([]CartItem, error) {
	var res []CartItem
	for _, item := range m.items {
		for _, id := range ids {
			if item.ID == id {
				res = append(res, item)
			}
		}
	}
	return res, nil
}

// ---- mock SKURepo ----

type mockSKURepo struct {
	skus map[int64]product.SKU
}

func (m *mockSKURepo) FindByProduct(_ context.Context, _ int64) ([]product.SKU, error) {
	return nil, nil
}
func (m *mockSKURepo) FindByID(_ context.Context, id int64) (*product.SKU, error) {
	if sk, ok := m.skus[id]; ok {
		return &sk, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockSKURepo) BatchUpdatePrice(_ context.Context, _ []int64, _ int64) error { return nil }
func (m *mockSKURepo) FindByIDs(_ context.Context, ids []int64) ([]product.SKU, error) {
	var res []product.SKU
	for _, id := range ids {
		if sk, ok := m.skus[id]; ok {
			res = append(res, sk)
		}
	}
	return res, nil
}

// ---- mock ProductRepo ----

type mockProductRepo struct {
	products map[int64]product.Product
}

func (m *mockProductRepo) List(_ context.Context, _ product.ProductFilter) ([]product.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) FindByID(_ context.Context, id int64) (*product.Product, error) {
	if p, ok := m.products[id]; ok {
		return &p, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockProductRepo) FindWithSpecs(_ context.Context, _ int64) (*product.ProductDetail, error) {
	return nil, nil
}
func (m *mockProductRepo) Create(_ context.Context, _ *product.Product, _ []product.ProductSpec, _ []product.ProductSpecValue, _ []product.SKU) error {
	return nil
}
func (m *mockProductRepo) Update(_ context.Context, _ *product.Product) error { return nil }
func (m *mockProductRepo) SoftDelete(_ context.Context, _ int64) error        { return nil }
func (m *mockProductRepo) Copy(_ context.Context, _ int64) (*product.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) UpdateStatus(_ context.Context, _ int64, _ string) error { return nil }
func (m *mockProductRepo) BatchUpdateStatus(_ context.Context, _ []int64, _ string) error {
	return nil
}
func (m *mockProductRepo) UpdatePriceRange(_ context.Context, _ int64, _, _ int64) error {
	return nil
}
func (m *mockProductRepo) ReplaceSpecsAndSKUs(_ context.Context, _ int64, _ []product.ProductSpec, _ []product.ProductSpecValue, _ []product.SKU) error {
	return nil
}

func (m *mockProductRepo) FindByIDs(_ context.Context, ids []int64) ([]product.Product, error) {
	var result []product.Product
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProductRepo) ListAdmin(_ context.Context, _ product.ProductFilter) ([]product.ProductAdminRow, int64, error) {
	return nil, 0, nil
}

// ---- 测试 ----

func newTestService(repo CartRepo, skuCount int64) *Service {
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		10: {ID: 10, ProductID: 1, Status: "active", PriceCents: 1000, Stock: 100},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		1: {ID: 1, Status: "onsale"},
	}}
	cartRepo := &mockCartRepo{countByUser: skuCount}
	if repo != nil {
		cartRepo = repo.(*mockCartRepo)
	}
	return NewService(cartRepo, skuRepo, prodRepo, nil)
}

// TestAdd_MergeExisting 重复加同一 SKU 应触发 UPSERT（合并数量）。
func TestAdd_MergeExisting(t *testing.T) {
	upsertCalled := 0
	repo := &mockCartRepo{
		countByUser: 1, // 已有 1 行
		upsertFn: func(_, _ int64, _ int, _ int64) error {
			upsertCalled++
			return nil
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		10: {ID: 10, ProductID: 1, Status: "active", PriceCents: 1000, Stock: 100},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		1: {ID: 1, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	if err := svc.Add(context.Background(), 100, 10, 2); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if upsertCalled != 1 {
		t.Fatalf("expected upsert called 1 time, got %d", upsertCalled)
	}
}

// TestAdd_ExceedLimit 超过 100 行应被拒绝。
func TestAdd_ExceedLimit(t *testing.T) {
	repo := &mockCartRepo{countByUser: maxCartRows} // 已满
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		10: {ID: 10, ProductID: 1, Status: "active", PriceCents: 1000, Stock: 100},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		1: {ID: 1, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	err := svc.Add(context.Background(), 100, 10, 1)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	ae, ok := err.(*errs.AppError)
	if !ok || ae.Code != errs.ErrConflict.Code {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// TestPrecheck_PriceChanged 快照价格与当前价不一致时应返回冲突。
func TestPrecheck_PriceChanged(t *testing.T) {
	repo := &mockCartRepo{
		items: []CartItem{
			{ID: 1, UserID: 100, SkuID: 10, Qty: 2, SnapshotPriceCents: 900}, // 快照 900，当前 1000
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		10: {ID: 10, ProductID: 1, Status: "active", PriceCents: 1000, Stock: 100},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		1: {ID: 1, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	resp, err := svc.Precheck(context.Background(), 100, []int64{1})
	if err != nil {
		t.Fatalf("Precheck error: %v", err)
	}
	if resp.OK {
		t.Fatal("expected OK=false")
	}
	if len(resp.Conflicts) == 0 {
		t.Fatal("expected conflicts")
	}
	if resp.Conflicts[0].Reason != "price_changed" {
		t.Fatalf("expected price_changed, got %s", resp.Conflicts[0].Reason)
	}
}

// TestPrecheck_StockInsufficient 库存不足时应返回 stock_insufficient 冲突。
func TestPrecheck_StockInsufficient(t *testing.T) {
	repo := &mockCartRepo{
		items: []CartItem{
			{ID: 2, UserID: 100, SkuID: 20, Qty: 10, SnapshotPriceCents: 500},
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		// Stock=5, LockedStock=3 → available=2 < qty=10
		20: {ID: 20, ProductID: 2, Status: "active", PriceCents: 500, Stock: 5, LockedStock: 3},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		2: {ID: 2, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	resp, err := svc.Precheck(context.Background(), 100, []int64{2})
	if err != nil {
		t.Fatalf("Precheck error: %v", err)
	}
	if resp.OK {
		t.Fatal("expected OK=false")
	}
	found := false
	for _, c := range resp.Conflicts {
		if c.Reason == "stock_insufficient" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected stock_insufficient conflict, got %+v", resp.Conflicts)
	}
}

// TestPrecheck_ProductOffsale 下架商品应在 precheck 中被阻断。
func TestPrecheck_ProductOffsale(t *testing.T) {
	repo := &mockCartRepo{
		items: []CartItem{
			{ID: 3, UserID: 100, SkuID: 30, Qty: 1, SnapshotPriceCents: 800},
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		30: {ID: 30, ProductID: 3, Status: "active", PriceCents: 800, Stock: 10},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		3: {ID: 3, Status: "draft"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	resp, err := svc.Precheck(context.Background(), 100, []int64{3})
	if err != nil {
		t.Fatalf("Precheck error: %v", err)
	}
	if resp.OK {
		t.Fatal("expected OK=false")
	}
	if len(resp.Conflicts) == 0 || resp.Conflicts[0].Reason != "product_offsale" {
		t.Fatalf("expected product_offsale, got %+v", resp.Conflicts)
	}
}

// TestUpdate_RejectsQtyAboveAvailable 数量修改不能超过当前可售库存。
func TestUpdate_RejectsQtyAboveAvailable(t *testing.T) {
	repo := &mockCartRepo{
		items: []CartItem{
			{ID: 4, UserID: 100, SkuID: 40, Qty: 1, SnapshotPriceCents: 1000},
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		40: {ID: 40, ProductID: 4, Status: "active", PriceCents: 1000, Stock: 5, LockedStock: 2},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		4: {ID: 4, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	err := svc.Update(context.Background(), 4, 100, 4)
	if err == nil {
		t.Fatal("expected stock limit error")
	}
	ae, ok := err.(*errs.AppError)
	if !ok || ae.Code != errs.ErrParam.Code {
		t.Fatalf("expected ErrParam, got %v", err)
	}
}

func TestPrecheck_NoConflictsReturnsEmptySlice(t *testing.T) {
	repo := &mockCartRepo{
		items: []CartItem{
			{ID: 3, UserID: 100, SkuID: 30, Qty: 1, SnapshotPriceCents: 1200},
		},
	}
	skuRepo := &mockSKURepo{skus: map[int64]product.SKU{
		30: {ID: 30, ProductID: 3, Status: "active", PriceCents: 1200, Stock: 10},
	}}
	prodRepo := &mockProductRepo{products: map[int64]product.Product{
		3: {ID: 3, Status: "onsale"},
	}}
	svc := NewService(repo, skuRepo, prodRepo, nil)

	resp, err := svc.Precheck(context.Background(), 100, []int64{3})
	if err != nil {
		t.Fatalf("Precheck error: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected OK=true")
	}
	if resp.Conflicts == nil {
		t.Fatal("expected conflicts to be an empty slice, got nil")
	}
	if len(resp.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %+v", resp.Conflicts)
	}
}

func TestList_ReturnsProductVOAndFallsBackToMainImage(t *testing.T) {
	repo := &mockCartRepo{
		details: []CartItemDetail{
			{
				ID:                   1,
				UserID:               100,
				SkuID:                10,
				Qty:                  2,
				SnapshotPriceCents:   900,
				SkuPriceCents:        1000,
				SkuStatus:            "active",
				SkuStock:             100,
				SkuLockedStock:       0,
				SkuImage:             "",
				SkuAttrs:             []byte(`["红色","L"]`),
				ProductID:            1,
				ProductCategoryID:    9,
				ProductTitle:         "测试商品",
				ProductSubtitle:      "副标题",
				ProductMainImage:     "https://cdn.example.com/main.jpg",
				ProductTags:          []byte(`["新品"]`),
				ProductStatus:        "onsale",
				ProductSales:         22,
				ProductPriceMinCents: 1000,
				ProductPriceMaxCents: 2000,
			},
		},
	}
	svc := newTestService(repo, 0)

	resp, err := svc.List(context.Background(), 100)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	item := resp.Items[0]
	if item.Product.MainImage != "https://cdn.example.com/main.jpg" {
		t.Fatalf("expected product main image, got %s", item.Product.MainImage)
	}
	if item.SkuImage != "https://cdn.example.com/main.jpg" {
		t.Fatalf("expected sku image fallback to product main image, got %s", item.SkuImage)
	}
	if item.Product.Title != "测试商品" {
		t.Fatalf("expected product title, got %s", item.Product.Title)
	}
}
