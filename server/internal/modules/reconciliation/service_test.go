package reconciliation

import (
	"context"
	"testing"
	"time"
)

// mockRepo 内存实现，模拟 upsert 行为（按业务唯一键合并）。
type mockRepo struct {
	byKey map[string]*Diff
	byID  map[int64]*Diff
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		byKey: make(map[string]*Diff),
		byID:  make(map[int64]*Diff),
	}
}

func (m *mockRepo) key(d *Diff) string {
	return d.Job + "|" + d.BizDate.Format("2006-01-02") + "|" +
		d.RefType + "|" + d.RefID + "|" + d.Field
}

func (m *mockRepo) Upsert(_ context.Context, d *Diff) error {
	k := m.key(d)
	if existing, ok := m.byKey[k]; ok {
		// 模拟 ON CONFLICT DO UPDATE：刷新值字段，保留状态字段
		existing.ExpectedValue = d.ExpectedValue
		existing.ActualValue = d.ActualValue
		existing.DiffCents = d.DiffCents
		existing.Severity = d.Severity
		existing.UpdatedAt = time.Now()
		return nil
	}
	clone := *d
	m.byKey[k] = &clone
	m.byID[d.ID] = &clone
	return nil
}

func (m *mockRepo) List(_ context.Context, f Filter) ([]Diff, int64, error) {
	var out []Diff
	for _, d := range m.byID {
		if f.Job != "" && d.Job != f.Job {
			continue
		}
		if f.Status != "" && d.Status != f.Status {
			continue
		}
		out = append(out, *d)
	}
	return out, int64(len(out)), nil
}

func (m *mockRepo) Get(_ context.Context, id int64) (*Diff, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, errNotFound
	}
	return d, nil
}

func (m *mockRepo) UpdateStatus(_ context.Context, id int64, status string, op int64, note *string) error {
	d, err := m.Get(context.Background(), id)
	if err != nil {
		return err
	}
	d.Status = status
	if note != nil {
		d.Note = note
	}
	if status == StatusAcknowledged {
		d.AckedBy = &op
		now := time.Now()
		d.AckedAt = &now
	}
	if status == StatusResolved {
		d.ResolvedBy = &op
		now := time.Now()
		d.ResolvedAt = &now
	}
	return nil
}

var errNotFound = stubErr("not found")

type stubErr string

func (s stubErr) Error() string { return string(s) }

func TestRecordDiff_UpsertDeduplication(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	bizDate := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	exp1 := "10000"
	act1 := "9900"
	err := svc.RecordDiff(ctx, &Diff{
		Job: JobPayment, BizDate: bizDate,
		RefType: RefTypeOrder, RefID: "1001", Field: "amount_cents",
		ExpectedValue: &exp1, ActualValue: &act1, Severity: SeverityCritical,
	})
	if err != nil {
		t.Fatalf("first record: %v", err)
	}
	if got := len(repo.byKey); got != 1 {
		t.Fatalf("expect 1 row, got %d", got)
	}

	// 第二次写入相同唯一键 → 应更新而非新增
	act2 := "9800"
	err = svc.RecordDiff(ctx, &Diff{
		Job: JobPayment, BizDate: bizDate,
		RefType: RefTypeOrder, RefID: "1001", Field: "amount_cents",
		ExpectedValue: &exp1, ActualValue: &act2, Severity: SeverityCritical,
	})
	if err != nil {
		t.Fatalf("second record: %v", err)
	}
	if got := len(repo.byKey); got != 1 {
		t.Fatalf("expect dedup, got %d rows", got)
	}
	stored := repo.byKey[JobPayment+"|2026-05-30|"+RefTypeOrder+"|1001|amount_cents"]
	if stored.ActualValue == nil || *stored.ActualValue != "9800" {
		t.Fatalf("expect actual updated to 9800, got %v", stored.ActualValue)
	}

	// 不同 field 应新增
	err = svc.RecordDiff(ctx, &Diff{
		Job: JobPayment, BizDate: bizDate,
		RefType: RefTypeOrder, RefID: "1001", Field: "state",
	})
	if err != nil {
		t.Fatalf("third record: %v", err)
	}
	if got := len(repo.byKey); got != 2 {
		t.Fatalf("expect 2 rows for different field, got %d", got)
	}
}

func TestRecordDiff_ValidationAndDefaults(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.RecordDiff(ctx, &Diff{Job: "", RefType: "x", RefID: "1", Field: "f"}); err == nil {
		t.Fatal("expect validation error for empty job")
	}

	// 业务日缺省 → 昨天
	d := &Diff{Job: JobInventory, RefType: RefTypeSKU, RefID: "55", Field: "stock_balance"}
	if err := svc.RecordDiff(ctx, d); err != nil {
		t.Fatalf("record: %v", err)
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	if d.BizDate.Day() != yesterday.Day() {
		t.Errorf("expect default biz_date=yesterday, got %v", d.BizDate)
	}
	if d.Severity != SeverityWarn || d.Status != StatusOpen {
		t.Errorf("expect default warn/open, got %s/%s", d.Severity, d.Status)
	}
	if d.ID == 0 {
		t.Error("expect ID assigned")
	}
}

func TestService_AcknowledgeAndResolve(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	d := &Diff{Job: JobCommission, RefType: RefTypeCommissionRecord, RefID: "77", Field: "amount_cents"}
	if err := svc.RecordDiff(ctx, d); err != nil {
		t.Fatal(err)
	}

	if err := svc.Acknowledge(ctx, d.ID, 999); err != nil {
		t.Fatalf("ack: %v", err)
	}
	stored, _ := repo.Get(ctx, d.ID)
	if stored.Status != StatusAcknowledged || stored.AckedBy == nil || *stored.AckedBy != 999 {
		t.Errorf("ack status not applied: %+v", stored)
	}

	note := "已人工补单"
	if err := svc.Resolve(ctx, d.ID, 1000, &note); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	stored, _ = repo.Get(ctx, d.ID)
	if stored.Status != StatusResolved || stored.Note == nil || *stored.Note != note {
		t.Errorf("resolve not applied: %+v", stored)
	}

	// 再次 resolve 应拒绝
	if err := svc.Resolve(ctx, d.ID, 1000, nil); err == nil {
		t.Error("expect error on double resolve")
	}
}
