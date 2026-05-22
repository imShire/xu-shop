package tag

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// Service 标签服务。
type Service struct {
	repo Repo
	db   *gorm.DB
}

// NewService 构造服务。
func NewService(repo Repo, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// ===== 字典 CRUD =====

// ListTags 列字典。
func (s *Service) ListTags(ctx context.Context, category, source string) ([]UserTag, error) {
	return s.repo.ListTags(ctx, category, source)
}

// CreateTag 创建标签字典（仅允许 manual 来源 + business/member 类目）。
func (s *Service) CreateTag(ctx context.Context, req CreateTagReq) (*UserTag, error) {
	if req.Code == "" {
		return nil, errs.ErrParam
	}
	t := &UserTag{
		Code:        req.Code,
		Name:        req.Name,
		Category:    req.Category,
		ParentCode:  req.ParentCode,
		Color:       req.Color,
		Description: req.Description,
		Source:      SourceManual,
		Config:      JSONMap{},
		Enabled:     true,
		Sort:        req.Sort,
	}
	if err := s.repo.CreateTag(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// UpdateTag 更新字典（仅 manual 来源允许编辑核心字段；auto 仅允许改 enabled / sort）。
func (s *Service) UpdateTag(ctx context.Context, code string, req UpdateTagReq) error {
	t, err := s.repo.GetTag(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	fields := map[string]any{}
	if req.Enabled != nil {
		fields["enabled"] = *req.Enabled
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if t.Source == SourceManual {
		if req.Name != nil {
			fields["name"] = *req.Name
		}
		if req.Color != nil {
			fields["color"] = *req.Color
		}
		if req.Description != nil {
			fields["description"] = *req.Description
		}
	}
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = time.Now()
	return s.repo.UpdateTag(ctx, code, fields)
}

// DeleteTag 删除字典（仅 manual 且无关系绑定时允许）。
func (s *Service) DeleteTag(ctx context.Context, code string) error {
	t, err := s.repo.GetTag(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if t.Source != SourceManual {
		return errs.ErrForbidden.WithMsg("auto 标签不可删除")
	}
	cnt, err := s.repo.CountRelationsByTag(ctx, code)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errs.ErrConflict.WithMsg("标签存在用户关系，无法删除")
	}
	return s.repo.DeleteTag(ctx, code)
}

// ===== 用户标签 =====

// GetUserTags 查询某用户全部生效标签（含字典名称）。
func (s *Service) GetUserTags(ctx context.Context, userID int64) ([]UserTagResp, error) {
	rels, err := s.repo.ListUserRelations(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return []UserTagResp{}, nil
	}
	codes := make([]string, 0, len(rels))
	for i := range rels {
		codes = append(codes, rels[i].TagCode)
	}
	dict := make(map[string]UserTag)
	var tags []UserTag
	if err := s.repo.DB().WithContext(ctx).Where("code IN ?", codes).Find(&tags).Error; err != nil {
		return nil, err
	}
	for i := range tags {
		dict[tags[i].Code] = tags[i]
	}
	resp := make([]UserTagResp, 0, len(rels))
	for i := range rels {
		r := rels[i]
		item := UserTagResp{
			TagCode:   r.TagCode,
			Source:    r.Source,
			Score:     r.Score,
			ExpireAt:  r.ExpireAt,
			CreatedAt: r.CreatedAt,
		}
		item.ID = types.Int64Str(r.ID)
		if t, ok := dict[r.TagCode]; ok {
			item.TagName = t.Name
			item.Category = t.Category
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// AddManualTag 手动给用户加标签。
func (s *Service) AddManualTag(ctx context.Context, userID int64, req AddManualTagReq) error {
	t, err := s.repo.GetTag(ctx, req.TagCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound.WithMsg("标签不存在")
		}
		return err
	}
	if !t.Enabled {
		return errs.ErrForbidden.WithMsg("标签已禁用")
	}
	rel := &UserTagRelation{
		ID:       snowflake.NextID(),
		UserID:   userID,
		TagCode:  req.TagCode,
		Score:    1,
		Source:   SourceManual,
		ExpireAt: req.ExpireAt,
	}
	return s.repo.UpsertRelation(ctx, nil, rel)
}

// RemoveManualTag 删除某用户某标签。
func (s *Service) RemoveManualTag(ctx context.Context, userID int64, tagCode string) error {
	return s.repo.DeleteRelation(ctx, userID, tagCode)
}

// ===== 人群预估 =====

// PreviewAudience 估算人群规模 + 抽样。
func (s *Service) PreviewAudience(ctx context.Context, filter AudienceFilter) (PreviewAudienceResp, error) {
	count, ids, err := s.repo.PreviewAudience(ctx, filter, 30)
	if err != nil {
		return PreviewAudienceResp{}, err
	}
	resp := PreviewAudienceResp{Count: count}
	resp.SampleUsers = make([]types.Int64Str, 0, len(ids))
	for _, id := range ids {
		resp.SampleUsers = append(resp.SampleUsers, types.Int64Str(id))
	}
	return resp, nil
}

// ListAudience 召回执行用：按 cursor 拉人群。
func (s *Service) ListAudience(ctx context.Context, filter AudienceFilter, lastID int64, size int) ([]int64, error) {
	return s.repo.ListAudience(ctx, filter, lastID, size)
}

// ===== 自动重算 =====

// Recompute 单用户增量重算（OrderPaid 等事件触发）。
//
// 计算 RFM + 生命周期。以全量 delete + insert 形式覆盖该用户的 auto/RFM/lifecycle 标签。
func (s *Service) Recompute(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errs.ErrParam
	}
	cnt, lastPaid, gmv, err := s.repo.AggregateUserOrder(ctx, userID)
	if err != nil {
		return err
	}
	codes := computeRFMCodes(cnt, lastPaid, gmv, time.Now())

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除该用户的 auto RFM/lifecycle 标签
		if err := tx.Where(
			"user_id = ? AND source = ? AND (tag_code LIKE ? OR tag_code LIKE ? OR tag_code LIKE ?)",
			userID, SourceAuto, "rfm_%", "lifecycle_%", "value_%",
		).Delete(&UserTagRelation{}).Error; err != nil {
			return err
		}
		for _, code := range codes {
			rel := &UserTagRelation{
				ID:      snowflake.NextID(),
				UserID:  userID,
				TagCode: code,
				Score:   1,
				Source:  SourceAuto,
			}
			if err := s.repo.UpsertRelation(ctx, tx, rel); err != nil {
				return err
			}
		}
		return nil
	})
}

// RecomputeAll 全量重算所有 active 用户的 RFM + 生命周期标签。
//
// 简化实现：分批拉用户 ID，逐个调 Recompute。生产环境可优化为聚合 SQL。
func (s *Service) RecomputeAll(ctx context.Context) error {
	var lastID int64
	const batch = 1000
	for {
		ids, err := s.repo.ListAllUserIDs(ctx, lastID, batch)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, id := range ids {
			if err := s.Recompute(ctx, id); err != nil {
				// 单用户失败不阻断全量
				continue
			}
		}
		lastID = ids[len(ids)-1]
		if len(ids) < batch {
			return nil
		}
	}
}

// MonthlySnapshot 月度快照：把 user_tag_relation 按 user 聚合写入快照表。
func (s *Service) MonthlySnapshot(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	var lastID int64
	const batch = 500
	for {
		ids, err := s.repo.ListAllUserIDs(ctx, lastID, batch)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, uid := range ids {
			rels, err := s.repo.ListUserRelations(ctx, uid)
			if err != nil {
				continue
			}
			tags := make(JSONStrArray, 0, len(rels))
			for i := range rels {
				tags = append(tags, rels[i].TagCode)
			}
			snap := &UserTagSnapshot{
				ID:           snowflake.NextID(),
				SnapshotDate: today,
				UserID:       uid,
				Tags:         tags,
			}
			_ = s.repo.WriteSnapshot(ctx, snap)
		}
		lastID = ids[len(ids)-1]
		if len(ids) < batch {
			return nil
		}
	}
}

// ===== RFM 分桶逻辑 =====

// computeRFMCodes 计算单用户 RFM + 生命周期 + 价值带（简化版，无 P50/P75/P90 动态分位数）。
func computeRFMCodes(orderCount int64, lastPaidAt *time.Time, gmvCents int64, now time.Time) []string {
	codes := make([]string, 0, 4)

	// R 值：最近一次下单距今天数
	rDays := -1
	if lastPaidAt != nil {
		rDays = int(now.Sub(*lastPaidAt).Hours() / 24)
	}
	switch {
	case rDays < 0:
		codes = append(codes, "rfm_r_never")
	case rDays <= 30:
		codes = append(codes, "rfm_r_0_30")
	case rDays <= 90:
		codes = append(codes, "rfm_r_31_90")
	case rDays <= 180:
		codes = append(codes, "rfm_r_91_180")
	default:
		codes = append(codes, "rfm_r_181_plus")
	}

	// F 值：累计成单数
	switch {
	case orderCount == 0:
		codes = append(codes, "rfm_f_0")
	case orderCount == 1:
		codes = append(codes, "rfm_f_1")
	case orderCount <= 5:
		codes = append(codes, "rfm_f_2_5")
	case orderCount <= 10:
		codes = append(codes, "rfm_f_6_10")
	default:
		codes = append(codes, "rfm_f_11_plus")
	}

	// M 值：累计 GMV（简化档位：< 100 / 100-500 / 500-2000 / 2000+）
	switch {
	case gmvCents < 10000:
		codes = append(codes, "rfm_m_low")
	case gmvCents < 50000:
		codes = append(codes, "rfm_m_mid")
	case gmvCents < 200000:
		codes = append(codes, "rfm_m_high")
	default:
		codes = append(codes, "rfm_m_top")
	}

	// 生命周期
	switch {
	case orderCount == 0:
		codes = append(codes, "lifecycle_new_user")
	case rDays >= 0 && rDays <= 30:
		codes = append(codes, "lifecycle_active")
	case rDays > 30 && rDays <= 90:
		codes = append(codes, "lifecycle_dormant")
	default:
		codes = append(codes, "lifecycle_churned")
	}

	return codes
}


