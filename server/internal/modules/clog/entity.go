// Package clog 接收并落库前端（admin / H5 / 小程序）日志上报。
package clog

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// ClientLog 前端上报日志记录。对应表 client_log。
type ClientLog struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Source    string    `gorm:"column:source;size:16;not null"`
	Level     string    `gorm:"column:level;size:8;not null"`
	Message   string    `gorm:"column:message;type:text;not null"`
	Stack     *string   `gorm:"column:stack;type:text"`
	URL       *string   `gorm:"column:url;size:512"`
	UserAgent *string   `gorm:"column:user_agent;size:255"`
	Release   *string   `gorm:"column:release;size:64"`
	UserID    *int64    `gorm:"column:user_id"`
	AdminID   *int64    `gorm:"column:admin_id"`
	Extra     []byte    `gorm:"column:extra;type:jsonb"`
	ClientIP  *string   `gorm:"column:client_ip;type:inet"`
	TraceID   *string   `gorm:"column:trace_id;size:64"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名。
func (ClientLog) TableName() string { return "client_log" }

// Repository 前端日志写入接口。
type Repository interface {
	Insert(ctx context.Context, logs []*ClientLog) error
}

type repoImpl struct {
	db *gorm.DB
}

// NewRepository 构造仓储。
func NewRepository(db *gorm.DB) Repository {
	return &repoImpl{db: db}
}

// Insert 批量写入。
func (r *repoImpl) Insert(ctx context.Context, logs []*ClientLog) error {
	if len(logs) == 0 {
		return nil
	}
	for _, l := range logs {
		if l.ID == 0 {
			l.ID = snowflake.NextID()
		}
		if l.CreatedAt.IsZero() {
			l.CreatedAt = time.Now()
		}
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 50).Error
}
