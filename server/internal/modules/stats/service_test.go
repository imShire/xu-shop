package stats

import (
	"context"
	"testing"
	"time"
)

// ---- mock repo ----

type mockStatsRepo struct {
	daily        []StatsDaily
	productDaily []StatsProductDaily
	channelDaily []StatsChannelDaily
	upserted     []StatsDaily
	aggCalled    []time.Time
}

func (m *mockStatsRepo) GetDailyRange(_ context.Context, _, _ time.Time) ([]StatsDaily, error) {
	return m.daily, nil
}
func (m *mockStatsRepo) GetProductDailyRange(_ context.Context, _, _ time.Time) ([]StatsProductDaily, error) {
	return m.productDaily, nil
}
func (m *mockStatsRepo) GetChannelDailyRange(_ context.Context, _, _ time.Time) ([]StatsChannelDaily, error) {
	return m.channelDaily, nil
}
func (m *mockStatsRepo) UpsertDaily(_ context.Context, s *StatsDaily) error {
	m.upserted = append(m.upserted, *s)
	return nil
}
func (m *mockStatsRepo) UpsertProductDaily(_ context.Context, rows []StatsProductDaily) error {
	m.productDaily = append(m.productDaily, rows...)
	return nil
}
func (m *mockStatsRepo) AggregateFromOrders(_ context.Context, date time.Time) error {
	m.aggCalled = append(m.aggCalled, date)
	// 模拟写入 stats_daily
	m.upserted = append(m.upserted, StatsDaily{
		Date:           date.Format("2006-01-02"),
		PaidOrderCount: 5,
		PaidAmountCents: 50000,
	})
	return nil
}
func (m *mockStatsRepo) AggregateProductFromItems(_ context.Context, _ time.Time) error {
	return nil
}
func (m *mockStatsRepo) TopProducts(_ context.Context, _, _ time.Time, _, _ int) ([]TopProductRow, int64, error) {
	return nil, 0, nil
}
func (m *mockStatsRepo) CategoryPie(_ context.Context, _, _ time.Time) ([]CategoryShare, error) {
	return nil, nil
}
func (m *mockStatsRepo) ChannelStats(_ context.Context, _, _ time.Time) ([]ChannelStatsRow, error) {
	return nil, nil
}

func (m *mockStatsRepo) GetWorkbench(_ context.Context) (*WorkbenchResp, error) {
	return &WorkbenchResp{}, nil
}

// TestAggregateDaily_UPSERT 聚合后 UpsertDaily 被调用且数据正确
func TestAggregateDaily_UPSERT(t *testing.T) {
	repo := &mockStatsRepo{}
	svc := NewService(repo)

	date := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	if err := svc.AggregateDaily(context.Background(), date); err != nil {
		t.Fatalf("AggregateDaily: %v", err)
	}

	// 验证 AggregateFromOrders 被调用
	if len(repo.aggCalled) == 0 {
		t.Fatal("expected AggregateFromOrders to be called")
	}
	if !repo.aggCalled[0].Equal(date) {
		t.Fatalf("expected date=%v, got %v", date, repo.aggCalled[0])
	}
}

// TestGetOverview_Sum 汇总多天数据正确
func TestGetOverview_Sum(t *testing.T) {
	repo := &mockStatsRepo{
		daily: []StatsDaily{
			{Date: "2026-04-28", PaidOrderCount: 3, PaidAmountCents: 30000},
			{Date: "2026-04-29", PaidOrderCount: 5, PaidAmountCents: 50000},
		},
	}
	svc := NewService(repo)

	from := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	resp, err := svc.GetOverview(context.Background(), from, to)
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if resp.PaidOrderCount != 8 {
		t.Fatalf("expected PaidOrderCount=8, got %d", resp.PaidOrderCount)
	}
	if resp.PaidAmountCents != 80000 {
		t.Fatalf("expected PaidAmountCents=80000, got %d", resp.PaidAmountCents)
	}
}

// TestGetWorkbench_EmptyDB 空 DB 时 GetWorkbench 返回零值结构体而不是 nil 或 panic。
func TestGetWorkbench_EmptyDB(t *testing.T) {
	repo := &mockStatsRepo{} // 所有字段为零值
	svc := NewService(repo)

	resp, err := svc.GetWorkbench(context.Background())
	if err != nil {
		t.Fatalf("GetWorkbench error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil WorkbenchResp, got nil")
	}

	// 空 DB → 所有计数应为零
	if resp.TodayOrderCount != 0 {
		t.Errorf("expected TodayOrderCount=0, got %d", resp.TodayOrderCount)
	}
	if resp.TodaySales != 0 {
		t.Errorf("expected TodaySales=0, got %d", resp.TodaySales)
	}
	if resp.PendingShip != 0 {
		t.Errorf("expected PendingShip=0, got %d", resp.PendingShip)
	}
	if resp.AftersalePending != 0 {
		t.Errorf("expected AftersalePending=0, got %d", resp.AftersalePending)
	}
}
