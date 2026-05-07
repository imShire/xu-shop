package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/wxsubscribe"
)

// ---- 事件类型 ----

const (
	EventOrderPaid      = "order_paid"
	EventOrderShipped   = "order_shipped"
	EventOrderDelivered = "order_delivered"
	EventRefundSuccess  = "refund_success"
	EventLowStockAlert  = "low_stock_alert"
)

// Event 通知事件。
type Event struct {
	Type   string
	UserID int64
	// Target 强制指定接收目标（openid 或 webhook URL），为空时从 UserID 查
	Target     string
	TargetType string
	RefID      string // 关联业务 ID（用于去重 key）
	Params     map[string]any
}

// UpdateTemplateReq 更新模板请求。
type UpdateTemplateReq struct {
	TemplateIDExternal string         `json:"template_id_external"`
	Fields             map[string]any `json:"fields"`
	Enabled            *bool          `json:"enabled"`
}

// ---- cooldown 配置 ----

var cooldownRules = map[int]time.Duration{
	43101: 7 * 24 * time.Hour,
	43004: 24 * time.Hour,
}

// Service 通知服务。
type Service struct {
	repo     NotificationRepo
	wxClient wxsubscribe.Client
	rdb      *redis.Client
	httpCli  *http.Client
}

// NewService 构造 Service。
func NewService(repo NotificationRepo, wxClient wxsubscribe.Client, rdb *redis.Client) *Service {
	return &Service{
		repo:     repo,
		wxClient: wxClient,
		rdb:      rdb,
		httpCli:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Dispatch 按事件分发消息（查模板 → 去重 → 入队）。
// 返回 taskID（已存在时为已有任务 ID），enqueuer 由 caller 负责（此处仅入库）。
func (s *Service) Dispatch(ctx context.Context, event Event) (int64, error) {
	tpl, err := s.repo.FindTemplate(ctx, event.Type)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errs.ErrNotFound.WithMsg("通知模板不存在: " + event.Type)
		}
		return 0, fmt.Errorf("notification: find template: %w", err)
	}
	if !tpl.Enabled {
		return 0, nil
	}

	target := event.Target
	targetType := event.TargetType
	if target == "" {
		return 0, fmt.Errorf("notification: target is required")
	}
	if targetType == "" {
		targetType = TargetTypeUser
	}

	// 检查 cooldown
	if targetType == TargetTypeUser && event.UserID > 0 {
		key := fmt.Sprintf("notif:cooldown:%d:%s", event.UserID, tpl.Code)
		if v, err := s.rdb.Exists(ctx, key).Result(); err == nil && v > 0 {
			return 0, nil
		}
	}

	dedupKey := ""
	if event.RefID != "" {
		dedupKey = fmt.Sprintf("%s:%s:%s", tpl.Code, target, event.RefID)
	}

	params := event.Params
	if params == nil {
		params = make(map[string]any)
	}

	task := &NotificationTask{
		ID:           snowflake.NextID(),
		TemplateCode: tpl.Code,
		TargetType:   targetType,
		Target:       target,
		Params:       JSONMap(params),
		Status:       TaskStatusPending,
	}

	t, created, err := s.repo.UpsertTaskByDedup(ctx, dedupKey, task)
	if err != nil {
		return 0, fmt.Errorf("notification: upsert task: %w", err)
	}
	if !created {
		return t.ID, nil
	}
	return t.ID, nil
}

