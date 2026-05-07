// Package account 实现账号和权限相关的领域逻辑。
package account

import "time"

// User C 端用户表。
type User struct {
	ID               int64      `gorm:"primaryKey"`
	OpenidMP         *string    `gorm:"column:openid_mp;uniqueIndex"`
	OpenidH5         *string    `gorm:"column:openid_h5;uniqueIndex"`
	Unionid          *string    `gorm:"column:unionid;uniqueIndex"`
	Phone            *string    `gorm:"column:phone"`
	PhoneCountry     string     `gorm:"column:phone_country;default:86"`
	PasswordHash     *string    `gorm:"column:password_hash"`
	Source           string     `gorm:"column:source;default:mp"`
	Nickname         *string    `gorm:"column:nickname"`
	Avatar           *string    `gorm:"column:avatar"`
	Gender           int        `gorm:"column:gender;default:0"`
	Birthday         *time.Time `gorm:"column:birthday"`
	Status           string     `gorm:"column:status;default:active"`
	DeactivateAt     *time.Time `gorm:"column:deactivate_at"`
	InvitedByUserID  *int64     `gorm:"column:invited_by_user_id"`
	DistributorID    *int64     `gorm:"column:distributor_id"`
	Points           int        `gorm:"column:points;default:0"`
	BalanceCents     int64      `gorm:"column:balance_cents;not null;default:0"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 对应 PostgreSQL 的 "user" 表（关键字需加引号）。
func (User) TableName() string { return `"user"` }

// Admin 后台管理员表。
type Admin struct {
	ID             int64      `gorm:"primaryKey"`
	Username       string     `gorm:"column:username;uniqueIndex;not null"`
	PasswordHash   string     `gorm:"column:password_hash;not null"`
	RealName       *string    `gorm:"column:real_name"`
	Phone          *string    `gorm:"column:phone"`
	Status         string     `gorm:"column:status;default:active"`
	FailedAttempts int        `gorm:"column:failed_attempts;default:0"`
	LockedUntil    *time.Time `gorm:"column:locked_until"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at"`
	LastLoginIP    *string    `gorm:"column:last_login_ip"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (Admin) TableName() string { return "admin" }

// Role 角色表。
type Role struct {
	ID        int64     `gorm:"primaryKey"`
	Code      string    `gorm:"column:code;uniqueIndex;not null"`
	Name      string    `gorm:"column:name;not null"`
	IsSystem  bool      `gorm:"column:is_system;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	Permissions []Permission `gorm:"many2many:role_permission;joinForeignKey:role_id;joinReferences:permission_code"`
}

func (Role) TableName() string { return "role" }

// Permission 权限点表。
type Permission struct {
	Code   string `gorm:"primaryKey"`
	Module string `gorm:"column:module;not null"`
	Action string `gorm:"column:action;not null"`
	Name   string `gorm:"column:name;not null"`
}

func (Permission) TableName() string { return "permission" }

// AdminRole 管理员-角色关联表。
type AdminRole struct {
	AdminID int64 `gorm:"primaryKey;column:admin_id"`
	RoleID  int64 `gorm:"primaryKey;column:role_id"`
}

func (AdminRole) TableName() string { return "admin_role" }

// RolePermission 角色-权限关联表。
type RolePermission struct {
	RoleID         int64  `gorm:"primaryKey;column:role_id"`
	PermissionCode string `gorm:"primaryKey;column:permission_code"`
}

func (RolePermission) TableName() string { return "role_permission" }

// LoginLog 登录日志。
type LoginLog struct {
	ID          int64     `gorm:"primaryKey"`
	SubjectType string    `gorm:"column:subject_type;not null"` // user / admin
	SubjectID   *int64    `gorm:"column:subject_id"`
	IP          *string   `gorm:"column:ip"`
	UA          *string   `gorm:"column:ua"`
	Success     bool      `gorm:"column:success;not null"`
	FailReason  *string   `gorm:"column:fail_reason"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (LoginLog) TableName() string { return "login_log" }

// BalanceLog 余额流水。
type BalanceLog struct {
	ID                 int64     `gorm:"primaryKey"     json:"id"`
	UserID             int64     `gorm:"not null"       json:"user_id"`
	ChangeCents        int64     `gorm:"not null"       json:"change_cents"`
	Type               string    `gorm:"size:16;not null" json:"type"` // recharge/spend/refund
	RefType            string    `gorm:"size:16"        json:"ref_type,omitempty"`
	RefID              *int64    `json:"ref_id,omitempty"`
	BalanceBeforeCents int64     `gorm:"not null"       json:"balance_before_cents"`
	BalanceAfterCents  int64     `gorm:"not null"       json:"balance_after_cents"`
	OperatorID         *int64    `json:"operator_id,omitempty"`
	Remark             string    `gorm:"size:200"       json:"remark,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

func (BalanceLog) TableName() string { return "balance_log" }
