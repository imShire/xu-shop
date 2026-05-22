package coupon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/modules/marketing/shared"
	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// =====================================================================================
// 测试基建：mock Repo + 空 SQLite DB（仅用于 db.Transaction 包装；mock repo 不读写 SQL）。
// =====================================================================================

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// mockRepo 内存仓储。
type mockRepo struct {
	tpls         map[int64]*CouponTemplate
	ucs          map[int64]*UserCoupon
	codes        map[string]*RedeemCode
	codeByID     map[int64]*RedeemCode
	tasks        map[int64]*GrantTask
	expireScan   []UserCoupon // ExpireScan 返回值
	idSeq        int64
	claimCounter map[string]int64 // (userID,tplID) -> count
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		tpls:         make(map[int64]*CouponTemplate),
		ucs:          make(map[int64]*UserCoupon),
		codes:        make(map[string]*RedeemCode),
		codeByID:     make(map[int64]*RedeemCode),
		tasks:        make(map[int64]*GrantTask),
		claimCounter: make(map[string]int64),
	}
}

func (m *mockRepo) nextID() int64 { m.idSeq++; return m.idSeq }

func (m *mockRepo) FindTemplate(_ context.Context, id int64) (*CouponTemplate, error) {
	if t, ok := m.tpls[id]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) FindTemplateForUpdate(_ context.Context, _ *gorm.DB, id int64) (*CouponTemplate, error) {
	return m.FindTemplate(context.Background(), id)
}

