package distribution

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
)

// ===== 错误 =====

var (
	ErrDistributorNotActive = errs.New(70001, "您不是分销员或已被停用", 403)
	ErrDistributorExists    = errs.New(70002, "您已申请分销员，无需重复提交", 409)
	ErrInvalidScene         = errs.New(70003, "分享场景非法", 400)
	ErrShareLinkExpired     = errs.New(70004, "分享链接已过期", 410)
	ErrSelfPurchase         = errs.New(70005, "禁止自购返佣", 400)
	ErrSuspectFraud         = errs.New(70006, "本单触发反作弊规则", 403)
	ErrInsufficientLocked   = errs.New(70007, "可提现余额不足", 400)
	ErrSmsCodeInvalid       = errs.New(70008, "短信验证码错误", 400)
	ErrWithdrawTooSmall     = errs.New(70009, "单笔提现金额不得小于 10 元", 400)
	ErrWithdrawDuplicate    = errs.New(70010, "同一申请已在处理中", 409)
	ErrCommissionNotSuspect = errs.New(70011, "佣金状态不匹配", 409)
	ErrInvalidTransition    = errs.New(70012, "状态变更不被允许", 409)
	ErrAlreadyHasInviter    = errs.New(70013, "已存在邀请关系", 409)
)

// ===== 常量 =====

const (
	defaultBaseRate          = 0.05
	relationDefaultDays      = 90 // 邀请关系有效期 90 天
	commissionFreezeDays     = 7
	withdrawMinCents         = 1000 // 单笔最小 10 元
	smsCodeTTL               = 5 * time.Minute
	smsCodeDailyLimit        = 5
	smsCodeMaxFailures       = 5
	smsLockDuration          = time.Hour
	defaultShortTokenLen     = 8
	defaultAttributionWindow = 7
)

// ===== 依赖接口 =====

// UserOpenidGetter 取分销员收款 openid。
type UserOpenidGetter interface {
	GetOpenidForTransfer(ctx context.Context, userID int64) (openid, appID string, err error)
}

// SmsSender 仅用于发送提现验证码（mock 友好）。
type SmsSender interface {
	SendWithdrawSms(ctx context.Context, userID int64, code string) error
}

// PasswordVerifier 校验业务密码（可选，传 nil 时仅走 sms）。
type PasswordVerifier interface {
	Verify(ctx context.Context, userID int64, password string) error
}

// Service 分销模块业务服务。
type Service struct {
	repo            Repo
	rdb             *redis.Client
	wxpay           pkgwxpay.TransferClient
	openid          UserOpenidGetter
	sms             SmsSender
	pwd             PasswordVerifier
	transferAppID   string
	transferScene   string
	transferNotify  string
	shareBaseURL    string
}

// Config 服务配置。
type Config struct {
	TransferAppID  string // 转账接收侧 appid（公众号 / 小程序）
	TransferScene  string // 商家转账场景 id
	TransferNotify string // 商家转账回调地址
	ShareBaseURL   string // 短链前缀（含 scheme+host）
}

// NewService 构造服务。
func NewService(repo Repo, rdb *redis.Client, wxpay pkgwxpay.TransferClient, openid UserOpenidGetter, sms SmsSender, pwd PasswordVerifier, cfg Config) *Service {
	return &Service{
		repo:           repo,
		rdb:            rdb,
		wxpay:          wxpay,
		openid:         openid,
		sms:            sms,
		pwd:            pwd,
		transferAppID:  cfg.TransferAppID,
		transferScene:  cfg.TransferScene,
		transferNotify: cfg.TransferNotify,
		shareBaseURL:   cfg.ShareBaseURL,
	}
}

func snowID() int64 { return snowflake.NextID() }

// ===== 分享 / 溯源 =====

