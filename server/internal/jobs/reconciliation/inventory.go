package reconciliation

import (
	"context"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/gorm"

	rec "github.com/xushop/xu-shop/internal/modules/reconciliation"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// SKURow 库存对账单行输入。
type SKURow struct {
	SKUID       int64
	Stock       int // 可售
	LockedStock int // 已锁
}

// InventoryDiff 库存差异中间结构。
type InventoryDiff struct {
	SKUID    int64
	Field    string // 'locked_stock_balance'
	Expected int    // 应锁库存 = 未发货且 status in (paid, processing) 的订单 item 合计
	Actual   int    // sku.locked_stock
	Severity string
}

// diffInventory 比较单个 SKU 的锁定库存与订单端推导值。纯函数，便于单测。
// 算法（docs/arch/04-inventory.md）：
//   sku.locked_stock 应 = 所有 status ∈ {paid, processing} 且 shipped_at IS NULL 的订单中该 SKU item.qty 总和
// 偏差即记差异；金额无关，severity 默认 warn。
func diffInventory(sku SKURow, pendingShipQty int) []InventoryDiff {
	if sku.LockedStock == pendingShipQty {
		return nil
	}
	return []InventoryDiff{{
		SKUID:    sku.SKUID,
		Field:    "locked_stock_balance",
		Expected: pendingShipQty,
		Actual:   sku.LockedStock,
		Severity: rec.SeverityWarn,
	}}
}

// InventoryDeps 库存对账依赖。
type InventoryDeps struct {
	DB  *gorm.DB
	Svc *rec.Service
}

// NewInventoryHandler 构造 asynq handler。
func NewInventoryHandler(deps InventoryDeps) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		return RunInventory(ctx, deps)
	}
}

// RunInventory 执行库存对账。
func RunInventory(ctx context.Context, deps InventoryDeps) error {
	ctx, span := otel.Tracer("reconciliation").Start(ctx, "inventory_reconcile")
	defer span.End()

	bizDate := time.Now().In(bizLoc).AddDate(0, 0, -1)
	dayStart := time.Date(bizDate.Year(), bizDate.Month(), bizDate.Day(), 0, 0, 0, 0, bizLoc)

	start := time.Now()
	logger.L().Info("inventory_reconcile start",
		zap.String("biz_date", bizDate.Format("2006-01-02")))

	const batch = 500
	var lastID int64
	var checked, diffCount int
	for {
		type row struct {
			ID          int64
			Stock       int
			LockedStock int
		}
		var skus []row
		if err := deps.DB.WithContext(ctx).
			Table("sku").
			Select("id, stock, locked_stock").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batch).
			Find(&skus).Error; err != nil {
			logger.L().Error("inventory_reconcile query sku failed", zap.Error(err))
			return err
		}
		if len(skus) == 0 {
			break
		}

		ids := make([]int64, len(skus))
		for i, s := range skus {
			ids[i] = s.ID
		}

		// 聚合未发货订单 item 数量
		type aggRow struct {
			SkuID   int64
			Pending int64
		}
		var aggs []aggRow
		if err := deps.DB.WithContext(ctx).
			Table("order_item AS oi").
			Select("oi.sku_id AS sku_id, COALESCE(SUM(oi.qty),0) AS pending").
			Joins(`JOIN "order" o ON o.id = oi.order_id`).
			Where("oi.sku_id IN ?", ids).
			Where("o.status IN ?", []string{"paid", "processing"}).
			Where("o.shipped_at IS NULL").
			Group("oi.sku_id").
			Scan(&aggs).Error; err != nil {
			logger.L().Error("inventory_reconcile query order_item failed", zap.Error(err))
			return err
		}
		pendingBySKU := make(map[int64]int, len(aggs))
		for _, a := range aggs {
			pendingBySKU[a.SkuID] = int(a.Pending)
		}

		for _, s := range skus {
			checked++
			pending := pendingBySKU[s.ID]
			diffs := diffInventory(SKURow{SKUID: s.ID, Stock: s.Stock, LockedStock: s.LockedStock}, pending)
			for _, d := range diffs {
				diffCount++
				expected := strconv.Itoa(d.Expected)
				actual := strconv.Itoa(d.Actual)
				err := deps.Svc.RecordDiff(ctx, &rec.Diff{
					Job: rec.JobInventory, BizDate: dayStart,
					RefType:       rec.RefTypeSKU,
					RefID:         strconv.FormatInt(d.SKUID, 10),
					Field:         d.Field,
					ExpectedValue: &expected,
					ActualValue:   &actual,
					Severity:      d.Severity,
				})
				if err != nil {
					logger.L().Warn("inventory_reconcile record diff failed",
						zap.Int64("sku_id", d.SKUID), zap.Error(err))
				}
			}
		}

		lastID = skus[len(skus)-1].ID
		if len(skus) < batch {
			break
		}
	}

	logger.L().Info("inventory_reconcile done",
		zap.String("biz_date", bizDate.Format("2006-01-02")),
		zap.Int("checked", checked),
		zap.Int("diff_count", diffCount),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()))
	return nil
}
