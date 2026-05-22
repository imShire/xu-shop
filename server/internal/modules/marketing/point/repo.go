package point

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repo 积分仓储。
type Repo interface {
	GetOrCreateAccount(ctx context.Context, tx *gorm.DB, userID int64) (*Account, error)
	UpdateAccount(ctx context.Context, tx *gorm.DB, userID int64, deltaBalance, deltaLocked, deltaEarned, deltaSpent int64) error
	InsertTransaction(ctx context.Context, tx *gorm.DB, t *Transaction) error
	FindTransactionByIdem(ctx context.Context, idemKey string) (*Transaction, error)
	ListTransactions(ctx context.Context, userID int64, page, size int) ([]Transaction, int64, error)

	// FIFO 扫描可消费余额（earn 类未消费且未过期）
	ListAvailableEarn(ctx context.Context, tx *gorm.DB, userID int64, limit int) ([]Transaction, error)
	MarkTransactionConsumed(ctx context.Context, tx *gorm.DB, ids []int64) error

	// 过期扫描
	ScanExpired(ctx context.Context, before time.Time, limit int) ([]Transaction, error)

	// 规则
	GetRule(ctx context.Context, code string) (*Rule, error)

	// 调整工单
	CreateTicket(ctx context.Context, t *AdjustTicket) error
	FindTicket(ctx context.Context, id int64) (*AdjustTicket, error)
	UpdateTicketStatus(ctx context.Context, tx *gorm.DB, id int64, fromStatus, toStatus string, approverID int64) (int64, error)
	HasPendingTicket(ctx context.Context, userID int64) (bool, error)
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造仓储。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) GetOrCreateAccount(ctx context.Context, tx *gorm.DB, userID int64) (*Account, error) {
	var a Account
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).First(&a).Error
	if err == nil {
		return &a, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	a = Account{UserID: userID}
	if err := tx.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repoImpl) UpdateAccount(ctx context.Context, tx *gorm.DB, userID int64, deltaBalance, deltaLocked, deltaEarned, deltaSpent int64) error {
	updates := map[string]any{
		"balance":      gorm.Expr("balance + ?", deltaBalance),
		"locked":       gorm.Expr("locked + ?", deltaLocked),
		"total_earned": gorm.Expr("total_earned + ?", deltaEarned),
		"total_spent":  gorm.Expr("total_spent + ?", deltaSpent),
		"updated_at":   time.Now(),
	}
	return tx.WithContext(ctx).Model(&Account{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *repoImpl) InsertTransaction(ctx context.Context, tx *gorm.DB, t *Transaction) error {
	return tx.WithContext(ctx).Create(t).Error
}

func (r *repoImpl) FindTransactionByIdem(ctx context.Context, idemKey string) (*Transaction, error) {
	var t Transaction
	err := r.db.WithContext(ctx).Where("idem_key = ?", idemKey).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *repoImpl) ListTransactions(ctx context.Context, userID int64, page, size int) ([]Transaction, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	var total int64
	if err := r.db.WithContext(ctx).Model(&Transaction{}).Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Transaction
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}

func (r *repoImpl) ListAvailableEarn(ctx context.Context, tx *gorm.DB, userID int64, limit int) ([]Transaction, error) {
	var list []Transaction
	q := tx.WithContext(ctx).Where("user_id = ? AND type = ? AND consumed = false", userID, TxnTypeEarn)
	q = q.Where("(expire_at IS NULL OR expire_at > now())")
	err := q.Order("created_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *repoImpl) MarkTransactionConsumed(ctx context.Context, tx *gorm.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Model(&Transaction{}).Where("id IN ?", ids).
		Update("consumed", true).Error
}

func (r *repoImpl) ScanExpired(ctx context.Context, before time.Time, limit int) ([]Transaction, error) {
	var list []Transaction
	err := r.db.WithContext(ctx).
		Where("type = ? AND consumed = false AND expire_at IS NOT NULL AND expire_at <= ?", TxnTypeEarn, before).
		Order("expire_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *repoImpl) GetRule(ctx context.Context, code string) (*Rule, error) {
	var rule Rule
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *repoImpl) CreateTicket(ctx context.Context, t *AdjustTicket) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repoImpl) FindTicket(ctx context.Context, id int64) (*AdjustTicket, error) {
	var t AdjustTicket
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repoImpl) UpdateTicketStatus(ctx context.Context, tx *gorm.DB, id int64, fromStatus, toStatus string, approverID int64) (int64, error) {
	now := time.Now()
	res := tx.WithContext(ctx).Model(&AdjustTicket{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]any{
			"status":            toStatus,
			"approver_admin_id": approverID,
			"approved_at":       now,
		})
	return res.RowsAffected, res.Error
}

func (r *repoImpl) HasPendingTicket(ctx context.Context, userID int64) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&AdjustTicket{}).
		Where("user_id = ? AND status = ?", userID, TicketStatusPending).
		Count(&n).Error
	return n > 0, err
}