// CreateShareLink 创建分享链接。
//
// 任何登录用户可创建（非分销员也允许，分享 inv_register 场景）。
// 短链 token 6-10 位 base62，30 天默认 TTL，唯一冲突自动重试。
func (s *Service) CreateShareLink(ctx context.Context, userID int64, req CreateShareLinkReq) (*ShareLink, error) {
	if userID <= 0 {
		return nil, errs.ErrUnauth
	}
	switch req.Scene {
	case ShareSceneProduct, ShareSceneActivity, ShareSceneBrand, ShareSceneInviteRegister:
	default:
		return nil, ErrInvalidScene
	}
	ttl := req.TTLDays
	if ttl <= 0 {
		ttl = 30
	}
	if req.ChannelCode == "" {
		req.ChannelCode = "other"
	}
	link := &ShareLink{
		ID:          snowID(),
		UserID:      userID,
		Scene:       req.Scene,
		TargetID:    req.TargetID,
		ChannelCode: req.ChannelCode,
		ExpireAt:    time.Now().Add(time.Duration(ttl) * 24 * time.Hour),
		CreatedAt:   time.Now(),
	}
	// 短链冲突重试
	for i := 0; i < 5; i++ {
		token, err := genShortToken(defaultShortTokenLen)
		if err != nil {
			return nil, errs.ErrInternal.WithMsg("rand: " + err.Error())
		}
		link.ShortToken = token
		if err := s.repo.CreateShareLink(ctx, link); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		return link, nil
	}
	return nil, errs.ErrInternal.WithMsg("short_token collision")
}

// ResolveShortToken 解析短链 -> ShareLink，可触发一次点击记录。
func (s *Service) ResolveShortToken(ctx context.Context, token, traceID, ua, ip, device, referer, fingerprint string) (*ShareLink, error) {
	if len(token) < 4 || len(token) > 16 {
		return nil, errs.ErrNotFound
	}
	l, err := s.repo.GetShareLinkByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	if time.Now().After(l.ExpireAt) {
		return nil, ErrShareLinkExpired
	}
	if traceID == "" {
		traceID = newTraceID()
	}
	_ = s.TrackClick(ctx, l, traceID, ua, ip, device, referer, fingerprint)
	return l, nil
}

// TrackClick 记录一次点击：写 share_click + upsert share_attribution + 自增 click_count。
func (s *Service) TrackClick(ctx context.Context, l *ShareLink, traceID, ua, ip, device, referer, fingerprint string) error {
	click := &ShareClick{
		ID:          snowID(),
		TraceID:     traceID,
		ShareLinkID: l.ID,
		TS:          time.Now(),
	}
	if ua != "" {
		click.UA = &ua
	}
	if ip != "" {
		click.IP = &ip
	}
	if device != "" {
		click.Device = &device
	}
	if referer != "" {
		click.Referer = &referer
	}
	if fingerprint != "" {
		click.VisitorFingerprint = &fingerprint
	}
	if err := s.repo.CreateShareClick(ctx, click); err != nil {
		return err
	}
	_ = s.repo.IncShareLinkCounter(ctx, l.ID, "click_count", 1)
	// upsert attribution（仅首次触达写入 share_link_id）
	now := time.Now()
	att := &ShareAttribution{
		ID:                    snowID(),
		ShareLinkID:           l.ID,
		TraceID:               traceID,
		FirstTouchTS:          now,
		LastTouchTS:           now,
		AttributionWindowDays: defaultAttributionWindow,
	}
	_ = s.repo.UpsertAttribution(ctx, att)
	return nil
}

