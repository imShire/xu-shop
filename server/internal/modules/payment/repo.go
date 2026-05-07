package payment

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// PaymentRepo 支付数据访问接口。
type PaymentRepo interface {
	// UpsertByTxn 按 transaction_id 幂等插入支付记录；bool=true 表示新插入。
	UpsertByTxn(ctx context.Context, txnID, orderNo string, orderID, amtCents int64, raw []byte) (*Payment, bool, error)
	// FindByOrderID 按订单 ID 查找最新支付记录。
	FindByOrderID(ctx context.Context, orderID int64) (*Payment, error)
	// FindSuccessByOrder 查找订单的成功支付记录。
	FindSuccessByOrder(ctx context.Context, orderID int64) (*Payment, error)
	// MarkSuccess 标记支付成功。
	MarkSuccess(ctx context.Context, id int64, raw []byte, paidAt time.Time) error
	// MarkFailed 标记支付失败。
	MarkFailed(ctx context.Context, id int64) error
	// MarkOrphan 标记为金额异常（已入队自动退款）。
	MarkOrphan(ctx context.Context, id int64) error
	// CreatePayment 创建支付记录（预支付时调用）。
	CreatePayment(ctx context.Context, p *Payment) error

	// CreateRefund 创建退款记录。
	CreateRefund(ctx context.Context, r *Refund) error
	// FindRefundByNo 按退款单号查找。
	FindRefundByNo(ctx context.Context, refundNo string) (*Refund, error)
	// SumRefundSuccess 计算订单已成功退款累计金额。
	SumRefundSuccess(ctx context.Context, orderID int64) (int64, error)
	// MarkRefundSuccess 标记退款成功。
	MarkRefundSuccess(ctx context.Context, id int64, raw []byte, refundedAt time.Time) error
	// MarkRefundFailed 标记退款失败。
	MarkRefundFailed(ctx context.Context, id int64, raw []byte) error
	// ListRefunds 订单退款列表。
	ListRefunds(ctx context.Context, orderID int64) ([]Refund, error)

	// CreateReconciliationDiff 创建对账差异记录。
	CreateReconciliationDiff(ctx context.Context, d *ReconciliationDiff) error
	// ListDiffs 对账差异列表。
	ListDiffs(ctx context.Context, filter DiffFilter) ([]ReconciliationDiff, int64, error)
	// ResolveDiff 标记差异已处理。
	ResolveDiff(ctx context.Context, id, adminID int64) error

	// ListPayments 支付记录列表。
	ListPayments(ctx context.Context, filter PaymentFilter) ([]Payment, int64, error)

	// ListAuditLogs 操作审计日志列表。
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error)

	// DB 返回底层 gorm.DB（事务内使用）。
	DB() *gorm.DB
}

type paymentRepoImpl struct{ db *gorm.DB }

// NewPaymentRepo 构造 PaymentRepo。
func NewPaymentRepo(db *gorm.DB) PaymentRepo {
	return &paymentRepoImpl{db: db}
}

func (r *paymentRepoImpl) DB() *gorm.DB { return r.db }

func (r *paymentRepoImpl) UpsertByTxn(ctx context.Context, txnID, orderNo string, orderID, amtCents int64, raw []byte) (*Payment, bool, error) {
	p := &Payment{
		ID:            snowflake.NextID(),
		OrderID:       orderID,
		Channel:       "wxpay",
		TradeType:     "JSAPI",
		TransactionID: &txnID,
		AmountCents:   amtCents,
		Status:        PayStatusPending,
		RawNotify:     rawNotifyFromBytes(raw),
	}

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "transaction_id"}},
		DoNothing: true,
	}).Create(p)

	if result.Error != nil {
		return nil, false, result.Error
	}

	isNew := result.RowsAffected > 0
	if !isNew {
		// 已存在，查询返回
		var existing Payment
		if err := r.db.WithContext(ctx).Where("transaction_id = ?", txnID).First(&existing).Error; err != nil {
			return nil, false, err
		}
		return &existing, false, nil
	}

	// 如果 orderNo 非空，尝试关联到已有支付行（更新 order_id）
	_ = orderNo
	return p, true, nil
}

