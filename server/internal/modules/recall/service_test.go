package recall

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// =============================================================================
// mock Repo
// =============================================================================

type mockRecallRepo struct {
	campaigns    map[int64]*RecallCampaign
	logs         []*RecallLog
	onlineCron   []RecallCampaign
	onlineEvent  []RecallCampaign
	listErr      error
	updateFields map[int64]map[string]any
	idSeq        int64
	hasLogToday  bool
	countTotal   int64
}

func newMockRecallRepo() *mockRecallRepo {
	return &mockRecallRepo{
		campaigns:    make(map[int64]*RecallCampaign),
		updateFields: make(map[int64]map[string]any),
	}
}

func (m *mockRecallRepo) nextID() int64 { m.idSeq++; return m.idSeq }
func (m *mockRecallRepo) DB() *gorm.DB  { return nil }

func (m *mockRecallRepo) CreateCampaign(_ context.Context, c *RecallCampaign) error {
	cp := *c
	m.campaigns[c.ID] = &cp
	return nil
}

func (m *mockRecallRepo) UpdateCampaign(_ context.Context, id int64, fields map[string]any) error {
	m.updateFields[id] = fields
	if c, ok := m.campaigns[id]; ok {
		if s, ok := fields["status"].(string); ok {
			c.Status = s
		}
	}
	return nil
}