func (m *mockRepo) IncrTemplateClaimed(_ context.Context, _ *gorm.DB, id int64, delta int64) error {
	if t, ok := m.tpls[id]; ok {
		t.ClaimedCount += delta
	}
	return nil
}
func (m *mockRepo) IncrTemplateUsed(_ context.Context, _ *gorm.DB, id int64, delta int64) error {
	if t, ok := m.tpls[id]; ok {
		t.UsedCount += delta
	}
	return nil
}
func (m *mockRepo) ListOnlineTemplates(_ context.Context, _, _ int) ([]CouponTemplate, int64, error) {
	out := make([]CouponTemplate, 0, len(m.tpls))
	for _, t := range m.tpls {
		if t.Status == TplStatusOnline {
			out = append(out, *t)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockRepo) CreateUserCoupon(_ context.Context, _ *gorm.DB, uc *UserCoupon) error {
	cp := *uc
	m.ucs[uc.ID] = &cp
	key := keyOf(uc.UserID, uc.CouponTemplateID)
	m.claimCounter[key]++
	return nil
}

func (m *mockRepo) FindUserCoupon(_ context.Context, id int64) (*UserCoupon, error) {
	if uc, ok := m.ucs[id]; ok {
		cp := *uc
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) FindUserCouponForUpdate(_ context.Context, _ *gorm.DB, id int64) (*UserCoupon, error) {
	return m.FindUserCoupon(context.Background(), id)
}

func (m *mockRepo) UpdateUserCouponStatus(_ context.Context, _ *gorm.DB, id int64, fromStatus, toStatus string, fields map[string]any) (int64, error) {
	uc, ok := m.ucs[id]
	if !ok {
		return 0, nil
	}
	if uc.Status != fromStatus {
		return 0, nil
	}
	uc.Status = toStatus
	for k, v := range fields {
		switch k {
		case "order_id":
			if v == nil {
				uc.OrderID = nil
			} else if x, ok := v.(int64); ok {
				uc.OrderID = &x
			}
		case "locked_at":
			if v == nil {
				uc.LockedAt = nil
			} else if t, ok := v.(time.Time); ok {
				uc.LockedAt = &t
			}
		case "used_at":
			if v == nil {
				uc.UsedAt = nil
			} else if t, ok := v.(time.Time); ok {
				uc.UsedAt = &t
			}
		}
	}
	return 1, nil
}

func (m *mockRepo) CountUserClaim(_ context.Context, userID, templateID int64) (int64, error) {
	return m.claimCounter[keyOf(userID, templateID)], nil
}

func (m *mockRepo) ListMyCoupons(_ context.Context, userID int64, status string, _, _ int) ([]UserCoupon, int64, error) {
	var out []UserCoupon
	for _, uc := range m.ucs {
		if uc.UserID == userID && (status == "" || uc.Status == status) {
			out = append(out, *uc)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockRepo) FindUserCouponByOrder(_ context.Context, orderID int64) (*UserCoupon, error) {
	for _, uc := range m.ucs {
		if uc.OrderID != nil && *uc.OrderID == orderID && (uc.Status == UCStatusLocked || uc.Status == UCStatusUsed) {
			cp := *uc
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) FindRedeemCode(_ context.Context, code string) (*RedeemCode, error) {
	if rc, ok := m.codes[code]; ok {
		cp := *rc
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) UseRedeemCode(_ context.Context, _ *gorm.DB, codeID int64, userID int64) (int64, error) {
	rc, ok := m.codeByID[codeID]
	if !ok {
		return 0, nil
	}
	if rc.Status != RCStatusUnused {
		return 0, nil
	}
	now := time.Now()
	rc.Status = RCStatusUsed
	rc.UsedByUserID = &userID
	rc.UsedAt = &now
	return 1, nil
}

func (m *mockRepo) CreateGrantTask(_ context.Context, t *GrantTask) error {
	cp := *t
	m.tasks[t.ID] = &cp
	return nil
}
func (m *mockRepo) UpdateGrantTaskProgress(_ context.Context, _ int64, _, _ int64, _ string) error {
	return nil
}

func (m *mockRepo) ScanExpire(_ context.Context, _ time.Time, _ int) ([]UserCoupon, error) {
	return m.expireScan, nil
}

func keyOf(a, b int64) string {
	return formatInt(a) + ":" + formatInt(b)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// 工厂
func tplOnline(id int64, mod func(*CouponTemplate)) *CouponTemplate {
	t := &CouponTemplate{
		ID:               id,
		Name:             "T",
		Type:             TypeAmount,
		ValueCents:       1000,
		ScopeType:        ScopeAll,
		ValidityMode:     ValidityRelative,
		Status:           TplStatusOnline,
		PerUserLimit:     1,
	}
	v := 30
	t.ValidDays = &v
	if mod != nil {
		mod(t)
	}
	return t
}

func snapshotFromTpl(t *CouponTemplate) JSONMap {
	return JSONMap{
		"name":               t.Name,
		"type":               t.Type,
		"value_cents":        float64(t.ValueCents),
		"max_discount_cents": float64(t.MaxDiscountCents),
		"min_amount_cents":   float64(t.MinAmountCents),
		"scope_type":         t.ScopeType,
		"include_freight":    t.IncludeFreight,
	}
}

// 注入一个 unused 的用户券。
func seedUC(m *mockRepo, id, userID int64, status string, mod func(*UserCoupon)) *UserCoupon {
	uc := &UserCoupon{
		ID:               id,
		UserID:           userID,
		CouponTemplateID: 1,
		Status:           status,
		ClaimedAt:        time.Now(),
		ExpireAt:         time.Now().Add(24 * time.Hour),
		Snapshot: JSONMap{
			"type":             TypeAmount,
			"value_cents":      float64(1000),
			"min_amount_cents": float64(0),
			"scope_type":       ScopeAll,
		},
	}
	if mod != nil {
		mod(uc)
	}
	cp := *uc
	m.ucs[id] = &cp
	return uc
}

// =====================================================================================
// Claim
// =====================================================================================

func TestService_Claim(t *testing.T) {
	t.Run("success_creates_user_coupon_and_increments_claimed", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, nil)
		s := NewService(repo, newTestDB(t))

		uc, err := s.Claim(context.Background(), 100, 1, "active", nil)
		if err != nil {
			t.Fatalf("claim: %v", err)
		}
		if uc.UserID != 100 || uc.CouponTemplateID != 1 || uc.Status != UCStatusUnused {
			t.Fatalf("unexpected uc: %+v", uc)
		}
		if repo.tpls[1].ClaimedCount != 1 {
			t.Fatalf("claimed_count want 1, got %d", repo.tpls[1].ClaimedCount)
		}
	})

	t.Run("per_user_limit_exceeded", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) { t.PerUserLimit = 1 })
		// 预置一次领取记录
		repo.claimCounter[keyOf(100, 1)] = 1
		s := NewService(repo, newTestDB(t))

		_, err := s.Claim(context.Background(), 100, 1, "active", nil)
		if !errors.Is(err, shared.ErrCouponClaimLimit) {
			t.Fatalf("want ErrCouponClaimLimit, got %v", err)
		}
	})

	t.Run("template_offline_rejected", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) { t.Status = TplStatusOffline })
		s := NewService(repo, newTestDB(t))

		_, err := s.Claim(context.Background(), 100, 1, "active", nil)
		if !errors.Is(err, shared.ErrCouponTemplateOffline) {
			t.Fatalf("want offline err, got %v", err)
		}
	})

	t.Run("quota_exhausted", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) {
			t.TotalQuota = 5
			t.ClaimedCount = 5
		})
		s := NewService(repo, newTestDB(t))

		_, err := s.Claim(context.Background(), 100, 1, "active", nil)
		if !errors.Is(err, shared.ErrCouponQuotaExhausted) {
			t.Fatalf("want quota err, got %v", err)
		}
	})

	t.Run("claim_window_not_started", func(t *testing.T) {
		repo := newMockRepo()
		future := time.Now().Add(24 * time.Hour)
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) { t.ClaimStartAt = &future })
		s := NewService(repo, newTestDB(t))

		_, err := s.Claim(context.Background(), 100, 1, "active", nil)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("invalid_param", func(t *testing.T) {
		s := NewService(newMockRepo(), newTestDB(t))
		if _, err := s.Claim(context.Background(), 0, 1, "x", nil); !errors.Is(err, errs.ErrParam) {
			t.Fatalf("want ErrParam, got %v", err)
		}
	})
}