// HandleSend worker 执行发送。
func (s *Service) HandleSend(ctx context.Context, taskID int64) error {
	task, err := s.repo.FindTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("notification: find task %d: %w", taskID, err)
	}

	if task.Status == TaskStatusSent || task.Status == TaskStatusSkipped {
		return nil
	}

	tpl, err := s.repo.FindTemplate(ctx, task.TemplateCode)
	if err != nil {
		return fmt.Errorf("notification: find template %s: %w", task.TemplateCode, err)
	}

	var sendErr error
	switch task.TargetType {
	case TargetTypeUser:
		sendErr = s.sendSubscribe(ctx, task, tpl)
	case TargetTypeWebhook:
		sendErr = s.sendWebhook(ctx, task)
	default:
		sendErr = fmt.Errorf("unknown target type: %s", task.TargetType)
	}

	if sendErr != nil {
		_ = s.repo.IncrRetry(ctx, taskID)
		_ = s.repo.MarkTaskFailed(ctx, taskID, sendErr.Error())

		// 处理微信 cooldown 错误码
		if wxErr, ok := sendErr.(*wxsubscribe.WxAPIError); ok {
			if ttl, exists := cooldownRules[wxErr.Code]; exists {
				key := fmt.Sprintf("notif:cooldown:%s:%s", task.Target, task.TemplateCode)
				_ = s.rdb.Set(ctx, key, "1", ttl).Err()
				_ = s.repo.MarkTaskFailed(ctx, taskID, fmt.Sprintf("cooldown: %s", sendErr.Error()))
			}
		}

		logger.L().Warn("notification: send failed",
			zap.Int64("task_id", taskID),
			zap.String("template", task.TemplateCode),
			zap.Error(sendErr))
		return sendErr
	}

	if err := s.repo.MarkTaskSuccess(ctx, taskID); err != nil {
		logger.L().Warn("notification: mark success failed", zap.Int64("task_id", taskID), zap.Error(err))
	}
	logger.L().Info("notification: send success", zap.Int64("task_id", taskID))
	return nil
}

func (s *Service) sendSubscribe(ctx context.Context, task *NotificationTask, tpl *NotificationTemplate) error {
	data := buildSubscribeData(task.Params, tpl.Fields)
	return s.wxClient.Send(ctx, wxsubscribe.SendReq{
		ToUser:     task.Target,
		TemplateID: tpl.TemplateIDExternal,
		Data:       data,
	})
}

// buildSubscribeData 将 task.Params 按 tpl.Fields 映射组装成微信订阅消息 data。
func buildSubscribeData(params JSONMap, fields JSONMap) map[string]any {
	data := make(map[string]any)
	for k, fieldKey := range fields {
		if fk, ok := fieldKey.(string); ok {
			if v, exists := params[fk]; exists {
				data[k] = map[string]any{"value": fmt.Sprintf("%v", v)}
			}
		}
	}
	return data
}

func (s *Service) sendWebhook(ctx context.Context, task *NotificationTask) error {
	paramsJSON, err := json.Marshal(task.Params)
	if err != nil {
		return fmt.Errorf("marshal webhook params: %w", err)
	}

	content := string(paramsJSON)
	if v, ok := task.Params["content"]; ok {
		content = fmt.Sprintf("%v", v)
	}

	payload := map[string]any{
		"msgtype":  "markdown",
		"markdown": map[string]string{"content": content},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, task.Target, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// ListNotifications 通知任务列表。
func (s *Service) ListNotifications(ctx context.Context, filter TaskFilter) ([]NotificationTaskResp, int64, error) {
	list, total, err := s.repo.ListTasks(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]NotificationTaskResp, len(list))
	for i := range list {
		resps[i] = toNotificationTaskResp(&list[i])
	}
	return resps, total, nil
}

// ListTemplates 通知模板列表。
func (s *Service) ListTemplates(ctx context.Context) ([]NotificationTemplateResp, error) {
	list, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	resps := make([]NotificationTemplateResp, len(list))
	for i := range list {
		resps[i] = toNotificationTemplateResp(&list[i])
	}
	return resps, nil
}

// UpdateTemplate 更新通知模板。
func (s *Service) UpdateTemplate(ctx context.Context, code string, req UpdateTemplateReq) error {
	tpl, err := s.repo.FindTemplate(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("模板不存在")
		}
		return fmt.Errorf("notification: find template: %w", err)
	}

	if req.TemplateIDExternal != "" {
		tpl.TemplateIDExternal = req.TemplateIDExternal
	}
	if req.Fields != nil {
		tpl.Fields = JSONMap(req.Fields)
	}
	if req.Enabled != nil {
		tpl.Enabled = *req.Enabled
	}
	return s.repo.UpdateTemplate(ctx, tpl)
}

// TestSend 测试发送（直接发送，不去重不记录）。
func (s *Service) TestSend(ctx context.Context, code, openid string) error {
	tpl, err := s.repo.FindTemplate(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("模板不存在")
		}
		return fmt.Errorf("notification: find template: %w", err)
	}

	testParams := JSONMap{"content": "【测试】这是一条测试通知消息"}
	return s.wxClient.Send(ctx, wxsubscribe.SendReq{
		ToUser:     openid,
		TemplateID: tpl.TemplateIDExternal,
		Data:       buildSubscribeData(testParams, tpl.Fields),
	})
}