func (m *mockRecallRepo) GetCampaign(_ context.Context, id int64) (*RecallCampaign, error) {
	if c, ok := m.campaigns[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRecallRepo) ListCampaigns(_ context.Context, _ string, _, _ int) ([]RecallCampaign, int64, error) {
	out := make([]RecallCampaign, 0, len(m.campaigns))
	for _, c := range m.campaigns {
		out = append(out, *c)
	}
	return out, int64(len(out)), nil
}

func (m *mockRecallRepo) ListOnlineByTrigger(_ context.Context, triggerType string) ([]RecallCampaign, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	switch triggerType {
	case TriggerCron:
		return m.onlineCron, nil
	case TriggerEvent:
		return m.onlineEvent, nil
	}
	return nil, nil
}

func (m *mockRecallRepo) InsertLog(_ context.Context, l *RecallLog) error {
	cp := *l
	m.logs = append(m.logs, &cp)
	return nil
}

func (m *mockRecallRepo) HasLogToday(_ context.Context, _, _ int64) (bool, error) {
	return m.hasLogToday, nil
}

func (m *mockRecallRepo) CountLogToday(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (m *mockRecallRepo) CountLogTotal(_ context.Context, _ int64) (int64, error) {
	return m.countTotal, nil
}

func (m *mockRecallRepo) FunnelStats(_ context.Context, _ int64) (int64, int64, int64, int64, error) {
	return 0, 0, 0, 0, nil
}

func (m *mockRecallRepo) ListLogs(_ context.Context, _ int64, _, _ int) ([]RecallLog, int64, error) {
	return nil, 0, nil
}

func (m *mockRecallRepo) AttributeOrder(_ context.Context, _, _ int64, _ time.Time, _ int64, _ int) (int64, error) {
	return 0, nil
}

// =============================================================================
// helpers
// =============================================================================

// newTestRecallService 构造无 Redis、无 coupon/notif 的测试服务。tagSvc 可为 nil。
func newTestRecallService(repo *mockRecallRepo) *Service {
	return NewService(repo, nil, nil, nil, nil, nil)
}

func makeCampaign(id int64, status, triggerType string) *RecallCampaign {
	return &RecallCampaign{
		ID:              id,
		Name:            "测试活动",
		Status:          status,
		TriggerType:     triggerType,
		TriggerConfig:   JSONMap{},
		AudienceFilter:  JSONMap{},
		Actions:         JSONArray{},
		CreatedBy:       1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// =============================================================================
// Transition — 召回活动状态机
// =============================================================================

// Given draft 活动，When Transition → online，Then 状态变为 online
func Test_Transition_DraftToOnline_OK(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(1, StatusDraft, TriggerCron)
	repo.campaigns[1] = c

	err := svc.Transition(context.Background(), 1, StatusOnline, 999)
	require.NoError(t, err)
	assert.Equal(t, StatusOnline, repo.campaigns[1].Status)
}

// Given online 活动，When Transition → paused，Then 状态变为 paused
func Test_Transition_OnlineToPaused_OK(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(2, StatusOnline, TriggerCron)
	repo.campaigns[2] = c

	err := svc.Transition(context.Background(), 2, StatusPaused, 999)
	require.NoError(t, err)
	assert.Equal(t, StatusPaused, repo.campaigns[2].Status)
}

// Given draft 活动，When Transition → paused（非法），Then 返回 ErrConflict
func Test_Transition_DraftToPaused_ErrConflict(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(3, StatusDraft, TriggerCron)
	repo.campaigns[3] = c

	err := svc.Transition(context.Background(), 3, StatusPaused, 999)
	assert.ErrorIs(t, err, errs.ErrConflict)
	assert.Equal(t, StatusDraft, repo.campaigns[3].Status, "状态不应被修改")
}

// Given closed 活动，When 任意 Transition，Then 返回 ErrConflict（closed 无出边）
func Test_Transition_FromClosed_ErrConflict(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(4, StatusClosed, TriggerCron)
	repo.campaigns[4] = c

	err := svc.Transition(context.Background(), 4, StatusOnline, 999)
	assert.ErrorIs(t, err, errs.ErrConflict)
}

// Given 不存在活动，When Transition，Then 返回 ErrNotFound
func Test_Transition_NotFound_ErrNotFound(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	err := svc.Transition(context.Background(), 999, StatusOnline, 1)
	assert.ErrorIs(t, err, errs.ErrNotFound)
}

// =============================================================================
// CreateCampaign
// =============================================================================

// Given 合法表单，When CreateCampaign，Then 以草稿状态落库
func Test_CreateCampaign_ValidForm_DraftStatus(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	resp, err := svc.CreateCampaign(context.Background(), CampaignForm{
		Name:        "夏日唤回",
		Goal:        "复购",
		TriggerType: TriggerCron,
	}, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Len(t, repo.campaigns, 1)
	for _, c := range repo.campaigns {
		assert.Equal(t, StatusDraft, c.Status)
		assert.Equal(t, int64(10), c.CreatedBy)
	}
}

// =============================================================================
// UpdateCampaign
// =============================================================================

// Given 非 draft 活动，When UpdateCampaign，Then 返回 ErrConflict
func Test_UpdateCampaign_NonDraft_ErrConflict(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(5, StatusOnline, TriggerCron)
	repo.campaigns[5] = c

	err := svc.UpdateCampaign(context.Background(), 5, CampaignForm{Name: "新名字"})
	assert.ErrorIs(t, err, errs.ErrConflict)
}

// Given draft 活动，When UpdateCampaign，Then 更新成功
func Test_UpdateCampaign_Draft_OK(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(6, StatusDraft, TriggerCron)
	repo.campaigns[6] = c

	err := svc.UpdateCampaign(context.Background(), 6, CampaignForm{Name: "修改后"})
	require.NoError(t, err)
	fields, ok := repo.updateFields[6]
	assert.True(t, ok)
	assert.Equal(t, "修改后", fields["name"])
}

// =============================================================================
// ScheduleScan
// =============================================================================

// Given 无在线 cron 活动，When ScheduleScan，Then 返回 nil
func Test_ScheduleScan_NoCampaigns_ReturnsNil(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	err := svc.ScheduleScan(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, repo.logs)
}

// Given repo.ListOnlineByTrigger 返回错误，When ScheduleScan，Then 透传错误
func Test_ScheduleScan_RepoError_PropagatesError(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)
	repo.listErr = errors.New("db timeout")

	err := svc.ScheduleScan(context.Background())
	assert.Error(t, err)
	assert.EqualError(t, err, "db timeout")
}

// Given 有在线 cron 活动但 tagSvc 为 nil，When ScheduleScan，Then 返回 nil（单活动失败不阻断）
func Test_ScheduleScan_WithCampaignButNoTagSvc_NoGlobalError(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo) // tagSvc = nil

	repo.onlineCron = []RecallCampaign{
		*makeCampaign(10, StatusOnline, TriggerCron),
	}

	err := svc.ScheduleScan(context.Background())
	assert.NoError(t, err) // 单活动失败被吞，不阻断整体扫描
}

// =============================================================================
// OnEvent
// =============================================================================

// Given 无在线 event 活动，When OnEvent，Then 返回 nil
func Test_OnEvent_NoCampaigns_ReturnsNil(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	err := svc.OnEvent(context.Background(), EventCartAbandoned2H, 42)
	assert.NoError(t, err)
}

// Given repo 返回错误，When OnEvent，Then 透传错误
func Test_OnEvent_RepoError_PropagatesError(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)
	repo.listErr = errors.New("network error")

	err := svc.OnEvent(context.Background(), EventOrderCompleted, 0)
	assert.Error(t, err)
}

// Given 活动 trigger_config.event 不匹配，When OnEvent，Then 不执行活动
func Test_OnEvent_EventMismatch_NoCampaignExecuted(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	c := makeCampaign(20, StatusOnline, TriggerEvent)
	c.TriggerConfig = JSONMap{"event": EventBirthdayToday}
	repo.onlineEvent = []RecallCampaign{*c}

	err := svc.OnEvent(context.Background(), EventCartAbandoned2H, 42)
	require.NoError(t, err)
	assert.Empty(t, repo.logs, "事件不匹配时不应写日志")
}

// =============================================================================
// OnOrderPaid — 归因
// =============================================================================

// Given userID 或 orderID 为 0，When OnOrderPaid，Then 返回 nil（防卫性提前返回）
func Test_OnOrderPaid_InvalidIDs_ReturnsNil(t *testing.T) {
	repo := newMockRecallRepo()
	svc := newTestRecallService(repo)

	assert.NoError(t, svc.OnOrderPaid(context.Background(), 0, 1001, time.Now(), 5000))
	assert.NoError(t, svc.OnOrderPaid(context.Background(), 42, 0, time.Now(), 5000))
}