// =====================================================================================
// Quote
// =====================================================================================

func TestService_Quote(t *testing.T) {
	repo := newMockRepo()
	s := NewService(repo, newTestDB(t))
	ctx := context.Background()

	// 满减 100 元，门槛 50 元
	seedUC(repo, 1, 100, UCStatusUnused, func(uc *UserCoupon) {
		uc.Snapshot = JSONMap{
			"type":             TypeAmount,
			"value_cents":      float64(10000),
			"min_amount_cents": float64(5000),
			"scope_type":       ScopeAll,
		}
	})
	// 折扣 8 折，封顶 50 元
	seedUC(repo, 2, 100, UCStatusUnused, func(uc *UserCoupon) {
		rate := 0.8
		uc.Snapshot = JSONMap{
			"type":               TypeDiscount,
			"discount_rate":      rate,
			"max_discount_cents": float64(5000),
			"min_amount_cents":   float64(0),
			"scope_type":         ScopeAll,
		}
	})
	// 已过期
	seedUC(repo, 3, 100, UCStatusUnused, func(uc *UserCoupon) {
		uc.ExpireAt = time.Now().Add(-time.Hour)
	})
	// 已锁定
	seedUC(repo, 4, 100, UCStatusLocked, nil)
	// 类目限定
	seedUC(repo, 5, 100, UCStatusUnused, func(uc *UserCoupon) {
		uc.Snapshot = JSONMap{
			"type":             TypeAmount,
			"value_cents":      float64(2000),
			"min_amount_cents": float64(0),
			"scope_type":       ScopeCategory,
			"scope_targets":    []any{float64(7), float64(8)},
		}
	})

	t.Run("zero_id_returns_zero", func(t *testing.T) {
		got, err := s.Quote(ctx, QuoteReq{UserID: 100})
		if err != nil || got != 0 {
			t.Fatalf("want 0/nil, got %d/%v", got, err)
		}
	})

	t.Run("amount_below_threshold", func(t *testing.T) {
		_, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 1, OrderAmountCents: 4000})
		if !errors.Is(err, shared.ErrCouponNotEligible) {
			t.Fatalf("want not eligible, got %v", err)
		}
	})

	t.Run("amount_coupon_full_deduct", func(t *testing.T) {
		got, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 1, OrderAmountCents: 8000})
		if err != nil || got != 8000 { // value 10000 但订单只 8000 → 抵扣全部
			t.Fatalf("want 8000, got %d/%v", got, err)
		}
	})

	t.Run("amount_coupon_normal", func(t *testing.T) {
		got, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 1, OrderAmountCents: 20000})
		if err != nil || got != 10000 {
			t.Fatalf("want 10000, got %d/%v", got, err)
		}
	})

	t.Run("discount_coupon_with_max_cap", func(t *testing.T) {
		// 100000 * (1 - 0.8) = 20000，但封顶 5000
		got, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 2, OrderAmountCents: 100000})
		if err != nil || got != 5000 {
			t.Fatalf("want 5000, got %d/%v", got, err)
		}
	})

	t.Run("discount_coupon_below_cap", func(t *testing.T) {
		// 10000 * 0.2 = 2000，未达封顶
		got, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 2, OrderAmountCents: 10000})
		if err != nil || got != 2000 {
			t.Fatalf("want 2000, got %d/%v", got, err)
		}
	})

	t.Run("expired_rejected", func(t *testing.T) {
		_, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 3, OrderAmountCents: 1000})
		if !errors.Is(err, shared.ErrCouponExpired) {
			t.Fatalf("want expired, got %v", err)
		}
	})

	t.Run("locked_rejected", func(t *testing.T) {
		_, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 4, OrderAmountCents: 1000})
		if !errors.Is(err, shared.ErrCouponLocked) {
			t.Fatalf("want locked, got %v", err)
		}
	})

	t.Run("forbidden_other_user", func(t *testing.T) {
		_, err := s.Quote(ctx, QuoteReq{UserID: 999, UserCouponID: 1, OrderAmountCents: 10000})
		if !errors.Is(err, errs.ErrForbidden) {
			t.Fatalf("want forbidden, got %v", err)
		}
	})

	t.Run("scope_category_match", func(t *testing.T) {
		got, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 5, OrderAmountCents: 10000, ItemCategoryIDs: []int64{7}})
		if err != nil || got != 2000 {
			t.Fatalf("want 2000, got %d/%v", got, err)
		}
	})

	t.Run("scope_category_miss", func(t *testing.T) {
		_, err := s.Quote(ctx, QuoteReq{UserID: 100, UserCouponID: 5, OrderAmountCents: 10000, ItemCategoryIDs: []int64{99}})
		if !errors.Is(err, shared.ErrCouponNotEligible) {
			t.Fatalf("want not eligible, got %v", err)
		}
	})
}

