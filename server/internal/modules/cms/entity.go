// Package cms 实现内容管理（文章）。
package cms

import "time"

// Article 文章。
type Article struct {
	ID          int64      `gorm:"primaryKey"                          json:"id"`
	Title       string     `gorm:"size:255;not null"                   json:"title"`
	Cover       string     `gorm:"size:512"                            json:"cover,omitempty"`
	Content     string     `gorm:"type:text"                           json:"content,omitempty"`
	Status      string     `gorm:"size:16;not null;default:'draft'"    json:"status"` // draft/published
	Sort        int        `gorm:"not null;default:0"                  json:"sort"`
	CreatedBy   *int64     `gorm:"column:created_by"                   json:"created_by,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index"                               json:"-"`
}

// TableName 指定数据库表名。
func (Article) TableName() string { return "article" }
