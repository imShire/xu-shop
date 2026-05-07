package address_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/modules/address"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/wxlogin"
)

func init() {
	snowflake.Init(1)
}

// ---- mock AddressRepo ----

type mockAddressRepo struct {
	addrs  map[int64]*address.Address
	nextID int64
}

func newMockAddressRepo() *mockAddressRepo {
	return &mockAddressRepo{addrs: make(map[int64]*address.Address), nextID: 1}
}

func (m *mockAddressRepo) List(_ context.Context, userID int64) ([]address.Address, error) {
	var list []address.Address
	for _, a := range m.addrs {
		if a.UserID == userID {
			list = append(list, *a)
		}
	}
	return list, nil
}

func (m *mockAddressRepo) FindByID(_ context.Context, id, userID int64) (*address.Address, error) {
	if a, ok := m.addrs[id]; ok && a.UserID == userID {
		return a, nil
	}
	return nil, errors.New("record not found")
}

func (m *mockAddressRepo) Count(_ context.Context, userID int64) (int64, error) {
	var cnt int64
	for _, a := range m.addrs {
		if a.UserID == userID {
			cnt++
		}
	}
	return cnt, nil
}

func (m *mockAddressRepo) Create(_ context.Context, a *address.Address) error {
	m.addrs[a.ID] = a
	return nil
}

func (m *mockAddressRepo) Update(_ context.Context, id, _ int64, updates map[string]any) error {
	if a, ok := m.addrs[id]; ok {
		if v, ok := updates["is_default"].(bool); ok {
			a.IsDefault = v
		}
		if v, ok := updates["name"].(string); ok {
			a.Name = v
		}
	}
	return nil
}

func (m *mockAddressRepo) Delete(_ context.Context, id, _ int64) error {
	delete(m.addrs, id)
	return nil
}

func (m *mockAddressRepo) ClearDefault(_ context.Context, userID int64) error {
	for _, a := range m.addrs {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
	return nil
}

func (m *mockAddressRepo) SetDefault(_ context.Context, id, userID int64) error {
	for _, a := range m.addrs {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
	if a, ok := m.addrs[id]; ok {
		a.IsDefault = true
	}
	return nil
}

// ---- mock RegionRepo ----

type mockRegionRepo struct {
	listByParent map[string][]address.Region
}

func (m *mockRegionRepo) ListByParent(_ context.Context, parentCode string) ([]address.Region, error) {
	if m.listByParent == nil {
		return nil, nil
	}
	return m.listByParent[parentCode], nil
}
func (m *mockRegionRepo) FindByCode(_ context.Context, _ string) (*address.Region, error) {
	return nil, nil
}

// ---- helpers ----

func newTestRdb(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func newTestService(t *testing.T, addrRepo address.AddressRepo) *address.Service {
	t.Helper()
	rdb := newTestRdb(t)
	return address.NewService(addrRepo, &mockRegionRepo{}, rdb, wxlogin.NewMockClient())
}

func newTestServiceWithRegionRepo(t *testing.T, addrRepo address.AddressRepo, regionRepo address.RegionRepo) *address.Service {
	t.Helper()
	rdb := newTestRdb(t)
	return address.NewService(addrRepo, regionRepo, rdb, wxlogin.NewMockClient())
}

// ---- 测试用例 ----

func TestCreateAddress_Limit20(t *testing.T) {
	repo := newMockAddressRepo()
	svc := newTestService(t, repo)

	ctx := context.Background()
	const userID = int64(1001)

	// 插入 20 条
	for i := 0; i < 20; i++ {
		_, err := svc.Create(ctx, userID, address.CreateAddressReq{
			Name:   "测试用户",
			Phone:  "13800138000",
			Detail: "测试地址",
		})
		if err != nil {
			t.Fatalf("Create #%d error: %v", i+1, err)
		}
	}

	// 第 21 条应返回冲突错误
	_, err := svc.Create(ctx, userID, address.CreateAddressReq{
		Name:   "超出用户",
		Phone:  "13800138000",
		Detail: "超出地址",
	})
	if err == nil {
		t.Fatal("expected error when exceeding 20 address limit, got nil")
	}
}

func TestCreateAddress_DefaultUnique(t *testing.T) {
	repo := newMockAddressRepo()
	svc := newTestService(t, repo)

	ctx := context.Background()
	const userID = int64(1002)

	// 创建默认地址 1
	a1, err := svc.Create(ctx, userID, address.CreateAddressReq{
		Name: "地址1", Phone: "13800138001", Detail: "详情1", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("Create addr1 error: %v", err)
	}

	// 创建默认地址 2（应清除 addr1 的默认标记）
	a2, err := svc.Create(ctx, userID, address.CreateAddressReq{
		Name: "地址2", Phone: "13800138002", Detail: "详情2", IsDefault: true,
	})
	if err != nil {
		t.Fatalf("Create addr2 error: %v", err)
	}

	// 验证 addr1 已不是默认
	stored1 := repo.addrs[int64(a1.ID)]
	if stored1 != nil && stored1.IsDefault {
		t.Error("expected addr1 is_default=false after addr2 set as default")
	}
	// addr2 是默认
	stored2 := repo.addrs[int64(a2.ID)]
	if stored2 == nil || !stored2.IsDefault {
		t.Error("expected addr2 is_default=true")
	}
}

func TestSetDefault(t *testing.T) {
	repo := newMockAddressRepo()
	svc := newTestService(t, repo)

	ctx := context.Background()
	const userID = int64(1003)

	a1, _ := svc.Create(ctx, userID, address.CreateAddressReq{
		Name: "addr1", Phone: "13800138003", Detail: "d1", IsDefault: true,
	})
	a2, _ := svc.Create(ctx, userID, address.CreateAddressReq{
		Name: "addr2", Phone: "13800138004", Detail: "d2", IsDefault: false,
	})

	// 将 addr2 设为默认
	if err := svc.SetDefault(ctx, int64(a2.ID), userID); err != nil {
		t.Fatalf("SetDefault error: %v", err)
	}

	// 验证 addr1 不再是默认，addr2 是默认
	if repo.addrs[int64(a1.ID)].IsDefault {
		t.Error("expected addr1 is_default=false")
	}
	if !repo.addrs[int64(a2.ID)].IsDefault {
		t.Error("expected addr2 is_default=true")
	}
}

func TestListRegions_HasChildren(t *testing.T) {
	svc := newTestServiceWithRegionRepo(t, newMockAddressRepo(), &mockRegionRepo{
		listByParent: map[string][]address.Region{
			"": {
				{Code: "1", Name: "北京", Level: 1, HasChildren: true},
				{Code: "2", Name: "上海", Level: 1, HasChildren: true},
			},
			"1": {
				{Code: "72", ParentCode: ptr("1"), Name: "朝阳区", Level: 2, HasChildren: true},
				{Code: "2901", ParentCode: ptr("1"), Name: "昌平区", Level: 2, HasChildren: true},
			},
		},
	})

	root, err := svc.ListRegions(context.Background(), "")
	if err != nil {
		t.Fatalf("ListRegions root error: %v", err)
	}
	if len(root) != 2 || !root[0].HasChildren {
		t.Fatalf("unexpected root result: %+v", root)
	}

	children, err := svc.ListRegions(context.Background(), "1")
	if err != nil {
		t.Fatalf("ListRegions child error: %v", err)
	}
	if len(children) != 2 || children[0].ParentCode == nil || *children[0].ParentCode != "1" {
		t.Fatalf("unexpected child result: %+v", children)
	}
}

func ptr(v string) *string { return &v }
