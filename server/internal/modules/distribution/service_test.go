package distribution

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// mock Repo
// =============================================================================

type mockDistRepo struct {
	distributors  map[int64]*Distributor      // by distributor.ID
	distByUser    map[int64]*Distributor      // by user_id
	commissions   map[int64]*CommissionRecord // by id
	withdraws     map[string]*WithdrawOrder   // by idem_key
	createDistErr error
	updateDistErr error
	createCommErr error
	sumStatus     struct{ pending, locked, settled int64 }
	lockedList    []CommissionRecord
}

func newMockDistRepo() *mockDistRepo {
	return &mockDistRepo{
		distributors: make(map[int64]*Distributor),
		distByUser:   make(map[int64]*Distributor),
		commissions:  make(map[int64]*CommissionRecord),
		withdraws:    make(map[string]*WithdrawOrder),
	}
}

func (m *mockDistRepo) DB() *gorm.DB { return nil }

// share_link
func (m *mockDistRepo) CreateShareLink(_ context.Context, _ *ShareLink) error { return nil }
func (m *mockDistRepo) GetShareLinkByToken(_ context.Context, _ string) (*ShareLink, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) GetShareLinkByID(_ context.Context, _ int64) (*ShareLink, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) IncShareLinkCounter(_ context.Context, _ int64, _ string, _ int64) error {
	return nil
}
func (m *mockDistRepo) ListShareLinks(_ context.Context, _ int64, _ string, _ int) ([]ShareLink, error) {
	return nil, nil
}

// share_click
func (m *mockDistRepo) CreateShareClick(_ context.Context, _ *ShareClick) error { return nil }

// share_attribution
func (m *mockDistRepo) GetAttributionByTrace(_ context.Context, _ string) (*ShareAttribution, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) UpsertAttribution(_ context.Context, _ *ShareAttribution) error { return nil }
func (m *mockDistRepo) BindAttributionUser(_ context.Context, _ string, _ int64) error { return nil }

