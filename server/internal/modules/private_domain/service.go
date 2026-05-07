package private_domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/logger"
	"github.com/xushop/xu-shop/internal/pkg/qywx"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	pkgtypes "github.com/xushop/xu-shop/internal/pkg/types"
)

// ---- 请求 DTO ----

// CreateChannelCodeReq 创建渠道码请求。
type CreateChannelCodeReq struct {
	Name            string                 `json:"name" binding:"required"`
	CustomerServers []string               `json:"customer_servers"`
	TagIDs          []pkgtypes.Int64Str    `json:"tag_ids"`
	WelcomeText     string                 `json:"welcome_text"`
	State           string                 `json:"state"`
}

// UpdateChannelCodeReq 更新渠道码请求。
type UpdateChannelCodeReq struct {
	Name        *string              `json:"name"`
	WelcomeText *string              `json:"welcome_text"`
	TagIDs      []pkgtypes.Int64Str  `json:"tag_ids"`
}

// CreateTagReq 创建标签请求。
type CreateTagReq struct {
	Name    string `json:"name" binding:"required"`
	Source  string `json:"source"`
	GroupID string `json:"group_id"`
}

// UpdateTagReq 更新标签请求。
type UpdateTagReq struct {
	Name string `json:"name" binding:"required"`
}

// Service 私域服务。
type Service struct {
	channelRepo ChannelCodeRepo
	tagRepo     TagRepo
	userTagRepo UserTagRepo
	shareRepo   ShareRepo
	qywxClient  qywx.Client
	rdb         *redis.Client
}

// NewService 构造 Service。
func NewService(
	channelRepo ChannelCodeRepo,
	tagRepo TagRepo,
	userTagRepo UserTagRepo,
	shareRepo ShareRepo,
	qywxClient qywx.Client,
	rdb *redis.Client,
) *Service {
	return &Service{
		channelRepo: channelRepo,
		tagRepo:     tagRepo,
		userTagRepo: userTagRepo,
		shareRepo:   shareRepo,
		qywxClient:  qywxClient,
		rdb:         rdb,
	}
}

// CreateChannelCode 创建渠道码（调企微 API）。
func (s *Service) CreateChannelCode(ctx context.Context, _ int64, req CreateChannelCodeReq) (*ChannelCodeResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errs.ErrParam.WithMsg("渠道码名称不能为空")
	}

	state := req.State
	if state == "" {
		state = strconv.FormatInt(snowflake.NextID(), 36)
	}

	resp, err := s.qywxClient.AddContactWay(ctx, qywx.AddContactWayReq{
		Type:       2,
		Scene:      1,
		Remark:     name,
		SkipVerify: true,
		State:      state,
		User:       req.CustomerServers,
	})
	if err != nil {
		logger.L().Warn("private_domain: add_contact_way failed", zap.Error(err))
	}

	cc := &ChannelCode{
		ID:              snowflake.NextID(),
		Name:            name,
		CustomerServers: JSONStrings(req.CustomerServers),
		TagIDs:          int64StrSliceToInt64s(req.TagIDs),
	}
	if req.WelcomeText != "" {
		cc.WelcomeText = &req.WelcomeText
	}
	if resp != nil {
		cc.QYWXConfigID = &resp.ConfigID
		if resp.QrCode != "" {
			cc.QRImageURL = &resp.QrCode
		}
	}

	if err := s.channelRepo.Create(ctx, cc); err != nil {
		if isUniqueViolation(err) {
			return nil, errs.ErrConflict.WithMsg("渠道码名称已存在")
		}
		return nil, fmt.Errorf("private_domain: create channel code: %w", err)
	}
	ccResp := toChannelCodeResp(cc)
	return &ccResp, nil
}

// UpdateChannelCode 更新渠道码。
func (s *Service) UpdateChannelCode(ctx context.Context, id int64, req UpdateChannelCodeReq) error {
	cc, err := s.channelRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound
		}
		return fmt.Errorf("private_domain: find channel code: %w", err)
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return errs.ErrParam.WithMsg("渠道码名称不能为空")
		}
		cc.Name = name
	}
	if req.WelcomeText != nil {
		cc.WelcomeText = req.WelcomeText
	}
	if req.TagIDs != nil {
		cc.TagIDs = int64StrSliceToInt64s(req.TagIDs)
	}
	if err := s.channelRepo.Update(ctx, cc); err != nil {
		if isUniqueViolation(err) {
			return errs.ErrConflict.WithMsg("渠道码名称已存在")
		}
		return fmt.Errorf("private_domain: update channel code: %w", err)
	}
	return nil
}