// =====================================================================================
// Lock / Consume / Release / RefundRestore
// =====================================================================================

func TestService_LockReleaseConsumeRefund(t *testing.T) {
	t.Run("lock_then_consume_marks_used", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, nil)
		seedUC(repo, 10, 100, UCStatusUnused, nil)
		s := NewService(repo, newTestDB(t))

		deduct, err := s.Lock(context.Background(), nil, 5001, 10, 100, 5000)
		if err != nil {
			t.Fatalf("lock: %v", err)
		}
		if deduct != 1000 {
			t.Fatalf("deduct want 1000, got %d", deduct)
		}
		if repo.ucs[10].Status != UCStatusLocked {
			t.Fatalf("status want locked, got %s", repo.ucs[10].Status)
		}

		if err := s.Consume(context.Background(), nil, 5001); err != nil {
			t.Fatalf("consume: %v", err)
		}
		if repo.ucs[10].Status != UCStatusUsed {
			t.Fatalf("status want used, got %s", repo.ucs[10].Status)
		}
		if repo.tpls[1].UsedCount != 1 {
			t.Fatalf("used_count want 1, got %d", repo.tpls[1].UsedCount)
		}
	})

	t.Run("lock_already_locked", func(t *testing.T) {
		repo := newMockRepo()
		seedUC(repo, 10, 100, UCStatusLocked, nil)
		s := NewService(repo, newTestDB(t))
		_, err := s.Lock(context.Background(), nil, 5001, 10, 100, 5000)
		if !errors.Is(err, shared.ErrCouponLocked) {
			t.Fatalf("want locked err, got %v", err)
		}
	})

	t.Run("lock_zero_id_returns_zero", func(t *testing.T) {
		s := NewService(newMockRepo(), newTestDB(t))
		got, err := s.Lock(context.Background(), nil, 1, 0, 100, 5000)
		if err != nil || got != 0 {
			t.Fatalf("want 0/nil, got %d/%v", got, err)
		}
	})

	t.Run("lock_then_release_back_to_unused", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, nil)
		seedUC(repo, 10, 100, UCStatusUnused, nil)
		s := NewService(repo, newTestDB(t))

		if _, err := s.Lock(context.Background(), nil, 5001, 10, 100, 5000); err != nil {
			t.Fatalf("lock: %v", err)
		}
		if err := s.Release(context.Background(), nil, 5001); err != nil {
			t.Fatalf("release: %v", err)
		}
		if repo.ucs[10].Status != UCStatusUnused {
			t.Fatalf("status want unused, got %s", repo.ucs[10].Status)
		}
	})

	t.Run("refund_restore_full_revives_used_to_unused", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) { t.UsedCount = 1 })
		oid := int64(5001)
		seedUC(repo, 10, 100, UCStatusUsed, func(uc *UserCoupon) { uc.OrderID = &oid })
		s := NewService(repo, newTestDB(t))

		if err := s.RefundRestore(context.Background(), nil, 5001, true); err != nil {
			t.Fatalf("refund: %v", err)
		}
		if repo.ucs[10].Status != UCStatusUnused {
			t.Fatalf("want unused, got %s", repo.ucs[10].Status)
		}
		if repo.tpls[1].UsedCount != 0 {
			t.Fatalf("used_count want 0, got %d", repo.tpls[1].UsedCount)
		}
	})

	t.Run("refund_restore_partial_no_change", func(t *testing.T) {
		repo := newMockRepo()
		oid := int64(5001)
		seedUC(repo, 10, 100, UCStatusUsed, func(uc *UserCoupon) { uc.OrderID = &oid })
		s := NewService(repo, newTestDB(t))

		if err := s.RefundRestore(context.Background(), nil, 5001, false); err != nil {
			t.Fatalf("refund: %v", err)
		}
		if repo.ucs[10].Status != UCStatusUsed {
			t.Fatalf("status should remain used, got %s", repo.ucs[10].Status)
		}
	})

	t.Run("refund_restore_expired_no_change", func(t *testing.T) {
		repo := newMockRepo()
		oid := int64(5001)
		seedUC(repo, 10, 100, UCStatusUsed, func(uc *UserCoupon) {
			uc.OrderID = &oid
			uc.ExpireAt = time.Now().Add(-time.Hour)
		})
		s := NewService(repo, newTestDB(t))

		if err := s.RefundRestore(context.Background(), nil, 5001, true); err != nil {
			t.Fatalf("refund: %v", err)
		}
		if repo.ucs[10].Status != UCStatusUsed {
			t.Fatalf("expired coupon should not restore, got %s", repo.ucs[10].Status)
		}
	})
}

