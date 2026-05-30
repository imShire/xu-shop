package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/jobs"
	"github.com/xushop/xu-shop/internal/modules/account"
	"github.com/xushop/xu-shop/internal/modules/marketing"
	mkcoupon "github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	mkmember "github.com/xushop/xu-shop/internal/modules/marketing/member"
	mkpoint "github.com/xushop/xu-shop/internal/modules/marketing/point"
	"github.com/xushop/xu-shop/internal/modules/notification"
	pkglogger "github.com/xushop/xu-shop/internal/pkg/logger"
)

// wireMarketingJobs 把 marketing 8 个 worker job 注册到 mux + scheduler。
//
// 8 个 task：
//   - coupon:expire-scan      cron 0 2  * * *
//   - coupon:expire-warning   cron 0 10 * * *
//   - coupon:birthday-cron    cron 0 9  * * *
//   - point:earn-grant        cron 0 4  * * *
//   - point:expire-scan       cron 30 2 * * *
//   - point:expire-warning    cron 0 10 * * *
//   - point:rollback          事件触发，无 cron
//   - member:level-recompute  cron 0 3  * * *
func wireMarketingJobs(
	mux *asynq.ServeMux,
	scheduler *asynq.Scheduler,
	mkSvc *marketing.Service,
	notifSvc *notification.Service,
	userRepo account.UserRepo,
	asynqClient *asynq.Client,
	rdb *redis.Client,
	db *gorm.DB,
) {
	disp := &marketingNotifyDispatcher{notif: notifSvc, users: userRepo, enqueuer: asynqClient}
	birthdaySrc := &birthdayUserSource{db: db}
	birthdayResolver := &birthdayTplResolver{db: db}
	earnSrc := &pendingEarnOrderSource{db: db}
	memberSrc := &memberUserSource{db: db}
	memberNotifier := &memberLevelChangeNotifier{disp: disp}
	setNX := func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
		return rdb.SetNX(ctx, key, "1", ttl).Result()
	}

	mux.Handle(jobs.TaskCouponExpireScan, jobs.NewCouponExpireScanHandler(mkSvc.Coupon))
	mux.Handle(jobs.TaskCouponExpireWarning, jobs.NewCouponExpireWarningHandler(mkSvc.Coupon, disp))
	mux.Handle(jobs.TaskCouponBirthdayCron, jobs.NewCouponBirthdayCronHandler(mkSvc.Coupon, birthdaySrc, birthdayResolver, setNX))
	mux.Handle(jobs.TaskPointEarnGrant, jobs.NewPointEarnGrantHandler(mkSvc.Point, earnSrc))
	mux.Handle(jobs.TaskPointExpireScan, jobs.NewPointExpireScanHandler(mkSvc.Point))
	mux.Handle(jobs.TaskPointExpireWarning, jobs.NewPointExpireWarningHandler(mkSvc.Point, disp))
	mux.Handle(jobs.TaskPointRollback, jobs.NewPointRollbackHandler(mkSvc.Point))
	mux.Handle(jobs.TaskMemberLevelRecompute, jobs.NewMemberLevelRecomputeHandler(mkSvc.Member, memberSrc, memberNotifier))

	type cronSpec struct {
		spec string
		name string
	}
	specs := []cronSpec{
		{"0 2 * * *", jobs.TaskCouponExpireScan},
		{"0 10 * * *", jobs.TaskCouponExpireWarning},
		{"0 9 * * *", jobs.TaskCouponBirthdayCron},
		{"0 4 * * *", jobs.TaskPointEarnGrant},
		{"30 2 * * *", jobs.TaskPointExpireScan},
		{"0 10 * * *", jobs.TaskPointExpireWarning},
		{"0 3 * * *", jobs.TaskMemberLevelRecompute},
	}
	for _, s := range specs {
		if _, err := scheduler.Register(s.spec, asynq.NewTask(s.name, nil)); err != nil {
			pkglogger.L().Fatal("scheduler register marketing failed",
				zap.String("task", s.name), zap.Error(err))
		}
	}
	pkglogger.L().Info("marketing jobs wired",
		zap.Int("mux_handlers", 8), zap.Int("scheduler_entries", len(specs)))
}

// ---- 通知分发适配器 ----

// marketingNotifyDispatcher 同时实现 mkcoupon / mkpoint 的 NotificationDispatcher。
type marketingNotifyDispatcher struct {
	notif    *notification.Service
	users    account.UserRepo
	enqueuer *asynq.Client
}