// DeleteChannelCode 删除渠道码。
func (s *Service) DeleteChannelCode(ctx context.Context, id int64) error {
	return s.channelRepo.Delete(ctx, id)
}

// ListChannelCodes 渠道码列表。
func (s *Service) ListChannelCodes(ctx context.Context, page, size int) ([]ChannelCodeResp, int64, error) {
	list, total, err := s.channelRepo.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]ChannelCodeResp, len(list))
	for i := range list {
		resps[i] = toChannelCodeResp(&list[i])
	}
	return resps, total, nil
}

// GetChannelCodeStats 渠道码统计（简单封装，stats 聚合由 stats 模块完成）。
func (s *Service) GetChannelCodeStats(ctx context.Context, id int64) (*ChannelCodeStatsResp, error) {
	cc, err := s.channelRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &ChannelCodeStatsResp{ChannelCodeResp: toChannelCodeResp(cc)}, nil
}

// CreateTag 创建客户标签（同步企微标签）。
func (s *Service) CreateTag(ctx context.Context, req CreateTagReq) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errs.ErrParam.WithMsg("标签名不能为空")
	}

	tagID, err := s.qywxClient.AddCorpTag(ctx, name, req.GroupID)
	if err != nil {
		logger.L().Warn("private_domain: add corp tag failed", zap.Error(err))
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	t := &CustomerTag{
		ID:     snowflake.NextID(),
		Name:   name,
		Source: source,
	}
	if tagID != "" {
		t.QYWXTagID = &tagID
	}
	if err := s.tagRepo.Create(ctx, t); err != nil {
		if isUniqueViolation(err) {
			return errs.ErrConflict.WithMsg("标签名称已存在")
		}
		return fmt.Errorf("private_domain: create tag: %w", err)
	}
	return nil
}

// UpdateTag 更新客户标签。
func (s *Service) UpdateTag(ctx context.Context, id int64, req UpdateTagReq) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errs.ErrParam.WithMsg("标签名不能为空")
	}

	tag, err := s.tagRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("标签不存在")
		}
		return fmt.Errorf("private_domain: find tag: %w", err)
	}

	tag.Name = name
	if err := s.tagRepo.Update(ctx, tag); err != nil {
		if isUniqueViolation(err) {
			return errs.ErrConflict.WithMsg("标签名称已存在")
		}
		return fmt.Errorf("private_domain: update tag: %w", err)
	}
	return nil
}

// DeleteTag 删除客户标签（同步企微）。
func (s *Service) DeleteTag(ctx context.Context, id int64) error {
	tag, err := s.tagRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("标签不存在")
		}
		return err
	}
	if tag.QYWXTagID != nil && *tag.QYWXTagID != "" {
		if delErr := s.qywxClient.DeleteCorpTag(ctx, *tag.QYWXTagID); delErr != nil {
			logger.L().Warn("private_domain: delete corp tag failed", zap.Error(delErr))
		}
	}
	return s.tagRepo.Delete(ctx, id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicated key")
}

// ListTags 标签列表。
func (s *Service) ListTags(ctx context.Context) ([]TagResp, error) {
	list, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	resps := make([]TagResp, len(list))
	for i := range list {
		resps[i] = toTagResp(&list[i])
	}
	return resps, nil
}

// AddUserTag 给用户打标签。
func (s *Service) AddUserTag(ctx context.Context, userID, tagID int64, source string) error {
	return s.userTagRepo.Add(ctx, &UserTag{
		UserID: userID,
		TagID:  tagID,
		Source: source,
	})
}

// RemoveUserTag 移除用户标签。
func (s *Service) RemoveUserTag(ctx context.Context, userID, tagID int64) error {
	return s.userTagRepo.Remove(ctx, userID, tagID)
}

// ListUserTags 用户标签列表。
func (s *Service) ListUserTags(ctx context.Context, userID int64) ([]TagResp, error) {
	list, err := s.userTagRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	resps := make([]TagResp, len(list))
	for i := range list {
		resps[i] = toTagResp(&list[i])
	}
	return resps, nil
}

