package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xushop/xu-shop/internal/modules/admin/audit"
)

// fakeRepo 收集插入的审计日志。
type fakeRepo struct {
	mu    sync.Mutex
	logs  []*audit.AuditLog
	err   error
}

func (r *fakeRepo) Insert(_ context.Context, log *audit.AuditLog) error {
	if r.err != nil {
		return r.err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, log)
	return nil
}

func (r *fakeRepo) all() []*audit.AuditLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*audit.AuditLog, len(r.logs))
	copy(out, r.logs)
	return out
}

func newAuditEngine(repo audit.Repository, handler gin.HandlerFunc, opts ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuditLog(repo))
	for _, o := range opts {
		r.Use(o)
	}
	r.POST("/api/v1/admin/banners", handler)
	r.POST("/api/v1/admin/products/:id", handler)
	r.GET("/api/v1/admin/banners", handler)
	r.POST("/api/v1/c/orders", handler)
	r.POST("/api/v1/admin/auth/login", handler)
	return r
}

// waitForLogs 轮询直到 repo 出现期望条数或超时，规避异步写入竞态。
func waitForLogs(t *testing.T, repo *fakeRepo, want int) []*audit.AuditLog {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := repo.all(); len(got) >= want {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	return repo.all()
}

func TestAuditLog_GETSkipped(t *testing.T) {
	repo := &fakeRepo{}
	r := newAuditEngine(repo, func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/banners", nil)
	r.ServeHTTP(w, req)

	// 留点时间给可能误触发的 goroutine
	time.Sleep(20 * time.Millisecond)
	if logs := repo.all(); len(logs) != 0 {
		t.Errorf("GET should not trigger audit, got %d logs", len(logs))
	}
}

func TestAuditLog_NonAdminPathSkipped(t *testing.T) {
	repo := &fakeRepo{}
	r := newAuditEngine(repo, func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/c/orders",
		strings.NewReader(`{"x":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	time.Sleep(20 * time.Millisecond)
	if logs := repo.all(); len(logs) != 0 {
		t.Errorf("/c/ path should not be audited, got %d", len(logs))
	}
}

func TestAuditLog_POSTRecordsSanitizedBody(t *testing.T) {
	repo := &fakeRepo{}
	handler := func(c *gin.Context) {
		// 确认 handler 仍能读到 body
		b, _ := io.ReadAll(c.Request.Body)
		if !bytes.Contains(b, []byte("p@ss")) {
			t.Errorf("handler lost access to body: %s", b)
		}
		c.JSON(http.StatusCreated, gin.H{"id": "1"})
	}
	r := newAuditEngine(repo, handler)

	body := `{"username":"alice","password":"p@ss"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/banners",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	logs := waitForLogs(t, repo, 1)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	got := logs[0]
	if got.ResponseStatus != http.StatusCreated {
		t.Errorf("status mismatch: %d", got.ResponseStatus)
	}
	if got.Method != http.MethodPost {
		t.Errorf("method: %s", got.Method)
	}
	// password 必须脱敏
	var parsed map[string]any
	if err := json.Unmarshal(got.RequestBody, &parsed); err != nil {
		t.Fatalf("request_body not valid JSON: %v / %s", err, got.RequestBody)
	}
	if parsed["password"] != "***" {
		t.Errorf("password not masked: %v", parsed)
	}
	if parsed["username"] != "alice" {
		t.Errorf("username changed: %v", parsed)
	}
	if got.ResponseExcerpt == nil || !strings.Contains(*got.ResponseExcerpt, `"id":"1"`) {
		t.Errorf("response excerpt missing: %v", got.ResponseExcerpt)
	}
	if got.Action != "POST:/api/v1/admin/banners" {
		t.Errorf("default action wrong: %s", got.Action)
	}
}

func TestAuditLog_BodyTruncated(t *testing.T) {
	repo := &fakeRepo{}
	handler := func(c *gin.Context) { c.Status(http.StatusOK) }
	r := newAuditEngine(repo, handler)

	huge := strings.Repeat("a", auditMaxBodyBytes+5000)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/banners",
		strings.NewReader(huge))
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)

	logs := waitForLogs(t, repo, 1)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	// 非 JSON、超长 → 会走 jsonQuote 包成 JSON 字符串；长度被截到 fallbackMaxKeep + 边角
	if len(logs[0].RequestBody) > 5000 {
		t.Errorf("request_body should be truncated, got %d", len(logs[0].RequestBody))
	}
}

func TestAuditLog_CustomActionOverride(t *testing.T) {
	repo := &fakeRepo{}
	handler := func(c *gin.Context) {
		c.Set(CtxKeyAuditAction, "banner.create")
		c.Set(CtxKeyAuditTargetType, "banner")
		c.Set(CtxKeyAuditTargetID, "123")
		c.Status(http.StatusOK)
	}
	r := newAuditEngine(repo, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/banners",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logs := waitForLogs(t, repo, 1)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log")
	}
	if logs[0].Action != "banner.create" {
		t.Errorf("action override failed: %s", logs[0].Action)
	}
	if logs[0].TargetType == nil || *logs[0].TargetType != "banner" {
		t.Errorf("target_type override failed: %v", logs[0].TargetType)
	}
	if logs[0].TargetID == nil || *logs[0].TargetID != "123" {
		t.Errorf("target_id override failed: %v", logs[0].TargetID)
	}
}

func TestAuditLog_RepoErrorDoesNotBreak(t *testing.T) {
	repo := &fakeRepo{err: errors.New("db down")}
	called := false
	handler := func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
	r := newAuditEngine(repo, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/banners",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Errorf("handler not invoked")
	}
	if w.Code != http.StatusOK {
		t.Errorf("response status changed by repo error: %d", w.Code)
	}
	if w.Body.String() != `{"ok":true}` {
		t.Errorf("response body changed by repo error: %s", w.Body.String())
	}
}

func TestAuditLog_AdminIDOverride(t *testing.T) {
	repo := &fakeRepo{}
	handler := func(c *gin.Context) {
		c.Set(CtxKeyAuditAdminID, int64(42))
		c.Status(http.StatusOK)
	}
	r := newAuditEngine(repo, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login",
		strings.NewReader(`{"username":"a","password":"p"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logs := waitForLogs(t, repo, 1)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log")
	}
	if logs[0].AdminID != 42 {
		t.Errorf("admin_id override failed: %d", logs[0].AdminID)
	}
}
