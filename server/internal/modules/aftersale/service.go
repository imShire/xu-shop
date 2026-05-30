package aftersale

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/lock"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 依赖接口 ----

// PaymentService 售后服务使用的支付接口（既有签名，保持不变）。
type PaymentService interface {
	ApplyRefund(ctx context.Context, orderID, adminID int64, amtCents int64, reason string) error
}

// OrderInfo 售后查询用的订单快照。
type OrderInfo struct {
	ID            int64
	OrderNo       string
	UserID        int64
	Status        string
	PayCents      int64
	RefundedCents int64
	DeliveredAt   *time.Time
	PaidAt        *time.Time
}

// OrderItemInfo 售后查询用的订单行快照。
type OrderItemInfo struct {
	ID            int64
	OrderID       int64
	SubtotalCents int64
}

// OrderAccessor 售后依赖的订单访问接口。由 cmd/api 用 *gorm.DB 适配 order 模块。
type OrderAccessor interface {
	FindOrder(ctx context.Context, orderID int64) (*OrderInfo, error)
	FindOrderItem(ctx context.Context, itemID int64) (*OrderItemInfo, error)
}

// DistributionHook 退款完成时调整佣金（distribution.OnOrderRefund）。允许 nil。
type DistributionHook interface {
	OnOrderRefund(ctx context.Context, orderID int64, refundCents, payCents int64, fullRefund bool) error
}

// Notifier 通知派发接口（aftersale 内部用，cmd/api 适配 notification.Service + asynq）。允许 nil。
type Notifier interface {
	NotifyAftersaleEvent(ctx context.Context, eventCode string, userID int64, refID string, params map[string]any) error
}

// ---- 配置 ----

// Windows 自动同意 / 关闭窗口。
type Windows struct {
	ApplyTimeout         time.Duration
	SellerAgreedTimeout  time.Duration
	BuyerReturnedTimeout time.Duration
	PostRefundApplyDays  int
}

// DefaultWindows 默认窗口。
func DefaultWindows() Windows {
	return Windows{
		ApplyTimeout:         5 * 24 * time.Hour,
		SellerAgreedTimeout:  7 * 24 * time.Hour,
		BuyerReturnedTimeout: 5 * 24 * time.Hour,
		PostRefundApplyDays:  7,
	}
}

// ---- Service ----

// Service 售后核心服务（状态机 + 自动扫描）。
type Service struct {
	repo       Repo
	orderAcc   OrderAccessor
	paymentSvc PaymentService
	distHook   DistributionHook
	notifier   Notifier
	locker     lock.Lock
	db         *gorm.DB
	rdb        *redis.Client
	windows    Windows
	now        func() time.Time

	legacyRepo LegacyOrderRepo
}

// NewService 构造 Service。
func NewService(
	repo Repo,
	orderAcc OrderAccessor,
	paymentSvc PaymentService,
	db *gorm.DB,
	rdb *redis.Client,
) *Service {
	s := &Service{
		repo:       repo,
		orderAcc:   orderAcc,
		paymentSvc: paymentSvc,
		db:         db,
		rdb:        rdb,
		windows:    DefaultWindows(),
		now:        time.Now,
	}
	if rdb != nil {
		s.locker = lock.New(rdb)
	}
	return s
}

// WithDistributionHook 注入分销 hook。
func (s *Service) WithDistributionHook(h DistributionHook) *Service { s.distHook = h; return s }

// WithNotifier 注入通知派发。
func (s *Service) WithNotifier(n Notifier) *Service { s.notifier = n; return s }

// WithWindows 自定义窗口。
func (s *Service) WithWindows(w Windows) *Service { s.windows = w; return s }

// WithLocker 自定义 lock（测试用）。
func (s *Service) WithLocker(l lock.Lock) *Service { s.locker = l; return s }

// WithNow 自定义当前时间（测试用）。
func (s *Service) WithNow(fn func() time.Time) *Service { s.now = fn; return s }

// WithLegacyOrderRepo 注入旧 cancel_request_pending 依赖。
func (s *Service) WithLegacyOrderRepo(r LegacyOrderRepo) *Service { s.legacyRepo = r; return s }

// ---- 锁 ----

func (s *Service) lockAftersale(ctx context.Context, id int64) (func(), error) {
	if s.locker == nil {
		return func() {}, nil
	}
	key := fmt.Sprintf("aftersale:%d", id)
	ok, err := s.locker.TryLock(ctx, key, 30*time.Second)
	if err != nil {
		return nil, errs.ErrInternal.WithMsg("lock error")
	}
	if !ok {
		return nil, errs.ErrConflict.WithMsg("操作进行中，请稍后重试")
	}
	return func() { _ = s.locker.Unlock(ctx, key) }, nil
}