// OnUserRegister 新用户注册时：若 trace_id 命中分享归因，写入 distributor_relation（90 天有效）。
//
// 老用户不绑定（依赖 traceID 仅出现在新设备首登场景）。
// 已存在邀请关系时直接跳过（首次绑定为准）。
func (s *Service) OnUserRegister(ctx context.Context, userID int64, traceID string) error {
	if userID <= 0 || traceID == "" {
		return nil
	}
	att, err := s.repo.GetAttributionByTrace(ctx, traceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	link, err := s.repo.GetShareLinkByID(ctx, att.ShareLinkID)
	if err != nil {
		return nil
	}
	if link.UserID == userID {
		// 自己分享自己注册，忽略
		return nil
	}
	if existing, _ := s.repo.GetRelationByInvitee(ctx, userID); existing != nil {
		return nil // 老关系优先
	}
	rel := &DistributorRelation{
		ID:            snowID(),
		InviteeUserID: userID,
		InviterUserID: link.UserID,
		ShareLinkID:   link.ID,
		BoundAt:       time.Now(),
		ExpireAt:      time.Now().Add(time.Duration(relationDefaultDays) * 24 * time.Hour),
		LastRenewedAt: time.Now(),
	}
	if err := s.repo.CreateRelation(ctx, nil, rel); err != nil {
		if isUniqueViolation(err) {
			return nil
		}
		return err
	}
	_ = s.repo.BindAttributionUser(ctx, traceID, userID)
	_ = s.repo.IncShareLinkCounter(ctx, link.ID, "register_count", 1)
	return nil
}

// ResolveTraceForOrder 订单创建时：根据 (userID, traceID) 解析归属分销员。
//
// 规则：
//  1. 查 distributor_relation：在有效期内 → 返回 inviterUserID（同时续签 90 天）。
//  2. 否则若 traceID 命中 attribution 且 link.user 是激活分销员且为新用户 → 写 relation。
//  3. 否则不归因。
func (s *Service) ResolveTraceForOrder(ctx context.Context, userID int64, traceID string) (distributorUserID int64, shareLinkID int64) {
	if userID <= 0 {
		return 0, 0
	}
	rel, err := s.repo.GetRelationByInvitee(ctx, userID)
	if err == nil {
		if time.Now().Before(rel.ExpireAt) {
			// 续签
			_ = s.repo.RenewRelation(ctx, userID, time.Now().Add(time.Duration(relationDefaultDays)*24*time.Hour))
			return rel.InviterUserID, rel.ShareLinkID
		}
		// 已过期，不归因
		return 0, 0
	}
	if traceID == "" {
		return 0, 0
	}
	att, err := s.repo.GetAttributionByTrace(ctx, traceID)
	if err != nil {
		return 0, 0
	}
	link, err := s.repo.GetShareLinkByID(ctx, att.ShareLinkID)
	if err != nil {
		return 0, 0
	}
	if link.UserID == userID {
		return 0, 0
	}
	d, err := s.repo.GetDistributorByUserID(ctx, link.UserID)
	if err != nil || d.Status != DistStatusActive {
		return 0, 0
	}
	// 写 relation（首次绑定）
	newRel := &DistributorRelation{
		ID:            snowID(),
		InviteeUserID: userID,
		InviterUserID: link.UserID,
		ShareLinkID:   link.ID,
		BoundAt:       time.Now(),
		ExpireAt:      time.Now().Add(time.Duration(relationDefaultDays) * 24 * time.Hour),
		LastRenewedAt: time.Now(),
	}
	if err := s.repo.CreateRelation(ctx, nil, newRel); err != nil && !isUniqueViolation(err) {
		return 0, 0
	}
	return link.UserID, link.ID
}

// ===== 分销员 =====

// Apply 申请成为分销员（pending）。
func (s *Service) Apply(ctx context.Context, userID int64, _ ApplyDistributorReq) (*Distributor, error) {
	if existing, err := s.repo.GetDistributorByUserID(ctx, userID); err == nil && existing != nil {
		return nil, ErrDistributorExists
	}
	d := &Distributor{
		ID:        snowID(),
		UserID:    userID,
		Level:     DistLevelNormal,
		Rate:      defaultBaseRate,
		Status:    DistStatusPending,
		ApplyAt:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.CreateDistributor(ctx, d); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDistributorExists
		}
		return nil, err
	}
	return d, nil
}

