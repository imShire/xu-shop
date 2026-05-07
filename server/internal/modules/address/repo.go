package address

import (
	"context"

	"gorm.io/gorm"
)

// AddressRepo 地址数据访问接口。
type AddressRepo interface {
	List(ctx context.Context, userID int64) ([]Address, error)
	FindByID(ctx context.Context, id, userID int64) (*Address, error)
	Count(ctx context.Context, userID int64) (int64, error)
	Create(ctx context.Context, addr *Address) error
	Update(ctx context.Context, id, userID int64, updates map[string]any) error
	Delete(ctx context.Context, id, userID int64) error
	ClearDefault(ctx context.Context, userID int64) error
	SetDefault(ctx context.Context, id, userID int64) error
}

// RegionRepo 行政区划数据访问接口。
type RegionRepo interface {
	ListByParent(ctx context.Context, parentCode string) ([]Region, error)
	FindByCode(ctx context.Context, code string) (*Region, error)
}

// ---- AddressRepoImpl ----

type addressRepoImpl struct{ db *gorm.DB }

// NewAddressRepo 构造 AddressRepo 实现。
func NewAddressRepo(db *gorm.DB) AddressRepo {
	return &addressRepoImpl{db: db}
}

func (r *addressRepoImpl) List(ctx context.Context, userID int64) ([]Address, error) {
	var list []Address
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, updated_at DESC").
		Find(&list).Error
	return list, err
}

func (r *addressRepoImpl) FindByID(ctx context.Context, id, userID int64) (*Address, error) {
	var addr Address
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *addressRepoImpl) Count(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&Address{}).Where("user_id = ?", userID).Count(&cnt).Error
	return cnt, err
}

func (r *addressRepoImpl) Create(ctx context.Context, addr *Address) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *addressRepoImpl) Update(ctx context.Context, id, userID int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Address{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

func (r *addressRepoImpl) Delete(ctx context.Context, id, userID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&Address{}).Error
}

// ClearDefault 清除该用户所有地址的默认标记。
func (r *addressRepoImpl) ClearDefault(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&Address{}).
		Where("user_id = ? AND is_default = true", userID).
		Update("is_default", false).Error
}

// SetDefault 先清除再设置默认（事务保证唯一性）。
func (r *addressRepoImpl) SetDefault(ctx context.Context, id, userID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Address{}).
			Where("user_id = ? AND is_default = true", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&Address{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("is_default", true).Error
	})
}

// ---- RegionRepoImpl ----

type regionRepoImpl struct{ db *gorm.DB }

// NewRegionRepo 构造 RegionRepo 实现。
func NewRegionRepo(db *gorm.DB) RegionRepo {
	return &regionRepoImpl{db: db}
}

func (r *regionRepoImpl) ListByParent(ctx context.Context, parentCode string) ([]Region, error) {
	var list []Region
	query := r.db.WithContext(ctx).
		Table("region AS r").
		Select(`
			r.code,
			r.parent_code,
			r.name,
			r.level,
			r.sort,
			EXISTS (
				SELECT 1
				FROM region AS child
				WHERE child.parent_code = r.code
			) AS has_children
		`).
		Order("r.sort ASC, r.code ASC")
	if parentCode == "" {
		query = query.Where("r.parent_code IS NULL OR r.parent_code = ''")
	} else {
		query = query.Where("r.parent_code = ?", parentCode)
	}
	err := query.Scan(&list).Error
	return list, err
}

func (r *regionRepoImpl) FindByCode(ctx context.Context, code string) (*Region, error) {
	var region Region
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}
