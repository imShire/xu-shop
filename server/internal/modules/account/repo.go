package account

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// UserRepo user 数据访问接口。
type UserRepo interface {
	FindByOpenidMP(ctx context.Context, openid string) (*User, error)
	FindByOpenidH5(ctx context.Context, openid string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	UpsertByOpenidMP(ctx context.Context, user *User) error
	UpsertByOpenidH5(ctx context.Context, user *User) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	CountByPhone(ctx context.Context, phone string) (int64, error)
	SessionKeyKey(userID int64) string
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	SetPasswordHash(ctx context.Context, userID int64, hash string) error
	CreateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, phone, nickname, status string, page, pageSize int) ([]User, int64, error)
	GetBalance(ctx context.Context, userID int64) (int64, error)
	RechargeBalance(ctx context.Context, userID int64, amountCents int64, operatorID int64, remark string) error
	DeductBalance(ctx context.Context, userID int64, amountCents int64, refType string, refID int64, remark string) error
	RefundBalance(ctx context.Context, userID int64, amountCents int64, refType string, refID int64, remark string) error
	ListBalanceLogs(ctx context.Context, userID int64, page, size int) ([]BalanceLog, int64, error)
}

// AdminRepo admin 数据访问接口。
type AdminRepo interface {
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	FindByID(ctx context.Context, id int64) (*Admin, error)
	Create(ctx context.Context, admin *Admin) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	List(ctx context.Context, page, pageSize int) ([]Admin, int64, error)
	SaveLoginLog(ctx context.Context, log *LoginLog) error
}

// RoleRepo 角色数据访问接口。
type RoleRepo interface {
	ListRoles(ctx context.Context) ([]Role, error)
	FindRoleByID(ctx context.Context, id int64) (*Role, error)
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, id int64, updates map[string]any) error
	DeleteRole(ctx context.Context, id int64) error
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetAdminPermCodes(ctx context.Context, adminID int64) ([]string, error)
	GetAdminRoleCodes(ctx context.Context, adminID int64) ([]string, error)
	SetAdminRoles(ctx context.Context, adminID int64, roleIDs []int64) error
	SetRolePerms(ctx context.Context, roleID int64, permCodes []string) error
}

// ------- UserRepoImpl -------

type userRepoImpl struct{ db *gorm.DB }

// NewUserRepo 构造 UserRepo 实现。
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepoImpl{db: db}
}

func (r *userRepoImpl) FindByOpenidMP(ctx context.Context, openid string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("openid_mp = ?", openid).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoImpl) FindByOpenidH5(ctx context.Context, openid string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("openid_h5 = ?", openid).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoImpl) FindByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpsertByOpenidMP 以 openid_mp 为唯一键 upsert。
func (r *userRepoImpl) UpsertByOpenidMP(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "openid_mp"}},
			DoUpdates: clause.AssignmentColumns([]string{"unionid", "updated_at"}),
		}).
		Create(user).Error
}

// UpsertByOpenidH5 以 openid_h5 为唯一键 upsert。
func (r *userRepoImpl) UpsertByOpenidH5(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "openid_h5"}},
			DoUpdates: clause.AssignmentColumns([]string{"unionid", "updated_at"}),
		}).
		Create(user).Error
}

func (r *userRepoImpl) Update(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepoImpl) CountByPhone(ctx context.Context, phone string) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&User{}).
		Where("phone = ? AND status != 'deactivated'", phone).
		Count(&cnt).Error
	return cnt, err
}

func (r *userRepoImpl) SessionKeyKey(userID int64) string {
	return ""
}

// GetUserByPhone 按手机号查询活跃用户。
func (r *userRepoImpl) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("phone = ? AND status != 'deactivated'", phone).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SetPasswordHash 更新用户密码 hash。
func (r *userRepoImpl) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

// CreateUser 创建新用户。
func (r *userRepoImpl) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// ListUsers 分页查询用户列表。
func (r *userRepoImpl) ListUsers(ctx context.Context, phone, nickname, status string, page, pageSize int) ([]User, int64, error) {
	q := r.db.WithContext(ctx).Model(&User{})
	if phone != "" {
		q = q.Where("phone LIKE ?", "%"+phone+"%")
	}
	if nickname != "" {
		q = q.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []User
	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// GetBalance 查询用户当前余额。
func (r *userRepoImpl) GetBalance(ctx context.Context, userID int64) (int64, error) {
	var u User
	err := r.db.WithContext(ctx).Select("balance_cents").First(&u, userID).Error
	return u.BalanceCents, err
}

// RechargeBalance Admin 充值：原子递增余额并记录流水。
func (r *userRepoImpl) RechargeBalance(ctx context.Context, userID int64, amountCents int64, operatorID int64, remark string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u User
		if err := tx.Select("id, balance_cents").First(&u, userID).Error; err != nil {
			return err
		}
		newBalance := u.BalanceCents + amountCents
		if err := tx.Model(&User{}).Where("id = ?", userID).
			Update("balance_cents", newBalance).Error; err != nil {
			return err
		}
		var opIDPtr *int64
		if operatorID > 0 {
			opIDPtr = &operatorID
		}
		bl := &BalanceLog{
			ID:                 snowflake.NextID(),
			UserID:             userID,
			ChangeCents:        amountCents,
			Type:               "recharge",
			BalanceBeforeCents: u.BalanceCents,
			BalanceAfterCents:  newBalance,
			OperatorID:         opIDPtr,
			Remark:             remark,
		}
		return tx.Create(bl).Error
	})
}

