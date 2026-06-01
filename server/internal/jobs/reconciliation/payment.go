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
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// PaymentRow 支付对账输入行（本地侧）。
type PaymentRow struct {
	OrderID       int64
	OrderNo       string
	AmountCents   int64
	TransactionID string
}

// PaymentDiff 一条支付差异结果（中间结构，方便单测）。
type PaymentDiff struct {
	OrderID       int64
	OrderNo       string
	Field         string // 'state' | 'amount_cents' | 'transaction_id'
	Expected      string
	Actual        string
	DiffCents     *int64
	Severity      string
}

// diffPayment 比对单笔本地支付与远端查单结果，返回差异（可能 0~3 条）。
// 纯函数，不依赖 DB / 网络。
func diffPayment(local PaymentRow, remote *pkgwxpay.QueryResp) []PaymentDiff {
	var out []PaymentDiff
	if remote == nil {
		out = append(out, PaymentDiff{
			OrderID: local.OrderID, OrderNo: local.OrderNo,
			Field:    "state",
			Expected: "SUCCESS",
			Actual:   "NOT_FOUND",
			Severity: rec.SeverityCritical,
		})
		return out
	}
	// state：本地 succeeded，对应远端必须 SUCCESS
	if remote.TradeState != "SUCCESS" {
		out = append(out, PaymentDiff{
			OrderID: local.OrderID, OrderNo: local.OrderNo,
			Field:    "state",
			Expected: "SUCCESS",
			Actual:   remote.TradeState,
			Severity: rec.SeverityCritical,
		})
	}
	// amount
	if remote.AmtCents != local.AmountCents {
		diff := local.AmountCents - remote.AmtCents
		out = append(out, PaymentDiff{
			OrderID: local.OrderID, OrderNo: local.OrderNo,
			Field:     "amount_cents",
			Expected:  strconv.FormatInt(remote.AmtCents, 10),
			Actual:    strconv.FormatInt(local.AmountCents, 10),
			DiffCents: &diff,
			Severity:  rec.SeverityCritical,
		})
	}
	// transaction_id
	if local.TransactionID != "" && remote.TransactionID != "" &&
		local.TransactionID != remote.TransactionID {
		out = append(out, PaymentDiff{
			OrderID: local.OrderID, OrderNo: local.OrderNo,
			Field:    "transaction_id",
			Expected: remote.TransactionID,
			Actual:   local.TransactionID,
			Severity: rec.SeverityWarn,
		})
	}
	return out
}

// PaymentDeps 支付对账依赖。
type PaymentDeps struct {
	DB    *gorm.DB
	Wxpay pkgwxpay.Client
	Svc   *rec.Service
}

// NewPaymentHandler 构造 asynq handler。
func NewPaymentHandler(deps PaymentDeps) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		return RunPayment(ctx, deps)
	}
}

// RunPayment 执行一次支付对账（业务日 = 昨天）。
func RunPayment(ctx context.Context, deps PaymentDeps) error {
	ctx, span := otel.Tracer("reconciliation").Start(ctx, "payment_reconcile")
	defer span.End()

	bizDate := time.Now().In(bizLoc).AddDate(0, 0, -1)
	dayStart := time.Date(bizDate.Year(), bizDate.Month(), bizDate.Day(), 0, 0, 0, 0, bizLoc)
	dayEnd := dayStart.Add(24 * time.Hour)

	type localRow struct {
		OrderID       int64
		AmountCents   int64
		TransactionID *string
	}
	var rows []localRow
	if err := deps.DB.WithContext(ctx).
		Table("payment").
		Select("order_id, amount_cents, transaction_id").
		Where("status = ?", "succeeded").
		Where("paid_at >= ? AND paid_at < ?", dayStart, dayEnd).
		Find(&rows).Error; err != nil {
		logger.L().Error("payment_reconcile query local failed", zap.Error(err))
		return err
	}

	logger.L().Info("payment_reconcile start",
		zap.String("biz_date", bizDate.Format("2006-01-02")),
		zap.Int("checked", len(rows)))

	start := time.Now()
	limiter := time.NewTicker(100 * time.Millisecond) // ≤10/s
	defer limiter.Stop()

	var diffCount, criticalCount int
	for _, row := range rows {
		<-limiter.C

		// 取 orderNo
		var orderNo string
		if err := deps.DB.WithContext(ctx).Table(`"order"`).
			Select("order_no").Where("id = ?", row.OrderID).
			Scan(&orderNo).Error; err != nil || orderNo == "" {
			continue
		}

		remote, queryErr := deps.Wxpay.QueryByOutTradeNo(ctx, orderNo)
		if queryErr != nil {
			logger.L().Warn("payment_reconcile query remote failed",
				zap.String("order_no", orderNo), zap.Error(queryErr))
			// 网络错误不算差异，跳过
			continue
		}

		var txn string
		if row.TransactionID != nil {
			txn = *row.TransactionID
		}
		local := PaymentRow{
			OrderID: row.OrderID, OrderNo: orderNo,
			AmountCents: row.AmountCents, TransactionID: txn,
		}
		for _, d := range diffPayment(local, remote) {
			diffCount++
			if d.Severity == rec.SeverityCritical {
				criticalCount++
			}
			expected := d.Expected
			actual := d.Actual
			err := deps.Svc.RecordDiff(ctx, &rec.Diff{
				Job: rec.JobPayment, BizDate: dayStart,
				RefType: rec.RefTypeOrder, RefID: strconv.FormatInt(d.OrderID, 10),
				Field:         d.Field,
				ExpectedValue: &expected,
				ActualValue:   &actual,
				DiffCents:     d.DiffCents,
				Severity:      d.Severity,
			})
			if err != nil {
				logger.L().Warn("payment_reconcile record diff failed",
					zap.Int64("order_id", d.OrderID), zap.Error(err))
			}
		}
	}

	logger.L().Info("payment_reconcile done",
		zap.String("biz_date", bizDate.Format("2006-01-02")),
		zap.Int("checked", len(rows)),
		zap.Int("diff_count", diffCount),
		zap.Int("critical_count", criticalCount),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()))
	return nil
}