// Dispatch 查 openid → notif.Dispatch（写库 + dedup）→ 入队 notification:send。
func (d *marketingNotifyDispatcher) Dispatch(ctx context.Context, eventType string, userID int64, refID string, params map[string]any) error {
	u, err := d.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		pkglogger.L().Debug("marketing notify: user not found",
			zap.Int64("user_id", userID), zap.String("event", eventType))
		return nil
	}
	target := ""
	if u.OpenidMP != nil && *u.OpenidMP != "" {
		target = *u.OpenidMP
	} else if u.OpenidH5 != nil && *u.OpenidH5 != "" {
		target = *u.OpenidH5
	}
	if target == "" {
		pkglogger.L().Debug("marketing notify: no openid",
			zap.Int64("user_id", userID), zap.String("event", eventType))
		return nil
	}
	taskID, err := d.notif.Dispatch(ctx, notification.Event{
		Type:       eventType,
		UserID:     userID,
		Target:     target,
		TargetType: notification.TargetTypeUser,
		RefID:      refID,
		Params:     params,
	})
	if err != nil {
		pkglogger.L().Warn("marketing notify dispatch failed",
			zap.String("event", eventType), zap.Int64("user_id", userID), zap.Error(err))
		return err
	}
	if taskID == 0 {
		// 模板未启用 / cooldown / 已存在 → 不再入队
		return nil
	}
	payload, _ := json.Marshal(jobs.NotificationSendPayload{TaskID: taskID})
	if _, err := d.enqueuer.EnqueueContext(ctx, asynq.NewTask(jobs.TaskNotificationSend, payload)); err != nil {
		pkglogger.L().Warn("marketing notify enqueue failed",
			zap.Int64("task_id", taskID), zap.Error(err))
	}
	return nil
}

var (
	_ mkcoupon.NotificationDispatcher = (*marketingNotifyDispatcher)(nil)
	_ mkpoint.NotificationDispatcher  = (*marketingNotifyDispatcher)(nil)
)

// ---- 生日用户源 ----

type birthdayUserSource struct{ db *gorm.DB }

func (s *birthdayUserSource) ListTodayBirthday(ctx context.Context, lastID int64, limit int) ([]mkcoupon.BirthdayUser, error) {
	now := time.Now()
	month, day := int(now.Month()), now.Day()
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, COALESCE(member_level_code, 'normal') AS level_code
		FROM "user"
		WHERE status = 'active'
		  AND birthday IS NOT NULL
		  AND EXTRACT(MONTH FROM birthday) = ?
		  AND EXTRACT(DAY   FROM birthday) = ?
		  AND id > ?
		ORDER BY id ASC
		LIMIT ?`, month, day, lastID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []mkcoupon.BirthdayUser
	for rows.Next() {
		var u mkcoupon.BirthdayUser
		if err := rows.Scan(&u.UserID, &u.LevelCode); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

var _ mkcoupon.BirthdayUserSource = (*birthdayUserSource)(nil)

// ---- 生日券模板解析器 ----

type birthdayTplResolver struct{ db *gorm.DB }

func (r *birthdayTplResolver) BirthdayCouponTplFor(ctx context.Context, levelCode string) (int64, error) {
	if levelCode == "" {
		levelCode = "normal"
	}
	var tplID *int64
	err := r.db.WithContext(ctx).
		Raw(`SELECT birthday_coupon_tpl_id FROM member_level WHERE code = ? AND enabled = true`, levelCode).
		Scan(&tplID).Error
	if err != nil {
		return 0, err
	}
	if tplID == nil {
		return 0, nil
	}
	return *tplID, nil
}

var _ mkcoupon.BirthdayTemplateResolver = (*birthdayTplResolver)(nil)

// ---- 兜底入账订单源 ----

type pendingEarnOrderSource struct{ db *gorm.DB }

func (s *pendingEarnOrderSource) ListPendingEarnOrders(ctx context.Context, lastID int64, limit int) ([]mkpoint.PendingEarnOrder, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT o.id, o.user_id, o.pay_cents, COALESCE(o.completed_at, o.updated_at)
		FROM "order" o
		WHERE o.status = 'completed'
		  AND o.id > ?
		  AND NOT EXISTS (
		    SELECT 1 FROM point_transaction pt
		    WHERE pt.ref_type = 'order' AND pt.ref_id = o.id AND pt.type = 'earn'
		  )
		ORDER BY o.id ASC
		LIMIT ?`, lastID, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []mkpoint.PendingEarnOrder
	for rows.Next() {
		var p mkpoint.PendingEarnOrder
		if err := rows.Scan(&p.OrderID, &p.UserID, &p.PayCents, &p.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

var _ mkpoint.OrderEarnSource = (*pendingEarnOrderSource)(nil)

// ---- 会员用户迭代器 ----

type memberUserSource struct{ db *gorm.DB }

func (s *memberUserSource) ListAllUserIDs(ctx context.Context, lastID int64, batchSize int) ([]int64, error) {
	var ids []int64
	err := s.db.WithContext(ctx).Raw(
		`SELECT id FROM "user" WHERE status = 'active' AND id > ? ORDER BY id ASC LIMIT ?`,
		lastID, batchSize,
	).Scan(&ids).Error
	return ids, err
}

var _ mkmember.UserSource = (*memberUserSource)(nil)

// ---- 会员等级变更通知 ----

type memberLevelChangeNotifier struct{ disp *marketingNotifyDispatcher }

func (n *memberLevelChangeNotifier) OnLevelChanged(ctx context.Context, userID int64, oldCode, newCode string) {
	refID := fmt.Sprintf("level_change:%d:%s:%s", userID, oldCode, newCode)
	_ = n.disp.Dispatch(ctx, "member_level_changed", userID, refID, map[string]any{
		"user_id":    userID,
		"old_code":   oldCode,
		"new_code":   newCode,
		"changed_at": time.Now().Format("2006-01-02 15:04"),
	})
}

var _ mkmember.LevelChangeNotifier = (*memberLevelChangeNotifier)(nil)
