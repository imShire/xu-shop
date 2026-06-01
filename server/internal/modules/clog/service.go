package clog

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

const (
	// MaxMessageBytes 单条 message 长度上限（4KB）。
	MaxMessageBytes = 4 * 1024
	// MaxStackBytes stack 长度上限（16KB）。
	MaxStackBytes = 16 * 1024
	// MaxExtraBytes extra JSON 序列化后长度上限（8KB）。
	MaxExtraBytes = 8 * 1024
	// MaxURLBytes url 长度上限（512）。
	MaxURLBytes = 512
	// MaxUserAgentBytes user_agent 长度上限（255）。
	MaxUserAgentBytes = 255
	// MaxReleaseBytes release 长度上限（64）。
	MaxReleaseBytes = 64
	// MaxBatchSize 单批次最多条数。
	MaxBatchSize = 50
)

// AllowedSources 允许的 source 取值（与 DB CHECK 一致）。
var AllowedSources = map[string]struct{}{
	"admin":         {},
	"client_h5":     {},
	"client_weapp": {},
}

// AllowedLevels 允许的 level 取值。
var AllowedLevels = map[string]struct{}{
	"error": {},
	"warn":  {},
	"info":  {},
}

// Service 前端日志服务。
type Service struct {
	repo Repository
}

// NewService 构造。
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Meta 服务端从请求上下文采集的附加信息。
type Meta struct {
	ClientIP string
	TraceID  string
	UserID   *int64
	AdminID  *int64
}

// SubmitBatch 校验、截断并落库；req 顺序保持。
//
//   - 单批次 > MaxBatchSize 返回 ErrParam
//   - source / level 校验已交由 binding 标签 + 此处兜底
//   - meta.UserID/AdminID 来自服务端解析的 token；若非 nil，则覆盖 req 中的值
func (s *Service) SubmitBatch(ctx context.Context, reqs []SubmitReq, meta Meta) error {
	if len(reqs) == 0 {
		return errs.ErrParam
	}
	if len(reqs) > MaxBatchSize {
		return errs.ErrParam
	}

	logs := make([]*ClientLog, 0, len(reqs))
	for i := range reqs {
		l, err := s.build(&reqs[i], meta)
		if err != nil {
			return err
		}
		logs = append(logs, l)
	}
	if err := s.repo.Insert(ctx, logs); err != nil {
		return errs.ErrInternal
	}
	return nil
}

func (s *Service) build(req *SubmitReq, meta Meta) (*ClientLog, error) {
	if _, ok := AllowedSources[req.Source]; !ok {
		return nil, errs.ErrParam
	}
	if _, ok := AllowedLevels[req.Level]; !ok {
		return nil, errs.ErrParam
	}
	if req.Message == "" {
		return nil, errs.ErrParam
	}

	l := &ClientLog{
		Source:  req.Source,
		Level:   req.Level,
		Message: truncate(req.Message, MaxMessageBytes),
	}
	if req.Stack != "" {
		s := truncate(req.Stack, MaxStackBytes)
		l.Stack = &s
	}
	if req.URL != "" {
		u := truncate(req.URL, MaxURLBytes)
		l.URL = &u
	}
	if req.UserAgent != "" {
		ua := truncate(req.UserAgent, MaxUserAgentBytes)
		l.UserAgent = &ua
	}
	if req.Release != "" {
		r := truncate(req.Release, MaxReleaseBytes)
		l.Release = &r
	}

	// user_id / admin_id 仅信任 meta（来自服务端 token 解析）。
	// 前端任何形式上传的 user_id / admin_id 都已在 DTO 移除，service 不再回退接受。
	if meta.UserID != nil {
		uid := *meta.UserID
		l.UserID = &uid
	}
	if meta.AdminID != nil {
		aid := *meta.AdminID
		l.AdminID = &aid
	}

	if len(req.Extra) > 0 {
		raw, err := json.Marshal(req.Extra)
		if err != nil {
			return nil, errs.ErrParam
		}
		if len(raw) > MaxExtraBytes {
			return nil, errs.ErrParam
		}
		l.Extra = raw
	}

	if meta.ClientIP != "" {
		ip := meta.ClientIP
		l.ClientIP = &ip
	}
	if meta.TraceID != "" {
		tid := meta.TraceID
		l.TraceID = &tid
	}

	return l, nil
}

// truncate 把字符串截到 max 字节（按字节，不保证 UTF-8 边界，
// 但 PG text/varchar 不强制；上层用于安全兜底）。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// ErrEmptyBatch 显式空批次。
var ErrEmptyBatch = errors.New("clog: empty batch")
