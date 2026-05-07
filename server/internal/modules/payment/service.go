package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/types"
	"github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// ---- 依赖接口（避免导入循环） ----

// OrderSnapshot 支付服务所需的订单信息快照。
type OrderSnapshot struct {
	ID                    int64
	OrderNo               string
	UserID                int64
	Status                string
	PayCents              int64
	ExpireAt              time.Time
	CurrentPrepayID       *string
	CurrentPrepayExpireAt *time.Time
}

// OrderAccessor 支付服务对订单的读写接口（由 main 适配 order.Service）。
type OrderAccessor interface {
	// FindByOrderNo 按订单号查找。
	FindByOrderNo(ctx context.Context, orderNo string) (*OrderSnapshot, error)
	// FindByID 按 ID 查找。
	FindByID(ctx context.Context, id int64) (*OrderSnapshot, error)
	// SetPrepayID 更新 current_prepay_id 和过期时间。
	SetPrepayID(ctx context.Context, orderID int64, prepayID string, expireAt time.Time) error
	// Transition 触发订单状态机变更。
	Transition(ctx context.Context, orderID int64, trigger, opType string, opID int64, reason string) error
	// DB 返回 gorm.DB（用于事务）。
	DB() *gorm.DB
}

// UserOpenidGetter 按 scene 获取用户 openid 的接口。
type UserOpenidGetter interface {
	GetOpenid(ctx context.Context, userID int64, scene string) (string, error)
}

// AsynqEnqueuer asynq 任务入队接口（便于 mock）。
type AsynqEnqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// WxPaySceneConfig 各 scene 对应的 AppID 配置。
type WxPaySceneConfig struct {
	AppIDMP string
	AppIDOA string
	AppIDH5 string
	MchID   string
}

// ---- Service ----

// Service 支付服务。
type Service struct {
	repo       PaymentRepo
	orderDB    OrderAccessor
	userOpenid UserOpenidGetter
	wxpayClient wxpay.Client
	enqueuer   AsynqEnqueuer
	sceneCfg   WxPaySceneConfig
}

// NewService 构造支付服务。
func NewService(
	repo PaymentRepo,
	orderDB OrderAccessor,
	userOpenid UserOpenidGetter,
	wxpayClient wxpay.Client,
	enqueuer AsynqEnqueuer,
	sceneCfg WxPaySceneConfig,
) *Service {
	return &Service{
		repo:        repo,
		orderDB:     orderDB,
		userOpenid:  userOpenid,
		wxpayClient: wxpayClient,
		enqueuer:    enqueuer,
		sceneCfg:    sceneCfg,
	}
}

// ---- Prepay ----

// PrepayResult 预支付返回给前端的参数。
type PrepayResult struct {
	wxpay.PrepayResp
	PaymentID types.Int64Str `json:"payment_id"`
}