// DeductBalance 扣减余额（原子 CAS + 流水）。
func (r *userRepoImpl) DeductBalance(ctx context.Context, userID int64, amountCents int64, refType string, refID int64, remark string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u User
		if err := tx.Select("id, balance_cents").First(&u, userID).Error; err != nil {
			return err
		}
		if u.BalanceCents < amountCents {
			return errs.ErrParam.WithMsg("余额不足")
		}
		newBalance := u.BalanceCents - amountCents
		res := tx.Model(&User{}).
			Where("id = ? AND balance_cents >= ?", userID, amountCents).
			Update("balance_cents", newBalance)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errs.ErrParam.WithMsg("余额不足")
		}
		var refIDPtr *int64
		if refID > 0 {
			refIDPtr = &refID
		}
		bl := &BalanceLog{
			ID:                 snowflake.NextID(),
			UserID:             userID,
			ChangeCents:        -amountCents,
			Type:               "spend",
			RefType:            refType,
			RefID:              refIDPtr,
			BalanceBeforeCents: u.BalanceCents,
			BalanceAfterCents:  newBalance,
			Remark:             remark,
		}
		return tx.Create(bl).Error
	})
}

// RefundBalance 退款到余额（原子递增 + 流水）。
func (r *userRepoImpl) RefundBalance(ctx context.Context, userID int64, amountCents int64, refType string, refID int64, remark string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u User
		if err := tx.Select("id, balance_cents").First(&u, userID).Error; err != nil {
			return err
		}
		newBalance := u.BalanceCents + amountCents
		if err := tx.Model(&User{}).Where("id = ?", userID).
			Update("balance_cents", newBalance).Error; err != nil {
			return err
		}
		var refIDPtr *int64
		if refID > 0 {
			refIDPtr = &refID
		}
		bl := &BalanceLog{
			ID:                 snowflake.NextID(),
			UserID:             userID,
			ChangeCents:        amountCents,
			Type:               "refund",
			RefType:            refType,
			RefID:              refIDPtr,
			BalanceBeforeCents: u.BalanceCents,
			BalanceAfterCents:  newBalance,
			Remark:             remark,
		}
		return tx.Create(bl).Error
	})
}

// ListBalanceLogs 分页查询余额流水。
func (r *userRepoImpl) ListBalanceLogs(ctx context.Context, userID int64, page, size int) ([]BalanceLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	var list []BalanceLog
	q := r.db.WithContext(ctx).Model(&BalanceLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ------- AdminRepoImpl -------

type adminRepoImpl struct{ db *gorm.DB }

// NewAdminRepo 构造 AdminRepo 实现。
func NewAdminRepo(db *gorm.DB) AdminRepo {
	return &adminRepoImpl{db: db}
}

func (r *adminRepoImpl) FindByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *adminRepoImpl) FindByID(ctx context.Context, id int64) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *adminRepoImpl) Create(ctx context.Context, admin *Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *adminRepoImpl) Update(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Admin{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r *adminRepoImpl) List(ctx context.Context, page, pageSize int) ([]Admin, int64, error) {
	var list []Admin
	var total int64
	query := r.db.WithContext(ctx).Model(&Admin{}).Where("deleted_at IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *adminRepoImpl) SaveLoginLog(ctx context.Context, log *LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ------- RoleRepoImpl -------

type roleRepoImpl struct{ db *gorm.DB }

// NewRoleRepo 构造 RoleRepo 实现。
func NewRoleRepo(db *gorm.DB) RoleRepo {
	return &roleRepoImpl{db: db}
}

func (r *roleRepoImpl) ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepoImpl) FindRoleByID(ctx context.Context, id int64) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepoImpl) CreateRole(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepoImpl) UpdateRole(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Role{}).Where("id = ?", id).Updates(updates).Error
}

func (r *roleRepoImpl) DeleteRole(ctx context.Context, id int64) error {
	// 系统角色不可删
	var role Role
	if err := r.db.WithContext(ctx).Select("is_system").First(&role, id).Error; err != nil {
		return err
	}
	if role.IsSystem {
		return &deleteSystemRoleErr{}
	}
	return r.db.WithContext(ctx).Delete(&Role{}, id).Error
}

func (r *roleRepoImpl) ListPermissions(ctx context.Context) ([]Permission, error) {
	var perms []Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

// GetAdminPermCodes 查询管理员所有权限点 code（通过角色关联）。
func (r *roleRepoImpl) GetAdminPermCodes(ctx context.Context, adminID int64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("role_permission rp").
		Select("DISTINCT rp.permission_code").
		Joins("JOIN admin_role ar ON ar.role_id = rp.role_id").
		Where("ar.admin_id = ?", adminID).
		Pluck("permission_code", &codes).Error
	return codes, err
}

// GetAdminRoleCodes 查询管理员所有角色 code。
func (r *roleRepoImpl) GetAdminRoleCodes(ctx context.Context, adminID int64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("role r").
		Select("r.code").
		Joins("JOIN admin_role ar ON ar.role_id = r.id").
		Where("ar.admin_id = ?", adminID).
		Pluck("code", &codes).Error
	return codes, err
}

// SetAdminRoles 重新设置管理员角色绑定（先删后插）。
func (r *roleRepoImpl) SetAdminRoles(ctx context.Context, adminID int64, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("admin_id = ?", adminID).Delete(&AdminRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			if err := tx.Create(&AdminRole{AdminID: adminID, RoleID: rid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SetRolePerms 重新设置角色权限绑定（先删后插）。
func (r *roleRepoImpl) SetRolePerms(ctx context.Context, roleID int64, permCodes []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		for _, code := range permCodes {
			if err := tx.Create(&RolePermission{RoleID: roleID, PermissionCode: code}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// deleteSystemRoleErr 系统角色不可删除错误。
type deleteSystemRoleErr struct{}

func (e *deleteSystemRoleErr) Error() string { return "系统角色不可删除" }
