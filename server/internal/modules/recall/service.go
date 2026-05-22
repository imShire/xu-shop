// Package recall 召回活动服务。
package recall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	coupon "github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	"github.com/xushop/xu-shop/internal/modules/notification"
	"github.com/xushop/xu-shop/internal/modules/tag"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

// Service 召回活动服务。
type Service struct {
	repo      Repo
	db        *gorm.DB
	rdb       *redis.Client
	tagSvc    *tag.Service
	couponSvc *coupon.Service
	notifSvc  *notification.Service
}

// NewService 构造。tagSvc 为必需；couponSvc / notifSvc 可为 nil（测试场景）。
func NewService(repo Repo, db *gorm.DB, rdb *redis.Client, tagSvc *tag.Service, couponSvc *coupon.Service, notifSvc *notification.Service) *Service {
	return &Service{repo: repo, db: db, rdb: rdb, tagSvc: tagSvc, couponSvc: couponSvc, notifSvc: notifSvc}
}

// ===== CRUD =====

// ListCampaigns 列活动。
func (s *Service) ListCampaigns(ctx context.Context, status string, page, size int) ([]CampaignResp, int64, error) {
	list, total, err := s.repo.ListCampaigns(ctx, status, page, size)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]CampaignResp, 0, len(list))
	for i := range list {
		resp = append(resp, toCampaignResp(&list[i]))
	}
	return resp, total, nil
}

// GetCampaign 详情。
func (s *Service) GetCampaign(ctx context.Context, id int64) (CampaignResp, error) {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CampaignResp{}, errs.ErrNotFound
		}
		return CampaignResp{}, err
	}
	return toCampaignResp(c), nil
}

// CreateCampaign 创建（草稿态）。
func (s *Service) CreateCampaign(ctx context.Context, form CampaignForm, adminID int64) (CampaignResp, error) {
	c := &RecallCampaign{
		ID:                    snowflake.NextID(),
		Name:                  form.Name,
		Goal:                  form.Goal,
		AudienceFilter:        ensureMap(form.AudienceFilter),
		Actions:               ensureArr(form.Actions),
		TriggerType:           form.TriggerType,
		TriggerConfig:         ensureMap(form.TriggerConfig),
		EffectiveFrom:         form.EffectiveFrom,
		EffectiveTo:           form.EffectiveTo,
		ThrottlePerUserDays:   form.ThrottlePerUserDays,
		DailyQuota:            form.DailyQuota,
		TotalQuota:            form.TotalQuota,
		AttributionWindowDays: form.AttributionWindowDays,
		Status:                StatusDraft,
		CreatedBy:             adminID,
	}
	if err := s.repo.CreateCampaign(ctx, c); err != nil {
		return CampaignResp{}, err
	}
	return toCampaignResp(c), nil
}

// UpdateCampaign 修改（仅 draft 可改）。
func (s *Service) UpdateCampaign(ctx context.Context, id int64, form CampaignForm) error {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if c.Status != StatusDraft {
		return errs.ErrConflict.WithMsg("仅草稿态可修改")
	}
	fields := map[string]any{
		"name":                    form.Name,
		"goal":                    form.Goal,
		"audience_filter":         ensureMap(form.AudienceFilter),
		"actions":                 ensureArr(form.Actions),
		"trigger_type":            form.TriggerType,
		"trigger_config":          ensureMap(form.TriggerConfig),
		"effective_from":          form.EffectiveFrom,
		"effective_to":            form.EffectiveTo,
		"throttle_per_user_days":  form.ThrottlePerUserDays,
		"daily_quota":             form.DailyQuota,
		"total_quota":             form.TotalQuota,
		"attribution_window_days": form.AttributionWindowDays,
		"updated_at":              time.Now(),
	}
	return s.repo.UpdateCampaign(ctx, id, fields)
}

