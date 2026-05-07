package stats

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/types"
)

// StatsRepo 数据看板聚合查询接口。
type StatsRepo interface {
	// GetDailyRange 查询日期范围内的每日汇总。
	GetDailyRange(ctx context.Context, from, to time.Time) ([]StatsDaily, error)
	// GetProductDailyRange 查询日期范围内的商品每日数据。
	GetProductDailyRange(ctx context.Context, from, to time.Time) ([]StatsProductDaily, error)
	// GetChannelDailyRange 查询日期范围内的渠道每日数据。
	GetChannelDailyRange(ctx context.Context, from, to time.Time) ([]StatsChannelDaily, error)
	// UpsertDaily UPSERT 每日汇总。
	UpsertDaily(ctx context.Context, s *StatsDaily) error
	// UpsertProductDaily UPSERT 商品每日数据。
	UpsertProductDaily(ctx context.Context, rows []StatsProductDaily) error
	// AggregateFromOrders 从 order 表聚合一天的数据写入 stats_daily。
	AggregateFromOrders(ctx context.Context, date time.Time) error
	// AggregateProductFromItems 从 order_item 聚合写入 stats_product_daily。
	AggregateProductFromItems(ctx context.Context, date time.Time) error
	// TopProducts 商品销售排行（指定日期范围+分页）。
	TopProducts(ctx context.Context, from, to time.Time, page, size int) ([]TopProductRow, int64, error)
	// CategoryPie 商品分类占比 Top5+其他。
	CategoryPie(ctx context.Context, from, to time.Time) ([]CategoryShare, error)
	// ChannelStats 渠道统计。
	ChannelStats(ctx context.Context, from, to time.Time) ([]ChannelStatsRow, error)
	// GetWorkbench 工作台实时数据：今日订单、待发货、售后待处理。
	GetWorkbench(ctx context.Context) (*WorkbenchResp, error)
}

type statsRepoImpl struct{ db *gorm.DB }

// NewStatsRepo 构造 StatsRepo。
func NewStatsRepo(db *gorm.DB) StatsRepo { return &statsRepoImpl{db: db} }

func (r *statsRepoImpl) GetDailyRange(ctx context.Context, from, to time.Time) ([]StatsDaily, error) {
	var list []StatsDaily
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("date").Find(&list).Error
	return list, err
}

func (r *statsRepoImpl) GetProductDailyRange(ctx context.Context, from, to time.Time) ([]StatsProductDaily, error) {
	var list []StatsProductDaily
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&list).Error
	return list, err
}

func (r *statsRepoImpl) GetChannelDailyRange(ctx context.Context, from, to time.Time) ([]StatsChannelDaily, error) {
	var list []StatsChannelDaily
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Find(&list).Error
	return list, err
}

func (r *statsRepoImpl) UpsertDaily(ctx context.Context, s *StatsDaily) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *statsRepoImpl) UpsertProductDaily(ctx context.Context, rows []StatsProductDaily) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Save(&rows).Error
}

// AggregateFromOrders 从 "order" 表聚合写入 stats_daily（INSERT ... ON CONFLICT DO UPDATE）。
func (r *statsRepoImpl) AggregateFromOrders(ctx context.Context, date time.Time) error {
	dateStr := date.Format("2006-01-02")
	sql := `
INSERT INTO stats_daily (date, paid_order_count, paid_amount_cents, refund_amount_cents, net_amount_cents, paid_user_count, new_user_count, updated_at)
SELECT
    ?::date                                                        AS date,
    COUNT(*) FILTER (WHERE status IN ('paid','shipped','completed')) AS paid_order_count,
    COALESCE(SUM(pay_cents) FILTER (WHERE status IN ('paid','shipped','completed')), 0) AS paid_amount_cents,
    COALESCE(SUM(pay_cents) FILTER (WHERE status = 'refunded'), 0)                     AS refund_amount_cents,
    COALESCE(SUM(pay_cents) FILTER (WHERE status IN ('paid','shipped','completed')), 0)
      - COALESCE(SUM(pay_cents) FILTER (WHERE status = 'refunded'), 0)                AS net_amount_cents,
    COUNT(DISTINCT user_id) FILTER (WHERE status IN ('paid','shipped','completed'))    AS paid_user_count,
    0                                                              AS new_user_count,
    NOW()                                                          AS updated_at
FROM "order"
WHERE created_at::date = ?::date
ON CONFLICT (date) DO UPDATE SET
    paid_order_count    = EXCLUDED.paid_order_count,
    paid_amount_cents   = EXCLUDED.paid_amount_cents,
    refund_amount_cents = EXCLUDED.refund_amount_cents,
    net_amount_cents    = EXCLUDED.net_amount_cents,
    paid_user_count     = EXCLUDED.paid_user_count,
    updated_at          = NOW()
`
	return r.db.WithContext(ctx).Exec(sql, dateStr, dateStr).Error
}

