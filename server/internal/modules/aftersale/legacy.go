// 旧 cancel_request_pending 流程（v1.3 之前），保留供 admin 处理库存中残留的取消申请。
// 新业务请走 aftersale_order 表。

package aftersale

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// LegacyOrder 旧 cancel_request_pending 售后视图。
type LegacyOrder struct {
	ID                   int64
	OrderNo              string
	UserID               int64
	Status               string
	PayCents             int64
	CancelRequestPending bool
	CancelRequestReason  *string
	CancelRequestAt      *time.Time
}

// LegacyOrderResp 旧售后响应。
type LegacyOrderResp struct {
	ID                   types.Int64Str `json:"id"`
	OrderNo              string         `json:"order_no"`
	UserID               types.Int64Str `json:"user_id"`
	Status               string         `json:"status"`
	PayCents             int64          `json:"pay_cents"`
	CancelRequestPending bool           `json:"cancel_request_pending"`
	CancelRequestReason  *string        `json:"cancel_request_reason,omitempty"`
	CancelRequestAt      *time.Time     `json:"cancel_request_at,omitempty"`
}

func toLegacyResp(o LegacyOrder) LegacyOrderResp {
	return LegacyOrderResp{
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

// LegacyOrderRepo 旧 cancel_request_pending 数据访问。
type LegacyOrderRepo interface {
	ListAftersale(ctx context.Context, page, size int) ([]LegacyOrder, int64, error)
	FindByID(ctx context.Context, id int64) (*LegacyOrder, error)
	UpdateCancelRequest(ctx context.Context, id int64, pending bool, cancelledAt *time.Time) error
}

// ListLegacyCancelRequests 后台旧取消申请列表。
func (s *Service) ListLegacyCancelRequests(ctx context.Context, page, size int) ([]LegacyOrderResp, int64, error) {
	if s.legacyRepo == nil {
		return []LegacyOrderResp{}, 0, nil
	}
	list, total, err := s.legacyRepo.ListAftersale(ctx, page, size)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	out := make([]LegacyOrderResp, len(list))
	for i := range list {
		out[i] = toLegacyResp(list[i])
	}
	return out, total, nil
}

// LegacyApproveCancel 同意旧取消申请（全额退款）。
func (s *Service) LegacyApproveCancel(ctx context.Context, orderID, adminID int64) error {
	if s.legacyRepo == nil {
		return errs.ErrNotFound
	}
	o, err := s.legacyRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if err := s.paymentSvc.ApplyRefund(ctx, orderID, adminID, o.PayCents, reason); err != nil {
		return err
	}
	if err := s.legacyRepo.UpdateCancelRequest(ctx, orderID, false, nil); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// LegacyRejectCancel 拒绝旧取消申请。
func (s *Service) LegacyRejectCancel(ctx context.Context, orderID, adminID int64, reason string) error {
	if s.legacyRepo == nil {
		return errs.ErrNotFound
	}
	o, err := s.legacyRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if !o.CancelRequestPending {
		return errs.ErrParam.WithMsg("当前订单无待处理的取消申请")
	}
	_ = adminID
	_ = reason
	if err := s.legacyRepo.UpdateCancelRequest(ctx, orderID, false, nil); err != nil {
		return errs.ErrInternal
	}
	return nil
}
