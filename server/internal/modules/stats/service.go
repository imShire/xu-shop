package stats

import (
	"context"
	"fmt"
	"io"
	"time"

	pkgcsv "github.com/xushop/xu-shop/internal/pkg/csv"
)

// Service 数据看板服务。
type Service struct {
	repo StatsRepo
}

// NewService 构造 Service。
func NewService(repo StatsRepo) *Service {
	return &Service{repo: repo}
}

// GetOverview 总览核心指标（含同比）。
func (s *Service) GetOverview(ctx context.Context, from, to time.Time) (*OverviewResp, error) {
	rows, err := s.repo.GetDailyRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats: get daily range: %w", err)
	}

	var resp OverviewResp
	for _, r := range rows {
		resp.PaidOrderCount += r.PaidOrderCount
		resp.PaidAmountCents += r.PaidAmountCents
		resp.RefundAmountCents += r.RefundAmountCents
		resp.NetAmountCents += r.NetAmountCents
		resp.PaidUserCount += r.PaidUserCount
		resp.NewUserCount += r.NewUserCount
	}

	// 同比（前一年同期）
	diff := to.Sub(from)
	prevFrom := from.AddDate(-1, 0, 0)
	prevTo := to.AddDate(-1, 0, 0)
	_ = diff

	prevRows, err := s.repo.GetDailyRange(ctx, prevFrom, prevTo)
	if err == nil {
		var prevOrders int
		var prevAmount int64
		for _, r := range prevRows {
			prevOrders += r.PaidOrderCount
			prevAmount += r.PaidAmountCents
		}
		if prevOrders > 0 {
			resp.OrderCountYoY = float64(resp.PaidOrderCount-prevOrders) / float64(prevOrders)
		}
		if prevAmount > 0 {
			resp.AmountYoY = float64(resp.PaidAmountCents-prevAmount) / float64(prevAmount)
		}
	}
	return &resp, nil
}

// GetSalesTrend 销售趋势折线图。
func (s *Service) GetSalesTrend(ctx context.Context, from, to time.Time) ([]DailyPoint, error) {
	rows, err := s.repo.GetDailyRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats: get sales trend: %w", err)
	}

	points := make([]DailyPoint, len(rows))
	for i, r := range rows {
		points[i] = DailyPoint{
			Date:        r.Date,
			AmountCents: r.PaidAmountCents,
			OrderCount:  r.PaidOrderCount,
		}
	}
	return points, nil
}

// GetCategoryPie 商品分类占比（Top5+其他）。
func (s *Service) GetCategoryPie(ctx context.Context, from, to time.Time) ([]CategoryShare, error) {
	return s.repo.CategoryPie(ctx, from, to)
}

// GetTopProducts 商品销售排行。
func (s *Service) GetTopProducts(ctx context.Context, from, to time.Time, page, size int) ([]TopProductRow, int64, error) {
	return s.repo.TopProducts(ctx, from, to, page, size)
}

// GetUserStats 用户统计。
func (s *Service) GetUserStats(ctx context.Context, from, to time.Time) (*UserStatsResp, error) {
	rows, err := s.repo.GetDailyRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats: get user stats: %w", err)
	}
	var resp UserStatsResp
	for _, r := range rows {
		resp.NewUserCount += r.NewUserCount
		resp.PaidUserCount += r.PaidUserCount
	}
	return &resp, nil
}

// GetChannelStats 渠道统计。
func (s *Service) GetChannelStats(ctx context.Context, from, to time.Time) ([]ChannelStatsRow, error) {
	return s.repo.ChannelStats(ctx, from, to)
}

// GetWorkbench 工作台实时概览。
func (s *Service) GetWorkbench(ctx context.Context) (*WorkbenchResp, error) {
	return s.repo.GetWorkbench(ctx)
}

// ExportProducts 导出商品销售 CSV。
func (s *Service) ExportProducts(ctx context.Context, w io.Writer, from, to time.Time) error {
	rows, _, err := s.repo.TopProducts(ctx, from, to, 1, 10000)
	if err != nil {
		return fmt.Errorf("stats: export products: %w", err)
	}

	sw := pkgcsv.NewSafeWriter(w)
	_ = sw.Write([]string{"商品ID", "商品名称", "销量", "销售额(分)"})
	for _, r := range rows {
		_ = sw.Write([]string{
			fmt.Sprintf("%d", r.ProductID),
			r.ProductName,
			fmt.Sprintf("%d", r.Qty),
			fmt.Sprintf("%d", r.AmountCents),
		})
	}
	sw.Flush()
	return sw.Error()
}

// AggregateDaily 聚合指定日期数据（幂等 UPSERT）。
func (s *Service) AggregateDaily(ctx context.Context, date time.Time) error {
	if err := s.repo.AggregateFromOrders(ctx, date); err != nil {
		return fmt.Errorf("stats: aggregate orders: %w", err)
	}
	if err := s.repo.AggregateProductFromItems(ctx, date); err != nil {
		return fmt.Errorf("stats: aggregate products: %w", err)
	}
	return nil
}
