package payment

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// ---- mock PaymentRepo ----

type mockPaymentRepo struct {
	payments map[int64]*Payment
	refunds  map[string]*Refund
	diffs    map[int64]*ReconciliationDiff
	txnIndex map[string]int64 // txnID -> payment.ID
	idSeq    int64
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{
		payments: make(map[int64]*Payment),
		refunds:  make(map[string]*Refund),
		diffs:    make(map[int64]*ReconciliationDiff),
		txnIndex: make(map[string]int64),
	}
}

func (m *mockPaymentRepo) nextID() int64 {
	m.idSeq++
	return m.idSeq
}

func (m *mockPaymentRepo) DB() *gorm.DB { return nil }

func (m *mockPaymentRepo) UpsertByTxn(_ context.Context, txnID, _ string, orderID, amtCents int64, raw []byte) (*Payment, bool, error) {
	if existID, ok := m.txnIndex[txnID]; ok {
		return m.payments[existID], false, nil
	}
	p := &Payment{
		ID:            m.nextID(),
		OrderID:       orderID,
		Channel:       "wxpay",
		TradeType:     "JSAPI",
		TransactionID: &txnID,
		AmountCents:   amtCents,
		Status:        PayStatusPending,
		RawNotify:     rawNotifyFromBytes(raw),
	}
	m.payments[p.ID] = p
	m.txnIndex[txnID] = p.ID
	return p, true, nil
}

