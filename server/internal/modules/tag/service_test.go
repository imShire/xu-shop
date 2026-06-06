package tag

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// =============================================================================
// 测试基建
// =============================================================================

// newTestSQLiteDB 创建内存 SQLite，仅用于需要 db.Transaction 的方法。
func newTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&UserTagRelation{}))
	t.Cleanup(func() {
		s, _ := db.DB()
		if s != nil {
			_ = s.Close()
		}
	})
	return db
}

// mockTagRepo 仅用接口 mock。
type mockTagRepo struct {
	tags       map[string]*UserTag
	relations  map[int64][]UserTagRelation // by userID
	snapshots  []*UserTagSnapshot
	allUserIDs []int64
	// 控制错误注入
	createTagErr   error
	getTagErr      error
	upsertRelErr   error
	aggErr         error
	aggResult      struct {
		count    int64
		lastPaid *time.Time
		gmv      int64
	}
	countRelations int64
}

func newMockTagRepo() *mockTagRepo {
	return &mockTagRepo{
		tags:      make(map[string]*UserTag),
		relations: make(map[int64][]UserTagRelation),
	}
}

func (m *mockTagRepo) DB() *gorm.DB { return nil }

func (m *mockTagRepo) CreateTag(_ context.Context, t *UserTag) error {
	if m.createTagErr != nil {
		return m.createTagErr
	}
	cp := *t
	m.tags[t.Code] = &cp
	return nil
}
func (m *mockTagRepo) UpdateTag(_ context.Context, code string, fields map[string]any) error {
	if t, ok := m.tags[code]; ok {
		if v, ok := fields["enabled"].(bool); ok {
			t.Enabled = v
		}
	}
	return nil
}
func (m *mockTagRepo) DeleteTag(_ context.Context, code string) error {
	delete(m.tags, code)
	return nil
}
func (m *mockTagRepo) GetTag(_ context.Context, code string) (*UserTag, error) {
	if m.getTagErr != nil {
		return nil, m.getTagErr
	}
	if t, ok := m.tags[code]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockTagRepo) ListTags(_ context.Context, _, _ string) ([]UserTag, error) {
	out := make([]UserTag, 0, len(m.tags))
	for _, t := range m.tags {
		out = append(out, *t)
	}
	return out, nil
}
func (m *mockTagRepo) CountRelationsByTag(_ context.Context, _ string) (int64, error) {
	return m.countRelations, nil
}
func (m *mockTagRepo) UpsertRelation(_ context.Context, _ *gorm.DB, rel *UserTagRelation) error {
	if m.upsertRelErr != nil {
		return m.upsertRelErr
	}
	m.relations[rel.UserID] = append(m.relations[rel.UserID], *rel)
	return nil
}
func (m *mockTagRepo) DeleteRelation(_ context.Context, userID int64, tagCode string) error {
	rels := m.relations[userID]
	out := rels[:0]
	for _, r := range rels {
		if r.TagCode != tagCode {
			out = append(out, r)
		}
	}
	m.relations[userID] = out
	return nil
}
func (m *mockTagRepo) DeleteAutoByPrefix(_ context.Context, _ *gorm.DB, _ string) error {
	return nil
}
func (m *mockTagRepo) DeleteAllUserAuto(_ context.Context, _ *gorm.DB, _ int64) error { return nil }
func (m *mockTagRepo) ListUserRelations(_ context.Context, userID int64) ([]UserTagRelation, error) {
	return m.relations[userID], nil
}
func (m *mockTagRepo) ListAllUserIDs(_ context.Context, lastID int64, batchSize int) ([]int64, error) {
	out := make([]int64, 0)
	for _, id := range m.allUserIDs {
		if id > lastID {
			out = append(out, id)
		}
		if len(out) >= batchSize {
			break
		}
	}
	return out, nil
}
func (m *mockTagRepo) WriteSnapshot(_ context.Context, snap *UserTagSnapshot) error {
	m.snapshots = append(m.snapshots, snap)
	return nil
}
func (m *mockTagRepo) AggregateUserOrder(_ context.Context, _ int64) (int64, *time.Time, int64, error) {
	if m.aggErr != nil {
		return 0, nil, 0, m.aggErr
	}
	return m.aggResult.count, m.aggResult.lastPaid, m.aggResult.gmv, nil
}
func (m *mockTagRepo) PreviewAudience(_ context.Context, _ AudienceFilter, _ int) (int64, []int64, error) {
	return 0, nil, nil
}
func (m *mockTagRepo) ListAudience(_ context.Context, _ AudienceFilter, _ int64, _ int) ([]int64, error) {
	return nil, nil
}