// Prepay 发起预支付。
func (s *Service) Prepay(ctx context.Context, userID, orderID int64, scene, clientIP string) (*PrepayResult, error) {
	// 1. 加载订单
	o, err := s.orderDB.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound.WithMsg("订单不存在")
		}
		return nil, errs.ErrInternal
	}
	if o.UserID != userID {
		return nil, errs.ErrForbidden
	}
	if o.Status != "pending" {
		return nil, errs.ErrParam.WithMsg(fmt.Sprintf("订单状态 %s 不可发起支付", o.Status))
	}
	if time.Now().After(o.ExpireAt) {
		return nil, errs.ErrParam.WithMsg("订单已过期")
	}

	// 2. scene/appid/openid 三元组校验
	var openID string
	switch scene {
	case "jsapi_mp", "jsapi_oa":
		openID, err = s.userOpenid.GetOpenid(ctx, userID, scene)
		if err != nil {
			return nil, errs.ErrInternal
		}
		if openID == "" {
			return nil, errs.ErrOpenIDMissing
		}
	case "h5":
		// h5 不需要 openid，但需要 clientIP
		if clientIP == "" {
			clientIP = "127.0.0.1"
		}
	default:
		return nil, errs.ErrParam.WithMsg("不支持的支付场景: " + scene)
	}

	// 3. 复用判断：current_prepay_id 非空 && expire > now+5min → 重用
	if o.CurrentPrepayID != nil && o.CurrentPrepayExpireAt != nil &&
		o.CurrentPrepayExpireAt.After(time.Now().Add(5*time.Minute)) {
		// 已有有效 prepay_id，重新计算签名返回
		resp, buildErr := s.rebuildPrepayResp(ctx, scene, *o.CurrentPrepayID)
		if buildErr == nil {
			return &PrepayResult{PrepayResp: *resp}, nil
		}
		// 重建失败则继续走新建流程
	}

	// 4. 调用微信支付预下单
	expire := time.Now().Add(14 * time.Minute) // 留 1min 余量
	wxReq := wxpay.PrepayReq{
		Scene:    scene,
		OrderNo:  o.OrderNo,
		PayCents: o.PayCents,
		OpenID:   openID,
		ClientIP: clientIP,
		Expire:   expire,
	}
	wxResp, err := s.wxpayClient.Prepay(ctx, wxReq)
	if err != nil {
		logger.Ctx(ctx).Error("wxpay prepay failed", zap.Error(err))
		return nil, errs.ErrInternal.WithMsg("预支付失败: " + err.Error())
	}

	// 5. 写 payment 行
	appID := s.sceneAppID(scene)
	prepayID := extractPrepayID(wxResp)
	pmt := &Payment{
		ID:          snowflake.NextID(),
		OrderID:     orderID,
		Channel:     "wxpay",
		TradeType:   sceneTradeType(scene),
		AmountCents: o.PayCents,
		Status:      PayStatusPending,
	}
	if appID != "" {
		pmt.AppID = &appID
	}
	if prepayID != "" {
		pmt.PrepayID = &prepayID
	}
	if err := s.repo.CreatePayment(ctx, pmt); err != nil {
		logger.Ctx(ctx).Error("create payment record failed", zap.Error(err))
	}

	// 6. 更新订单 current_prepay_id/expire_at
	if prepayID != "" {
		if err := s.orderDB.SetPrepayID(ctx, orderID, prepayID, expire); err != nil {
			logger.Ctx(ctx).Warn("set prepay_id on order failed", zap.Error(err))
		}
	}

	return &PrepayResult{PrepayResp: *wxResp, PaymentID: types.Int64Str(pmt.ID)}, nil
}

// rebuildPrepayResp 基于已有 prepay_id 重建签名（调用 wxpay 接口重签）。
func (s *Service) rebuildPrepayResp(ctx context.Context, scene, prepayID string) (*wxpay.PrepayResp, error) {
	// 创建一个 dummy 请求以触发签名重建
	req := wxpay.PrepayReq{
		Scene:    scene,
		OrderNo:  "rebuild_" + prepayID,
		PayCents: 1,
		Expire:   time.Now().Add(10 * time.Minute),
	}
	// 注意：此处调用 Prepay 会产生新的预下单请求，正式实现应使用 SDK 的签名方法
	// 作为降级处理，直接返回 error 让上层走重新预下单逻辑
	_ = ctx
	_ = req
	return nil, fmt.Errorf("rebuild not supported, create new prepay")
}

// ---- WxPay Notify ----