// ---- Apply ----

// Apply 用户申请售后。
func (s *Service) Apply(ctx context.Context, userID int64, req ApplyReq) (*ApplyResp, error) {
	orderID := req.OrderID.Int64()
	if orderID <= 0 {
		return nil, errs.ErrParam.WithMsg("order_id required")
	}
	o, err := s.orderAcc.FindOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound.WithMsg("订单不存在")
		}
		return nil, errs.ErrInternal
	}
	if o.UserID != userID {
		return nil, errs.ErrForbidden
	}
	if err := s.validateOrderForApply(o, req); err != nil {
		return nil, err
	}
	var orderItemID *int64
	if req.OrderItemID != nil {
		itemID := req.OrderItemID.Int64()
		item, err := s.orderAcc.FindOrderItem(ctx, itemID)
		if err != nil || item == nil || item.OrderID != orderID {
			return nil, errs.ErrParam.WithMsg("order_item invalid")
		}
		if req.RefundAmountCents > item.SubtotalCents {
			return nil, errs.ErrParam.WithMsg("退款金额超过行金额")
		}
		orderItemID = &itemID
	} else {
		if req.RefundAmountCents > o.PayCents-o.RefundedCents {
			return nil, errs.ErrParam.WithMsg("退款金额超过可退余额")
		}
	}
	if len(req.Evidence) > 6 {
		return nil, errs.ErrParam.WithMsg("最多 6 张凭证")
	}

	exists, err := s.repo.FindActiveByOrder(ctx, orderID, orderItemID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	if exists != nil {
		return nil, errs.ErrConflict.WithMsg("已存在进行中的售后单")
	}

	now := s.now()
	id := snowflake.NextID()
	a := &AftersaleOrder{
		ID:                id,
		AftersaleNo:       GenAftersaleNo(id, now),
		OrderID:           orderID,
		OrderItemID:       orderItemID,
		UserID:            userID,
		Type:              req.Type,
		Reason:            req.Reason,
		RefundAmountCents: req.RefundAmountCents,
		Status:            StatusApplying,
		BuyerEvidence:     JSONStrArray(req.Evidence),
		AppliedAt:         now,
		AutoCloseAt:       now.Add(s.windows.ApplyTimeout),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(ctx, tx, a); err != nil {
			return err
		}
		return s.repo.AddNegotiation(ctx, tx, &AftersaleNegotiation{
			ID:          snowflake.NextID(),
			AftersaleID: a.ID,
			Role:        RoleBuyer,
			Content:     req.Reason,
			Evidence:    JSONStrArray(req.Evidence),
			CreatedAt:   now,
		})
	})
	if err != nil {
		return nil, errs.ErrInternal.WithMsg(err.Error())
	}
	s.notify(ctx, "aftersale.applied", userID, a)
	return &ApplyResp{ID: types.Int64Str(a.ID), AftersaleNo: a.AftersaleNo}, nil
}

func (s *Service) validateOrderForApply(o *OrderInfo, req ApplyReq) error {
	switch o.Status {
	case "paid":
		if req.Type != TypeRefundOnly {
			return errs.ErrParam.WithMsg("未发货订单仅支持仅退款")
		}
	case "shipped":
		// 已发货后所有售后类型均可申请
	case "delivered":
		if o.DeliveredAt != nil &&
			s.now().Sub(*o.DeliveredAt) > time.Duration(s.windows.PostRefundApplyDays)*24*time.Hour {
			return errs.ErrParam.WithMsg("已超过售后申请期限")
		}
	default:
		return errs.ErrParam.WithMsg("当前订单状态不支持申请售后")
	}
	return nil
}

// ---- Buyer Cancel ----

// Cancel 用户撤销售后申请。
func (s *Service) Cancel(ctx context.Context, userID, id int64) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if a.UserID != userID {
		return errs.ErrForbidden
	}
	switch a.Status {
	case StatusApplying, StatusSellerAgreed, StatusBuyerReturned:
	default:
		return errs.ErrConflict.WithMsg("当前状态不可撤销")
	}
	now := s.now()
	if err := s.repo.UpdateMap(ctx, nil, id, a.Status, map[string]any{
		"status":        StatusCancelled,
		"closed_at":     now,
		"auto_close_at": farFuture,
	}); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, id, RoleBuyer, nil, "用户撤销", nil)
	s.notify(ctx, "aftersale.cancelled", a.UserID, a)
	return nil
}