// Transition 状态机：draft→online / online↔paused / *→closed。
func (s *Service) Transition(ctx context.Context, id int64, to string, approverAdminID int64) error {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	allow := map[string]map[string]bool{
		StatusDraft:  {StatusOnline: true, StatusClosed: true},
		StatusOnline: {StatusPaused: true, StatusClosed: true},
		StatusPaused: {StatusOnline: true, StatusClosed: true},
	}
	if !allow[c.Status][to] {
		return errs.ErrConflict.WithMsg(fmt.Sprintf("不允许 %s → %s", c.Status, to))
	}
	fields := map[string]any{"status": to, "updated_at": time.Now()}
	if to == StatusOnline && c.ApproverAdminID == nil {
		fields["approver_admin_id"] = approverAdminID
	}
	return s.repo.UpdateCampaign(ctx, id, fields)
}

// ===== 漏斗 =====

// FunnelReport 单活动漏斗。
func (s *Service) FunnelReport(ctx context.Context, campaignID int64) (FunnelResp, error) {
	c, err := s.repo.GetCampaign(ctx, campaignID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FunnelResp{}, errs.ErrNotFound
		}
		return FunnelResp{}, err
	}
	t, o, cv, gmv, err := s.repo.FunnelStats(ctx, campaignID)
	if err != nil {
		return FunnelResp{}, err
	}
	resp := FunnelResp{Triggered: t, Opened: o, Converted: cv, GMVCents: gmv}
	resp.CampaignID = toCampaignResp(c).ID
	if t > 0 {
		resp.OpenRate = float64(o) / float64(t)
		resp.ConvertRate = float64(cv) / float64(t)
	}
	return resp, nil
}

// ListLogs 单活动触达日志。
func (s *Service) ListLogs(ctx context.Context, campaignID int64, page, size int) ([]RecallLog, int64, error) {
	return s.repo.ListLogs(ctx, campaignID, page, size)
}

// ===== 触发器 =====

// ScheduleScan 调度扫描：cron 类型活动定时触发。
func (s *Service) ScheduleScan(ctx context.Context) error {
	list, err := s.repo.ListOnlineByTrigger(ctx, TriggerCron)
	if err != nil {
		return err
	}
	for i := range list {
		c := list[i]
		if !s.cronShouldFire(&c) {
			continue
		}
		if err := s.runCampaign(ctx, &c); err != nil {
			// 单活动失败不阻断
			continue
		}
	}
	return nil
}

// OnEvent 事件触发：匹配 event 类型的在线活动并执行。
//
// targetUserID > 0 时只对单用户执行（如 cart_abandoned 事件）。
func (s *Service) OnEvent(ctx context.Context, eventName string, targetUserID int64) error {
	list, err := s.repo.ListOnlineByTrigger(ctx, TriggerEvent)
	if err != nil {
		return err
	}
	for i := range list {
		c := list[i]
		if cfg, _ := c.TriggerConfig["event"].(string); cfg != eventName {
			continue
		}
		if targetUserID > 0 {
			if err := s.executeForUser(ctx, &c, targetUserID); err != nil {
				continue
			}
		} else {
			if err := s.runCampaign(ctx, &c); err != nil {
				continue
			}
		}
	}
	return nil
}

// OnOrderPaid 订单支付成功 → 归因最近一次召回触达。
func (s *Service) OnOrderPaid(ctx context.Context, userID, orderID int64, paidAt time.Time, gmvCents int64) error {
	if userID <= 0 || orderID <= 0 {
		return nil
	}
	_, err := s.repo.AttributeOrder(ctx, userID, orderID, paidAt, gmvCents, 7)
	return err
}

// runCampaign 完整人群拉取 + 执行。
func (s *Service) runCampaign(ctx context.Context, c *RecallCampaign) error {
	filter, err := decodeAudienceFilter(c.AudienceFilter)
	if err != nil {
		return err
	}
	if s.tagSvc == nil {
		return errs.ErrInternal.WithMsg("tag service not initialized")
	}
	const batch = 500
	var lastID int64
	for {
		ids, err := s.tagSvc.ListAudience(ctx, filter, lastID, batch)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		for _, uid := range ids {
			_ = s.executeForUser(ctx, c, uid)
		}
		lastID = ids[len(ids)-1]
		if len(ids) < batch {
			return nil
		}
	}
}