// HandleWxpayNotify 处理微信支付回调。
func (s *Service) HandleWxpayNotify(ctx context.Context, body []byte, headers map[string]string) error {
	// 1. 校验签名 + 解密
	result, err := s.wxpayClient.VerifyNotify(ctx, body, headers)
	if err != nil {
		return errs.ErrParam.WithMsg("回调签名校验失败")
	}

	// 2. 加载订单（SELECT FOR UPDATE）
	var orderID int64
	var orderPayCents int64
	var orderStatus string

	err = s.orderDB.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var o struct {
			ID       int64  `gorm:"column:id"`
			PayCents int64  `gorm:"column:pay_cents"`
			Status   string `gorm:"column:status"`
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Table(`"order"`).
			Select("id, pay_cents, status").
			Where("order_no = ?", result.OutTradeNo).
			First(&o).Error; err != nil {
			return err
		}
		orderID = o.ID
		orderPayCents = o.PayCents
		orderStatus = o.Status
		return nil
	})
	if err != nil {
		logger.Ctx(ctx).Error("wxpay notify: find order failed",
			zap.String("order_no", result.OutTradeNo), zap.Error(err))
		return errs.ErrInternal
	}

	// 3. 幂等：UpsertByTxn
	pmt, isNew, err := s.repo.UpsertByTxn(ctx, result.TransactionID, result.OutTradeNo, orderID, result.AmtCents, result.Raw)
	if err != nil {
		logger.Ctx(ctx).Error("wxpay notify: upsert payment failed", zap.Error(err))
		return errs.ErrInternal
	}
	if !isNew {
		// 已处理，幂等返回
		logger.Ctx(ctx).Info("wxpay notify: duplicate, skip", zap.String("txn_id", result.TransactionID))
		return nil
	}

	// 4. 金额校验
	if result.AmtCents != orderPayCents {
		paidAt := time.Now()
		if result.PaidAt != nil {
			paidAt = *result.PaidAt
		}
		_ = s.repo.MarkSuccess(ctx, pmt.ID, result.Raw, paidAt)
		_ = s.repo.MarkOrphan(ctx, pmt.ID)

		// 自动入队退款任务
		payload, _ := json.Marshal(map[string]any{
			"payment_id":  pmt.ID,
			"order_id":    orderID,
			"order_no":    result.OutTradeNo,
			"amount":      result.AmtCents,
			"total_cents": result.AmtCents,
			"reason":      "支付金额与订单金额不一致，自动原路退款",
		})
		if s.enqueuer != nil {
			task := asynq.NewTask(TaskPaymentAutoRefund, payload)
			if _, enqErr := s.enqueuer.EnqueueContext(ctx, task, asynq.MaxRetry(3)); enqErr != nil {
				logger.Ctx(ctx).Error("enqueue auto_refund failed", zap.Error(enqErr))
			}
		}
		logger.Ctx(ctx).Warn("wxpay notify: amount mismatch, auto refund enqueued",
			zap.String("order_no", result.OutTradeNo),
			zap.Int64("expect", orderPayCents), zap.Int64("actual", result.AmtCents))
		return nil
	}

	// 5. 状态分支
	if orderStatus != "pending" {
		logger.Ctx(ctx).Info("wxpay notify: order not pending, skip transition",
			zap.String("status", orderStatus))
		return nil
	}

	paidAt := time.Now()
	if result.PaidAt != nil {
		paidAt = *result.PaidAt
	}
	if err := s.repo.MarkSuccess(ctx, pmt.ID, result.Raw, paidAt); err != nil {
		logger.Ctx(ctx).Error("mark payment success failed", zap.Error(err))
	}

	if err := s.orderDB.Transition(ctx, orderID, "pay_success", "system", 0, "微信支付成功"); err != nil {
		logger.Ctx(ctx).Error("order transition pay_success failed",
			zap.Int64("order_id", orderID), zap.Error(err))
		return errs.ErrInternal
	}

	logger.Ctx(ctx).Info("wxpay notify: success", zap.String("txn_id", result.TransactionID))
	return nil
}

// ---- Refund Notify ----

// HandleWxpayRefundNotify 处理微信退款回调。
func (s *Service) HandleWxpayRefundNotify(ctx context.Context, body []byte, headers map[string]string) error {
	result, err := s.wxpayClient.VerifyRefundNotify(ctx, body, headers)
	if err != nil {
		return errs.ErrParam.WithMsg("退款回调校验失败")
	}

	ref, err := s.repo.FindRefundByNo(ctx, result.OutRefundNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Ctx(ctx).Warn("refund notify: refund record not found",
				zap.String("refund_no", result.OutRefundNo))
			return nil
		}
		return errs.ErrInternal
	}

	if ref.Status != RefundStatusPending {
		return nil // 幂等
	}

	now := time.Now()
	switch result.Status {
	case "SUCCESS":
		if err := s.repo.MarkRefundSuccess(ctx, ref.ID, result.Raw, now); err != nil {
			return errs.ErrInternal
		}
		// 更新订单状态为已退款
		if err := s.orderDB.Transition(ctx, ref.OrderID, "refund_success", "system", 0, "微信退款成功"); err != nil {
			logger.Ctx(ctx).Warn("order refund_success transition failed", zap.Error(err))
		}
	default:
		if err := s.repo.MarkRefundFailed(ctx, ref.ID, result.Raw); err != nil {
			return errs.ErrInternal
		}
		if err := s.orderDB.Transition(ctx, ref.OrderID, "refund_failed", "system", 0, "微信退款失败"); err != nil {
			logger.Ctx(ctx).Warn("order refund_failed transition failed", zap.Error(err))
		}
	}

	return nil
}

// ---- ApplyRefund ----

