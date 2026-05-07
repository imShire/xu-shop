package notification

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaskFilter 通知任务列表过滤条件。
type TaskFilter struct {
	Status       string
	TemplateCode string
	Page         int
	Size         int
}

// NotificationRepo 通知数据访问接口。
type NotificationRepo interface {
	FindTemplate(ctx context.Context, code string) (*NotificationTemplate, error)
	ListTemplates(ctx context.Context) ([]NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, t *NotificationTemplate) error

	UpsertTaskByDedup(ctx context.Context, dedupKey string, task *NotificationTask) (*NotificationTask, bool, error)
	FindTaskByID(ctx context.Context, id int64) (*NotificationTask, error)
	MarkTaskSuccess(ctx context.Context, id int64) error
	MarkTaskFailed(ctx context.Context, id int64, reason string) error
	IncrRetry(ctx context.Context, id int64) error
	ListTasks(ctx context.Context, filter TaskFilter) ([]NotificationTask, int64, error)
}

type notificationRepoImpl struct{ db *gorm.DB }

// NewNotificationRepo 构造 NotificationRepo。
func NewNotificationRepo(db *gorm.DB) NotificationRepo {
	return &notificationRepoImpl{db: db}
}

func (r *notificationRepoImpl) FindTemplate(ctx context.Context, code string) (*NotificationTemplate, error) {
	var t NotificationTemplate
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *notificationRepoImpl) ListTemplates(ctx context.Context) ([]NotificationTemplate, error) {
	var list []NotificationTemplate
	if err := r.db.WithContext(ctx).Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *notificationRepoImpl) UpdateTemplate(ctx context.Context, t *NotificationTemplate) error {
	t.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(t).Error
}

// UpsertTaskByDedup 按 dedup_key 幂等插入任务。
// 返回 (task, created, error)。若已存在则 created=false，返回已有记录。
func (r *notificationRepoImpl) UpsertTaskByDedup(ctx context.Context, dedupKey string, task *NotificationTask) (*NotificationTask, bool, error) {
	if dedupKey == "" {
		if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
			return nil, false, err
		}
		return task, true, nil
	}

	task.DedupKey = &dedupKey
	result := r.db.WithContext(ctx).
		Where(clause.OnConflict{
			Columns:   []clause.Column{{Name: "dedup_key"}},
			DoNothing: true,
		}).
		Create(task)
	if result.Error != nil {
		return nil, false, result.Error
	}
	if result.RowsAffected == 0 {
		// 已存在，查回
		var existing NotificationTask
		if err := r.db.WithContext(ctx).Where("dedup_key = ?", dedupKey).First(&existing).Error; err != nil {
			return nil, false, err
		}
		return &existing, false, nil
	}
	return task, true, nil
}

func (r *notificationRepoImpl) FindTaskByID(ctx context.Context, id int64) (*NotificationTask, error) {
	var t NotificationTask
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *notificationRepoImpl) MarkTaskSuccess(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&NotificationTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":  TaskStatusSent,
			"sent_at": now,
		}).Error
}

func (r *notificationRepoImpl) MarkTaskFailed(ctx context.Context, id int64, reason string) error {
	return r.db.WithContext(ctx).Model(&NotificationTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":     TaskStatusFailed,
			"last_error": reason,
		}).Error
}

func (r *notificationRepoImpl) IncrRetry(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&NotificationTask{}).Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (r *notificationRepoImpl) ListTasks(ctx context.Context, filter TaskFilter) ([]NotificationTask, int64, error) {
	page, size := filter.Page, filter.Size
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	q := r.db.WithContext(ctx).Model(&NotificationTask{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.TemplateCode != "" {
		q = q.Where("template_code = ?", filter.TemplateCode)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []NotificationTask
	if err := q.Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