// Approve 审核通过。
func (s *Service) Approve(ctx context.Context, distributorID, adminID int64) error {
	d, err := s.repo.GetDistributorByID(ctx, distributorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if d.Status != DistStatusPending {
		return ErrInvalidTransition
	}
	now := time.Now()
	return s.repo.UpdateDistributor(ctx, distributorID, map[string]any{
		"status":            DistStatusActive,
		"approved_at":       now,
		"approver_admin_id": adminID,
	})
}

// Reject 拒绝申请。
func (s *Service) Reject(ctx context.Context, distributorID int64, reason string) error {
	d, err := s.repo.GetDistributorByID(ctx, distributorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if d.Status != DistStatusPending {
		return ErrInvalidTransition
	}
	now := time.Now()
	return s.repo.UpdateDistributor(ctx, distributorID, map[string]any{
		"status":           DistStatusDisabled,
		"suspended_at":     now,
		"suspended_reason": reason,
	})
}

// Ban 停用分销员。
func (s *Service) Ban(ctx context.Context, distributorID int64, reason string) error {
	d, err := s.repo.GetDistributorByID(ctx, distributorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if d.Status == DistStatusDisabled {
		return nil
	}
	now := time.Now()
	return s.repo.UpdateDistributor(ctx, distributorID, map[string]any{
		"status":           DistStatusDisabled,
		"suspended_at":     now,
		"suspended_reason": reason,
	})
}

// AdjustLevel 调整等级。
func (s *Service) AdjustLevel(ctx context.Context, distributorID int64, level string) error {
	if level != DistLevelNormal && level != DistLevelSenior {
		return errs.ErrParam
	}
	return s.repo.UpdateDistributor(ctx, distributorID, map[string]any{"level": level})
}

// AdjustRate 调整专属费率（nil 表示清空 override）。
func (s *Service) AdjustRate(ctx context.Context, distributorID int64, rate *float64) error {
	if rate != nil && (*rate <= 0 || *rate > 1) {
		return errs.ErrParam.WithMsg("rate must be (0,1]")
	}
	return s.repo.UpdateDistributor(ctx, distributorID, map[string]any{"rate_override": rate})
}

// GetMyProfile 我的分销员资料 + 余额。
func (s *Service) GetMyProfile(ctx context.Context, userID int64) (*MyProfileResp, error) {
	d, err := s.repo.GetDistributorByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	pending, locked, settled, err := s.repo.SumByUserStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	withdrawn, err := s.repo.SumWithdrawnByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &MyProfileResp{
		PendingCents:     pending,
		LockedCents:      locked,
		SettledCents:     settled,
		TotalEarnedCents: pending + locked + settled,
		WithdrawnCents:   withdrawn,
	}
	if d != nil {
		r := toDistributorResp(d)
		resp.Distributor = &r
	}
	return resp, nil
}

// ListDistributors admin 列表。
func (s *Service) ListDistributors(ctx context.Context, status, level string, limit, offset int) ([]Distributor, int64, error) {
	return s.repo.ListDistributors(ctx, status, level, limit, offset)
}

// ===== 佣金 =====

// OrderInfo 用于注入订单信息（避免循环依赖 order 包）。
type OrderInfo struct {
	OrderID         int64
	UserID          int64 // 下单用户
	PayCents        int64
	DistributorUser int64 // order.distributor_user_id（已由订单创建时 ResolveTraceForOrder 注入）
	ShareLinkID     int64
}

// OnOrderPaid 订单支付成功时生成佣金记录（pending）。
//
// 红线：
//   - 自购拒绝
//   - 分销员未激活拒绝
//   - 反作弊：fingerprint 30 天 5 次、24h 同收件人 3 单（此处仅做基础检查，巡检 job 做大批量扫描）
//   - 幂等：(order_id, distributor_user_id) 唯一索引
func (s *Service) OnOrderPaid(ctx context.Context, info OrderInfo) error {
	if info.DistributorUser <= 0 || info.PayCents <= 0 {
		return nil
	}
	if info.DistributorUser == info.UserID {
		return ErrSelfPurchase
	}
	d, err := s.repo.GetDistributorByUserID(ctx, info.DistributorUser)
	if err != nil {
		return nil
	}
	if d.Status != DistStatusActive {
		return nil
	}
	rate := d.EffectiveRate()
	amount := int64(float64(info.PayCents) * rate)
	if amount <= 0 {
		return nil
	}
	rec := &CommissionRecord{
		ID:                snowID(),
		OrderID:           info.OrderID,
		DistributorUserID: info.DistributorUser,
		Level:             d.Level,
		Rate:              rate,
		BaseAmountCents:   info.PayCents,
		AmountCents:       amount,
		Status:            CommissionStatusPending,
		FreezeUntil:       time.Now().Add(time.Duration(commissionFreezeDays) * 24 * time.Hour),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.repo.CreateCommission(ctx, rec); err != nil {
		if isUniqueViolation(err) {
			return nil // 幂等
		}
		return err
	}
	if info.ShareLinkID > 0 {
		_ = s.repo.IncShareLinkCounter(ctx, info.ShareLinkID, "order_count", 1)
		_ = s.repo.IncShareLinkCounter(ctx, info.ShareLinkID, "gmv_cents", info.PayCents)
	}
	return nil
}

// OnOrderRefund 订单（部分/全额）退款时调整佣金。
//
//	部分退款 → 标记 partial_refund，按比例缩减
//	全额退款 → 作废
func (s *Service) OnOrderRefund(ctx context.Context, orderID int64, refundCents, payCents int64, fullRefund bool) error {
	if orderID <= 0 {
		return nil
	}
	var list []CommissionRecord
	if err := s.repo.DB().WithContext(ctx).Where("order_id = ?", orderID).Find(&list).Error; err != nil {
		return err
	}
	for i := range list {
		c := &list[i]
		if c.Status == CommissionStatusCanceled || c.Status == CommissionStatusSettled {
			continue
		}
		if fullRefund || refundCents >= payCents {
			_ = s.Transition(ctx, c, "order_full_refund", "全额退款")
			continue
		}
		// 部分退款：缩减金额（按退款占比）
		ratio := float64(refundCents) / float64(payCents)
		newAmount := c.AmountCents - int64(float64(c.AmountCents)*ratio)
		if newAmount < 0 {
			newAmount = 0
		}
		_ = s.repo.UpdateCommission(ctx, c.ID, map[string]any{
			"amount_cents":    newAmount,
			"canceled_reason": "partial_refund",
		})
	}
	return nil
}

// FreezeReleaseScan 巡检：把 freeze_until 已到期的 pending → locked。
//
// 由 asynq scheduler 每日触发。
func (s *Service) FreezeReleaseScan(ctx context.Context) (int, error) {
	list, err := s.repo.ListPendingExpired(ctx, time.Now(), 1000)
	if err != nil {
		return 0, err
	}
	n := 0
	for i := range list {
		c := &list[i]
		if err := s.Transition(ctx, c, "pass_freeze", "freeze_until 到期"); err == nil {
			n++
		}
	}
	return n, nil
}

// MyCommissions C 端：我的佣金列表。
func (s *Service) MyCommissions(ctx context.Context, userID int64, status string, limit, offset int) ([]CommissionRecord, int64, error) {
	return s.repo.ListCommissionsByUser(ctx, userID, status, limit, offset)
}

// AdminListCommissions admin 列表。
func (s *Service) AdminListCommissions(ctx context.Context, status string, distributorUserID int64, limit, offset int) ([]CommissionRecord, int64, error) {
	return s.repo.ListCommissionsAdmin(ctx, status, distributorUserID, limit, offset)
}

// AdminAuditCommission admin 审核：release 解除 suspect / cancel 作废。
func (s *Service) AdminAuditCommission(ctx context.Context, commissionID int64, req AuditCommissionReq) error {
	c, err := s.repo.GetCommission(ctx, commissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	switch req.Action {
	case "release":
		if c.Status != CommissionStatusSuspect {
			return ErrCommissionNotSuspect
		}
		return s.Transition(ctx, c, "unsuspect", req.Reason)
	case "cancel":
		return s.Transition(ctx, c, "manual_cancel", req.Reason)
	}
	return errs.ErrParam
}

// Transition 佣金状态机：唯一变更入口。
//
// 事件支持：mark_suspect / unsuspect / pass_freeze / partial_refund / order_full_refund / manual_cancel / withdraw / withdraw_failed
func (s *Service) Transition(ctx context.Context, c *CommissionRecord, event, reason string) error {
	prev := c.Status
	next := prev
	fields := map[string]any{}
	switch event {
	case "mark_suspect":
		if prev != CommissionStatusPending {
			return ErrInvalidTransition
		}
		next = CommissionStatusSuspect
		fields["suspect_reason"] = reason
	case "unsuspect":
		if prev != CommissionStatusSuspect {
			return ErrInvalidTransition
		}
		next = CommissionStatusPending
		var nilStr *string
		fields["suspect_reason"] = nilStr
	case "pass_freeze":
		if prev != CommissionStatusPending {
			return ErrInvalidTransition
		}
		next = CommissionStatusLocked
	case "withdraw":
		if prev != CommissionStatusLocked {
			return ErrInvalidTransition
		}
		next = CommissionStatusSettled
		now := time.Now()
		fields["settled_at"] = now
	case "withdraw_failed":
		if prev != CommissionStatusSettled {
			return ErrInvalidTransition
		}
		next = CommissionStatusLocked
		var nilT *time.Time
		fields["settled_at"] = nilT
	case "order_full_refund", "manual_cancel":
		if prev == CommissionStatusSettled {
			return ErrInvalidTransition
		}
		next = CommissionStatusCanceled
		fields["canceled_reason"] = reason
	default:
		return ErrInvalidTransition
	}
	fields["status"] = next
	if err := s.repo.UpdateCommission(ctx, c.ID, fields); err != nil {
		return err
	}
	c.Status = next
	return nil
}

// ===== 提现 =====

// WithdrawSmsRequest 发送验证码，记 Redis（5min TTL，5/天）。
func (s *Service) WithdrawSmsRequest(ctx context.Context, userID int64) (string, error) {
	if s.rdb == nil {
		return "", errs.ErrServiceDegraded
	}
	dailyKey := fmt.Sprintf("dist:wd:sms:cnt:%d:%s", userID, time.Now().Format("20060102"))
	cnt, _ := s.rdb.Incr(ctx, dailyKey).Result()
	_ = s.rdb.Expire(ctx, dailyKey, 24*time.Hour).Err()
	if cnt > smsCodeDailyLimit {
		return "", errs.ErrRateLimit
	}
	code := genSmsCode()
	codeKey := fmt.Sprintf("dist:wd:sms:code:%d", userID)
	_ = s.rdb.Set(ctx, codeKey, code, smsCodeTTL).Err()
	if s.sms != nil {
		_ = s.sms.SendWithdrawSms(ctx, userID, code)
	}
	return code, nil
}

// RequestWithdraw 申请提现。
//
// 二次校验：sms_code 必传或 password 必传（取决于业务密码是否启用）。
func (s *Service) RequestWithdraw(ctx context.Context, userID int64, idemKey string, req WithdrawReq) (*WithdrawOrder, error) {
	if req.AmountCents < withdrawMinCents {
		return nil, ErrWithdrawTooSmall
	}
	// 二次校验：至少一种凭据通过
	verified := false
	if req.SmsCode != "" && s.rdb != nil {
		codeKey := fmt.Sprintf("dist:wd:sms:code:%d", userID)
		stored, _ := s.rdb.Get(ctx, codeKey).Result()
		if stored != "" && stored == req.SmsCode {
			_ = s.rdb.Del(ctx, codeKey).Err()
			verified = true
		} else {
			failKey := fmt.Sprintf("dist:wd:sms:fail:%d", userID)
			fails, _ := s.rdb.Incr(ctx, failKey).Result()
			_ = s.rdb.Expire(ctx, failKey, smsLockDuration).Err()
			if fails >= smsCodeMaxFailures {
				return nil, errs.ErrRateLimit.WithMsg("已锁定，请 1 小时后重试")
			}
			return nil, ErrSmsCodeInvalid
		}
	}
	if !verified && req.Password != "" && s.pwd != nil {
		if err := s.pwd.Verify(ctx, userID, req.Password); err == nil {
			verified = true
		}
	}
	if !verified {
		return nil, ErrSmsCodeInvalid
	}

	// 分销员校验
	d, err := s.repo.GetDistributorByUserID(ctx, userID)
	if err != nil || d == nil || d.Status != DistStatusActive {
		return nil, ErrDistributorNotActive
	}

	// 余额校验
	_, lockedCents, _, err := s.repo.SumByUserStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.AmountCents > lockedCents {
		return nil, ErrInsufficientLocked
	}

	// 幂等：相同 idemKey 直接复用
	if idemKey != "" {
		if existing, _ := s.repo.GetWithdrawByIdem(ctx, idemKey); existing != nil {
			return existing, nil
		}
	} else {
		idemKey = fmt.Sprintf("auto:%d:%d:%d", userID, req.AmountCents, time.Now().Unix())
	}

	// 选取覆盖金额的 locked 佣金（FIFO）
	lockedList, err := s.repo.ListLockedForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	picked := make([]int64, 0, len(lockedList))
	covered := int64(0)
	for i := range lockedList {
		picked = append(picked, lockedList[i].ID)
		covered += lockedList[i].AmountCents
		if covered >= req.AmountCents {
			break
		}
	}
	if covered < req.AmountCents {
		return nil, ErrInsufficientLocked
	}

	withdrawNo := genWithdrawNo()
	w := &WithdrawOrder{
		ID:                snowID(),
		DistributorUserID: userID,
		WithdrawNo:        withdrawNo,
		AmountCents:       req.AmountCents,
		Channel:           "wx_transfer",
		Status:            WithdrawStatusPending,
		AppliedAt:         time.Now(),
		IdemKey:           idemKey,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.repo.CreateWithdraw(ctx, w); err != nil {
		if isUniqueViolation(err) {
			if existing, _ := s.repo.GetWithdrawByIdem(ctx, idemKey); existing != nil {
				return existing, nil
			}
			return nil, ErrWithdrawDuplicate
		}
		return nil, err
	}

	// 创建结算单 + 绑定佣金 → settled
	recordsJSON, _ := json.Marshal(picked)
	settlement := &CommissionSettlement{
		ID:                 snowID(),
		DistributorUserID:  userID,
		PeriodYYYYMM:       nil,
		RequestAmountCents: req.AmountCents,
		Records:            string(recordsJSON),
		WithdrawOrderID:    &w.ID,
		Status:             SettlementStatusProcessing,
		Channel:            "wx_transfer",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	_ = s.repo.CreateSettlement(ctx, settlement)
	_ = s.repo.BindCommissionsToSettlement(ctx, picked, settlement.ID)

	// 调起微信商家转账
	if err := s.dispatchTransfer(ctx, w, userID); err != nil {
		failReason := err.Error()
		_ = s.repo.UpdateWithdraw(ctx, w.ID, map[string]any{
			"status":      WithdrawStatusFailed,
			"fail_reason": failReason,
			"finished_at": time.Now(),
		})
		_ = s.repo.UpdateSettlement(ctx, settlement.ID, map[string]any{
			"status":      SettlementStatusFailed,
			"fail_reason": failReason,
		})
		// 回滚佣金
		s.rollbackPickedCommissions(ctx, picked)
		return nil, errs.ErrInternal.WithMsg("transfer: " + failReason)
	}
	return w, nil
}

func (s *Service) dispatchTransfer(ctx context.Context, w *WithdrawOrder, userID int64) error {
	if s.wxpay == nil {
		return errors.New("wxpay transfer client not configured")
	}
	openid, appID := "", s.transferAppID
	if s.openid != nil {
		oid, aid, err := s.openid.GetOpenidForTransfer(ctx, userID)
		if err != nil {
			return err
		}
		openid = oid
		if aid != "" {
			appID = aid
		}
	}
	if openid == "" {
		return errors.New("openid missing for transfer")
	}
	resp, err := s.wxpay.Transfer(ctx, pkgwxpay.TransferReq{
		OutBillNo:           w.WithdrawNo,
		AppID:               appID,
		OpenID:              openid,
		TransferSceneID:     s.transferScene,
		TransferAmountCents: w.AmountCents,
		TransferRemark:      "分销佣金提现",
		NotifyURL:           s.transferNotify,
		UserRecvPerception:  "劳务报酬",
	})
	if err != nil {
		return err
	}
	updates := map[string]any{
		"status":            WithdrawStatusProcessing,
		"wx_transfer_no":    resp.TransferBillNo,
		"wx_transfer_state": resp.State,
		"processed_at":      time.Now(),
	}
	return s.repo.UpdateWithdraw(ctx, w.ID, updates)
}

func (s *Service) rollbackPickedCommissions(ctx context.Context, ids []int64) {
	for _, id := range ids {
		_ = s.repo.UpdateCommission(ctx, id, map[string]any{
			"status":        CommissionStatusLocked,
			"settlement_id": nil,
			"settled_at":    nil,
		})
	}
}

// OnTransferNotify 商家转账回调处理。
func (s *Service) OnTransferNotify(ctx context.Context, n *pkgwxpay.TransferNotifyResult) error {
	if n == nil || n.OutBillNo == "" {
		return errs.ErrParam
	}
	w, err := s.repo.GetWithdrawByOutBillNo(ctx, n.OutBillNo)
	if err != nil {
		return errs.ErrNotFound
	}
	// 终态幂等
	if w.Status == WithdrawStatusSuccess || w.Status == WithdrawStatusCanceled {
		return nil
	}
	if pkgwxpay.IsTransferTerminal(n.State) {
		switch n.State {
		case pkgwxpay.TransferStateSuccess:
			_ = s.repo.UpdateWithdraw(ctx, w.ID, map[string]any{
				"status":            WithdrawStatusSuccess,
				"wx_transfer_state": n.State,
				"finished_at":       time.Now(),
			})
		default:
			fail := n.FailReason
			_ = s.repo.UpdateWithdraw(ctx, w.ID, map[string]any{
				"status":            WithdrawStatusFailed,
				"wx_transfer_state": n.State,
				"fail_reason":       fail,
				"finished_at":       time.Now(),
			})
			s.releaseWithdrawCommissions(ctx, w.ID)
		}
	} else {
		_ = s.repo.UpdateWithdraw(ctx, w.ID, map[string]any{
			"wx_transfer_state": n.State,
		})
	}
	return nil
}

// MyWithdraws C 端：我的提现工单。
func (s *Service) MyWithdraws(ctx context.Context, userID int64, limit, offset int) ([]WithdrawOrder, int64, error) {
	return s.repo.ListWithdrawByUser(ctx, userID, limit, offset)
}

// AdminListWithdraws admin。
func (s *Service) AdminListWithdraws(ctx context.Context, status string, limit, offset int) ([]WithdrawOrder, int64, error) {
	return s.repo.ListWithdrawAdmin(ctx, status, limit, offset)
}

// WithdrawTransition 提现状态机（admin 重试 / 系统对账）。
func (s *Service) WithdrawTransition(ctx context.Context, w *WithdrawOrder, event string) error {
	prev := w.Status
	next := prev
	switch event {
	case "retry":
		if prev != WithdrawStatusFailed {
			return ErrInvalidTransition
		}
		// 重新调起转账
		if err := s.dispatchTransfer(ctx, w, w.DistributorUserID); err != nil {
			return errs.ErrInternal.WithMsg("retry transfer: " + err.Error())
		}
		return nil
	case "cancel":
		if prev != WithdrawStatusPending && prev != WithdrawStatusFailed {
			return ErrInvalidTransition
		}
		next = WithdrawStatusCanceled
	default:
		return ErrInvalidTransition
	}
	return s.repo.UpdateWithdraw(ctx, w.ID, map[string]any{"status": next, "finished_at": time.Now()})
}

func (s *Service) releaseWithdrawCommissions(ctx context.Context, withdrawID int64) {
	// 把绑定到此 settlement 的佣金从 settled 回滚到 locked
	var sets []CommissionSettlement
	_ = s.repo.DB().WithContext(ctx).Where("withdraw_order_id = ?", withdrawID).Find(&sets).Error
	for _, st := range sets {
		var ids []int64
		_ = json.Unmarshal([]byte(st.Records), &ids)
		s.rollbackPickedCommissions(ctx, ids)
		_ = s.repo.UpdateSettlement(ctx, st.ID, map[string]any{
			"status": SettlementStatusFailed,
		})
	}
}

// AdminListSettlements admin 结算列表。
func (s *Service) AdminListSettlements(ctx context.Context, status string, limit, offset int) ([]CommissionSettlement, int64, error) {
	return s.repo.ListSettlements(ctx, status, limit, offset)
}

// MonthlySettlementJob 月结 cron：把所有 locked 佣金按用户汇总入 settlement（仅生成报表，不强制提现）。
func (s *Service) MonthlySettlementJob(ctx context.Context) error {
	// MVP：仅占位，实际策略由运营手动结算
	_ = ctx
	return nil
}

// ListShareLinks admin。
func (s *Service) ListShareLinks(ctx context.Context, userID int64, scene string, limit int) ([]ShareLink, error) {
	return s.repo.ListShareLinks(ctx, userID, scene, limit)
}

// FunnelReport 漏斗：分享 → 点击 → 注册 → 下单 → GMV → 佣金。
func (s *Service) FunnelReport(ctx context.Context, start, end time.Time) (*FunnelReportResp, error) {
	if end.IsZero() {
		end = time.Now()
	}
	if start.IsZero() {
		start = end.AddDate(0, 0, -30)
	}
	links, _ := s.repo.CountShareLinks(ctx, start, end)
	clicks, _ := s.repo.CountShareClicks(ctx, start, end)
	regs, _ := s.repo.CountAttributionRegisters(ctx, start, end)
	orders, gmv, _ := s.repo.SumDistributionGMV(ctx, start, end)
	pending, locked, settled, _ := s.repo.SumCommissionsByStatus(ctx, start, end)
	return &FunnelReportResp{
		StartDate:         start.Format("2006-01-02"),
		EndDate:           end.Format("2006-01-02"),
		ShareLinks:        links,
		Clicks:            clicks,
		Registers:         regs,
		Orders:            orders,
		GMVCents:          gmv,
		CommissionPending: pending,
		CommissionLocked:  locked,
		CommissionSettled: settled,
	}, nil
}

// ShareBaseURL 暴露 base url 给 handler 拼短链。
func (s *Service) ShareBaseURL() string { return s.shareBaseURL }

// ===== 工具 =====

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func genShortToken(n int) (string, error) {
	if n <= 0 {
		n = defaultShortTokenLen
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = base62[int(buf[i])%len(base62)]
	}
	return string(out), nil
}

func genSmsCode() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// 兜底：弱随机也可（非密码学需求）
		return strconv.Itoa(100000 + mathrand.Intn(900000))
	}
	n := (int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])) & 0x7fffffff
	return fmt.Sprintf("%06d", n%1000000)
}

func genWithdrawNo() string {
	return "WD" + time.Now().Format("20060102150405") + strconv.FormatInt(snowID()%100000, 10)
}

func newTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return fmt.Sprintf("%x", b)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint")
}