// =====================================================================================
// Transition
// =====================================================================================

func TestService_Transition(t *testing.T) {
	t.Run("invalid_trigger_returns_state_error", func(t *testing.T) {
		repo := newMockRepo()
		seedUC(repo, 10, 100, UCStatusUnused, nil)
		s := NewService(repo, newTestDB(t))
		err := s.Transition(context.Background(), nil, 10, "consume", nil) // unused 不可直接 consume
		if !errors.Is(err, shared.ErrInvalidStateTransition) {
			t.Fatalf("want invalid trans, got %v", err)
		}
	})

	t.Run("expire_from_unused_ok", func(t *testing.T) {
		repo := newMockRepo()
		seedUC(repo, 10, 100, UCStatusUnused, nil)
		s := NewService(repo, newTestDB(t))
		if err := s.Transition(context.Background(), nil, 10, "expire", nil); err != nil {
			t.Fatalf("want ok, got %v", err)
		}
		if repo.ucs[10].Status != UCStatusExpired {
			t.Fatalf("want expired, got %s", repo.ucs[10].Status)
		}
	})
}

// =====================================================================================
// ExpireScan
// =====================================================================================

func TestService_ExpireScan(t *testing.T) {
	repo := newMockRepo()
	seedUC(repo, 10, 100, UCStatusUnused, nil)
	seedUC(repo, 11, 100, UCStatusUnused, nil)
	repo.expireScan = []UserCoupon{*repo.ucs[10], *repo.ucs[11]}
	s := NewService(repo, newTestDB(t))

	n, err := s.ExpireScan(context.Background(), 100)
	if err != nil {
		t.Fatalf("expire: %v", err)
	}
	if n != 2 {
		t.Fatalf("processed want 2, got %d", n)
	}
	if repo.ucs[10].Status != UCStatusExpired || repo.ucs[11].Status != UCStatusExpired {
		t.Fatalf("statuses not expired: %s/%s", repo.ucs[10].Status, repo.ucs[11].Status)
	}
}

