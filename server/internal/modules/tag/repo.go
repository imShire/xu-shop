package tag

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repo 标签仓储接口。
type Repo interface {
	// ===== 字典 =====
	CreateTag(ctx context.Context, t *UserTag) error
	UpdateTag(ctx context.Context, code string, fields map[string]any) error
	DeleteTag(ctx context.Context, code string) error
	GetTag(ctx context.Context, code string) (*UserTag, error)
	ListTags(ctx context.Context, category, source string) ([]UserTag, error)
	CountRelationsByTag(ctx context.Context, code string) (int64, error)

	// ===== 关系 =====
	UpsertRelation(ctx context.Context, tx *gorm.DB, rel *UserTagRelation) error
	DeleteRelation(ctx context.Context, userID int64, tagCode string) error
	DeleteAutoByPrefix(ctx context.Context, tx *gorm.DB, prefix string) error
	DeleteAllUserAuto(ctx context.Context, tx *gorm.DB, userID int64) error
	ListUserRelations(ctx context.Context, userID int64) ([]UserTagRelation, error)
	ListAllUserIDs(ctx context.Context, lastID int64, batchSize int) ([]int64, error)

	// ===== 快照 =====
	WriteSnapshot(ctx context.Context, snap *UserTagSnapshot) error

	// ===== 单用户聚合查询 =====
	AggregateUserOrder(ctx context.Context, userID int64) (count int64, lastPaidAt *time.Time, gmvCents int64, err error)

	// ===== 人群预估 =====
	PreviewAudience(ctx context.Context, filter AudienceFilter, limit int) (count int64, sample []int64, err error)
	ListAudience(ctx context.Context, filter AudienceFilter, lastID int64, size int) ([]int64, error)

	// ===== 事务句柄 =====
	DB() *gorm.DB
}

type repoImpl struct{ db *gorm.DB }

// NewRepo 构造仓储。
func NewRepo(db *gorm.DB) Repo { return &repoImpl{db: db} }

func (r *repoImpl) DB() *gorm.DB { return r.db }

// ===== 字典 =====

func (r *repoImpl) CreateTag(ctx context.Context, t *UserTag) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repoImpl) UpdateTag(ctx context.Context, code string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&UserTag{}).Where("code = ?", code).Updates(fields).Error
}

func (r *repoImpl) DeleteTag(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Where("code = ?", code).Delete(&UserTag{}).Error
}

func (r *repoImpl) GetTag(ctx context.Context, code string) (*UserTag, error) {
	var t UserTag
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repoImpl) ListTags(ctx context.Context, category, source string) ([]UserTag, error) {
	q := r.db.WithContext(ctx).Model(&UserTag{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	var list []UserTag
	if err := q.Order("category ASC, sort ASC, code ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repoImpl) CountRelationsByTag(ctx context.Context, code string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&UserTagRelation{}).Where("tag_code = ?", code).Count(&n).Error
	return n, err
}

// ===== 关系 =====

func (r *repoImpl) UpsertRelation(ctx context.Context, tx *gorm.DB, rel *UserTagRelation) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "tag_code"}},
		DoUpdates: clause.Assignments(map[string]any{
			"score":      rel.Score,
			"source":     rel.Source,
			"source_ref": rel.SourceRef,
			"expire_at":  rel.ExpireAt,
			"updated_at": time.Now(),
		}),
	}).Create(rel).Error
}

func (r *repoImpl) DeleteRelation(ctx context.Context, userID int64, tagCode string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND tag_code = ?", userID, tagCode).
		Delete(&UserTagRelation{}).Error
}

func (r *repoImpl) DeleteAutoByPrefix(ctx context.Context, tx *gorm.DB, prefix string) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).
		Where("source = ? AND tag_code LIKE ?", SourceAuto, prefix+"%").
		Delete(&UserTagRelation{}).Error
}

func (r *repoImpl) DeleteAllUserAuto(ctx context.Context, tx *gorm.DB, userID int64) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).
		Where("user_id = ? AND source = ?", userID, SourceAuto).
		Delete(&UserTagRelation{}).Error
}

func (r *repoImpl) ListUserRelations(ctx context.Context, userID int64) ([]UserTagRelation, error) {
	var list []UserTagRelation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (expire_at IS NULL OR expire_at > now())", userID).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *repoImpl) ListAllUserIDs(ctx context.Context, lastID int64, batchSize int) ([]int64, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	var ids []int64
	err := r.db.WithContext(ctx).
		Table(`"user"`).
		Where("id > ? AND status = ?", lastID, "active").
		Order("id ASC").Limit(batchSize).
		Pluck("id", &ids).Error
	return ids, err
}

// ===== 快照 =====

func (r *repoImpl) WriteSnapshot(ctx context.Context, snap *UserTagSnapshot) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "snapshot_date"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tags"}),
	}).Create(snap).Error
}

// ===== 用户订单聚合 =====

func (r *repoImpl) AggregateUserOrder(ctx context.Context, userID int64) (int64, *time.Time, int64, error) {
	type row struct {
		Cnt        int64
		LastPaidAt *time.Time
		Gmv        int64
	}
	var rr row
	err := r.db.WithContext(ctx).
		Table(`"order"`).
		Select(`COUNT(*) AS cnt, MAX(paid_at) AS last_paid_at, COALESCE(SUM(pay_cents),0) AS gmv`).
		Where("user_id = ? AND status IN ?", userID, []string{"paid", "shipped", "completed"}).
		Scan(&rr).Error
	if err != nil {
		return 0, nil, 0, err
	}
	return rr.Cnt, rr.LastPaidAt, rr.Gmv, nil
}

// ===== 人群预估（参数化构造，禁止字符串拼接） =====

