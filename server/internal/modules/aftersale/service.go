// Package aftersale 售后协调模块，聚合 order/payment 操作。
package aftersale

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 依赖接口 ----

// AftersaleOrder 售后服务所需的订单信息。
type AftersaleOrder struct {
	ID                   int64
	OrderNo              string
	UserID               int64
	Status               string
	PayCents             int64
	CancelRequestPending bool
	CancelRequestReason  *string
	CancelRequestAt      *time.Time
}

// AftersaleOrderResp 售后订单响应 DTO。
type AftersaleOrderResp struct {
	ID                   types.Int64Str `json:"id"`
	OrderNo              string         `json:"order_no"`
	UserID               types.Int64Str `json:"user_id"`
	Status               string         `json:"status"`
	PayCents             int64          `json:"pay_cents"`
	CancelRequestPending bool           `json:"cancel_request_pending"`
	CancelRequestReason  *string        `json:"cancel_request_reason,omitempty"`
	CancelRequestAt      *time.Time     `json:"cancel_request_at,omitempty"`
}

func toAftersaleOrderResp(o AftersaleOrder) AftersaleOrderResp {
	return AftersaleOrderResp{
		ID:                   types.Int64Str(o.ID),
		OrderNo:              o.OrderNo,
		UserID:               types.Int64Str(o.UserID),
		Status:               o.Status,
		PayCents:             o.PayCents,
		CancelRequestPending: o.CancelRequestPending,
		CancelRequestReason:  o.CancelRequestReason,
		CancelRequestAt:      o.CancelRequestAt,
	}
}

// OrderRepo 售后服务对订单的访问接口。
type OrderRepo interface {
	// ListAftersale 查询待处理售后订单（cancel_request_pending=true 或 status IN refunding/refunded）。
	ListAftersale(ctx context.Context, page, size int) ([]AftersaleOrder, int64, error)
	// FindByID 查找订单。
	FindByID(ctx context.Context, id int64) (*AftersaleOrder, error)
	// UpdateCancelRequest 更新取消申请状态。
	UpdateCancelRequest(ctx context.Context, id int64, pending bool, cancelledAt *time.Time) error
}

// PaymentService 售后服务使用的支付接口。
type PaymentService interface {
	// ApplyRefund 发起退款（支持部分退款）。
	ApplyRefund(ctx context.Context, orderID, adminID int64, amtCents int64, reason string) error
}

// ---- Service ----

// Service 售后服务（薄壳，主要协调 order + payment）。
type Service struct {
	orderRepo      OrderRepo
	paymentService PaymentService
}

// NewService 构造售后服务。
func NewService(orderRepo OrderRepo, paymentService PaymentService) *Service {
	return &Service{
		orderRepo:      orderRepo,
		paymentService: paymentService,
	}
}

// ListAftersales 后台售后订单列表。
func (s *Service) ListAftersales(ctx context.Context, page, size int) ([]AftersaleOrderResp, int64, error) {
	list, total, err := s.orderRepo.ListAftersale(ctx, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resps := make([]AftersaleOrderResp, len(list))
	for i, o := range list {
		resps[i] = toAftersaleOrderResp(o)
	}
	return resps, total, nil
}

// ApproveCancel 同意取消申请（发起全额退款）。
func (s *Service) ApproveCancel(ctx context.Context, orderID, adminID int64) error {
	o, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if !o.CancelRequestPending {
		return errs.ErrParam.WithMsg("当前订单无待处理的取消申请")
	}

	reason := "管理员同意取消申请，全额退款"
	if o.CancelRequestReason != nil {
		reason = "买家申请原因：" + *o.CancelRequestReason + "；管理员同意"
	}

	// 发起全额退款
	if err := s.paymentService.ApplyRefund(ctx, orderID, adminID, o.PayCents, reason); err != nil {
		return err
	}

	// 清除取消申请标记
	if err := s.orderRepo.UpdateCancelRequest(ctx, orderID, false, nil); err != nil {
		return errs.ErrInternal
	}

	return nil
}

// RejectCancel 拒绝取消申请。
func (s *Service) RejectCancel(ctx context.Context, orderID, adminID int64, reason string) error {
	o, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if !o.CancelRequestPending {
		return errs.ErrParam.WithMsg("当前订单无待处理的取消申请")
	}

	_ = adminID
	_ = reason

	// 清除取消申请标记
	if err := s.orderRepo.UpdateCancelRequest(ctx, orderID, false, nil); err != nil {
		return errs.ErrInternal
	}

	return nil
}
