package clog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newHandlerEngine() (*Handler, *captureRepo, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	repo := &captureRepo{}
	svc := NewService(repo)
	h := NewHandler(svc)
	r := gin.New()
	r.POST("/api/v1/internal/clog", h.Submit)
	r.POST("/api/v1/internal/clog/batch", h.SubmitBatch)
	return h, repo, r
}

func TestHandler_Submit_OK(t *testing.T) {
	_, repo, r := newHandlerEngine()
	body := `{"source":"client_h5","level":"error","message":"oops"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/clog",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}
	if len(repo.logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(repo.logs))
	}
}

func TestHandler_Submit_InvalidSource(t *testing.T) {
	_, _, r := newHandlerEngine()
	body := `{"source":"ios","level":"error","message":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/clog",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_SubmitBatch_OK(t *testing.T) {
	_, repo, r := newHandlerEngine()
	items := []map[string]any{
		{"source": "admin", "level": "warn", "message": "a"},
		{"source": "admin", "level": "info", "message": "b"},
	}
	payload, _ := json.Marshal(map[string]any{"logs": items})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/clog/batch",
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}
	if len(repo.logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(repo.logs))
	}
}

func TestHandler_SubmitBatch_OverLimit(t *testing.T) {
	_, _, r := newHandlerEngine()
	items := make([]map[string]any, MaxBatchSize+1)
	for i := range items {
		items[i] = map[string]any{"source": "admin", "level": "info", "message": "x"}
	}
	payload, _ := json.Marshal(map[string]any{"logs": items})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/clog/batch",
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for over-limit batch, got %d body: %s", w.Code, w.Body.String())
	}
}