// PreviewAudience 返回 count + 最多 limit 个样例 user_id。
func (r *repoImpl) PreviewAudience(ctx context.Context, filter AudienceFilter, limit int) (int64, []int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	whereSQL, args, err := buildAudienceSQL(filter, "u")
	if err != nil {
		return 0, nil, err
	}

	// COUNT
	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM "user" u WHERE u.status = 'active' AND %s`, whereSQL)
	if err := r.db.WithContext(ctx).Raw(countQ, args...).Scan(&total).Error; err != nil {
		return 0, nil, err
	}

	// SAMPLE：固定 LIMIT，不可拼接
	var ids []int64
	listQ := fmt.Sprintf(`SELECT u.id FROM "user" u WHERE u.status = 'active' AND %s ORDER BY u.id DESC LIMIT ?`, whereSQL)
	allArgs := append([]any{}, args...)
	allArgs = append(allArgs, limit)
	if err := r.db.WithContext(ctx).Raw(listQ, allArgs...).Scan(&ids).Error; err != nil {
		return 0, nil, err
	}
	return total, ids, nil
}

// ListAudience 分页拉取 audience 用户 ID（按 id 升序，cursor 分页）。
func (r *repoImpl) ListAudience(ctx context.Context, filter AudienceFilter, lastID int64, size int) ([]int64, error) {
	if size <= 0 || size > 5000 {
		size = 500
	}
	whereSQL, args, err := buildAudienceSQL(filter, "u")
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(`SELECT u.id FROM "user" u WHERE u.status = 'active' AND u.id > ? AND %s ORDER BY u.id ASC LIMIT ?`, whereSQL)
	allArgs := []any{lastID}
	allArgs = append(allArgs, args...)
	allArgs = append(allArgs, size)
	var ids []int64
	if err := r.db.WithContext(ctx).Raw(q, allArgs...).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// ===== SQL 构造 =====

// buildAudienceSQL 递归构造 WHERE 片段；返回 (sql, args, err)，所有用户输入仅作为 ? 参数。
func buildAudienceSQL(f AudienceFilter, alias string) (string, []any, error) {
	op := strings.ToUpper(strings.TrimSpace(f.Op))
	if op == "" {
		op = "AND"
	}
	if op != "AND" && op != "OR" {
		return "", nil, errors.New("audience filter: op must be and|or")
	}

	parts := make([]string, 0, 4)
	args := make([]any, 0, 4)

	// 当前节点：标签包含 / 排除 / 行为
	if len(f.IncludeTags) > 0 {
		// IncludeTags 默认为 AND（必须同时拥有所有标签）
		for _, code := range f.IncludeTags {
			parts = append(parts, fmt.Sprintf(`EXISTS (SELECT 1 FROM user_tag_relation r WHERE r.user_id = %s.id AND r.tag_code = ? AND (r.expire_at IS NULL OR r.expire_at > now()))`, alias))
			args = append(args, code)
		}
	}
	if len(f.ExcludeTags) > 0 {
		parts = append(parts, fmt.Sprintf(`NOT EXISTS (SELECT 1 FROM user_tag_relation r WHERE r.user_id = %s.id AND r.tag_code IN (?) AND (r.expire_at IS NULL OR r.expire_at > now()))`, alias))
		args = append(args, f.ExcludeTags)
	}
	if f.Behavior != nil {
		bSQL, bArgs := buildBehaviorSQL(*f.Behavior, alias)
		if bSQL != "" {
			parts = append(parts, bSQL)
			args = append(args, bArgs...)
		}
	}

	// 子节点
	for i := range f.Children {
		s, a, err := buildAudienceSQL(f.Children[i], alias)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, "("+s+")")
		args = append(args, a...)
	}

	if len(parts) == 0 {
		return "TRUE", nil, nil
	}
	return strings.Join(parts, " "+op+" "), args, nil
}

// buildBehaviorSQL 构造行为条件子 SQL（使用 EXISTS / 子查询，全部参数化）。
func buildBehaviorSQL(b AudienceBehavior, alias string) (string, []any) {
	conds := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if b.LastOrderDaysGTE != nil {
		// 最近一次下单距今 >= N 天 → 流失类用户
		conds = append(conds, fmt.Sprintf(`COALESCE((SELECT EXTRACT(EPOCH FROM (now() - MAX(paid_at)))/86400 FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')), 99999) >= ?`, alias))
		args = append(args, *b.LastOrderDaysGTE)
	}
	if b.LastOrderDaysLTE != nil {
		conds = append(conds, fmt.Sprintf(`COALESCE((SELECT EXTRACT(EPOCH FROM (now() - MAX(paid_at)))/86400 FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')), 99999) <= ?`, alias))
		args = append(args, *b.LastOrderDaysLTE)
	}
	if b.OrderCountGTE != nil {
		conds = append(conds, fmt.Sprintf(`(SELECT COUNT(*) FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')) >= ?`, alias))
		args = append(args, *b.OrderCountGTE)
	}
	if b.OrderCountLTE != nil {
		conds = append(conds, fmt.Sprintf(`(SELECT COUNT(*) FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')) <= ?`, alias))
		args = append(args, *b.OrderCountLTE)
	}
	if b.GMVCentsGTE != nil {
		conds = append(conds, fmt.Sprintf(`COALESCE((SELECT SUM(pay_cents) FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')), 0) >= ?`, alias))
		args = append(args, *b.GMVCentsGTE)
	}
	if b.GMVCentsLTE != nil {
		conds = append(conds, fmt.Sprintf(`COALESCE((SELECT SUM(pay_cents) FROM "order" WHERE user_id=%s.id AND status IN ('paid','shipped','completed')), 0) <= ?`, alias))
		args = append(args, *b.GMVCentsLTE)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return strings.Join(conds, " AND "), args
}
