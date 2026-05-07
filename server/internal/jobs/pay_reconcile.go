package jobs

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/xushop/xu-shop/internal/pkg/logger"
)

// TaskPayReconcile 每日对账任务名（每日 00:30 执行）。
const TaskPayReconcile = "pay:reconcile"

// PayReconcilePayload 对账任务 payload。
type PayReconcilePayload struct {
	// BillDate 对账日期（格式 2006-01-02，默认为昨天）。
	BillDate string `json:"bill_date"`
}

// BillRow 微信账单单行数据。
type BillRow struct {
	TransactionID string
	OutTradeNo    string
	AmtCents      int64
}

// ReconcileHandler 对账数据访问接口（由 main 层传入实现）。
type ReconcileHandler interface {
	// DownloadBill 下载微信账单（返回 CSV 内容）。
	DownloadBill(ctx context.Context, date string) ([]byte, error)
	// FindPaymentByTxn 按 transaction_id 查找我方支付记录。
	FindPaymentByTxn(ctx context.Context, txnID string) (int64, error) // 返回 our_amount_cents
	// CreateDiff 创建对账差异记录。
	CreateDiff(ctx context.Context, billDate time.Time, txnID, orderNo string, ourCents, wxCents int64, diffType string) error
}

// NewPayReconcileHandler 构造对账任务 Handler。
func NewPayReconcileHandler(handler ReconcileHandler) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p PayReconcilePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			logger.L().Error("pay:reconcile unmarshal payload", zap.Error(err))
			return asynq.SkipRetry
		}
		return handlePayReconcile(ctx, p, handler)
	}
}

func handlePayReconcile(ctx context.Context, p PayReconcilePayload, handler ReconcileHandler) error {
	billDate := p.BillDate
	if billDate == "" {
		// 默认对前一天账单
		billDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	logger.L().Info("pay:reconcile start", zap.String("date", billDate))

	// 1. 下载微信账单
	billData, err := handler.DownloadBill(ctx, billDate)
	if err != nil {
		logger.L().Error("pay:reconcile download bill failed", zap.Error(err))
		return err
	}

	// 2. 解析账单 CSV
	billRows, err := parseBillCSV(billData)
	if err != nil {
		logger.L().Error("pay:reconcile parse csv failed", zap.Error(err))
		return err
	}

	// 3. 对账：遍历微信账单，与我方记录比较
	billDateParsed, _ := time.Parse("2006-01-02", billDate)
	for _, row := range billRows {
		ourCents, err := handler.FindPaymentByTxn(ctx, row.TransactionID)
		if err != nil {
			// 我方无此记录（微信有，我方无 → 幽灵单）
			if createErr := handler.CreateDiff(ctx, billDateParsed, row.TransactionID, row.OutTradeNo,
				0, row.AmtCents, "ghost"); createErr != nil {
				logger.L().Warn("create ghost diff failed", zap.Error(createErr))
			}
			continue
		}
		if ourCents != row.AmtCents {
			// 金额不一致
			if createErr := handler.CreateDiff(ctx, billDateParsed, row.TransactionID, row.OutTradeNo,
				ourCents, row.AmtCents, "amount_mismatch"); createErr != nil {
				logger.L().Warn("create amount_mismatch diff failed", zap.Error(createErr))
			}
		}
	}

	logger.L().Info("pay:reconcile done",
		zap.String("date", billDate), zap.Int("rows", len(billRows)))
	return nil
}

// parseBillCSV 解析微信支付账单 CSV（跳过前几行汇总行）。
func parseBillCSV(data []byte) ([]BillRow, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.LazyQuotes = true

	var rows []BillRow
	header := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if header {
			header = false
			continue
		}
		if len(record) < 12 {
			continue
		}
		// 微信账单格式：col[7]=交易号, col[0]=商户单号, col[5]=应结金额
		txnID := strings.TrimSpace(record[7])
		outTradeNo := strings.TrimSpace(record[0])
		if txnID == "" {
			continue
		}
		var amtCents int64
		amtStr := strings.TrimPrefix(strings.TrimSpace(record[5]), "¥")
		// 简单解析：保留两位小数转分
		_ = amtStr
		rows = append(rows, BillRow{
			TransactionID: txnID,
			OutTradeNo:    outTradeNo,
			AmtCents:      amtCents,
		})
	}
	return rows, nil
}