// newTestTagService 构造测试用 Service（使用 mock repo，DB 由调用方传入）。
func newTestTagService(repo *mockTagRepo, db *gorm.DB) *Service {
	return NewService(repo, db)
}

// =============================================================================
// computeRFMCodes — 纯函数测试
// =============================================================================

// Given 从未下单，Then rfm_r_never + rfm_f_0 + rfm_m_low + lifecycle_new_user
func Test_computeRFMCodes_NeverOrdered_ReturnsNeverCode(t *testing.T) {
	codes := computeRFMCodes(0, nil, 0, time.Now())

	assert.Contains(t, codes, "rfm_r_never")
	assert.Contains(t, codes, "rfm_f_0")
	assert.Contains(t, codes, "rfm_m_low")
	assert.Contains(t, codes, "lifecycle_new_user")
}

// Given 最近 7 天下单 3 次，GMV 100 元，Then active lifecycle + mid F + r_0_30
func Test_computeRFMCodes_RecentActiveMidValue(t *testing.T) {
	now := time.Now()
	recent := now.Add(-7 * 24 * time.Hour)
	codes := computeRFMCodes(3, &recent, 10000 /* 100 元 */, now)

	assert.Contains(t, codes, "rfm_r_0_30")
	assert.Contains(t, codes, "rfm_f_2_5")
	assert.Contains(t, codes, "rfm_m_mid")
	assert.Contains(t, codes, "lifecycle_active")
}

// Given 200 天前下单 1 次，GMV 50000 元（高价值），Then lifecycle_churned + rfm_m_top
func Test_computeRFMCodes_ChurnedHighValue(t *testing.T) {
	now := time.Now()
	old := now.Add(-200 * 24 * time.Hour)
	codes := computeRFMCodes(1, &old, 200000 /* 2000 元 */, now)

	assert.Contains(t, codes, "rfm_r_181_plus")
	assert.Contains(t, codes, "rfm_f_1")
	assert.Contains(t, codes, "rfm_m_top")
	assert.Contains(t, codes, "lifecycle_churned")
}

// Given 45 天前下单，Then lifecycle_dormant
func Test_computeRFMCodes_Dormant45Days(t *testing.T) {
	now := time.Now()
	paid := now.Add(-45 * 24 * time.Hour)
	codes := computeRFMCodes(2, &paid, 5000, now)

	assert.Contains(t, codes, "lifecycle_dormant")
	assert.Contains(t, codes, "rfm_r_31_90")
}

// Given 大单用户 12 单，Then rfm_f_11_plus
func Test_computeRFMCodes_PowerUser_FrequencyBucket(t *testing.T) {
	now := time.Now()
	paid := now.Add(-5 * 24 * time.Hour)
	codes := computeRFMCodes(12, &paid, 50000, now)

	assert.Contains(t, codes, "rfm_f_11_plus")
}

// =============================================================================
// CreateTag
// =============================================================================

// Given 合法请求，When CreateTag，Then 成功写入 repo
func Test_CreateTag_ValidRequest_OK(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	tag, err := svc.CreateTag(context.Background(), CreateTagReq{
		Code:     "test_code",
		Name:     "测试标签",
		Category: CategoryBusiness,
	})
	require.NoError(t, err)
	require.NotNil(t, tag)
	assert.Equal(t, "test_code", tag.Code)
	assert.Equal(t, SourceManual, tag.Source) // service 强制 SourceManual
	_, exists := repo.tags["test_code"]
	assert.True(t, exists)
}

// Given Code 为空，When CreateTag，Then 返回 ErrParam
func Test_CreateTag_EmptyCode_ErrParam(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	_, err := svc.CreateTag(context.Background(), CreateTagReq{
		Name:     "无 Code 标签",
		Category: CategoryBusiness,
	})
	assert.ErrorIs(t, err, errs.ErrParam)
}

