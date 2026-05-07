package shipping

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// SenderAddressRepo 发货地址数据访问接口。
type SenderAddressRepo interface {
	List(ctx context.Context) ([]SenderAddress, error)
	FindByID(ctx context.Context, id int64) (*SenderAddress, error)
	FindDefault(ctx context.Context) (*SenderAddress, error)
	Create(ctx context.Context, a *SenderAddress) error
	Update(ctx context.Context, a *SenderAddress) error
	Delete(ctx context.Context, id int64) error
	SetDefault(ctx context.Context, id int64) error
	DB() *gorm.DB
}

// CarrierRepo 快递公司数据访问接口。
type CarrierRepo interface {
	List(ctx context.Context) ([]Carrier, error)
	FindByCode(ctx context.Context, code string) (*Carrier, error)
	Update(ctx context.Context, c *Carrier) error
}

// ShipmentRepo 发货单数据访问接口。
type ShipmentRepo interface {
	Create(ctx context.Context, s *Shipment) error
	FindByID(ctx context.Context, id int64) (*Shipment, error)
	FindByOrderID(ctx context.Context, orderID int64) (*Shipment, error)
	FindByCarrierAndNo(ctx context.Context, carrierCode, trackingNo string) (*Shipment, error)
	Update(ctx context.Context, s *Shipment) error
	List(ctx context.Context, filter ShipmentFilter) ([]Shipment, int64, error)
	ListStale(ctx context.Context, before time.Time, excludeStatuses []string) ([]Shipment, error)

	CreateTrack(ctx context.Context, t *ShipmentTrack) error
	ListTracks(ctx context.Context, shipmentID int64) ([]ShipmentTrack, error)
	ExistsTrack(ctx context.Context, shipmentID int64, occurredAt time.Time, description string) (bool, error)
	DB() *gorm.DB
}

// ---- SenderAddressRepo 实现 ----

type senderAddrRepoImpl struct{ db *gorm.DB }

func NewSenderAddressRepo(db *gorm.DB) SenderAddressRepo {
	return &senderAddrRepoImpl{db: db}
}

func (r *senderAddrRepoImpl) DB() *gorm.DB { return r.db }

func (r *senderAddrRepoImpl) List(ctx context.Context) ([]SenderAddress, error) {
	var list []SenderAddress
	err := r.db.WithContext(ctx).Order("is_default DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *senderAddrRepoImpl) FindByID(ctx context.Context, id int64) (*SenderAddress, error) {
	var a SenderAddress
	err := r.db.WithContext(ctx).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *senderAddrRepoImpl) FindDefault(ctx context.Context) (*SenderAddress, error) {
	var a SenderAddress
	err := r.db.WithContext(ctx).Where("is_default = true").First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *senderAddrRepoImpl) Create(ctx context.Context, a *SenderAddress) error {
	if a.ID == 0 {
		a.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *senderAddrRepoImpl) Update(ctx context.Context, a *SenderAddress) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *senderAddrRepoImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&SenderAddress{}, id).Error
}

func (r *senderAddrRepoImpl) SetDefault(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先取消所有默认
		if err := tx.Model(&SenderAddress{}).Where("is_default = true").
			Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新默认
		return tx.Model(&SenderAddress{}).Where("id = ?", id).
			Update("is_default", true).Error
	})
}

// ---- CarrierRepo 实现 ----

type carrierRepoImpl struct{ db *gorm.DB }

func NewCarrierRepo(db *gorm.DB) CarrierRepo {
	return &carrierRepoImpl{db: db}
}

func (r *carrierRepoImpl) List(ctx context.Context) ([]Carrier, error) {
	var list []Carrier
	err := r.db.WithContext(ctx).Where("enabled = true").Order("sort ASC, code ASC").Find(&list).Error
	return list, err
}

func (r *carrierRepoImpl) FindByCode(ctx context.Context, code string) (*Carrier, error) {
	var c Carrier
	err := r.db.WithContext(ctx).First(&c, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *carrierRepoImpl) Update(ctx context.Context, c *Carrier) error {
	return r.db.WithContext(ctx).Save(c).Error
}

// ---- ShipmentRepo 实现 ----

type shipmentRepoImpl struct{ db *gorm.DB }

func NewShipmentRepo(db *gorm.DB) ShipmentRepo {
	return &shipmentRepoImpl{db: db}
}

func (r *shipmentRepoImpl) DB() *gorm.DB { return r.db }

func (r *shipmentRepoImpl) Create(ctx context.Context, s *Shipment) error {
	if s.ID == 0 {
		s.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *shipmentRepoImpl) FindByID(ctx context.Context, id int64) (*Shipment, error) {
	var s Shipment
	err := r.db.WithContext(ctx).First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepoImpl) FindByOrderID(ctx context.Context, orderID int64) (*Shipment, error) {
	var s Shipment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at DESC").First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepoImpl) FindByCarrierAndNo(ctx context.Context, carrierCode, trackingNo string) (*Shipment, error) {
	var s Shipment
	err := r.db.WithContext(ctx).
		Where("carrier_code = ? AND tracking_no = ?", carrierCode, trackingNo).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepoImpl) Update(ctx context.Context, s *Shipment) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *shipmentRepoImpl) List(ctx context.Context, filter ShipmentFilter) ([]Shipment, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}

	q := r.db.WithContext(ctx).Model(&Shipment{})
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
	var list []Shipment
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *shipmentRepoImpl) ListStale(ctx context.Context, before time.Time, excludeStatuses []string) ([]Shipment, error) {
	q := r.db.WithContext(ctx).Model(&Shipment{}).
		Where("status NOT IN ? AND (last_track_at IS NULL OR last_track_at < ?)", excludeStatuses, before)
	var list []Shipment
	err := q.Find(&list).Error
	return list, err
}

func (r *shipmentRepoImpl) CreateTrack(ctx context.Context, t *ShipmentTrack) error {
	if t.ID == 0 {
		t.ID = snowflake.NextID()
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(t).Error
}

func (r *shipmentRepoImpl) ListTracks(ctx context.Context, shipmentID int64) ([]ShipmentTrack, error) {
	var list []ShipmentTrack
	err := r.db.WithContext(ctx).
		Where("shipment_id = ?", shipmentID).
		Order("occurred_at DESC").Find(&list).Error
	return list, err
}

func (r *shipmentRepoImpl) ExistsTrack(ctx context.Context, shipmentID int64, occurredAt time.Time, description string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ShipmentTrack{}).
		Where("shipment_id = ? AND occurred_at = ? AND description = ?",
			shipmentID, occurredAt, description).
		Count(&count).Error
	return count > 0, err
}