// AggregateProductFromItems 从 order_item 聚合写入 stats_product_daily。
func (r *statsRepoImpl) AggregateProductFromItems(ctx context.Context, date time.Time) error {
	dateStr := date.Format("2006-01-02")
	sql := `
INSERT INTO stats_product_daily (date, product_id, qty, amount_cents)
SELECT
    ?::date        AS date,
    oi.product_id,
    SUM(oi.qty)    AS qty,
    SUM(oi.pay_cents) AS amount_cents
FROM order_item oi
JOIN "order" o ON o.id = oi.order_id
WHERE o.created_at::date = ?::date
  AND o.status IN ('paid','shipped','completed')
GROUP BY oi.product_id
ON CONFLICT (date, product_id) DO UPDATE SET
    qty          = EXCLUDED.qty,
    amount_cents = EXCLUDED.amount_cents
`
	return r.db.WithContext(ctx).Exec(sql, dateStr, dateStr).Error
}

func (r *statsRepoImpl) TopProducts(ctx context.Context, from, to time.Time, page, size int) ([]TopProductRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	type row struct {
		ProductID   int64  `gorm:"column:product_id"`
		Qty         int    `gorm:"column:qty"`
		AmountCents int64  `gorm:"column:amount_cents"`
	}

	var total int64
	r.db.WithContext(ctx).Model(&StatsProductDaily{}).
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Distinct("product_id").Count(&total)

	var rows []row
	err := r.db.WithContext(ctx).
		Table("stats_product_daily").
		Select("product_id, SUM(qty) AS qty, SUM(amount_cents) AS amount_cents").
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Group("product_id").
		Order("amount_cents DESC").
		Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]TopProductRow, len(rows))
	for i, r := range rows {
		result[i] = TopProductRow{
			ProductID:   types.Int64Str(r.ProductID),
			Qty:         r.Qty,
			AmountCents: r.AmountCents,
		}
	}
	return result, total, nil
}

func (r *statsRepoImpl) CategoryPie(ctx context.Context, from, to time.Time) ([]CategoryShare, error) {
	type row struct {
		CategoryID  int64  `gorm:"column:category_id"`
		AmountCents int64  `gorm:"column:amount_cents"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
SELECT p.category_id, SUM(spd.amount_cents) AS amount_cents
FROM stats_product_daily spd
JOIN product p ON p.id = spd.product_id
WHERE spd.date >= ?::date AND spd.date <= ?::date
GROUP BY p.category_id
ORDER BY amount_cents DESC
LIMIT 5
`, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var total int64
	for _, r := range rows {
		total += r.AmountCents
	}

	result := make([]CategoryShare, len(rows))
	for i, row := range rows {
		pct := 0.0
		if total > 0 {
			pct = float64(row.AmountCents) / float64(total)
		}
		result[i] = CategoryShare{
			CategoryID:  types.Int64Str(row.CategoryID),
			AmountCents: row.AmountCents,
			Percent:     pct,
		}
	}
	return result, nil
}

func (r *statsRepoImpl) ChannelStats(ctx context.Context, from, to time.Time) ([]ChannelStatsRow, error) {
	type row struct {
		ChannelCodeID int64  `gorm:"column:channel_code_id"`
		ScanCount     int    `gorm:"column:scan_count"`
		AddCount      int    `gorm:"column:add_count"`
		OrderCount    int    `gorm:"column:order_count"`
		AmountCents   int64  `gorm:"column:amount_cents"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("stats_channel_daily").
		Select("channel_code_id, SUM(scan_count) AS scan_count, SUM(add_count) AS add_count, SUM(order_count) AS order_count, SUM(amount_cents) AS amount_cents").
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Group("channel_code_id").
		Order("amount_cents DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]ChannelStatsRow, len(rows))
	for i, r := range rows {
		result[i] = ChannelStatsRow{
			ChannelCodeID: types.Int64Str(r.ChannelCodeID),
			ScanCount:     r.ScanCount,
			AddCount:      r.AddCount,
			OrderCount:    r.OrderCount,
			AmountCents:   r.AmountCents,
		}
	}
	return result, nil
}

func (r *statsRepoImpl) GetWorkbench(ctx context.Context) (*WorkbenchResp, error) {
	today := time.Now().Format("2006-01-02")

	var resp WorkbenchResp
	type dailyRow struct {
		OrderCount int   `gorm:"column:order_count"`
		SalesCents int64 `gorm:"column:sales_cents"`
	}
	var dr dailyRow
	r.db.WithContext(ctx).Raw(`
SELECT COUNT(*) AS order_count,
       COALESCE(SUM(pay_cents) FILTER (WHERE status IN ('paid','shipped','completed')), 0) AS sales_cents
FROM "order"
WHERE created_at::date = ?::date AND status NOT IN ('cancelled','pending_payment')
`, today).Scan(&dr)
	resp.TodayOrderCount = dr.OrderCount
	resp.TodaySales = dr.SalesCents

	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM "order" WHERE status = 'paid'`).Scan(&resp.PendingShip)
	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM "order" WHERE cancel_request_pending = true`).Scan(&resp.AftersalePending)

	return &resp, nil
}
