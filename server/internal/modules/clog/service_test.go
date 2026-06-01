package clog

import (
	"context"
	"strings"
	"sync"
	"testing"
)

type captureRepo struct {
	mu   sync.Mutex
	logs []*ClientLog
}

func (r *captureRepo) Insert(_ context.Context, logs []*ClientLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, logs...)
	return nil
}

func TestService_SubmitBatch_OK(t *testing.T) {
	r := &captureRepo{}
	s := NewService(r)
	reqs := []SubmitReq{{
		Source:  "client_h5",
		Level:   "error",
		Message: "boom",
	}}
	if err := s.SubmitBatch(context.Background(), reqs, Meta{ClientIP: "1.2.3.4"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(r.logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(r.logs))
	}
	if r.logs[0].Message != "boom" || r.logs[0].ClientIP == nil || *r.logs[0].ClientIP != "1.2.3.4" {
		t.Errorf("log mismatch: %+v", r.logs[0])
	}
}

func TestService_RejectInvalidSource(t *testing.T) {
	s := NewService(&captureRepo{})
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "ios", Level: "error", Message: "x",
	}}, Meta{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestService_RejectInvalidLevel(t *testing.T) {
	s := NewService(&captureRepo{})
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "admin", Level: "debug", Message: "x",
	}}, Meta{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestService_TruncateMessageAndStack(t *testing.T) {
	r := &captureRepo{}
	s := NewService(r)
	long := strings.Repeat("a", MaxMessageBytes+1000)
	longStack := strings.Repeat("b", MaxStackBytes+5000)
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "admin", Level: "warn", Message: long, Stack: longStack,
	}}, Meta{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	got := r.logs[0]
	if len(got.Message) != MaxMessageBytes {
		t.Errorf("message not truncated: %d", len(got.Message))
	}
	if got.Stack == nil || len(*got.Stack) != MaxStackBytes {
		t.Errorf("stack not truncated: %v", got.Stack)
	}
}

func TestService_RejectBatchOverLimit(t *testing.T) {
	s := NewService(&captureRepo{})
	reqs := make([]SubmitReq, MaxBatchSize+1)
	for i := range reqs {
		reqs[i] = SubmitReq{Source: "admin", Level: "info", Message: "m"}
	}
	if err := s.SubmitBatch(context.Background(), reqs, Meta{}); err == nil {
		t.Errorf("expected error for batch > %d", MaxBatchSize)
	}
}

func TestService_RejectEmptyBatch(t *testing.T) {
	s := NewService(&captureRepo{})
	if err := s.SubmitBatch(context.Background(), nil, Meta{}); err == nil {
		t.Errorf("expected error for empty batch")
	}
}

func TestService_RejectExtraTooLarge(t *testing.T) {
	s := NewService(&captureRepo{})
	big := strings.Repeat("z", MaxExtraBytes+500)
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "admin", Level: "warn", Message: "m",
		Extra: map[string]any{"big": big},
	}}, Meta{})
	if err == nil {
		t.Errorf("expected error for oversize extra")
	}
}

func TestService_MetaUserIDUsed(t *testing.T) {
	r := &captureRepo{}
	s := NewService(r)
	uid := int64(99)
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "client_h5", Level: "error", Message: "m",
	}}, Meta{UserID: &uid})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if r.logs[0].UserID == nil || *r.logs[0].UserID != 99 {
		t.Errorf("meta user_id should be used: %v", r.logs[0].UserID)
	}
}

func TestService_ClientUserIDIgnored(t *testing.T) {
	// CR Blocker #1: DTO 已移除 user_id，服务端不再信任前端传值。
	// meta 为空时，l.UserID 必须是 nil。
	r := &captureRepo{}
	s := NewService(r)
	err := s.SubmitBatch(context.Background(), []SubmitReq{{
		Source: "client_h5", Level: "info", Message: "m",
	}}, Meta{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if r.logs[0].UserID != nil {
		t.Errorf("user_id should be nil when meta absent: %v", *r.logs[0].UserID)
	}
}
