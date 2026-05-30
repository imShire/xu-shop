package aftersale

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repo 售后单数据访问接口。
type Repo interface {
	DB() *gorm.DB
	Create(ctx context.Context, tx *gorm.DB, a *AftersaleOrder) error
	FindByID(ctx context.Context, id int64) (*AftersaleOrder, error)
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*AftersaleOrder, error)
	UpdateMap(ctx context.Context, tx *gorm.DB, id int64, fromStatus string, fields map[string]any) error
	FindActiveByOrder(ctx context.Context, orderID int64, orderItemID *int64) (*AftersaleOrder, error)
	ListByUser(ctx context.Context, f UserListFilter) ([]AftersaleOrder, int64, error)
	ListByAdmin(ctx context.Context, f AdminListFilter) ([]AftersaleOrder, int64, error)
	ScanExpiring(ctx context.Context, now time.Time, limit int) ([]AftersaleOrder, error)
	ScanPendingRefund(ctx context.Context, limit int) ([]AftersaleOrder, error)

	AddNegotiation(ctx context.Context, tx *gorm.DB, n *AftersaleNegotiation) error
	ListNegotiations(ctx context.Context, aftersaleID int64) ([]AftersaleNegotiation, error)

	// FindRefundStatus 查询 aftersale 对应的 refund 记录状态。
	// 通过 refund.id = aftersale.refund_id 关联。
	FindRefundStatus(ctx context.Context, refundID int64) (status string, err error)
	// FindRefundByOrderReason 根据订单+原因前缀回查 refund_id（aftersale 申请退款后用于回填 refund_id）。
	FindRefundByOrderReason(ctx context.Context, orderID int64, reasonPrefix string) (int64, error)
}

type repoImpl struct {
	db *gorm.DB
}

// NewRepo 构造默认 Repo。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) DB() *gorm.DB { return r.db }

func (r *repoImpl) Create(ctx context.Context, tx *gorm.DB, a *AftersaleOrder) error {
	d := r.tx(tx)
	return d.WithContext(ctx).Create(a).Error
}

func (r *repoImpl) FindByID(ctx context.Context, id int64) (*AftersaleOrder, error) {
	var a AftersaleOrder
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repoImpl) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*AftersaleOrder, error) {
	var a AftersaleOrder
	d := r.tx(tx).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id)
	if err := d.Take(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repoImpl) UpdateMap(ctx context.Context, tx *gorm.DB, id int64, fromStatus string, fields map[string]any) error {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["updated_at"] = time.Now()
	d := r.tx(tx).WithContext(ctx).Model(&AftersaleOrder{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(fields)
	if d.Error != nil {
		return d.Error
	}
	if d.RowsAffected == 0 {
		return ErrStaleStatus
	}
	return nil
}

func (r *repoImpl) FindActiveByOrder(ctx context.Context, orderID int64, orderItemID *int64) (*AftersaleOrder, error) {
	q := r.db.WithContext(ctx).
		Where("order_id = ? AND status NOT IN ?", orderID,
			[]string{StatusCompleted, StatusSellerRejected, StatusCancelled, StatusClosed})
	if orderItemID != nil {
		q = q.Where("order_item_id = ?", *orderItemID)
	} else {
		q = q.Where("order_item_id IS NULL")
	}
	var a AftersaleOrder
	err := q.Take(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repoImpl) ListByUser(ctx context.Context, f UserListFilter) ([]AftersaleOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&AftersaleOrder{}).Where("user_id = ?", f.UserID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	return r.paged(q, f.Page, f.PageSize)
}

func (r *repoImpl) ListByAdmin(ctx context.Context, f AdminListFilter) ([]AftersaleOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&AftersaleOrder{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		like := "%" + escapeLike(kw) + "%"
		q = q.Where("aftersale_no LIKE ? ESCAPE '\\'", like)
	}
	if f.AppliedFrom != nil {
		q = q.Where("applied_at >= ?", *f.AppliedFrom)
	}
	if f.AppliedTo != nil {
		q = q.Where("applied_at < ?", *f.AppliedTo)
	}
	return r.paged(q, f.Page, f.PageSize)
}

func (r *repoImpl) paged(q *gorm.DB, page, pageSize int) ([]AftersaleOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AftersaleOrder
	if err := q.Order("applied_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *repoImpl) ScanExpiring(ctx context.Context, now time.Time, limit int) ([]AftersaleOrder, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []AftersaleOrder
	err := r.db.WithContext(ctx).
		Where("status IN ? AND auto_close_at <= ?",
			[]string{StatusApplying, StatusSellerAgreed, StatusBuyerReturned}, now).
		Order("auto_close_at ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *repoImpl) ScanPendingRefund(ctx context.Context, limit int) ([]AftersaleOrder, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []AftersaleOrder
	err := r.db.WithContext(ctx).
		Where("status = ? AND refund_id IS NOT NULL", StatusSellerReceived).
		Order("received_at ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *repoImpl) AddNegotiation(ctx context.Context, tx *gorm.DB, n *AftersaleNegotiation) error {
	return r.tx(tx).WithContext(ctx).Create(n).Error
}

func (r *repoImpl) ListNegotiations(ctx context.Context, aftersaleID int64) ([]AftersaleNegotiation, error) {
	var rows []AftersaleNegotiation
	err := r.db.WithContext(ctx).Where("aftersale_id = ?", aftersaleID).
		Order("created_at ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *repoImpl) FindRefundStatus(ctx context.Context, refundID int64) (string, error) {
	var row struct {
		Status string
	}
	err := r.db.WithContext(ctx).Table("refund").
		Select("status").Where("id = ?", refundID).Take(&row).Error
	if err != nil {
		return "", err
	}
	return row.Status, nil
}

func (r *repoImpl) FindRefundByOrderReason(ctx context.Context, orderID int64, reasonPrefix string) (int64, error) {
	var row struct {
		ID int64
	}
	like := escapeLike(reasonPrefix) + "%"
	err := r.db.WithContext(ctx).Table("refund").
		Select("id").
		Where("order_id = ? AND reason LIKE ? ESCAPE '\\'", orderID, like).
		Order("id DESC").Limit(1).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *repoImpl) tx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// ErrStaleStatus 状态机乐观锁冲突。
var ErrStaleStatus = errors.New("aftersale: stale status")

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// GenAftersaleNo 生成 AS + yyMMddHHmmss + 6 位 ID 后缀。
func GenAftersaleNo(id int64, now time.Time) string {
	return fmt.Sprintf("AS%s%06d", now.Format("060102150405"), id%1_000_000)
}