// distributor
func (m *mockDistRepo) CreateDistributor(_ context.Context, d *Distributor) error {
	if m.createDistErr != nil {
		return m.createDistErr
	}
	cp := *d
	m.distributors[d.ID] = &cp
	m.distByUser[d.UserID] = &cp
	return nil
}
func (m *mockDistRepo) GetDistributorByUserID(_ context.Context, userID int64) (*Distributor, error) {
	if d, ok := m.distByUser[userID]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) GetDistributorByID(_ context.Context, id int64) (*Distributor, error) {
	if d, ok := m.distributors[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) UpdateDistributor(_ context.Context, id int64, fields map[string]any) error {
	if m.updateDistErr != nil {
		return m.updateDistErr
	}
	if d, ok := m.distributors[id]; ok {
		if s, ok := fields["status"].(string); ok {
			d.Status = s
		}
	}
	return nil
}
func (m *mockDistRepo) ListDistributors(_ context.Context, _, _ string, _, _ int) ([]Distributor, int64, error) {
	return nil, 0, nil
}

// distributor_relation
func (m *mockDistRepo) CreateRelation(_ context.Context, _ *gorm.DB, _ *DistributorRelation) error {
	return nil
}
func (m *mockDistRepo) GetRelationByInvitee(_ context.Context, _ int64) (*DistributorRelation, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) RenewRelation(_ context.Context, _ int64, _ time.Time) error { return nil }

// commission
func (m *mockDistRepo) CreateCommission(_ context.Context, c *CommissionRecord) error {
	if m.createCommErr != nil {
		return m.createCommErr
	}
	cp := *c
	m.commissions[c.ID] = &cp
	return nil
}
func (m *mockDistRepo) GetCommission(_ context.Context, id int64) (*CommissionRecord, error) {
	if c, ok := m.commissions[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) UpdateCommission(_ context.Context, id int64, fields map[string]any) error {
	if c, ok := m.commissions[id]; ok {
		if s, ok := fields["status"].(string); ok {
			c.Status = s
		}
	}
	return nil
}
func (m *mockDistRepo) ListCommissionsByUser(_ context.Context, _ int64, _ string, _, _ int) ([]CommissionRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockDistRepo) ListCommissionsAdmin(_ context.Context, _ string, _ int64, _, _ int) ([]CommissionRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockDistRepo) ListPendingExpired(_ context.Context, _ time.Time, _ int) ([]CommissionRecord, error) {
	return nil, nil
}
func (m *mockDistRepo) SumByUserStatus(_ context.Context, _ int64) (pending, locked, settled int64, err error) {
	return m.sumStatus.pending, m.sumStatus.locked, m.sumStatus.settled, nil
}
func (m *mockDistRepo) ListLockedForUser(_ context.Context, _ int64) ([]CommissionRecord, error) {
	return m.lockedList, nil
}

// withdraw
func (m *mockDistRepo) CreateWithdraw(_ context.Context, w *WithdrawOrder) error {
	cp := *w
	m.withdraws[w.IdemKey] = &cp
	return nil
}
func (m *mockDistRepo) GetWithdraw(_ context.Context, _ int64) (*WithdrawOrder, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) GetWithdrawByIdem(_ context.Context, key string) (*WithdrawOrder, error) {
	if w, ok := m.withdraws[key]; ok {
		cp := *w
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) GetWithdrawByOutBillNo(_ context.Context, _ string) (*WithdrawOrder, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDistRepo) UpdateWithdraw(_ context.Context, _ int64, _ map[string]any) error {
	return nil
}
func (m *mockDistRepo) ListWithdrawByUser(_ context.Context, _ int64, _, _ int) ([]WithdrawOrder, int64, error) {
	return nil, 0, nil
}
func (m *mockDistRepo) ListWithdrawAdmin(_ context.Context, _ string, _, _ int) ([]WithdrawOrder, int64, error) {
	return nil, 0, nil
}
func (m *mockDistRepo) SumWithdrawnByUser(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

// settlement
func (m *mockDistRepo) CreateSettlement(_ context.Context, _ *CommissionSettlement) error {
	return nil
}
func (m *mockDistRepo) UpdateSettlement(_ context.Context, _ int64, _ map[string]any) error {
	return nil
}
func (m *mockDistRepo) ListSettlements(_ context.Context, _ string, _, _ int) ([]CommissionSettlement, int64, error) {
	return nil, 0, nil
}
func (m *mockDistRepo) BindCommissionsToSettlement(_ context.Context, _ []int64, _ int64) error {
	return nil
}

// 漏斗统计
func (m *mockDistRepo) CountShareLinks(_ context.Context, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockDistRepo) CountShareClicks(_ context.Context, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockDistRepo) CountAttributionRegisters(_ context.Context, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockDistRepo) SumDistributionGMV(_ context.Context, _, _ time.Time) (int64, int64, error) {
	return 0, 0, nil
}
func (m *mockDistRepo) SumCommissionsByStatus(_ context.Context, _, _ time.Time) (int64, int64, int64, error) {
	return 0, 0, 0, nil
}

// =============================================================================
// helpers
// =============================================================================

func newTestDistService(repo *mockDistRepo) *Service {
	return NewService(repo, nil, nil, nil, nil, nil, Config{})
}

func makeCommission(id int64, status string) *CommissionRecord {
	return &CommissionRecord{
		ID:              id,
		OrderID:         1001,
		DistributorUserID: 42,
		Status:          status,
		AmountCents:     500,
		BaseAmountCents: 10000,
		Rate:            0.05,
		FreezeUntil:     time.Now().Add(-24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// =============================================================================
// Transition — 佣金状态机
// =============================================================================

// Given 一条 pending 佣金，When mark_suspect，Then 状态变为 suspect
func Test_Transition_PendingToSuspect_OK(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(1, CommissionStatusPending)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "mark_suspect", "风控命中")
	require.NoError(t, err)
	assert.Equal(t, CommissionStatusSuspect, c.Status)
}

// Given 一条 pending 佣金，When pass_freeze，Then 状态变为 locked
func Test_Transition_PassFreeze_OK(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(2, CommissionStatusPending)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "pass_freeze", "")
	require.NoError(t, err)
	assert.Equal(t, CommissionStatusLocked, c.Status)
}

// Given 一条 pending 佣金，When pass_freeze from settled，Then 返回 ErrInvalidTransition
func Test_Transition_PassFreezeFromSettled_ErrInvalidTransition(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(3, CommissionStatusSettled)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "pass_freeze", "")
	assert.ErrorIs(t, err, ErrInvalidTransition)
	assert.Equal(t, CommissionStatusSettled, c.Status, "状态不应被修改")
}

// Given 未知事件，Then 返回 ErrInvalidTransition
func Test_Transition_UnknownEvent_ErrInvalidTransition(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(4, CommissionStatusPending)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "nonexistent_event", "")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

// Given suspect 佣金，When unsuspect，Then 回到 pending
func Test_Transition_UnsuspectFromSuspect_OK(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(5, CommissionStatusSuspect)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "unsuspect", "误判")
	require.NoError(t, err)
	assert.Equal(t, CommissionStatusPending, c.Status)
}

// Given 已 settled 佣金，When order_full_refund，Then 返回 ErrInvalidTransition
func Test_Transition_FullRefundFromSettled_ErrInvalidTransition(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(6, CommissionStatusSettled)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "order_full_refund", "全退")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

// Given pending 佣金，When order_full_refund，Then 变为 canceled
func Test_Transition_FullRefundFromPending_Canceled(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	c := makeCommission(7, CommissionStatusPending)
	repo.commissions[c.ID] = c

	err := svc.Transition(context.Background(), c, "order_full_refund", "全退")
	require.NoError(t, err)
	assert.Equal(t, CommissionStatusCanceled, c.Status)
}

// =============================================================================
// Apply — 申请分销员
// =============================================================================

// Given 未注册用户，When Apply，Then 创建 pending 分销员
func Test_Apply_NewDistributor_OK(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d, err := svc.Apply(context.Background(), 100, ApplyDistributorReq{})
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, DistStatusPending, d.Status)
	assert.Equal(t, int64(100), d.UserID)
}

// Given 已存在分销员，When Apply，Then 返回 ErrDistributorExists
func Test_Apply_AlreadyExists_ErrDistributorExists(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	existing := &Distributor{ID: 1, UserID: 200, Status: DistStatusPending}
	repo.distByUser[200] = existing

	_, err := svc.Apply(context.Background(), 200, ApplyDistributorReq{})
	assert.ErrorIs(t, err, ErrDistributorExists)
}

// =============================================================================
// Approve / Reject — 审核
// =============================================================================

// Given pending 分销员，When Approve，Then 状态变为 active
func Test_Approve_PendingDistributor_OK(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 10, UserID: 300, Status: DistStatusPending}
	repo.distributors[10] = d

	err := svc.Approve(context.Background(), 10, 999)
	require.NoError(t, err)
	assert.Equal(t, DistStatusActive, repo.distributors[10].Status)
}

// Given active 分销员，When Approve，Then 返回 ErrInvalidTransition
func Test_Approve_AlreadyActive_ErrInvalidTransition(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 11, UserID: 301, Status: DistStatusActive}
	repo.distributors[11] = d

	err := svc.Approve(context.Background(), 11, 999)
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

// Given pending 分销员，When Reject，Then 状态变为 disabled
func Test_Reject_PendingDistributor_Disabled(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 12, UserID: 302, Status: DistStatusPending}
	repo.distributors[12] = d

	err := svc.Reject(context.Background(), 12, "不符合条件")
	require.NoError(t, err)
	assert.Equal(t, DistStatusDisabled, repo.distributors[12].Status)
}