// RecordShareVisit 记录分享访问（30 天内去重）。
func (s *Service) RecordShareVisit(ctx context.Context, shareUserID, viewerUserID, productID int64, channel string) error {
	since := time.Now().AddDate(0, 0, -30)
	exists, err := s.shareRepo.ExistsAttribution(ctx, shareUserID, viewerUserID, productID, since)
	if err != nil {
		return fmt.Errorf("private_domain: check attribution: %w", err)
	}
	if exists {
		return nil
	}

	pid := productID
	vid := viewerUserID
	a := &ShareAttribution{
		ID:           snowflake.NextID(),
		ShareUserID:  shareUserID,
		ViewerUserID: &vid,
		ProductID:    &pid,
		Channel:      channel,
	}
	return s.shareRepo.CreateAttribution(ctx, a)
}

// GeneratePoster UPSERT 短码 + 构造海报 URL（OSS 缓存由外部实现，此处返回可分享链接）。
func (s *Service) GeneratePoster(ctx context.Context, userID, productID int64) (string, error) {
	// OSS 缓存 key
	cacheKey := fmt.Sprintf("poster:%d:%d", userID, productID)
	if s.rdb != nil {
		if url, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && url != "" {
			return url, nil
		}
	}

	pid := productID
	sc := &ShareShortCode{
		ID:          snowflake.NextID(),
		ShareUserID: userID,
		ProductID:   &pid,
	}
	created, err := s.shareRepo.UpsertShortCode(ctx, sc)
	if err != nil {
		return "", fmt.Errorf("private_domain: upsert short code: %w", err)
	}

	// 用 base36 编码 ID 作为 scene 参数
	scene := strconv.FormatInt(created.ID, 36)
	posterURL := fmt.Sprintf("https://poster.placeholder/share/%s", scene)

	if s.rdb != nil {
		_ = s.rdb.Set(ctx, cacheKey, posterURL, 24*time.Hour).Err()
	}
	return posterURL, nil
}

// ResolveShareScene base36 解码 → 查 share_short_code。
func (s *Service) ResolveShareScene(ctx context.Context, scene string) (*ResolveResp, error) {
	id, err := strconv.ParseInt(scene, 36, 64)
	if err != nil {
		return nil, errs.ErrParam.WithMsg("无效的分享场景码")
	}

	sc, err := s.shareRepo.FindShortCode(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("private_domain: find short code: %w", err)
	}

	if sc.ExpireAt != nil && sc.ExpireAt.Before(time.Now()) {
		return nil, errs.ErrNotFound.WithMsg("分享码已过期")
	}

	resp := &ResolveResp{
		ShareUserID: pkgtypes.Int64Str(sc.ShareUserID),
	}
	if sc.ProductID != nil {
		v := pkgtypes.Int64Str(*sc.ProductID)
		resp.ProductID = &v
	}
	if sc.ChannelCodeID != nil {
		v := pkgtypes.Int64Str(*sc.ChannelCodeID)
		resp.ChannelCodeID = &v
	}
	return resp, nil
}

// HandleQYWXCallback 处理企微回调事件。
func (s *Service) HandleQYWXCallback(ctx context.Context, body []byte, params map[string]string) error {
	event, err := s.qywxClient.DecryptCallback(ctx, body, params)
	if err != nil {
		return fmt.Errorf("private_domain: decrypt callback: %w", err)
	}

	logger.L().Info("private_domain: qywx callback",
		zap.String("event", event.Event),
		zap.String("msg_type", event.MsgType))

	switch event.Event {
	case "add_external_contact":
		if event.WelcomeCode != "" {
			_ = s.sendWelcomeByState(ctx, event.State, event.WelcomeCode)
		}
		_ = s.applyTagsByState(ctx, event.State, event.FromUser, event.ExternalUserID)
	case "del_external_contact":
		logger.L().Info("private_domain: external contact deleted",
			zap.String("external_user_id", event.ExternalUserID))
	}
	return nil
}

func (s *Service) sendWelcomeByState(ctx context.Context, state, welcomeCode string) error {
	if state == "" || welcomeCode == "" {
		return nil
	}
	// 根据 state 查找对应渠道码，发送欢迎语
	_ = state
	return s.qywxClient.SendWelcomeMsg(ctx, welcomeCode, qywx.WelcomeMsg{
		Text: &qywx.WelcomeText{Content: "感谢您的关注！"},
	})
}

func (s *Service) applyTagsByState(ctx context.Context, state, userID, externalUserID string) error {
	if state == "" {
		return nil
	}
	_ = userID
	_ = externalUserID
	return nil
}

func int64StrSliceToInt64s(ids []pkgtypes.Int64Str) JSONInt64s {
	out := make(JSONInt64s, len(ids))
	for i, id := range ids {
		out[i] = id.Int64()
	}
	return out
}
