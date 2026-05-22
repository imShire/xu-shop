package coupon

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repo 优惠券仓储接口。
type Repo interface {
	// 模板
	FindTemplate(ctx context.Context, id int64) (*CouponTemplate, error)
	FindTemplateForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*CouponTemplate, error)
	IncrTemplateClaimed(ctx context.Context, tx *gorm.DB, id int64, delta int64) error
	IncrTemplateUsed(ctx context.Context, tx *gorm.DB, id int64, delta int64) error
	ListOnlineTemplates(ctx context.Context, page, size int) ([]CouponTemplate, int64, error)

	// 用户券
	CreateUserCoupon(ctx context.Context, tx *gorm.DB, uc *UserCoupon) error
	FindUserCoupon(ctx context.Context, id int64) (*UserCoupon, error)
	FindUserCouponForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*UserCoupon, error)
	UpdateUserCouponStatus(ctx context.Context, tx *gorm.DB, id int64, fromStatus, toStatus string, fields map[string]any) (int64, error)
	CountUserClaim(ctx context.Context, userID, templateID int64) (int64, error)
	ListMyCoupons(ctx context.Context, userID int64, status string, page, size int) ([]UserCoupon, int64, error)
	FindUserCouponByOrder(ctx context.Context, orderID int64) (*UserCoupon, error)

	// 兑换码
	FindRedeemCode(ctx context.Context, code string) (*RedeemCode, error)
	UseRedeemCode(ctx context.Context, tx *gorm.DB, codeID int64, userID int64) (int64, error)

	// 发放任务
	CreateGrantTask(ctx context.Context, t *GrantTask) error
	UpdateGrantTaskProgress(ctx context.Context, id int64, granted, failed int64, status string) error

	// 过期扫描
	ScanExpire(ctx context.Context, before time.Time, limit int) ([]UserCoupon, error)
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造仓储。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) FindTemplate(ctx context.Context, id int64) (*CouponTemplate, error) {
	var t CouponTemplate
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repoImpl) FindTemplateForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*CouponTemplate, error) {
	var t CouponTemplate
	q := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", id)
	if err := q.First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repoImpl) IncrTemplateClaimed(ctx context.Context, tx *gorm.DB, id int64, delta int64) error {
	return tx.WithContext(ctx).Model(&CouponTemplate{}).Where("id = ?", id).
		UpdateColumn("claimed_count", gorm.Expr("claimed_count + ?", delta)).Error
}

func (r *repoImpl) IncrTemplateUsed(ctx context.Context, tx *gorm.DB, id int64, delta int64) error {
	return tx.WithContext(ctx).Model(&CouponTemplate{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", delta)).Error
}

func (r *repoImpl) ListOnlineTemplates(ctx context.Context, page, size int) ([]CouponTemplate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	var total int64
	if err := r.db.WithContext(ctx).Model(&CouponTemplate{}).
		Where("status = ? AND deleted_at IS NULL", TplStatusOnline).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []CouponTemplate
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", TplStatusOnline).
		Order("created_at DESC").
		Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}

func (r *repoImpl) CreateUserCoupon(ctx context.Context, tx *gorm.DB, uc *UserCoupon) error {
	return tx.WithContext(ctx).Create(uc).Error
}

func (r *repoImpl) FindUserCoupon(ctx context.Context, id int64) (*UserCoupon, error) {
	var uc UserCoupon
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&uc).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *repoImpl) FindUserCouponForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*UserCoupon, error) {
	var uc UserCoupon
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&uc).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *repoImpl) UpdateUserCouponStatus(ctx context.Context, tx *gorm.DB, id int64, fromStatus, toStatus string, fields map[string]any) (int64, error) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["status"] = toStatus
	res := tx.WithContext(ctx).Model(&UserCoupon{}).
		Where("id = ? AND status = ?", id, fromStatus).Updates(fields)
	return res.RowsAffected, res.Error
}

func (r *repoImpl) CountUserClaim(ctx context.Context, userID, templateID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&UserCoupon{}).
		Where("user_id = ? AND coupon_template_id = ?", userID, templateID).
		Count(&n).Error
	return n, err
}

func (r *repoImpl) ListMyCoupons(ctx context.Context, userID int64, status string, page, size int) ([]UserCoupon, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&UserCoupon{}).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []UserCoupon
	err := q.Order("claimed_at DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}

func (r *repoImpl) FindUserCouponByOrder(ctx context.Context, orderID int64) (*UserCoupon, error) {
	var uc UserCoupon
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Where("status IN ?", []string{UCStatusLocked, UCStatusUsed}).
		First(&uc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &uc, nil
}

func (r *repoImpl) FindRedeemCode(ctx context.Context, code string) (*RedeemCode, error) {
	var rc RedeemCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&rc).Error; err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *repoImpl) UseRedeemCode(ctx context.Context, tx *gorm.DB, codeID int64, userID int64) (int64, error) {
	now := time.Now()
	res := tx.WithContext(ctx).Model(&RedeemCode{}).
		Where("id = ? AND status = ?", codeID, RCStatusUnused).
		Updates(map[string]any{
			"status":           RCStatusUsed,
			"used_by_user_id":  userID,
			"used_at":          now,
		})
	return res.RowsAffected, res.Error
}

func (r *repoImpl) CreateGrantTask(ctx context.Context, t *GrantTask) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repoImpl) UpdateGrantTaskProgress(ctx context.Context, id int64, granted, failed int64, status string) error {
	now := time.Now()
	updates := map[string]any{
		"granted_count": granted,
		"failed_count":  failed,
		"status":        status,
	}
	if status == GTStatusDone || status == GTStatusFailed {
		updates["finished_at"] = now
	}
	if status == GTStatusRunning {
		updates["started_at"] = now
	}
	return r.db.WithContext(ctx).Model(&GrantTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repoImpl) ScanExpire(ctx context.Context, before time.Time, limit int) ([]UserCoupon, error) {
	var list []UserCoupon
	err := r.db.WithContext(ctx).
		Where("status = ? AND expire_at <= ?", UCStatusUnused, before).
		Limit(limit).Find(&list).Error
	return list, err
}