// ---- Admin Agree ----

// AdminAgree 商家同意。
func (s *Service) AdminAgree(ctx context.Context, adminID, id int64, req AgreeReq) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if a.Status != StatusApplying {
		return errs.ErrConflict.WithMsg("当前状态不可同意")
	}
	now := s.now()
	var opAdminPtr *int64
	if adminID > 0 {
		opAdminPtr = &adminID
	}
	fields := map[string]any{
		"status":            StatusSellerAgreed,
		"agreed_at":         now,
		"operator_admin_id": opAdminPtr,
		"seller_remark":     req.SellerRemark,
		"auto_close_at":     now.Add(s.windows.SellerAgreedTimeout),
	}
	if err := s.repo.UpdateMap(ctx, nil, id, StatusApplying, fields); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, id, RoleSeller, opAdminPtr, req.SellerRemark, nil)
	s.notify(ctx, "aftersale.agreed", a.UserID, a)

	if a.Type == TypeRefundOnly {
		// 仅退款：转入 seller_received 等同义路径，再触发 ApplyRefund，refund 回调驱动 completed。
		if err := s.repo.UpdateMap(ctx, nil, id, StatusSellerAgreed, map[string]any{
			"status":        StatusSellerReceived,
			"received_at":   now,
			"auto_close_at": farFuture,
		}); err != nil {
			return s.mapStaleErr(err)
		}
		a.Status = StatusSellerReceived
		return s.startRefund(ctx, a, adminID)
	}
	return nil
}

// ---- Admin Reject ----

// AdminReject 商家拒绝。
func (s *Service) AdminReject(ctx context.Context, adminID, id int64, req RejectReq) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if a.Status != StatusApplying {
		return errs.ErrConflict.WithMsg("当前状态不可拒绝")
	}
	now := s.now()
	var opAdminPtr *int64
	if adminID > 0 {
		opAdminPtr = &adminID
	}
	if err := s.repo.UpdateMap(ctx, nil, id, StatusApplying, map[string]any{
		"status":            StatusSellerRejected,
		"operator_admin_id": opAdminPtr,
		"seller_remark":     req.Reason,
		"closed_at":         now,
		"auto_close_at":     farFuture,
	}); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, id, RoleSeller, opAdminPtr, req.Reason, nil)
	s.notify(ctx, "aftersale.rejected", a.UserID, a)
	return nil
}

// ---- Buyer FillExpress ----

// FillExpress 用户回填寄回运单。
func (s *Service) FillExpress(ctx context.Context, userID, id int64, req ExpressReq) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if a.UserID != userID {
		return errs.ErrForbidden
	}
	if a.Status != StatusSellerAgreed {
		return errs.ErrConflict.WithMsg("当前状态不可寄回")
	}
	if a.Type == TypeRefundOnly {
		return errs.ErrParam.WithMsg("仅退款不需要寄回")
	}
	now := s.now()
	exp := &BuyerExpress{CarrierCode: req.CarrierCode, WaybillNo: req.WaybillNo, ShippedAt: &now}
	if err := s.repo.UpdateMap(ctx, nil, id, StatusSellerAgreed, map[string]any{
		"status":        StatusBuyerReturned,
		"returned_at":   now,
		"buyer_express": exp,
		"auto_close_at": now.Add(s.windows.BuyerReturnedTimeout),
	}); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, id, RoleBuyer, nil,
		fmt.Sprintf("已寄回 %s %s", req.CarrierCode, req.WaybillNo), nil)
	s.notify(ctx, "aftersale.returned", a.UserID, a)
	return nil
}

// ---- Admin ConfirmReceived ----

// AdminConfirmReceived 商家确认收货 → 触发退款 / 进入完成（exchange）。
func (s *Service) AdminConfirmReceived(ctx context.Context, adminID, id int64, req ConfirmReceivedReq) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if a.Status != StatusBuyerReturned {
		return errs.ErrConflict.WithMsg("当前状态不可确认")
	}
	now := s.now()
	var opAdminPtr *int64
	if adminID > 0 {
		opAdminPtr = &adminID
	}
	if err := s.repo.UpdateMap(ctx, nil, id, StatusBuyerReturned, map[string]any{
		"status":            StatusSellerReceived,
		"received_at":       now,
		"operator_admin_id": opAdminPtr,
		"seller_remark":     req.SellerRemark,
		"auto_close_at":     farFuture,
	}); err != nil {
		return s.mapStaleErr(err)
	}
	a.Status = StatusSellerReceived
	s.addAudit(ctx, id, RoleSeller, opAdminPtr, "已确认收货", nil)

	switch a.Type {
	case TypeRefundReturn:
		return s.startRefund(ctx, a, adminID)
	case TypeExchange:
		return s.completeAftersale(ctx, a, adminID)
	}
	return nil
}