// executeForUser 针对单用户执行召回（节流 + 配额 + 动作）。
func (s *Service) executeForUser(ctx context.Context, c *RecallCampaign, userID int64) error {
	// 节流 1：每用户每活动节流（默认按天，可配 throttle_per_user_days）
	throttleDays := c.ThrottlePerUserDays
	if throttleDays <= 0 {
		throttleDays = 1
	}
	throttleKey := fmt.Sprintf("recall:throttle:%d:%d", c.ID, userID)
	if s.rdb != nil {
		if ok, err := s.rdb.SetNX(ctx, throttleKey, "1", time.Duration(throttleDays)*24*time.Hour).Result(); err != nil {
			return err
		} else if !ok {
			return nil // 节流命中
		}
	} else {
		// 兜底：DB 查当日是否已发
		hit, err := s.repo.HasLogToday(ctx, c.ID, userID)
		if err != nil {
			return err
		}
		if hit {
			return nil
		}
	}

	// 节流 2：日配额
	if c.DailyQuota > 0 {
		dailyKey := fmt.Sprintf("recall:campaign_daily:%d:%s", c.ID, time.Now().Format("20060102"))
		if s.rdb != nil {
			n, err := s.rdb.Incr(ctx, dailyKey).Result()
			if err == nil {
				_ = s.rdb.Expire(ctx, dailyKey, 36*time.Hour).Err()
				if n > c.DailyQuota {
					return nil
				}
			}
		}
	}
	// 节流 3：总配额（粗略，依赖 DB）
	if c.TotalQuota > 0 {
		total, err := s.repo.CountLogTotal(ctx, c.ID)
		if err == nil && total >= c.TotalQuota {
			return nil
		}
	}

	// 执行动作
	results := JSONArray{}
	for _, raw := range c.Actions {
		actionMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		actionType, _ := actionMap["type"].(string)
		switch actionType {
		case ActionGrantCoupon:
			tid := parseInt64(actionMap["template_id"])
			if tid > 0 && s.couponSvc != nil {
				_, err := s.couponSvc.Claim(ctx, userID, tid, "recall", map[string]any{"campaign_id": c.ID})
				results = append(results, map[string]any{"type": actionType, "template_id": tid, "ok": err == nil})
			}
		case ActionWxSubscribe, ActionInbox:
			if s.notifSvc != nil {
				eventType, _ := actionMap["event_type"].(string)
				params, _ := actionMap["params"].(map[string]any)
				_, err := s.notifSvc.Dispatch(ctx, notification.Event{
					Type:   eventType,
					UserID: userID,
					RefID:  fmt.Sprintf("recall:%d", c.ID),
					Params: params,
				})
				results = append(results, map[string]any{"type": actionType, "ok": err == nil})
			}
		}
	}

	// 写日志
	log := &RecallLog{
		ID:               snowflake.NextID(),
		CampaignID:       c.ID,
		UserID:           userID,
		TriggeredAt:      time.Now(),
		AudienceSnapshot: JSONMap{"filter_hash": fingerprintAudience(c.AudienceFilter)},
		ActionsResult:    results,
	}
	return s.repo.InsertLog(ctx, log)
}

// cronShouldFire 极简 cron：trigger_config.cron_minutes 为整数（0 = 每次扫描都跑）。
func (s *Service) cronShouldFire(c *RecallCampaign) bool {
	// 默认每次扫描都触发；细粒度调度交由 asynq 任务节奏控制。
	return true
}

// ===== 辅助 =====

func ensureMap(m JSONMap) JSONMap {
	if m == nil {
		return JSONMap{}
	}
	return m
}

func ensureArr(a JSONArray) JSONArray {
	if a == nil {
		return JSONArray{}
	}
	return a
}

func decodeAudienceFilter(m JSONMap) (tag.AudienceFilter, error) {
	if len(m) == 0 {
		return tag.AudienceFilter{}, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return tag.AudienceFilter{}, err
	}
	var f tag.AudienceFilter
	if err := json.Unmarshal(b, &f); err != nil {
		return tag.AudienceFilter{}, err
	}
	return f, nil
}

func parseInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	case string:
		var n int64
		_, err := fmt.Sscanf(x, "%d", &n)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

func fingerprintAudience(m JSONMap) string {
	b, _ := json.Marshal(m)
	return fmt.Sprintf("%x", len(b))
}