func (m *mockPaymentRepo) FindByOrderID(_ context.Context, orderID int64) (*Payment, error) {
	for _, p := range m.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockPaymentRepo) FindSuccessByOrder(_ context.Context, orderID int64) (*Payment, error) {
	for _, p := range m.payments {
		if p.OrderID == orderID && p.Status == PayStatusSuccess {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockPaymentRepo) MarkSuccess(_ context.Context, id int64, raw []byte, paidAt time.Time) error {
	if p, ok := m.payments[id]; ok {
		p.Status = PayStatusSuccess
		p.RawNotify = rawNotifyFromBytes(raw)
		p.PaidAt = &paidAt
	}
	return nil
}

func (m *mockPaymentRepo) MarkFailed(_ context.Context, id int64) error {
	if p, ok := m.payments[id]; ok {
		p.Status = PayStatusFailed
	}
	return nil
}

func (m *mockPaymentRepo) MarkOrphan(_ context.Context, id int64) error {
	if p, ok := m.payments[id]; ok {
		p.Status = PayStatusOrphan
	}
	return nil
}

func (m *mockPaymentRepo) MarkSuccessTx(_ context.Context, _ *gorm.DB, id int64, raw []byte, paidAt time.Time) error {
	if p, ok := m.payments[id]; ok {
		p.Status = PayStatusSuccess
		now := paidAt
		p.PaidAt = &now
		p.RawNotify = rawNotifyFromBytes(raw)
	}
	return nil
}

func (m *mockPaymentRepo) MarkOrphanTx(_ context.Context, _ *gorm.DB, id int64) error {
	if p, ok := m.payments[id]; ok {
		p.Status = PayStatusOrphan
	}
	return nil
}

func (m *mockPaymentRepo) CreatePayment(_ context.Context, p *Payment) error {
	if p.ID == 0 {
		p.ID = m.nextID()
	}
	cp := *p
	m.payments[p.ID] = &cp
	return nil
}

func (m *mockPaymentRepo) CreateRefund(_ context.Context, r *Refund) error {
	if r.ID == 0 {
		r.ID = m.nextID()
	}
	cp := *r
	m.refunds[r.RefundNo] = &cp
	return nil
}

func (m *mockPaymentRepo) FindRefundByNo(_ context.Context, refundNo string) (*Refund, error) {
	if r, ok := m.refunds[refundNo]; ok {
		return r, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockPaymentRepo) SumRefundSuccess(_ context.Context, orderID int64) (int64, error) {
	var total int64
	for _, r := range m.refunds {
		if r.OrderID == orderID && r.Status == RefundStatusSuccess {
			total += r.AmountCents
		}
	}
	return total, nil
}

func (m *mockPaymentRepo) MarkRefundSuccess(_ context.Context, id int64, raw []byte, refundedAt time.Time) error {
	for _, r := range m.refunds {
		if r.ID == id {
			r.Status = RefundStatusSuccess
			r.RawNotify = rawNotifyFromBytes(raw)
			r.RefundedAt = &refundedAt
		}
	}
	return nil
}

func (m *mockPaymentRepo) MarkRefundFailed(_ context.Context, id int64, raw []byte) error {
	for _, r := range m.refunds {
		if r.ID == id {
			r.Status = RefundStatusFailed
			r.RawNotify = rawNotifyFromBytes(raw)
		}
	}
	return nil
}

func (m *mockPaymentRepo) ListRefunds(_ context.Context, orderID int64) ([]Refund, error) {
	var list []Refund
	for _, r := range m.refunds {
		if r.OrderID == orderID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (m *mockPaymentRepo) CreateReconciliationDiff(_ context.Context, d *ReconciliationDiff) error {
	if d.ID == 0 {
		d.ID = m.nextID()
	}
	m.diffs[d.ID] = d
	return nil
}

func (m *mockPaymentRepo) ListDiffs(_ context.Context, _ DiffFilter) ([]ReconciliationDiff, int64, error) {
	return nil, 0, nil
}

func (m *mockPaymentRepo) ResolveDiff(_ context.Context, id, _ int64) error {
	if d, ok := m.diffs[id]; ok {
		d.Status = DiffStatusResolved
	}
	return nil
}

func (m *mockPaymentRepo) ListPayments(_ context.Context, _ PaymentFilter) ([]Payment, int64, error) {
	return nil, 0, nil
}

func (m *mockPaymentRepo) ListAuditLogs(_ context.Context, _ AuditLogFilter) ([]AuditLog, int64, error) {
	return nil, 0, nil
}

// ---- mock OrderAccessor ----

type mockOrderAccessor struct {
	orders      map[int64]*OrderSnapshot
	transitions []string
}

func newMockOrderAccessor() *mockOrderAccessor {
	return &mockOrderAccessor{orders: make(map[int64]*OrderSnapshot)}
}

func (m *mockOrderAccessor) DB() *gorm.DB { return nil }

func (m *mockOrderAccessor) FindByOrderNo(_ context.Context, orderNo string) (*OrderSnapshot, error) {
	for _, o := range m.orders {
		if o.OrderNo == orderNo {
			cp := *o
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderAccessor) FindByID(_ context.Context, id int64) (*OrderSnapshot, error) {
	if o, ok := m.orders[id]; ok {
		cp := *o
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderAccessor) SetPrepayID(_ context.Context, orderID int64, prepayID string, expireAt time.Time) error {
	if o, ok := m.orders[orderID]; ok {
		o.CurrentPrepayID = &prepayID
		o.CurrentPrepayExpireAt = &expireAt
	}
	return nil
}

func (m *mockOrderAccessor) Transition(_ context.Context, _ int64, trigger, _ string, _ int64, _ string) error {
	m.transitions = append(m.transitions, trigger)
	return nil
}

func (m *mockOrderAccessor) TransitionInTx(_ context.Context, _ *gorm.DB, _ int64, trigger, _ string, _ int64, _ string) error {
	m.transitions = append(m.transitions, trigger)
	return nil
}

func (m *mockOrderAccessor) DeductStock(_ context.Context, _ int64, _ string) {}

// ---- mock UserOpenidGetter ----

type mockUserGetter struct {
	openids map[string]string // key: "scene:userID" -> openid
}

func (m *mockUserGetter) GetOpenid(_ context.Context, userID int64, scene string) (string, error) {
	key := fmt.Sprintf("%s:%d", scene, userID)
	return m.openids[key], nil
}

// ---- mock AsynqEnqueuer ----

type mockEnqueuer struct {
	tasks []*asynq.Task
}

func (m *mockEnqueuer) EnqueueContext(_ context.Context, t *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	m.tasks = append(m.tasks, t)
	return &asynq.TaskInfo{}, nil
}

// ---- 测试辅助 ----

func buildPaySvc(repo PaymentRepo, orderDB OrderAccessor, userGetter UserOpenidGetter,
	wxClient wxpay.Client, enqueuer AsynqEnqueuer) *Service {
	return &Service{
		repo:        repo,
		orderDB:     orderDB,
		userOpenid:  userGetter,
		wxpayClient: wxClient,
		enqueuer:    enqueuer,
		sceneCfg: WxPaySceneConfig{
			AppIDMP: "wx_mp_appid",
			AppIDOA: "wx_oa_appid",
			MchID:   "test_mchid",
		},
	}
}

func seededOrderSnap(accessor *mockOrderAccessor, id int64, orderNo string, payCents int64, status string) {
	expire := time.Now().Add(15 * time.Minute)
	accessor.orders[id] = &OrderSnapshot{
		ID:       id,
		OrderNo:  orderNo,
		UserID:   1,
		Status:   status,
		PayCents: payCents,
		ExpireAt: expire,
	}
}

func seededSuccessPayment(repo *mockPaymentRepo, id, orderID, amtCents int64) {
	txnID := fmt.Sprintf("txn_%d", id)
	repo.payments[id] = &Payment{
		ID:            id,
		OrderID:       orderID,
		Channel:       "wxpay",
		TradeType:     "JSAPI",
		TransactionID: &txnID,
		AmountCents:   amtCents,
		Status:        PayStatusSuccess,
	}
}

// ---- 正式测试 ----

// TestHandleWxpayNotify_Idempotent 重复回调只处理一次（UpsertByTxn 幂等）。
func TestHandleWxpayNotify_Idempotent(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1001, "ORDER001", 9900, "pending")

	ctx := context.Background()

	// 第一次插入
	pmt1, isNew1, err := repo.UpsertByTxn(ctx, "TXN_001", "ORDER001", 1001, 9900, []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if !isNew1 {
		t.Fatal("first insert should be new")
	}

	// 第二次插入（相同 txnID）
	pmt2, isNew2, err := repo.UpsertByTxn(ctx, "TXN_001", "ORDER001", 1001, 9900, []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if isNew2 {
		t.Fatal("second insert should NOT be new (idempotent)")
	}
	if pmt2.ID != pmt1.ID {
		t.Fatalf("idempotent: should return same payment id, got %d != %d", pmt2.ID, pmt1.ID)
	}

	// 状态机不应被重复触发
	if len(orderDB.transitions) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(orderDB.transitions))
	}
}

// TestHandleWxpayNotify_AmountMismatch 金额不一致应入队自动退款，订单不变更为 paid。
func TestHandleWxpayNotify_AmountMismatch(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1002, "ORDER002", 9900, "pending")

	ctx := context.Background()
	notifyAmtCents := int64(8800) // 实际收到 8800，订单应付 9900

	// 幂等插入
	pmt, isNew, err := repo.UpsertByTxn(ctx, "TXN_002", "ORDER002", 1002, notifyAmtCents, []byte("{}"))
	if err != nil || !isNew {
		t.Fatalf("upsert failed: err=%v isNew=%v", err, isNew)
	}

	// 金额不一致：MarkSuccess + MarkOrphan + 入队退款
	now := time.Now()
	_ = repo.MarkSuccess(ctx, pmt.ID, []byte("{}"), now)
	_ = repo.MarkOrphan(ctx, pmt.ID)

	enqueuer := &mockEnqueuer{}
	task := asynq.NewTask(TaskPaymentAutoRefund, []byte(`{"order_id":1002}`))
	_, _ = enqueuer.EnqueueContext(ctx, task)

	// 断言：payment 状态为 orphan（金额异常标记）
	p := repo.payments[pmt.ID]
	if p.Status != PayStatusOrphan {
		t.Fatalf("expected orphan, got %s", p.Status)
	}

	// 断言：自动退款任务已入队
	if len(enqueuer.tasks) != 1 {
		t.Fatalf("expected 1 auto_refund task, got %d", len(enqueuer.tasks))
	}
	if enqueuer.tasks[0].Type() != TaskPaymentAutoRefund {
		t.Fatalf("expected task type %s, got %s", TaskPaymentAutoRefund, enqueuer.tasks[0].Type())
	}

	// 断言：订单状态机未触发 pay_success
	for _, tr := range orderDB.transitions {
		if tr == "pay_success" {
			t.Fatalf("order should NOT transition to paid on amount mismatch")
		}
	}
}

// TestApplyRefund_PartialRefund 部分退款成功。
func TestApplyRefund_PartialRefund(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1003, "ORDER003", 10000, "paid")
	seededSuccessPayment(repo, 301, 1003, 10000)

	wxClient := wxpay.NewMockClient()
	enqueuer := &mockEnqueuer{}
	svc := buildPaySvc(repo, orderDB, nil, wxClient, enqueuer)

	err := svc.ApplyRefund(context.Background(), 1003, 1, 5000, "部分退款")
	if err != nil {
		t.Fatalf("partial refund failed: %v", err)
	}

	// 验证退款记录
	if len(repo.refunds) == 0 {
		t.Fatal("refund record should be created")
	}
	for _, r := range repo.refunds {
		if r.AmountCents != 5000 {
			t.Fatalf("expected 5000, got %d", r.AmountCents)
		}
		if r.Status != RefundStatusPending {
			t.Fatalf("expected pending, got %s", r.Status)
		}
	}

	// 验证 refund_apply 触发
	found := false
	for _, tr := range orderDB.transitions {
		if tr == "refund_apply" {
			found = true
		}
	}
	if !found {
		t.Fatal("refund_apply transition should be triggered")
	}
}

// TestApplyRefund_ExceedTotal 超过实付金额应被拒绝。
func TestApplyRefund_ExceedTotal(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1004, "ORDER004", 10000, "paid")
	seededSuccessPayment(repo, 401, 1004, 10000)

	// 预置 8000 分已成功退款
	repo.refunds["R_existing"] = &Refund{
		ID:          1,
		OrderID:     1004,
		RefundNo:    "R_existing",
		AmountCents: 8000,
		Status:      RefundStatusSuccess,
	}

	wxClient := wxpay.NewMockClient()
	enqueuer := &mockEnqueuer{}
	svc := buildPaySvc(repo, orderDB, nil, wxClient, enqueuer)

	// 再退 5000（已退 8000 + 5000 = 13000 > 10000）
	err := svc.ApplyRefund(context.Background(), 1004, 1, 5000, "超额测试")
	if err == nil {
		t.Fatal("expected error when refund exceeds total")
	}
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrParam.Code {
		t.Fatalf("expected ErrParam code=%d, got code=%d: %s", errs.ErrParam.Code, ae.Code, ae.Message)
	}
}

// TestPrepay_MissingOpenID jsapi_mp 场景缺少 openid 返回 50001。
func TestPrepay_MissingOpenID(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1005, "ORDER005", 9900, "pending")

	// 返回空 openid
	userGetter := &mockUserGetter{openids: map[string]string{}}
	wxClient := wxpay.NewMockClient()
	enqueuer := &mockEnqueuer{}
	svc := buildPaySvc(repo, orderDB, userGetter, wxClient, enqueuer)

	_, err := svc.Prepay(context.Background(), 1, 1005, "jsapi_mp", "")
	if err == nil {
		t.Fatal("expected ErrOpenIDMissing")
	}
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrOpenIDMissing.Code {
		t.Fatalf("expected code %d, got code=%d: %s", errs.ErrOpenIDMissing.Code, ae.Code, ae.Message)
	}
}

// TestApplyRefund_AlreadyFullyRefunded 订单已全额退款后再次申请退款应返回参数错误。
func TestApplyRefund_AlreadyFullyRefunded(t *testing.T) {
	repo := newMockPaymentRepo()
	orderDB := newMockOrderAccessor()
	seededOrderSnap(orderDB, 1006, "ORDER006", 10000, "refunded")
	seededSuccessPayment(repo, 601, 1006, 10000)

	// 预置全额退款记录（10000 分已成功退款）
	repo.refunds["R_full"] = &Refund{
		ID:          2,
		OrderID:     1006,
		RefundNo:    "R_full",
		AmountCents: 10000,
		Status:      RefundStatusSuccess,
	}

	wxClient := wxpay.NewMockClient()
	enqueuer := &mockEnqueuer{}
	svc := buildPaySvc(repo, orderDB, nil, wxClient, enqueuer)

	// 再申请任意金额退款，已退总额 10000 + 新申请 1 = 10001 > 10000（原付款金额）
	err := svc.ApplyRefund(context.Background(), 1006, 1, 1, "重复退款测试")
	if err == nil {
		t.Fatal("expected error when order already fully refunded, got nil")
	}

	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrParam.Code {
		t.Fatalf("expected ErrParam code=%d, got code=%d: %s", errs.ErrParam.Code, ae.Code, ae.Message)
	}

	// 不应写入新退款记录
	if len(repo.refunds) != 1 {
		t.Fatalf("expected 1 refund record (no new), got %d", len(repo.refunds))
	}
}
