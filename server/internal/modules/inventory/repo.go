package inventory

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// SKUStockRow 对账用 SKU 库存快照。
type SKUStockRow struct {
	ID          int64
	Stock       int
	LockedStock int
}

// InventoryRepo 库存数据访问接口。
type InventoryRepo interface {
	GetSKUStock(ctx context.Context, skuID int64) (stock, lockedStock int, err error)
	AdjustDB(ctx context.Context, skuID int64, change int, t, refType string, refID, operatorID int64, reason string) error
	DeductDB(ctx context.Context, skuID, qty int, orderID int64) error
	LockDB(ctx context.Context, skuID, qty int, orderID int64) error
	UnlockDB(ctx context.Context, skuID, qty int, orderID int64) error
	ListLogs(ctx context.Context, filter LogFilter, page, size int) ([]InventoryLog, int64, error)
	ListAlerts(ctx context.Context, status string, page, size int) ([]LowStockAlert, int64, error)
	MarkAlertRead(ctx context.Context, id, adminID int64) error
	MarkAllAlertsRead(ctx context.Context, adminID int64) error
	CreateAlert(ctx context.Context, alert *LowStockAlert) error
	FindAllSKUStocks(ctx context.Context) ([]SKUStockRow, error)
	// GetSKUThreshold 获取 SKU 的低库存预警阈值（来自 sku 表）。
	GetSKUThreshold(ctx context.Context, skuID int64) (int, error)
	// HasUnreadAlert 检查是否已有该 SKU 的未读预警（防止重复创建）。
	HasUnreadAlert(ctx context.Context, skuID int64) (bool, error)
}

type inventoryRepoImpl struct{ db *gorm.DB }

// NewInventoryRepo 构造 InventoryRepo。
func NewInventoryRepo(db *gorm.DB) InventoryRepo {
	return &inventoryRepoImpl{db: db}
}

func (r *inventoryRepoImpl) GetSKUStock(ctx context.Context, skuID int64) (int, int, error) {
	var row struct {
		Stock       int
		LockedStock int
	}
	err := r.db.WithContext(ctx).
		Table("sku").
		Select("stock, locked_stock").
		Where("id = ?", skuID).
		Scan(&row).Error
	return row.Stock, row.LockedStock, err
}

// AdjustDB 手动调整库存（DB 层），写日志。changeType 由 service 控制。
func (r *inventoryRepoImpl) AdjustDB(ctx context.Context, skuID int64, change int, t, refType string, refID, operatorID int64, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct {
			Stock       int
			LockedStock int
		}
		if err := tx.Table("sku").
			Select("stock, locked_stock").
			Where("id = ?", skuID).
			Scan(&row).Error; err != nil {
			return err
		}

		newStock := row.Stock + change
		if newStock < 0 {
			newStock = 0
		}

		if err := tx.Table("sku").
			Where("id = ?", skuID).
			Update("stock", newStock).Error; err != nil {
			return err
		}

		var refIDPtr *int64
		if refID != 0 {
			refIDPtr = &refID
		}
		var opIDPtr *int64
		if operatorID != 0 {
			opIDPtr = &operatorID
		}

		log := &InventoryLog{
			ID:            snowflake.NextID(),
			SkuID:         skuID,
			Change:        change,
			Type:          t,
			RefType:       refType,
			RefID:         refIDPtr,
			BalanceBefore: row.Stock,
			BalanceAfter:  newStock,
			LockedBefore:  row.LockedStock,
			LockedAfter:   row.LockedStock,
			OperatorType:  "admin",
			OperatorID:    opIDPtr,
			Reason:        reason,
		}
		return tx.Create(log).Error
	})
}

// DeductDB 支付成功后 DB 扣减（stock -= qty, locked_stock -= qty）。
func (r *inventoryRepoImpl) DeductDB(ctx context.Context, skuID, qty int, orderID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct {
			Stock       int
			LockedStock int
		}
		if err := tx.Table("sku").
			Select("stock, locked_stock").
			Where("id = ?", skuID).
			Scan(&row).Error; err != nil {
			return err
		}

		newStock := row.Stock - qty
		newLocked := row.LockedStock - qty
		if newStock < 0 {
			newStock = 0
		}
		if newLocked < 0 {
			newLocked = 0
		}

		if err := tx.Table("sku").Where("id = ?", skuID).
			Updates(map[string]any{
				"stock":        newStock,
				"locked_stock": newLocked,
			}).Error; err != nil {
			return err
		}

		log := &InventoryLog{
			ID:            snowflake.NextID(),
			SkuID:         int64(skuID),
			Change:        -qty,
			Type:          "deduct",
			RefType:       "order",
			RefID:         &orderID,
			BalanceBefore: row.Stock,
			BalanceAfter:  newStock,
			LockedBefore:  row.LockedStock,
			LockedAfter:   newLocked,
		}
		return tx.Create(log).Error
	})
}

