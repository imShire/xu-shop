package order

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// AdminOrderFilter 后台订单列表过滤条件。
type AdminOrderFilter struct {
	Status  string
	UserID  int64
	OrderNo string
	Page    int
	Size    int
}

// OrderRepo 订单数据访问接口。
type OrderRepo interface {
	Create(ctx context.Context, o *Order, items []OrderItem, logs []OrderLog) error
	FindByID(ctx context.Context, id int64) (*Order, error)
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*Order, error)
	Update(ctx context.Context, o *Order) error
	ListByUser(ctx context.Context, userID int64, status string, page, size int) ([]Order, int64, error)
	ListByAdmin(ctx context.Context, filter AdminOrderFilter) ([]Order, int64, error)
	FindItemsByOrderIDs(ctx context.Context, orderIDs []int64) ([]OrderItem, error)
	FindByOrderNo(ctx context.Context, orderNo string) (*Order, error)
	AddLog(ctx context.Context, log *OrderLog) error
	ListLogsByOrder(ctx context.Context, orderID int64) ([]OrderLog, error)
	AddRemark(ctx context.Context, remark *OrderRemark) error
	ListRemarks(ctx context.Context, orderID int64) ([]OrderRemark, error)
	FindDefaultFreight(ctx context.Context) (*FreightTemplate, error)
	GetFreightByID(ctx context.Context, id int64) (*FreightTemplate, error)
	ListFreightTemplates(ctx context.Context) ([]FreightTemplate, error)
	CreateFreightTemplate(ctx context.Context, t *FreightTemplate) error
	UpdateFreightTemplate(ctx context.Context, t *FreightTemplate) error
	DeleteFreightTemplate(ctx context.Context, id int64) error
	// DB 引用（事务内使用）
	DB() *gorm.DB
}

type orderRepoImpl struct{ db *gorm.DB }

// NewOrderRepo 构造 OrderRepo。
func NewOrderRepo(db *gorm.DB) OrderRepo { return &orderRepoImpl{db: db} }

func (r *orderRepoImpl) DB() *gorm.DB { return r.db }

func (r *orderRepoImpl) Create(ctx context.Context, o *Order, items []OrderItem, logs []OrderLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		if len(logs) > 0 {
			if err := tx.Create(&logs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepoImpl) FindByID(ctx context.Context, id int64) (*Order, error) {
	var o Order
	err := r.db.WithContext(ctx).First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepoImpl) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*Order, error) {
	var o Order
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepoImpl) Update(ctx context.Context, o *Order) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *orderRepoImpl) ListByUser(ctx context.Context, userID int64, status string, page, size int) ([]Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&Order{}).Where(`"order".user_id = ?`, userID)
	if status != "" {
		q = q.Where(`"order".status = ?`, status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Order
	err := q.Order(`"order".created_at DESC`).
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *orderRepoImpl) ListByAdmin(ctx context.Context, f AdminOrderFilter) ([]Order, int64, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.Size
	if size < 1 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&Order{})
	if f.Status != "" {
		q = q.Where(`"order".status = ?`, f.Status)
	}
	if f.UserID > 0 {
		q = q.Where(`"order".user_id = ?`, f.UserID)
	}
	if f.OrderNo != "" {
		q = q.Where(`"order".order_no = ?`, f.OrderNo)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []Order
	err := q.Order(`"order".created_at DESC`).
		Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *orderRepoImpl) FindItemsByOrderIDs(ctx context.Context, orderIDs []int64) ([]OrderItem, error) {
	var list []OrderItem
	err := r.db.WithContext(ctx).Where("order_id IN ?", orderIDs).Find(&list).Error
	return list, err
}

func (r *orderRepoImpl) FindByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	var o Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepoImpl) AddLog(ctx context.Context, log *OrderLog) error {
	log.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *orderRepoImpl) ListLogsByOrder(ctx context.Context, orderID int64) ([]OrderLog, error) {
	var list []OrderLog
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *orderRepoImpl) AddRemark(ctx context.Context, remark *OrderRemark) error {
	remark.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(remark).Error
}

func (r *orderRepoImpl) ListRemarks(ctx context.Context, orderID int64) ([]OrderRemark, error) {
	var list []OrderRemark
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *orderRepoImpl) FindDefaultFreight(ctx context.Context) (*FreightTemplate, error) {
	var t FreightTemplate
	err := r.db.WithContext(ctx).Where("is_default = true").First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *orderRepoImpl) GetFreightByID(ctx context.Context, id int64) (*FreightTemplate, error) {
	var t FreightTemplate
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *orderRepoImpl) ListFreightTemplates(ctx context.Context) ([]FreightTemplate, error) {
	var list []FreightTemplate
	err := r.db.WithContext(ctx).Order("is_default DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *orderRepoImpl) CreateFreightTemplate(ctx context.Context, t *FreightTemplate) error {
	t.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *orderRepoImpl) UpdateFreightTemplate(ctx context.Context, t *FreightTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *orderRepoImpl) DeleteFreightTemplate(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&FreightTemplate{}, id).Error
}
