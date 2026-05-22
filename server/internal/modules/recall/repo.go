package recall

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repo 召回模块持久层。
type Repo interface {
	CreateCampaign(ctx context.Context, c *RecallCampaign) error
	UpdateCampaign(ctx context.Context, id int64, fields map[string]any) error
	GetCampaign(ctx context.Context, id int64) (*RecallCampaign, error)
	ListCampaigns(ctx context.Context, status string, page, size int) ([]RecallCampaign, int64, error)
	ListOnlineByTrigger(ctx context.Context, triggerType string) ([]RecallCampaign, error)

	InsertLog(ctx context.Context, l *RecallLog) error
	HasLogToday(ctx context.Context, campaignID, userID int64) (bool, error)
	CountLogToday(ctx context.Context, campaignID int64) (int64, error)
	CountLogTotal(ctx context.Context, campaignID int64) (int64, error)
	FunnelStats(ctx context.Context, campaignID int64) (triggered, opened, converted, gmvCents int64, err error)
	ListLogs(ctx context.Context, campaignID int64, page, size int) ([]RecallLog, int64, error)

	// AttributeOrder 在归因窗口内找该用户最近一次未归因日志，更新转化字段。
	AttributeOrder(ctx context.Context, userID, orderID int64, paidAt time.Time, gmvCents int64, windowDaysFallback int) (int64, error)

	DB() *gorm.DB
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) DB() *gorm.DB { return r.db }

func (r *repoImpl) CreateCampaign(ctx context.Context, c *RecallCampaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *repoImpl) UpdateCampaign(ctx context.Context, id int64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&RecallCampaign{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repoImpl) GetCampaign(ctx context.Context, id int64) (*RecallCampaign, error) {
	var c RecallCampaign
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repoImpl) ListCampaigns(ctx context.Context, status string, page, size int) ([]RecallCampaign, int64, error) {
	q := r.db.WithContext(ctx).Model(&RecallCampaign{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	var list []RecallCampaign
	if err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) ListOnlineByTrigger(ctx context.Context, triggerType string) ([]RecallCampaign, error) {
	var list []RecallCampaign
	q := r.db.WithContext(ctx).Where("status = ?", StatusOnline)
	if triggerType != "" {
		q = q.Where("trigger_type = ?", triggerType)
	}
	now := time.Now()
	q = q.Where("(effective_from IS NULL OR effective_from <= ?)", now).
		Where("(effective_to IS NULL OR effective_to >= ?)", now)
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repoImpl) InsertLog(ctx context.Context, l *RecallLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *repoImpl) HasLogToday(ctx context.Context, campaignID, userID int64) (bool, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&RecallLog{}).
		Where("campaign_id = ? AND user_id = ? AND triggered_at::date = current_date", campaignID, userID).
		Count(&cnt).Error
	return cnt > 0, err
}

func (r *repoImpl) CountLogToday(ctx context.Context, campaignID int64) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&RecallLog{}).
		Where("campaign_id = ? AND triggered_at::date = current_date", campaignID).
		Count(&cnt).Error
	return cnt, err
}

func (r *repoImpl) CountLogTotal(ctx context.Context, campaignID int64) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&RecallLog{}).
		Where("campaign_id = ?", campaignID).
		Count(&cnt).Error
	return cnt, err
}

func (r *repoImpl) FunnelStats(ctx context.Context, campaignID int64) (triggered, opened, converted, gmvCents int64, err error) {
	row := struct {
		Triggered int64
		Opened    int64
		Converted int64
		GMVCents  int64
	}{}
	err = r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) AS triggered,
			COUNT(*) FILTER (WHERE opened_at IS NOT NULL) AS opened,
			COUNT(*) FILTER (WHERE converted_order_id IS NOT NULL) AS converted,
			COALESCE(SUM(converted_gmv_cents), 0) AS gmv_cents
		FROM recall_log
		WHERE campaign_id = ?
	`, campaignID).Scan(&row).Error
	return row.Triggered, row.Opened, row.Converted, row.GMVCents, err
}

func (r *repoImpl) ListLogs(ctx context.Context, campaignID int64, page, size int) ([]RecallLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&RecallLog{}).Where("campaign_id = ?", campaignID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	var list []RecallLog
	if err := q.Order("triggered_at DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) AttributeOrder(ctx context.Context, userID, orderID int64, paidAt time.Time, gmvCents int64, windowDaysFallback int) (int64, error) {
	if windowDaysFallback <= 0 {
		windowDaysFallback = 7
	}
	// 找最近一条 triggered_at 在归因窗口内、且尚未归因的日志
	var l RecallLog
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND converted_order_id IS NULL", userID).
		Where("triggered_at >= ?", paidAt.Add(-time.Duration(windowDaysFallback)*24*time.Hour)).
		Where("triggered_at <= ?", paidAt).
		Order("triggered_at DESC").
		Limit(1).
		First(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	now := paidAt
	res := r.db.WithContext(ctx).Model(&RecallLog{}).
		Where("id = ? AND converted_order_id IS NULL", l.ID).
		Updates(map[string]any{
			"converted_order_id":  orderID,
			"converted_at":        now,
			"converted_gmv_cents": gmvCents,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, nil
	}
	return l.CampaignID, nil
}