// LockDB 下单锁定库存（locked_stock += qty）。
func (r *inventoryRepoImpl) LockDB(ctx context.Context, skuID, qty int, orderID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct {
			Stock       int
			LockedStock int
		}
		if err := tx.Table("sku").
			Select("stock, locked_stock").
			Where("id = ?", skuID).
			Scan(&row).Error; err != nil {
			return err
		}

		newLocked := row.LockedStock + qty

		if err := tx.Table("sku").Where("id = ?", skuID).
			Update("locked_stock", newLocked).Error; err != nil {
			return err
		}

		log := &InventoryLog{
			ID:            snowflake.NextID(),
			SkuID:         int64(skuID),
			Change:        qty,
			Type:          "lock",
			RefType:       "order",
			RefID:         &orderID,
			BalanceBefore: row.Stock,
			BalanceAfter:  row.Stock,
			LockedBefore:  row.LockedStock,
			LockedAfter:   newLocked,
		}
		return tx.Create(log).Error
	})
}

// UnlockDB 关单解锁库存（locked_stock -= qty）。
func (r *inventoryRepoImpl) UnlockDB(ctx context.Context, skuID, qty int, orderID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row struct {
			Stock       int
			LockedStock int
		}
		if err := tx.Table("sku").
			Select("stock, locked_stock").
			Where("id = ?", skuID).
			Scan(&row).Error; err != nil {
			return err
		}

		newLocked := row.LockedStock - qty
		if newLocked < 0 {
			newLocked = 0
		}

		if err := tx.Table("sku").Where("id = ?", skuID).
			Update("locked_stock", newLocked).Error; err != nil {
			return err
		}

		log := &InventoryLog{
			ID:            snowflake.NextID(),
			SkuID:         int64(skuID),
			Change:        -qty,
			Type:          "unlock",
			RefType:       "order",
			RefID:         &orderID,
			BalanceBefore: row.Stock,
			BalanceAfter:  row.Stock,
			LockedBefore:  row.LockedStock,
			LockedAfter:   newLocked,
		}
		return tx.Create(log).Error
	})
}

// LogFilter 库存日志筛选条件。
type LogFilter struct {
	SkuCode    string
	ChangeType string
	StartDate  string
	EndDate    string
}

func (r *inventoryRepoImpl) ListLogs(ctx context.Context, filter LogFilter, page, size int) ([]InventoryLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&InventoryLog{})
	if filter.SkuCode != "" {
		q = q.Joins("JOIN sku ON sku.id = inventory_log.sku_id").Where("sku.sku_code = ?", filter.SkuCode)
	}
	if filter.ChangeType != "" {
		q = q.Where("inventory_log.type = ?", filter.ChangeType)
	}
	if filter.StartDate != "" {
		q = q.Where("inventory_log.created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		q = q.Where("inventory_log.created_at < ?", filter.EndDate+" 23:59:59")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []InventoryLog
	err := q.Order("inventory_log.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *inventoryRepoImpl) ListAlerts(ctx context.Context, status string, page, size int) ([]LowStockAlert, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&LowStockAlert{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []LowStockAlert
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *inventoryRepoImpl) MarkAlertRead(ctx context.Context, id, adminID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&LowStockAlert{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":  "read",
			"read_by": adminID,
			"read_at": now,
		}).Error
}

func (r *inventoryRepoImpl) MarkAllAlertsRead(ctx context.Context, adminID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&LowStockAlert{}).
		Where("status = ?", "unread").
		Updates(map[string]any{
			"status":  "read",
			"read_by": adminID,
			"read_at": now,
		}).Error
}

func (r *inventoryRepoImpl) CreateAlert(ctx context.Context, alert *LowStockAlert) error {
	alert.ID = snowflake.NextID()
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *inventoryRepoImpl) FindAllSKUStocks(ctx context.Context) ([]SKUStockRow, error) {
	var rows []SKUStockRow
	err := r.db.WithContext(ctx).
		Table("sku").
		Select("id, stock, locked_stock").
		Find(&rows).Error
	return rows, err
}

func (r *inventoryRepoImpl) GetSKUThreshold(ctx context.Context, skuID int64) (int, error) {
	var row struct {
		LowStockThreshold int `gorm:"column:low_stock_threshold"`
	}
	err := r.db.WithContext(ctx).
		Table("sku").
		Select("low_stock_threshold").
		Where("id = ?", skuID).
		Scan(&row).Error
	return row.LowStockThreshold, err
}

func (r *inventoryRepoImpl) HasUnreadAlert(ctx context.Context, skuID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LowStockAlert{}).
		Where("sku_id = ? AND status = ?", skuID, "unread").
		Count(&count).Error
	return count > 0, err
}