// =============================================================================
// DeleteTag
// =============================================================================

// Given auto 来源标签，When DeleteTag，Then 返回 ErrForbidden
func Test_DeleteTag_AutoSource_ErrForbidden(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	repo.tags["rfm_r_0_30"] = &UserTag{Code: "rfm_r_0_30", Source: SourceAuto, Enabled: true}

	err := svc.DeleteTag(context.Background(), "rfm_r_0_30")
	assert.ErrorIs(t, err, errs.ErrForbidden)
}

// Given manual 标签但存在用户关系，When DeleteTag，Then 返回 ErrConflict
func Test_DeleteTag_HasRelations_ErrConflict(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	repo.tags["manual_vip"] = &UserTag{Code: "manual_vip", Source: SourceManual, Enabled: true}
	repo.countRelations = 3 // 有关系

	err := svc.DeleteTag(context.Background(), "manual_vip")
	assert.ErrorIs(t, err, errs.ErrConflict)
}

// Given manual 标签且无关系，When DeleteTag，Then 成功删除
func Test_DeleteTag_ManualNoRelations_OK(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	repo.tags["manual_tag"] = &UserTag{Code: "manual_tag", Source: SourceManual, Enabled: true}
	repo.countRelations = 0

	err := svc.DeleteTag(context.Background(), "manual_tag")
	require.NoError(t, err)
	_, exists := repo.tags["manual_tag"]
	assert.False(t, exists)
}

// Given 不存在的标签，When DeleteTag，Then 返回 ErrNotFound
func Test_DeleteTag_NotFound_ErrNotFound(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	err := svc.DeleteTag(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// =============================================================================
// Recompute — 用户增量重算
// =============================================================================

// Given userID <= 0，When Recompute，Then 立即返回 ErrParam（无需 DB）
func Test_Recompute_InvalidUserID_ErrParam(t *testing.T) {
	repo := newMockTagRepo()
	svc := newTestTagService(repo, nil)

	err := svc.Recompute(context.Background(), 0)
	assert.ErrorIs(t, err, errs.ErrParam)

	err = svc.Recompute(context.Background(), -1)
	assert.ErrorIs(t, err, errs.ErrParam)
}

// Given repo.AggregateUserOrder 返回错误，When Recompute，Then 透传错误
func Test_Recompute_AggregateError_PropagatesErr(t *testing.T) {
	repo := newMockTagRepo()
	repo.aggErr = errors.New("db timeout")
	svc := newTestTagService(repo, nil)

	err := svc.Recompute(context.Background(), 42)
	assert.Error(t, err)
	assert.EqualError(t, err, "db timeout")
}

// Given 正常用户订单数据（sqlite 内存 DB），When Recompute，Then 写入 RFM 标签
func Test_Recompute_NormalUser_WritesRFMTags(t *testing.T) {
	db := newTestSQLiteDB(t)
	repo := newMockTagRepo()

	paid := time.Now().Add(-10 * 24 * time.Hour)
	repo.aggResult.count = 2
	repo.aggResult.lastPaid = &paid
	repo.aggResult.gmv = 5000

	svc := newTestTagService(repo, db)
	err := svc.Recompute(context.Background(), 100)
	require.NoError(t, err)

	// UpsertRelation 被 mock repo 捕获
	rels := repo.relations[100]
	assert.NotEmpty(t, rels, "应写入 RFM 标签关系")
	codes := make([]string, 0, len(rels))
	for _, r := range rels {
		codes = append(codes, r.TagCode)
	}
	assert.Contains(t, codes, "rfm_r_0_30")
	assert.Contains(t, codes, "lifecycle_active")
}

// =============================================================================
// RecomputeAll — 全量重算
// =============================================================================

// Given 无用户，When RecomputeAll，Then 返回 nil
func Test_RecomputeAll_NoUsers_ReturnsNil(t *testing.T) {
	repo := newMockTagRepo()
	repo.allUserIDs = []int64{}
	svc := newTestTagService(repo, nil)

	err := svc.RecomputeAll(context.Background())
	assert.NoError(t, err)
}
