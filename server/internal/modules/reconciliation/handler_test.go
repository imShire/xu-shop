package reconciliation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) (*gin.Engine, *Service, *mockRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := newMockRepo()
	svc := NewService(repo)
	h := NewHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("admin_id", int64(123))
		c.Next()
	})
	g := r.Group("/admin/reconciliation/diff")
	g.GET("", h.ListDiffs)
	g.POST("/:id/ack", h.AckDiff)
	g.POST("/:id/resolve", h.ResolveDiff)
	return r, svc, repo
}

func TestHandler_ListAndAckAndResolve(t *testing.T) {
	r, svc, repo := setupRouter(t)
	ctx := context.Background()

	bizDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	d := &Diff{Job: JobPayment, BizDate: bizDate, RefType: RefTypeOrder,
		RefID: "1001", Field: "state", Severity: SeverityCritical}
	if err := svc.RecordDiff(ctx, d); err != nil {
		t.Fatal(err)
	}

	// List
	req := httptest.NewRequest(http.MethodGet, "/admin/reconciliation/diff?job=payment", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []DiffResp `json:"items"`
			Total int64      `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 0 || resp.Data.Total != 1 || len(resp.Data.Items) != 1 {
		t.Fatalf("unexpected list resp: %+v", resp)
	}

	// Ack
	id := resp.Data.Items[0].ID.String()
	req = httptest.NewRequest(http.MethodPost, "/admin/reconciliation/diff/"+id+"/ack", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ack status=%d body=%s", w.Code, w.Body.String())
	}
	stored, _ := repo.Get(ctx, d.ID)
	if stored.Status != StatusAcknowledged {
		t.Fatalf("status=%s", stored.Status)
	}

	// Resolve with note
	req = httptest.NewRequest(http.MethodPost, "/admin/reconciliation/diff/"+id+"/resolve",
		strings.NewReader(`{"note":"已人工补单"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("resolve status=%d body=%s", w.Code, w.Body.String())
	}
	stored, _ = repo.Get(ctx, d.ID)
	if stored.Status != StatusResolved || stored.Note == nil || *stored.Note != "已人工补单" {
		t.Fatalf("resolve not applied: %+v", stored)
	}
}

func TestHandler_BadID(t *testing.T) {
	r, _, _ := setupRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/reconciliation/diff/abc/ack", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expect non-200 on bad id, got %d", w.Code)
	}
}

func TestHandler_BadBizDate(t *testing.T) {
	r, _, _ := setupRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/reconciliation/diff?biz_date=not-a-date", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expect non-200 on bad biz_date, got %d", w.Code)
	}
}