// ---- AppendMessage ----

// AppendMessage 任意非终态下追加协商。
func (s *Service) AppendMessage(ctx context.Context, role string, actorID, id int64, req MessageReq) error {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if role == RoleBuyer && a.UserID != actorID {
		return errs.ErrForbidden
	}
	if IsTerminal(a.Status) {
		return errs.ErrConflict.WithMsg("终态不可追加消息")
	}
	if len(req.Evidence) > 6 {
		return errs.ErrParam.WithMsg("最多 6 张凭证")
	}
	var adminID *int64
	if role == RoleSeller {
		adminID = &actorID
	}
	s.addAudit(ctx, id, role, adminID, req.Content, req.Evidence)
	return nil
}

// ---- Admin Close ----

// AdminClose 后台强制关闭。
func (s *Service) AdminClose(ctx context.Context, adminID, id int64, req CloseReq) error {
	unlock, err := s.lockAftersale(ctx, id)
	if err != nil {
		return err
	}
	defer unlock()

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal
	}
	if IsTerminal(a.Status) {
		return errs.ErrConflict.WithMsg("已是终态")
	}
	now := s.now()
	var opAdminPtr *int64
	if adminID > 0 {
		opAdminPtr = &adminID
	}
	if err := s.repo.UpdateMap(ctx, nil, id, a.Status, map[string]any{
		"status":            StatusClosed,
		"operator_admin_id": opAdminPtr,
		"seller_remark":     req.Reason,
		"closed_at":         now,
		"auto_close_at":     farFuture,
	}); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, id, RoleSeller, opAdminPtr, "管理员关闭："+req.Reason, nil)
	s.notify(ctx, "aftersale.closed", a.UserID, a)
	return nil
}

// ---- 退款两阶段 ----

func (s *Service) startRefund(ctx context.Context, a *AftersaleOrder, adminID int64) error {
	if a.RefundAmountCents <= 0 {
		return errs.ErrParam.WithMsg("退款金额必须 > 0")
	}
	reason := fmt.Sprintf("aftersale:%s:%s", a.AftersaleNo, a.Reason)
	if err := s.paymentSvc.ApplyRefund(ctx, a.OrderID, adminID, a.RefundAmountCents, reason); err != nil {
		return err
	}
	refundID, _ := s.repo.FindRefundByOrderReason(ctx, a.OrderID, fmt.Sprintf("aftersale:%s:", a.AftersaleNo))
	if refundID > 0 {
		_ = s.repo.UpdateMap(ctx, nil, a.ID, a.Status, map[string]any{"refund_id": refundID})
	}
	return nil
}

func (s *Service) completeAftersale(ctx context.Context, a *AftersaleOrder, adminID int64) error {
	now := s.now()
	var opAdminPtr *int64
	if adminID > 0 {
		opAdminPtr = &adminID
	}
	if err := s.repo.UpdateMap(ctx, nil, a.ID, a.Status, map[string]any{
		"status":            StatusCompleted,
		"completed_at":      now,
		"closed_at":         now,
		"operator_admin_id": opAdminPtr,
		"auto_close_at":     farFuture,
	}); err != nil {
		return s.mapStaleErr(err)
	}
	s.addAudit(ctx, a.ID, RoleSystem, nil, "completed", nil)

	if s.distHook != nil && a.RefundAmountCents > 0 {
		o, err := s.orderAcc.FindOrder(ctx, a.OrderID)
		if err == nil && o != nil {
			full := a.RefundAmountCents >= o.PayCents
			_ = s.distHook.OnOrderRefund(ctx, a.OrderID, a.RefundAmountCents, o.PayCents, full)
		}
	}
	s.notify(ctx, "aftersale.completed", a.UserID, a)
	return nil
}

// ---- 查询 ----

// UserGet C 端详情。
func (s *Service) UserGet(ctx context.Context, userID, id int64) (*DetailResp, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	if a.UserID != userID {
		return nil, errs.ErrForbidden
	}
	return s.detail(ctx, a)
}

// AdminGet 后台详情。
func (s *Service) AdminGet(ctx context.Context, id int64) (*DetailResp, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	return s.detail(ctx, a)
}

