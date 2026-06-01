package reconciliation

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Filter 差异列表过滤条件。
type Filter struct {
	Job      string
	BizDate  *time.Time
	Status   string
	Severity string
	Page     int
	Size     int
}

// Repository 对账差异仓储。
type Repository interface {
	Upsert(ctx context.Context, d *Diff) error
	List(ctx context.Context, f Filter) ([]Diff, int64, error)
	Get(ctx context.Context, id int64) (*Diff, error)
	UpdateStatus(ctx context.Context, id int64, status string, operatorID int64, note *string) error
}

type repoImpl struct{ db *gorm.DB }

// NewRepository 构造仓储。
func NewRepository(db *gorm.DB) Repository { return &repoImpl{db: db} }

// Upsert 按 (job, biz_date, ref_type, ref_id, field) 唯一键去重写入。
// 已存在则更新期望/实际/严重程度/差额（不覆盖人工处置状态）。
func (r *repoImpl) Upsert(ctx context.Context, d *Diff) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "job"}, {Name: "biz_date"}, {Name: "ref_type"},
			{Name: "ref_id"}, {Name: "field"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"expected_value": d.ExpectedValue,
			"actual_value":   d.ActualValue,
			"diff_cents":     d.DiffCents,
			"severity":       d.Severity,
			"updated_at":     time.Now(),
		}),
	}).Create(d).Error
}

func (r *repoImpl) List(ctx context.Context, f Filter) ([]Diff, int64, error) {
	q := r.db.WithContext(ctx).Model(&Diff{})
	if f.Job != "" {
		q = q.Where("job = ?", f.Job)
	}
	if f.BizDate != nil {
		q = q.Where("biz_date = ?", f.BizDate.Format("2006-01-02"))
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Severity != "" {
		q = q.Where("severity = ?", f.Severity)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 200 {
		f.Size = 20
	}
	var list []Diff
	if err := q.Order("created_at DESC").
		Offset((f.Page - 1) * f.Size).Limit(f.Size).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) Get(ctx context.Context, id int64) (*Diff, error) {
	var d Diff
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repoImpl) UpdateStatus(ctx context.Context, id int64, status string, operatorID int64, note *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	if note != nil {
		updates["note"] = *note
	}
	switch status {
	case StatusAcknowledged:
		updates["acked_by"] = operatorID
		updates["acked_at"] = now
	case StatusResolved:
		updates["resolved_by"] = operatorID
		updates["resolved_at"] = now
	}
	return r.db.WithContext(ctx).Model(&Diff{}).
		Where("id = ?", id).Updates(updates).Error
}