// ApplyRefund 申请退款（支持部分退款）。
func (s *Service) ApplyRefund(ctx context.Context, orderID, adminID int64, amtCents int64, reason string) error {
	// 1. 查找成功支付记录
	pmt, err := s.repo.FindSuccessByOrder(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("未找到成功的支付记录")
		}
		return errs.ErrInternal
	}

	// 2. 累计已退款金额
	sumRefunded, err := s.repo.SumRefundSuccess(ctx, orderID)
	if err != nil {
		return errs.ErrInternal
	}

	// 3. 校验：累计 + 本次 ≤ 实付
	if sumRefunded+amtCents > pmt.AmountCents {
		return errs.ErrParam.WithMsg(fmt.Sprintf(
			"退款金额超限：已退 %d 分，本次申请 %d 分，实付 %d 分",
			sumRefunded, amtCents, pmt.AmountCents,
		))
	}

	// 4. 创建退款记录
	refundNo := genRefundNo()
	reasonPtr := &reason
	ref := &Refund{
		OrderID:     orderID,
		PaymentID:   pmt.ID,
		RefundNo:    refundNo,
		AmountCents: amtCents,
		Reason:      reasonPtr,
		Status:      RefundStatusPending,
		OperatorID:  &adminID,
	}
	if err := s.repo.CreateRefund(ctx, ref); err != nil {
		return errs.ErrInternal
	}

	// 5. 获取订单号
	orderSnap, err := s.orderDB.FindByID(ctx, orderID)
	if err != nil {
		return errs.ErrInternal
	}

	// 6. 调用微信退款接口
	wxReq := wxpay.RefundReq{
		OrderNo:    orderSnap.OrderNo,
		RefundNo:   refundNo,
		AmtCents:   amtCents,
		TotalCents: pmt.AmountCents,
		Reason:     reason,
	}
	if err := s.wxpayClient.Refund(ctx, wxReq); err != nil {
		logger.Ctx(ctx).Error("wxpay refund failed", zap.Error(err))
		// 退款接口失败不回滚 refund 行（等待人工处理或重试）
		return errs.ErrInternal.WithMsg("微信退款接口失败: " + err.Error())
	}

	// 7. 触发订单状态 refund_apply
	if err := s.orderDB.Transition(ctx, orderID, "refund_apply", "admin", adminID, reason); err != nil {
		logger.Ctx(ctx).Warn("order refund_apply transition failed", zap.Error(err))
	}

	return nil
}

// ---- QueryPayStatus ----

// QueryPayStatus C 端查询支付状态。
func (s *Service) QueryPayStatus(ctx context.Context, orderID, userID int64) (*PayStatusResp, error) {
	o, err := s.orderDB.FindByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	if o.UserID != userID {
		return nil, errs.ErrForbidden
	}

	pmt, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &PayStatusResp{Status: "none"}, nil
		}
		return nil, errs.ErrInternal
	}

	return &PayStatusResp{
		Status:    pmt.Status,
		PaidAt:    pmt.PaidAt,
		AmtCents:  pmt.AmountCents,
		TradeType: pmt.TradeType,
	}, nil
}

// ---- 列表接口 ----

// ListPayments 后台支付列表。
func (s *Service) ListPayments(ctx context.Context, filter PaymentFilter) ([]Payment, int64, error) {
	list, total, err := s.repo.ListPayments(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	return list, total, nil
}

// ListRefunds 后台退款列表。
func (s *Service) ListRefunds(ctx context.Context, orderID int64) ([]Refund, error) {
	list, err := s.repo.ListRefunds(ctx, orderID)
	if err != nil {
		return nil, errs.ErrInternal
	}
	return list, nil
}

// ListDiffs 后台对账差异列表。
func (s *Service) ListDiffs(ctx context.Context, filter DiffFilter) ([]ReconciliationDiff, int64, error) {
	list, total, err := s.repo.ListDiffs(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	return list, total, nil
}

// ListAuditLogs 查询操作审计日志列表。
func (s *Service) ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error) {
	list, total, err := s.repo.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	return list, total, nil
}

// ResolveDiff 标记对账差异已处理。
func (s *Service) ResolveDiff(ctx context.Context, id, adminID int64) error {
	if err := s.repo.ResolveDiff(ctx, id, adminID); err != nil {
		return errs.ErrInternal
	}
	return nil
}

// ---- 内部辅助 ----

const TaskPaymentAutoRefund = "payment:auto_refund"

func (s *Service) sceneAppID(scene string) string {
	switch scene {
	case "jsapi_mp":
		return s.sceneCfg.AppIDMP
	case "jsapi_oa":
		return s.sceneCfg.AppIDOA
	case "h5":
		return s.sceneCfg.AppIDH5
	default:
		return ""
	}
}

func sceneTradeType(scene string) string {
	switch scene {
	case "jsapi_mp", "jsapi_oa":
		return "JSAPI"
	case "h5":
		return "MWEB"
	default:
		return "JSAPI"
	}
}

func extractPrepayID(resp *wxpay.PrepayResp) string {
	if resp == nil {
		return ""
	}
	pkg := resp.Package
	if len(pkg) > 9 && pkg[:9] == "prepay_id" {
		return pkg[10:]
	}
	return ""
}

func genRefundNo() string {
	return fmt.Sprintf("R%d", snowflake.NextID())
}