// Given active 分销员，When Reject，Then 返回 ErrInvalidTransition（只允许 pending 拒绝）
func Test_Reject_ActiveDistributor_ErrInvalidTransition(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 13, UserID: 303, Status: DistStatusActive}
	repo.distributors[13] = d

	err := svc.Reject(context.Background(), 13, "")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

// =============================================================================
// OnOrderPaid — 佣金入账
// =============================================================================

// Given 自购场景，When OnOrderPaid，Then 返回 ErrSelfPurchase
func Test_OnOrderPaid_SelfPurchase_ErrSelfPurchase(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	err := svc.OnOrderPaid(context.Background(), OrderInfo{
		OrderID:         5001,
		UserID:          42,
		PayCents:        10000,
		DistributorUser: 42, // 和 UserID 相同 → 自购
	})
	assert.ErrorIs(t, err, ErrSelfPurchase)
}

// Given 分销员已激活，When OnOrderPaid，Then 创建 pending 佣金
func Test_OnOrderPaid_ActiveDistributor_CreatesCommission(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 20, UserID: 50, Status: DistStatusActive, Rate: 0.10}
	repo.distByUser[50] = d

	err := svc.OnOrderPaid(context.Background(), OrderInfo{
		OrderID:         6001,
		UserID:          99,
		PayCents:        10000,
		DistributorUser: 50,
	})
	require.NoError(t, err)
	assert.Len(t, repo.commissions, 1)
	for _, c := range repo.commissions {
		assert.Equal(t, CommissionStatusPending, c.Status)
		assert.Equal(t, int64(1000), c.AmountCents) // 10000 * 0.10
	}
}

// Given 分销员未激活，When OnOrderPaid，Then 不创建佣金且无错误
func Test_OnOrderPaid_InactiveDistributor_NoCommission(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	d := &Distributor{ID: 21, UserID: 51, Status: DistStatusPending, Rate: 0.05}
	repo.distByUser[51] = d

	err := svc.OnOrderPaid(context.Background(), OrderInfo{
		OrderID:         7001,
		UserID:          100,
		PayCents:        20000,
		DistributorUser: 51,
	})
	require.NoError(t, err)
	assert.Empty(t, repo.commissions)
}

// =============================================================================
// RequestWithdraw — 提现申请
// =============================================================================

// Given 申请金额低于最小值，When RequestWithdraw，Then 立即返回 ErrWithdrawTooSmall（无需 Redis）
func Test_RequestWithdraw_AmountTooSmall_ErrWithdrawTooSmall(t *testing.T) {
	repo := newMockDistRepo()
	svc := newTestDistService(repo)

	_, err := svc.RequestWithdraw(context.Background(), 200, "idem-001", WithdrawReq{
		AmountCents: 500, // < 1000 (withdrawMinCents)
		SmsCode:     "123456",
	})
	errors.Is(err, ErrWithdrawTooSmall)
	assert.ErrorIs(t, err, ErrWithdrawTooSmall)
}