func (r *paymentRepoImpl) FindByOrderID(ctx context.Context, orderID int64) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at DESC").First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepoImpl) FindSuccessByOrder(ctx context.Context, orderID int64) (*Payment, error) {
	var p Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND status = ?", orderID, PayStatusSuccess).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepoImpl) MarkSuccess(ctx context.Context, id int64, raw []byte, paidAt time.Time) error {
	return r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Updates(map[string]any{
		"status":     PayStatusSuccess,
		"raw_notify": rawNotifyFromBytes(raw),
		"paid_at":    paidAt,
		"updated_at": time.Now(),
	}).Error
}

func (r *paymentRepoImpl) MarkFailed(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Updates(map[string]any{
		"status":     PayStatusFailed,
		"updated_at": time.Now(),
	}).Error
}

func (r *paymentRepoImpl) MarkOrphan(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Updates(map[string]any{
		"status":     PayStatusOrphan,
		"updated_at": time.Now(),
	}).Error
}

func (r *paymentRepoImpl) CreatePayment(ctx context.Context, p *Payment) error {
	if p.ID == 0 {
		p.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *paymentRepoImpl) CreateRefund(ctx context.Context, refund *Refund) error {
	if refund.ID == 0 {
		refund.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).Create(refund).Error
}

func (r *paymentRepoImpl) FindRefundByNo(ctx context.Context, refundNo string) (*Refund, error) {
	var ref Refund
	err := r.db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&ref).Error
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (r *paymentRepoImpl) SumRefundSuccess(ctx context.Context, orderID int64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&Refund{}).
		Where("order_id = ? AND status = ?", orderID, RefundStatusSuccess).
		Select("COALESCE(SUM(amount_cents), 0)").
		Scan(&total).Error
	return total, err
}

func (r *paymentRepoImpl) MarkRefundSuccess(ctx context.Context, id int64, raw []byte, refundedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&Refund{}).Where("id = ?", id).Updates(map[string]any{
		"status":      RefundStatusSuccess,
		"raw_notify":  rawNotifyFromBytes(raw),
		"refunded_at": refundedAt,
		"updated_at":  time.Now(),
	}).Error
}

func (r *paymentRepoImpl) MarkRefundFailed(ctx context.Context, id int64, raw []byte) error {
	return r.db.WithContext(ctx).Model(&Refund{}).Where("id = ?", id).Updates(map[string]any{
		"status":     RefundStatusFailed,
		"raw_notify": rawNotifyFromBytes(raw),
		"updated_at": time.Now(),
	}).Error
}

func (r *paymentRepoImpl) ListRefunds(ctx context.Context, orderID int64) ([]Refund, error) {
	var list []Refund
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *paymentRepoImpl) CreateReconciliationDiff(ctx context.Context, d *ReconciliationDiff) error {
	if d.ID == 0 {
		d.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *paymentRepoImpl) ListDiffs(ctx context.Context, filter DiffFilter) ([]ReconciliationDiff, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}

	q := r.db.WithContext(ctx).Model(&ReconciliationDiff{})
	if filter.Date != "" {
		q = q.Where("bill_date = ?", filter.Date)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []ReconciliationDiff
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *paymentRepoImpl) ResolveDiff(ctx context.Context, id, adminID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&ReconciliationDiff{}).Where("id = ?", id).Updates(map[string]any{
		"status":      DiffStatusResolved,
		"resolved_by": adminID,
		"resolved_at": now,
	}).Error
}

func (r *paymentRepoImpl) ListPayments(ctx context.Context, filter PaymentFilter) ([]Payment, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}

	q := r.db.WithContext(ctx).Model(&Payment{})
	if filter.OrderID > 0 {
		q = q.Where("order_id = ?", filter.OrderID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Payment
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *paymentRepoImpl) ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}

	q := r.db.WithContext(ctx).Model(&AuditLog{})
	if filter.Module != "" {
		q = q.Where("module = ?", filter.Module)
	}
	if filter.Operator != "" {
		q = q.Where("admin_username ILIKE ?", "%"+filter.Operator+"%")
	}
	if filter.StartDate != "" {
		q = q.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		q = q.Where("created_at < ?", filter.EndDate+" 23:59:59")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []AuditLog
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
