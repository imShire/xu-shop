package inventory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/modules/inventory"
	"github.com/xushop/xu-shop/internal/pkg/stock"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

func init() {
	snowflake.Init(1)
}

// mockInventoryRepo 内存版 InventoryRepo。
type mockInventoryRepo struct {
	skus map[int64]*skuRow
	logs []inventory.InventoryLog
}

type skuRow struct {
	stock       int
	lockedStock int
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{skus: make(map[int64]*skuRow)}
}

func (m *mockInventoryRepo) seed(id int64, stock, locked int) {
	m.skus[id] = &skuRow{stock: stock, lockedStock: locked}
}

func (m *mockInventoryRepo) GetSKUStock(_ context.Context, skuID int64) (int, int, error) {
	r, ok := m.skus[skuID]
	if !ok {
		return 0, 0, errors.New("not found")
	}
	return r.stock, r.lockedStock, nil
}

func (m *mockInventoryRepo) AdjustDB(_ context.Context, skuID int64, change int, _ string, _ string, _ int64, _ int64, _ string) error {
	r, ok := m.skus[skuID]
	if !ok {
		return errors.New("not found")
	}
	r.stock += change
	if r.stock < 0 {
		r.stock = 0
	}
	return nil
}

func (m *mockInventoryRepo) DeductDB(_ context.Context, skuID, qty int, _ int64) error {
	r, ok := m.skus[int64(skuID)]
	if !ok {
		return errors.New("not found")
	}
	r.stock -= qty
	r.lockedStock -= qty
	return nil
}

func (m *mockInventoryRepo) LockDB(_ context.Context, skuID, qty int, _ int64) error {
	r, ok := m.skus[int64(skuID)]
	if !ok {
		return errors.New("not found")
	}
	r.lockedStock += qty
	return nil
}

func (m *mockInventoryRepo) UnlockDB(_ context.Context, skuID, qty int, _ int64) error {
	r, ok := m.skus[int64(skuID)]
	if !ok {
		return errors.New("not found")
	}
	r.lockedStock -= qty
	if r.lockedStock < 0 {
		r.lockedStock = 0
	}
	return nil
}

func (m *mockInventoryRepo) ListLogs(_ context.Context, _ inventory.LogFilter, _, _ int) ([]inventory.InventoryLog, int64, error) {
	return m.logs, int64(len(m.logs)), nil
}

func (m *mockInventoryRepo) ListAlerts(_ context.Context, _ string, _, _ int) ([]inventory.LowStockAlert, int64, error) {
	return nil, 0, nil
}

func (m *mockInventoryRepo) MarkAlertRead(_ context.Context, _, _ int64) error { return nil }

func (m *mockInventoryRepo) MarkAllAlertsRead(_ context.Context, _ int64) error { return nil }

func (m *mockInventoryRepo) CreateAlert(_ context.Context, _ *inventory.LowStockAlert) error {
	return nil
}

func (m *mockInventoryRepo) GetSKUThreshold(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

func (m *mockInventoryRepo) HasUnreadAlert(_ context.Context, _ int64) (bool, error) {
	return false, nil
}

func (m *mockInventoryRepo) FindAllSKUStocks(_ context.Context) ([]inventory.SKUStockRow, error) {
	rows := make([]inventory.SKUStockRow, 0, len(m.skus))
	for id, r := range m.skus {
		rows = append(rows, inventory.SKUStockRow{ID: id, Stock: r.stock, LockedStock: r.lockedStock})
	}
	return rows, nil
}

func newTestService(t *testing.T, repo inventory.InventoryRepo) (*inventory.Service, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	stockClient := stock.New(rdb)

	svc := inventory.NewService(repo, stockClient, rdb, nil)
	return svc, rdb
}

// TestAdjust_In 增库存后 DB 和 Redis 都正确。
func TestAdjust_In(t *testing.T) {
	repo := newMockInventoryRepo()
	repo.seed(1001, 50, 5) // stock=50, locked=5

	svc, rdb := newTestService(t, repo)
	ctx := context.Background()

	// 预加载 Redis（available = 50 - 5 = 45）
	stockClient := stock.New(rdb)
	_ = stockClient.Load(ctx, 1001, 45)

	err := svc.Adjust(ctx, 100, 1001, "in", 20, "采购入库")
	if err != nil {
		t.Fatalf("Adjust in: %v", err)
	}

	// 验证 DB
	dbStock, _, _ := repo.GetSKUStock(ctx, 1001)
	if dbStock != 70 {
		t.Errorf("DB stock expected 70, got %d", dbStock)
	}

	// 验证 Redis
	redisStock, _ := stockClient.Get(ctx, 1001)
	if redisStock != 65 { // 45 + 20
		t.Errorf("Redis stock expected 65, got %d", redisStock)
	}
}

// TestAdjust_Out_Insufficient 减到负时被拒绝。
func TestAdjust_Out_Insufficient(t *testing.T) {
	repo := newMockInventoryRepo()
	repo.seed(1002, 10, 8) // stock=10, locked=8，可减只有 2

	svc, _ := newTestService(t, repo)
	ctx := context.Background()

	// 尝试减 5（new=5 < locked=8，应该报错）
	err := svc.Adjust(ctx, 100, 1002, "out", 5, "测试减库存")
	if err == nil {
		t.Fatal("expected error for insufficient stock, got nil")
	}
}

// TestAdjust_Set_BelowLocked set < locked_stock 被拒绝。
func TestAdjust_Set_BelowLocked(t *testing.T) {
	repo := newMockInventoryRepo()
	repo.seed(1003, 100, 30) // locked=30

	svc, _ := newTestService(t, repo)
	ctx := context.Background()

	// set=20 < locked=30，应该报错
	err := svc.Adjust(ctx, 100, 1003, "set", 20, "手动设置")
	if err == nil {
		t.Fatal("expected error when set < locked_stock, got nil")
	}
}
