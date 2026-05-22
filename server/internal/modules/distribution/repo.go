package distribution

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repo 分销模块仓储接口。
type Repo interface {
	DB() *gorm.DB

	// share_link
	CreateShareLink(ctx context.Context, l *ShareLink) error
	GetShareLinkByToken(ctx context.Context, token string) (*ShareLink, error)
	GetShareLinkByID(ctx context.Context, id int64) (*ShareLink, error)
	IncShareLinkCounter(ctx context.Context, id int64, field string, delta int64) error
	ListShareLinks(ctx context.Context, userID int64, scene string, limit int) ([]ShareLink, error)

	// share_click
	CreateShareClick(ctx context.Context, c *ShareClick) error

	// share_attribution
	GetAttributionByTrace(ctx context.Context, traceID string) (*ShareAttribution, error)
	UpsertAttribution(ctx context.Context, a *ShareAttribution) error
	BindAttributionUser(ctx context.Context, traceID string, userID int64) error

	// distributor
	CreateDistributor(ctx context.Context, d *Distributor) error
	GetDistributorByUserID(ctx context.Context, userID int64) (*Distributor, error)
	GetDistributorByID(ctx context.Context, id int64) (*Distributor, error)
	UpdateDistributor(ctx context.Context, id int64, fields map[string]any) error
	ListDistributors(ctx context.Context, status, level string, limit, offset int) ([]Distributor, int64, error)

	// distributor_relation
	CreateRelation(ctx context.Context, tx *gorm.DB, r *DistributorRelation) error
	GetRelationByInvitee(ctx context.Context, inviteeUserID int64) (*DistributorRelation, error)
	RenewRelation(ctx context.Context, inviteeUserID int64, expireAt time.Time) error

	// commission
	CreateCommission(ctx context.Context, c *CommissionRecord) error
	GetCommission(ctx context.Context, id int64) (*CommissionRecord, error)
	UpdateCommission(ctx context.Context, id int64, fields map[string]any) error
	ListCommissionsByUser(ctx context.Context, userID int64, status string, limit, offset int) ([]CommissionRecord, int64, error)
	ListCommissionsAdmin(ctx context.Context, status string, distributorUserID int64, limit, offset int) ([]CommissionRecord, int64, error)
	ListPendingExpired(ctx context.Context, before time.Time, limit int) ([]CommissionRecord, error)
	SumByUserStatus(ctx context.Context, userID int64) (pending, locked, settled int64, err error)
	ListLockedForUser(ctx context.Context, userID int64) ([]CommissionRecord, error)

	// withdraw
	CreateWithdraw(ctx context.Context, w *WithdrawOrder) error
	GetWithdraw(ctx context.Context, id int64) (*WithdrawOrder, error)
	GetWithdrawByIdem(ctx context.Context, idemKey string) (*WithdrawOrder, error)
	GetWithdrawByOutBillNo(ctx context.Context, outBillNo string) (*WithdrawOrder, error)
	UpdateWithdraw(ctx context.Context, id int64, fields map[string]any) error
	ListWithdrawByUser(ctx context.Context, userID int64, limit, offset int) ([]WithdrawOrder, int64, error)
	ListWithdrawAdmin(ctx context.Context, status string, limit, offset int) ([]WithdrawOrder, int64, error)
	SumWithdrawnByUser(ctx context.Context, userID int64) (int64, error)

	// settlement
	CreateSettlement(ctx context.Context, s *CommissionSettlement) error
	UpdateSettlement(ctx context.Context, id int64, fields map[string]any) error
	ListSettlements(ctx context.Context, status string, limit, offset int) ([]CommissionSettlement, int64, error)
	BindCommissionsToSettlement(ctx context.Context, ids []int64, settlementID int64) error

	// 漏斗统计
	CountShareLinks(ctx context.Context, start, end time.Time) (int64, error)
	CountShareClicks(ctx context.Context, start, end time.Time) (int64, error)
	CountAttributionRegisters(ctx context.Context, start, end time.Time) (int64, error)
	SumDistributionGMV(ctx context.Context, start, end time.Time) (orders int64, gmvCents int64, err error)
	SumCommissionsByStatus(ctx context.Context, start, end time.Time) (pending, locked, settled int64, err error)
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造仓储。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) DB() *gorm.DB { return r.db }

// ===== share_link =====

