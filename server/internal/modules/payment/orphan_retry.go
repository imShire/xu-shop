package payment

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// PaymentOrphanRetry 自动退款 enqueue 失败的兜底持久化记录。
// 由 reconciler/cron job 扫描 next_retry_at <= now 的行重新入队 TaskPaymentAutoRefund。
type PaymentOrphanRetry struct {
	ID           int64     `gorm:"primaryKey"`
	PaymentID    int64     `gorm:"not null"`
	WxTxID       string    `gorm:"column:wx_txid;size:64;not null"`
	AmountCents  int64     `gorm:"not null"`
	Reason       string    `gorm:"size:200;not null;default:''"`
	RetryCount   int       `gorm:"not null;default:0"`
	NextRetryAt  time.Time `gorm:"not null;default:now()"`
	LastError    *string   `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName 指定表名。
func (PaymentOrphanRetry) TableName() string { return "payment_orphan_retry" }

// OrphanRetryRepo 自动退款兜底数据访问接口。
type OrphanRetryRepo interface {
	// Enqueue 写入一条待重试记录；next_retry_at = now + delay。
	Enqueue(ctx context.Context, paymentID int64, wxTxID string, amountCents int64, reason string, delay time.Duration) error
}

type orphanRetryRepoImpl struct{ db *gorm.DB }

// NewOrphanRetryRepo 构造 OrphanRetryRepo。
func NewOrphanRetryRepo(db *gorm.DB) OrphanRetryRepo { return &orphanRetryRepoImpl{db: db} }

func (r *orphanRetryRepoImpl) Enqueue(ctx context.Context, paymentID int64, wxTxID string, amountCents int64, reason string, delay time.Duration) error {
	row := &PaymentOrphanRetry{
		ID:          snowflake.NextID(),
		PaymentID:   paymentID,
		WxTxID:      wxTxID,
		AmountCents: amountCents,
		Reason:      reason,
		RetryCount:  0,
		NextRetryAt: time.Now().Add(delay),
	}
	return r.db.WithContext(ctx).Create(row).Error
}
