// Package audit 提供后台审计日志的实体与持久化。
// 中间件 middleware.AuditLog 负责采集；本包负责落库。
package audit

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// AuditLog 后台审计日志条目。对应表 admin_audit_log。
type AuditLog struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	AdminID         int64     `gorm:"column:admin_id;not null;default:0"`
	Action          string    `gorm:"column:action;size:64;not null"`
	TargetType      *string   `gorm:"column:target_type;size:64"`
	TargetID        *string   `gorm:"column:target_id;size:64"`
	Method          string    `gorm:"column:method;size:8;not null"`
	Path            string    `gorm:"column:path;size:255;not null"`
	Query           *string   `gorm:"column:query;type:text"`
	RequestBody     []byte    `gorm:"column:request_body;type:jsonb"`
	ResponseStatus  int       `gorm:"column:response_status;not null"`
	ResponseExcerpt *string   `gorm:"column:response_excerpt;type:text"`
	ClientIP        *string   `gorm:"column:client_ip;type:inet"`
	UserAgent       *string   `gorm:"column:user_agent;size:255"`
	TraceID         *string   `gorm:"column:trace_id;size:64"`
	DurationMs      int       `gorm:"column:duration_ms;not null;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名。
func (AuditLog) TableName() string { return "admin_audit_log" }

// Repository 审计日志写入接口。
type Repository interface {
	Insert(ctx context.Context, log *AuditLog) error
}

type repoImpl struct {
	db *gorm.DB
}

// NewRepository 构造审计日志仓储。
func NewRepository(db *gorm.DB) Repository {
	return &repoImpl{db: db}
}

// Insert 写入一条审计日志；如未设置 ID 则用 snowflake 生成。
func (r *repoImpl) Insert(ctx context.Context, log *AuditLog) error {
	if log.ID == 0 {
		log.ID = snowflake.NextID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(log).Error
}