func (r *repoImpl) CreateShareLink(ctx context.Context, l *ShareLink) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *repoImpl) GetShareLinkByToken(ctx context.Context, token string) (*ShareLink, error) {
	var l ShareLink
	if err := r.db.WithContext(ctx).Where("short_token = ?", token).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *repoImpl) GetShareLinkByID(ctx context.Context, id int64) (*ShareLink, error) {
	var l ShareLink
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *repoImpl) IncShareLinkCounter(ctx context.Context, id int64, field string, delta int64) error {
	allowed := map[string]bool{"click_count": true, "register_count": true, "order_count": true, "gmv_cents": true}
	if !allowed[field] {
		return errors.New("invalid counter field")
	}
	return r.db.WithContext(ctx).Model(&ShareLink{}).Where("id = ?", id).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

func (r *repoImpl) ListShareLinks(ctx context.Context, userID int64, scene string, limit int) ([]ShareLink, error) {
	q := r.db.WithContext(ctx).Model(&ShareLink{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if scene != "" {
		q = q.Where("scene = ?", scene)
	}
	if limit <= 0 {
		limit = 50
	}
	var list []ShareLink
	if err := q.Order("created_at DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== share_click =====

func (r *repoImpl) CreateShareClick(ctx context.Context, c *ShareClick) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// ===== share_attribution =====

func (r *repoImpl) GetAttributionByTrace(ctx context.Context, traceID string) (*ShareAttribution, error) {
	var a ShareAttribution
	if err := r.db.WithContext(ctx).Where("trace_id = ?", traceID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repoImpl) UpsertAttribution(ctx context.Context, a *ShareAttribution) error {
	// 仅在不存在时插入；存在时仅刷新 last_touch_ts
	res := r.db.WithContext(ctx).Where("trace_id = ?", a.TraceID).
		Attrs(a).
		FirstOrCreate(&ShareAttribution{}, ShareAttribution{TraceID: a.TraceID})
	if res.Error != nil {
		return res.Error
	}
	return r.db.WithContext(ctx).Model(&ShareAttribution{}).
		Where("trace_id = ?", a.TraceID).
		UpdateColumn("last_touch_ts", time.Now()).Error
}

func (r *repoImpl) BindAttributionUser(ctx context.Context, traceID string, userID int64) error {
	return r.db.WithContext(ctx).Model(&ShareAttribution{}).
		Where("trace_id = ? AND user_id IS NULL", traceID).
		Update("user_id", userID).Error
}

// ===== distributor =====

func (r *repoImpl) CreateDistributor(ctx context.Context, d *Distributor) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *repoImpl) GetDistributorByUserID(ctx context.Context, userID int64) (*Distributor, error) {
	var d Distributor
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repoImpl) GetDistributorByID(ctx context.Context, id int64) (*Distributor, error) {
	var d Distributor
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repoImpl) UpdateDistributor(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&Distributor{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repoImpl) ListDistributors(ctx context.Context, status, level string, limit, offset int) ([]Distributor, int64, error) {
	q := r.db.WithContext(ctx).Model(&Distributor{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if level != "" {
		q = q.Where("level = ?", level)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []Distributor
	if err := q.Order("apply_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== distributor_relation =====

func (r *repoImpl) CreateRelation(ctx context.Context, tx *gorm.DB, rel *DistributorRelation) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(rel).Error
}

func (r *repoImpl) GetRelationByInvitee(ctx context.Context, inviteeUserID int64) (*DistributorRelation, error) {
	var rel DistributorRelation
	if err := r.db.WithContext(ctx).Where("invitee_user_id = ?", inviteeUserID).First(&rel).Error; err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *repoImpl) RenewRelation(ctx context.Context, inviteeUserID int64, expireAt time.Time) error {
	return r.db.WithContext(ctx).Model(&DistributorRelation{}).
		Where("invitee_user_id = ?", inviteeUserID).
		Updates(map[string]any{"expire_at": expireAt, "last_renewed_at": time.Now()}).Error
}

// ===== commission =====

func (r *repoImpl) CreateCommission(ctx context.Context, c *CommissionRecord) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *repoImpl) GetCommission(ctx context.Context, id int64) (*CommissionRecord, error) {
	var c CommissionRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repoImpl) UpdateCommission(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&CommissionRecord{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repoImpl) ListCommissionsByUser(ctx context.Context, userID int64, status string, limit, offset int) ([]CommissionRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&CommissionRecord{}).Where("distributor_user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []CommissionRecord
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) ListCommissionsAdmin(ctx context.Context, status string, distributorUserID int64, limit, offset int) ([]CommissionRecord, int64, error) {
	q := r.db.WithContext(ctx).Model(&CommissionRecord{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if distributorUserID > 0 {
		q = q.Where("distributor_user_id = ?", distributorUserID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []CommissionRecord
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) ListPendingExpired(ctx context.Context, before time.Time, limit int) ([]CommissionRecord, error) {
	if limit <= 0 {
		limit = 500
	}
	var list []CommissionRecord
	if err := r.db.WithContext(ctx).
		Where("status = ? AND freeze_until <= ?", CommissionStatusPending, before).
		Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repoImpl) SumByUserStatus(ctx context.Context, userID int64) (pending, locked, settled int64, err error) {
	type row struct {
		Status string
		Total  int64
	}
	var rows []row
	if err = r.db.WithContext(ctx).Model(&CommissionRecord{}).
		Select("status, COALESCE(SUM(amount_cents),0) as total").
		Where("distributor_user_id = ?", userID).
		Group("status").Scan(&rows).Error; err != nil {
		return
	}
	for _, x := range rows {
		switch x.Status {
		case CommissionStatusPending, CommissionStatusSuspect:
			pending += x.Total
		case CommissionStatusLocked:
			locked += x.Total
		case CommissionStatusSettled:
			settled += x.Total
		}
	}
	return
}

func (r *repoImpl) ListLockedForUser(ctx context.Context, userID int64) ([]CommissionRecord, error) {
	var list []CommissionRecord
	if err := r.db.WithContext(ctx).
		Where("distributor_user_id = ? AND status = ?", userID, CommissionStatusLocked).
		Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== withdraw =====

func (r *repoImpl) CreateWithdraw(ctx context.Context, w *WithdrawOrder) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *repoImpl) GetWithdraw(ctx context.Context, id int64) (*WithdrawOrder, error) {
	var w WithdrawOrder
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repoImpl) GetWithdrawByIdem(ctx context.Context, idemKey string) (*WithdrawOrder, error) {
	var w WithdrawOrder
	if err := r.db.WithContext(ctx).Where("idem_key = ?", idemKey).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repoImpl) GetWithdrawByOutBillNo(ctx context.Context, outBillNo string) (*WithdrawOrder, error) {
	var w WithdrawOrder
	if err := r.db.WithContext(ctx).Where("withdraw_no = ?", outBillNo).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repoImpl) UpdateWithdraw(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&WithdrawOrder{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repoImpl) ListWithdrawByUser(ctx context.Context, userID int64, limit, offset int) ([]WithdrawOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&WithdrawOrder{}).Where("distributor_user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []WithdrawOrder
	if err := q.Order("applied_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) ListWithdrawAdmin(ctx context.Context, status string, limit, offset int) ([]WithdrawOrder, int64, error) {
	q := r.db.WithContext(ctx).Model(&WithdrawOrder{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []WithdrawOrder
	if err := q.Order("applied_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) SumWithdrawnByUser(ctx context.Context, userID int64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&WithdrawOrder{}).
		Where("distributor_user_id = ? AND status = ?", userID, WithdrawStatusSuccess).
		Select("COALESCE(SUM(amount_cents),0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ===== settlement =====

func (r *repoImpl) CreateSettlement(ctx context.Context, s *CommissionSettlement) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repoImpl) UpdateSettlement(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&CommissionSettlement{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repoImpl) ListSettlements(ctx context.Context, status string, limit, offset int) ([]CommissionSettlement, int64, error) {
	q := r.db.WithContext(ctx).Model(&CommissionSettlement{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	var list []CommissionSettlement
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repoImpl) BindCommissionsToSettlement(ctx context.Context, ids []int64, settlementID int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&CommissionRecord{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"status":        CommissionStatusSettled,
			"settlement_id": settlementID,
			"settled_at":    time.Now(),
			"updated_at":    time.Now(),
		}).Error
}

// ===== 漏斗统计 =====

func (r *repoImpl) CountShareLinks(ctx context.Context, start, end time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&ShareLink{}).
		Where("created_at >= ? AND created_at < ?", start, end).Count(&n).Error
	return n, err
}

func (r *repoImpl) CountShareClicks(ctx context.Context, start, end time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&ShareClick{}).
		Where("ts >= ? AND ts < ?", start, end).Count(&n).Error
	return n, err
}

func (r *repoImpl) CountAttributionRegisters(ctx context.Context, start, end time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&ShareAttribution{}).
		Where("user_id IS NOT NULL AND first_touch_ts >= ? AND first_touch_ts < ?", start, end).
		Count(&n).Error
	return n, err
}

func (r *repoImpl) SumDistributionGMV(ctx context.Context, start, end time.Time) (orders int64, gmvCents int64, err error) {
	type row struct {
		Cnt int64
		GMV int64
	}
	var x row
	err = r.db.WithContext(ctx).Table(`"order"`).
		Select("COUNT(*) AS cnt, COALESCE(SUM(pay_cents),0) AS gmv").
		Where("distributor_user_id IS NOT NULL AND created_at >= ? AND created_at < ?", start, end).
		Scan(&x).Error
	return x.Cnt, x.GMV, err
}

func (r *repoImpl) SumCommissionsByStatus(ctx context.Context, start, end time.Time) (pending, locked, settled int64, err error) {
	type row struct {
		Status string
		Total  int64
	}
	var rows []row
	if err = r.db.WithContext(ctx).Model(&CommissionRecord{}).
		Select("status, COALESCE(SUM(amount_cents),0) as total").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("status").Scan(&rows).Error; err != nil {
		return
	}
	for _, x := range rows {
		switch x.Status {
		case CommissionStatusPending, CommissionStatusSuspect:
			pending += x.Total
		case CommissionStatusLocked:
			locked += x.Total
		case CommissionStatusSettled:
			settled += x.Total
		}
	}
	return
}

// 错误判断辅助。
func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