func (s *Service) detail(ctx context.Context, a *AftersaleOrder) (*DetailResp, error) {
	negs, err := s.repo.ListNegotiations(ctx, a.ID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	out := &DetailResp{AftersaleResp: toAftersaleResp(a)}
	out.Negotiations = make([]NegotiationResp, len(negs))
	for i := range negs {
		out.Negotiations[i] = toNegotiationResp(&negs[i])
	}
	return out, nil
}

// UserList C 端列表。
func (s *Service) UserList(ctx context.Context, f UserListFilter) ([]AftersaleResp, int64, error) {
	rows, total, err := s.repo.ListByUser(ctx, f)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]AftersaleResp, len(rows))
	for i := range rows {
		resp[i] = toAftersaleResp(&rows[i])
	}
	return resp, total, nil
}

// AdminList 后台列表。
func (s *Service) AdminList(ctx context.Context, f AdminListFilter) ([]AftersaleResp, int64, error) {
	rows, total, err := s.repo.ListByAdmin(ctx, f)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]AftersaleResp, len(rows))
	for i := range rows {
		resp[i] = toAftersaleResp(&rows[i])
	}
	return resp, total, nil
}

// ---- AutoScan ----

// AutoScan worker 定时调用：处理超时 + 监听 refund 回调结果。
func (s *Service) AutoScan(ctx context.Context) error {
	now := s.now()

	rows, err := s.repo.ScanExpiring(ctx, now, 200)
	if err != nil {
		return err
	}
	for i := range rows {
		s.handleExpiring(ctx, &rows[i])
	}

	pendings, err := s.repo.ScanPendingRefund(ctx, 200)
	if err != nil {
		return err
	}
	for i := range pendings {
		a := &pendings[i]
		if a.RefundID == nil {
			continue
		}
		st, err := s.repo.FindRefundStatus(ctx, *a.RefundID)
		if err != nil || st != "success" {
			continue
		}
		unlock, lerr := s.lockAftersale(ctx, a.ID)
		if lerr != nil {
			continue
		}
		_ = s.completeAftersale(ctx, a, derefInt64(a.OperatorAdminID))
		unlock()
	}
	return nil
}

func (s *Service) handleExpiring(ctx context.Context, a *AftersaleOrder) {
	unlock, err := s.lockAftersale(ctx, a.ID)
	if err != nil {
		return
	}
	defer unlock()

	cur, err := s.repo.FindByID(ctx, a.ID)
	if err != nil || cur == nil {
		return
	}
	now := s.now()
	switch cur.Status {
	case StatusApplying:
		_ = s.AdminAgree(ctx, 0, cur.ID, AgreeReq{SellerRemark: "系统超时自动同意"})
	case StatusSellerAgreed:
		_ = s.repo.UpdateMap(ctx, nil, cur.ID, StatusSellerAgreed, map[string]any{
			"status":        StatusClosed,
			"closed_at":     now,
			"auto_close_at": farFuture,
			"seller_remark": "买家超时未寄回",
		})
		s.addAudit(ctx, cur.ID, RoleSystem, nil, "超时未寄回，自动关闭", nil)
		s.notify(ctx, "aftersale.closed", cur.UserID, cur)
	case StatusBuyerReturned:
		_ = s.AdminConfirmReceived(ctx, 0, cur.ID, ConfirmReceivedReq{SellerRemark: "系统超时自动确认"})
	}
}

// ---- 辅助 ----

func (s *Service) addAudit(ctx context.Context, aftersaleID int64, role string, adminID *int64, content string, evidence []string) {
	_ = s.repo.AddNegotiation(ctx, nil, &AftersaleNegotiation{
		ID:          snowflake.NextID(),
		AftersaleID: aftersaleID,
		Role:        role,
		AdminID:     adminID,
		Content:     content,
		Evidence:    JSONStrArray(evidence),
		CreatedAt:   s.now(),
	})
}

func (s *Service) notify(ctx context.Context, eventCode string, userID int64, a *AftersaleOrder) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.NotifyAftersaleEvent(ctx, eventCode, userID,
		fmt.Sprintf("aftersale:%d", a.ID), map[string]any{
			"aftersale_no": a.AftersaleNo,
			"order_id":     fmt.Sprintf("%d", a.OrderID),
			"type":         a.Type,
			"status":       a.Status,
		})
}

func (s *Service) mapStaleErr(err error) error {
	if errors.Is(err, ErrStaleStatus) {
		return errs.ErrConflict.WithMsg("状态已变更，请刷新重试")
	}
	return errs.ErrInternal.WithMsg(err.Error())
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
