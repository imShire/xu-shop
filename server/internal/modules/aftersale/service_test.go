package aftersale

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// ---- mock OrderRepo ----

type mockOrderRepo struct {
	orders map[int64]*AftersaleOrder
	// track UpdateCancelRequest calls
	updatedID      int64
	updatedPending bool
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[int64]*AftersaleOrder)}
}

func (m *mockOrderRepo) ListAftersale(_ context.Context, _, _ int) ([]AftersaleOrder, int64, error) {
	return nil, 0, nil
}

func (m *mockOrderRepo) FindByID(_ context.Context, id int64) (*AftersaleOrder, error) {
	if o, ok := m.orders[id]; ok {
		cp := *o
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderRepo) UpdateCancelRequest(_ context.Context, id int64, pending bool, _ *time.Time) error {
	m.updatedID = id
	m.updatedPending = pending
	if o, ok := m.orders[id]; ok {
		o.CancelRequestPending = pending
	}
	return nil
}

// ---- mock PaymentService ----

type mockPaymentService struct {
	refundCalled bool
	refundErr    error
}

func (m *mockPaymentService) ApplyRefund(_ context.Context, _, _ int64, _ int64, _ string) error {
	m.refundCalled = true
	return m.refundErr
}

// ---- helpers ----

func buildAftersaleSvc(orderRepo OrderRepo, paySvc PaymentService) *Service {
	return NewService(orderRepo, paySvc)
}

func seededAftersaleOrder(repo *mockOrderRepo, id int64, pending bool, payCents int64) {
	reason := "不想要了"
	repo.orders[id] = &AftersaleOrder{
		ID:                   id,
		OrderNo:              "AS_TEST_001",
		UserID:               100,
		Status:               "paid",
		PayCents:             payCents,
		CancelRequestPending: pending,
		CancelRequestReason:  &reason,
	}
}

// ---- Tests ----

// TestAdminApproveCancel_Success 有效的待处理取消申请被同意后触发退款并清除 pending 标记。
func TestAdminApproveCancel_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{}

	seededAftersaleOrder(orderRepo, 2001, true, 9900)

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.ApproveCancel(context.Background(), 2001, 1)
	if err != nil {
		t.Fatalf("ApproveCancel error: %v", err)
	}

	// 退款接口必须被调用
	if !paySvc.refundCalled {
		t.Error("expected ApplyRefund to be called")
	}

	// cancel_request_pending 必须被更新为 false
	if orderRepo.updatedID != 2001 {
		t.Errorf("expected UpdateCancelRequest called for order 2001, got %d", orderRepo.updatedID)
	}
	if orderRepo.updatedPending {
		t.Error("expected pending=false after approve")
	}
}

// TestAdminApproveCancel_NotPending 无待处理取消申请的订单同意时返回参数错误。
func TestAdminApproveCancel_NotPending(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{}

	// CancelRequestPending = false
	seededAftersaleOrder(orderRepo, 2002, false, 9900)

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.ApproveCancel(context.Background(), 2002, 1)
	if err == nil {
		t.Fatal("expected error when no pending cancel request, got nil")
	}

	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrParam.Code {
		t.Errorf("expected ErrParam code=%d, got code=%d: %s", errs.ErrParam.Code, ae.Code, ae.Message)
	}

	// 退款不应被调用
	if paySvc.refundCalled {
		t.Error("ApplyRefund should NOT be called when no pending cancel request")
	}
}

// TestAdminApproveCancel_OrderNotFound 不存在的订单同意时返回 not-found 错误。
func TestAdminApproveCancel_OrderNotFound(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{}

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.ApproveCancel(context.Background(), 9999, 1)
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}

	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrNotFound.Code {
		t.Errorf("expected ErrNotFound code=%d, got code=%d", errs.ErrNotFound.Code, ae.Code)
	}
}

// TestAdminRejectCancel_Success 拒绝取消申请后仅清除 pending 标记，不调用退款。
func TestAdminRejectCancel_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{}

	seededAftersaleOrder(orderRepo, 2003, true, 9900)

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.RejectCancel(context.Background(), 2003, 1, "商品正常，拒绝退款")
	if err != nil {
		t.Fatalf("RejectCancel error: %v", err)
	}

	// 退款不应被调用
	if paySvc.refundCalled {
		t.Error("ApplyRefund should NOT be called on reject")
	}

	// pending 标记清除
	if orderRepo.updatedID != 2003 {
		t.Errorf("expected UpdateCancelRequest called for order 2003, got %d", orderRepo.updatedID)
	}
	if orderRepo.updatedPending {
		t.Error("expected pending=false after reject")
	}
}

// TestAdminRejectCancel_NotPending 拒绝无待处理申请的订单返回参数错误。
func TestAdminRejectCancel_NotPending(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{}

	seededAftersaleOrder(orderRepo, 2004, false, 9900)

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.RejectCancel(context.Background(), 2004, 1, "some reason")
	if err == nil {
		t.Fatal("expected error when no pending cancel request, got nil")
	}

	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrParam.Code {
		t.Errorf("expected ErrParam code=%d, got code=%d: %s", errs.ErrParam.Code, ae.Code, ae.Message)
	}
}

// TestAdminApproveCancel_RefundFails 退款服务失败时 ApproveCancel 应返回错误。
func TestAdminApproveCancel_RefundFails(t *testing.T) {
	orderRepo := newMockOrderRepo()
	paySvc := &mockPaymentService{refundErr: errors.New("mock refund error")}

	seededAftersaleOrder(orderRepo, 2005, true, 9900)

	svc := buildAftersaleSvc(orderRepo, paySvc)

	err := svc.ApproveCancel(context.Background(), 2005, 1)
	if err == nil {
		t.Fatal("expected error when refund fails, got nil")
	}
}