// =====================================================================================
// ClaimByCode
// =====================================================================================

func TestService_ClaimByCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newMockRepo()
		repo.tpls[1] = tplOnline(1, func(t *CouponTemplate) { t.PerUserLimit = 0 })
		rc := &RedeemCode{ID: 99, TemplateID: 1, Code: "ABC123", Status: RCStatusUnused}
		repo.codes["ABC123"] = rc
		repo.codeByID[99] = rc
		s := NewService(repo, newTestDB(t))

		uc, err := s.ClaimByCode(context.Background(), 100, "ABC123")
		if err != nil {
			t.Fatalf("claim by code: %v", err)
		}
		if uc.UserID != 100 {
			t.Fatalf("uc.user_id %d", uc.UserID)
		}
		if rc.Status != RCStatusUsed {
			t.Fatalf("code should be used")
		}
	})

	t.Run("invalid_code", func(t *testing.T) {
		s := NewService(newMockRepo(), newTestDB(t))
		_, err := s.ClaimByCode(context.Background(), 100, "NOPE")
		if !errors.Is(err, shared.ErrRedeemCodeInvalid) {
			t.Fatalf("want invalid, got %v", err)
		}
	})

	t.Run("already_used", func(t *testing.T) {
		repo := newMockRepo()
		uid := int64(50)
		rc := &RedeemCode{ID: 99, TemplateID: 1, Code: "USED", Status: RCStatusUsed, UsedByUserID: &uid}
		repo.codes["USED"] = rc
		repo.codeByID[99] = rc
		s := NewService(repo, newTestDB(t))

		_, err := s.ClaimByCode(context.Background(), 100, "USED")
		if !errors.Is(err, shared.ErrRedeemCodeUsed) {
			t.Fatalf("want used, got %v", err)
		}
	})
}
