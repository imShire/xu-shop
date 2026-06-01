package reconciliation

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gorm.io/gorm"

	rec "github.com/xushop/xu-shop/internal/modules/reconciliation"
	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// CommissionRow 佣金记录对账输入。
type CommissionRow struct {
	ID              int64
	BaseAmountCents int64
	Rate            float64
	AmountCents     int64 // 实际入账
}

// CommissionDiff 佣金差异中间结构。
type CommissionDiff struct {
	RecordID  int64
	Field     string // 'amount_cents'
	Expected  int64
	Actual    int64
	DiffCents int64
	Severity  string
}

// expectedCommission 计算预期佣金（向下取整，分级精度）。
func expectedCommission(baseCents int64, rate float64) int64 {
	if rate <= 0 || baseCents <= 0 {
		return 0
	}
	return int64(math.Floor(float64(baseCents) * rate))
}

// diffCommission 比较单条佣金记录的入账金额与重算结果。纯函数。
// > ¥1 (100 cents) 差额 → critical；其余 warn；0 差 → 无。
func diffCommission(row CommissionRow) []CommissionDiff {
	expected := expectedCommission(row.BaseAmountCents, row.Rate)
	if expected == row.AmountCents {
		return nil
	}
	delta := row.AmountCents - expected
	abs := delta
	if abs < 0 {
		abs = -abs
	}
	severity := rec.SeverityWarn
	if abs > 100 {
		severity = rec.SeverityCritical
	}
	return []CommissionDiff{{
		RecordID:  row.ID,
		Field:     "amount_cents",
		Expected:  expected,
		Actual:    row.AmountCents,
		DiffCents: delta,
		Severity:  severity,
	}}
}

// CommissionDeps 佣金对账依赖。
type CommissionDeps struct {
	DB  *gorm.DB
	Svc *rec.Service
}

// NewCommissionHandler 构造 asynq handler。
func NewCommissionHandler(deps CommissionDeps) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		return RunCommission(ctx, deps)
	}
}

// RunCommission 执行佣金对账：扫描昨日新增/变更的佣金记录。
func RunCommission(ctx context.Context, deps CommissionDeps) error {
	ctx, span := otel.Tracer("reconciliation").Start(ctx, "commission_reconcile")
	defer span.End()

	bizDate := time.Now().In(bizLoc).AddDate(0, 0, -1)
	dayStart := time.Date(bizDate.Year(), bizDate.Month(), bizDate.Day(), 0, 0, 0, 0, bizLoc)
	dayEnd := dayStart.Add(24 * time.Hour)

	start := time.Now()
	logger.L().Info("commission_reconcile start",
		zap.String("biz_date", bizDate.Format("2006-01-02")))

	const batch = 500
	var lastID int64
	var checked, diffCount, criticalCount int
	for {
		type row struct {
			ID              int64
			BaseAmountCents int64
			Rate            float64
			AmountCents     int64
		}
		var records []row
		if err := deps.DB.WithContext(ctx).
			Table("commission_record").
			Select("id, base_amount_cents, rate, amount_cents").
			Where("id > ?", lastID).
			Where("updated_at >= ? AND updated_at < ?", dayStart, dayEnd).
			Order("id ASC").
			Limit(batch).
			Find(&records).Error; err != nil {
			logger.L().Error("commission_reconcile query failed", zap.Error(err))
			return err
		}
		if len(records) == 0 {
			break
		}
		for _, r := range records {
			checked++
			diffs := diffCommission(CommissionRow{
				ID: r.ID, BaseAmountCents: r.BaseAmountCents,
				Rate: r.Rate, AmountCents: r.AmountCents,
			})
			for _, d := range diffs {
				diffCount++
				if d.Severity == rec.SeverityCritical {
					criticalCount++
				}
				expected := strconv.FormatInt(d.Expected, 10)
				actual := strconv.FormatInt(d.Actual, 10)
				delta := d.DiffCents
				err := deps.Svc.RecordDiff(ctx, &rec.Diff{
					Job: rec.JobCommission, BizDate: dayStart,
					RefType:       rec.RefTypeCommissionRecord,
					RefID:         strconv.FormatInt(d.RecordID, 10),
					Field:         d.Field,
					ExpectedValue: &expected,
					ActualValue:   &actual,
					DiffCents:     &delta,
					Severity:      d.Severity,
				})
				if err != nil {
					logger.L().Warn("commission_reconcile record diff failed",
						zap.Int64("record_id", d.RecordID), zap.Error(err))
				}
			}
		}

		lastID = records[len(records)-1].ID
		if len(records) < batch {
			break
		}
	}

	logger.L().Info("commission_reconcile done",
		zap.String("biz_date", bizDate.Format("2006-01-02")),
		zap.Int("checked", checked),
		zap.Int("diff_count", diffCount),
		zap.Int("critical_count", criticalCount),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()))
	return nil
}
