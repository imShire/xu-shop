package payment

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newOrphanTestDB 用 sqlite in-memory 模拟 payment_orphan_retry 表。
// 表 DDL 与 docs/arch/91-db-schema.md 中 PG 版本字段一致，但用 sqlite 兼容语法。
func newOrphanTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})

	if err := db.Exec(`CREATE TABLE payment_orphan_retry (
		id            INTEGER PRIMARY KEY,
		payment_id    INTEGER NOT NULL,
		wx_txid       TEXT NOT NULL DEFAULT '',
		amount_cents  INTEGER NOT NULL,
		reason        TEXT NOT NULL DEFAULT '',
		retry_count   INTEGER NOT NULL DEFAULT 0,
		next_retry_at DATETIME NOT NULL,
		last_error    TEXT,
		created_at    DATETIME,
		updated_at    DATETIME
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	return db
}

// TestOrphanRetryRepo_Enqueue_Success 正常插入返回 nil，所有字段写入正确。
func TestOrphanRetryRepo_Enqueue_Success(t *testing.T) {
	db := newOrphanTestDB(t)
	repo := NewOrphanRetryRepo(db)

	before := time.Now()
	delay := 1 * time.Minute
	err := repo.Enqueue(context.Background(), 1001, "TXN_OK_1", 9900, "金额不一致自动退款", delay)
	if err != nil {
		t.Fatalf("Enqueue should succeed, got %v", err)
	}
	after := time.Now()

	var row PaymentOrphanRetry
	if err := db.First(&row).Error; err != nil {
		t.Fatalf("read back row: %v", err)
	}

	if row.ID == 0 {
		t.Fatalf("id should be snowflake-generated non-zero")
	}
	if row.PaymentID != 1001 {
		t.Fatalf("payment_id: want 1001, got %d", row.PaymentID)
	}
	if row.WxTxID != "TXN_OK_1" {
		t.Fatalf("wx_txid: want TXN_OK_1, got %q", row.WxTxID)
	}
	if row.AmountCents != 9900 {
		t.Fatalf("amount_cents: want 9900, got %d", row.AmountCents)
	}
	if row.Reason != "金额不一致自动退款" {
		t.Fatalf("reason mismatch: got %q", row.Reason)
	}
	if row.RetryCount != 0 {
		t.Fatalf("retry_count: want 0 on initial enqueue, got %d", row.RetryCount)
	}

	// next_retry_at 应在 [before+delay, after+delay] 之间（允许 1s 偏差以兼容时区/精度）
	lo := before.Add(delay).Add(-1 * time.Second)
	hi := after.Add(delay).Add(1 * time.Second)
	if row.NextRetryAt.Before(lo) || row.NextRetryAt.After(hi) {
		t.Fatalf("next_retry_at out of range: got %v, want in [%v, %v]", row.NextRetryAt, lo, hi)
	}
}

// TestOrphanRetryRepo_Enqueue_DuplicateAllowed
// 当前实现未在 (payment_id, wx_txid) 上加唯一约束，重复入队应成功并产生多行。
// 由 reconciler/cron 侧做去重幂等。若未来加唯一索引，本测试要同步改成断言冲突错误。
func TestOrphanRetryRepo_Enqueue_DuplicateAllowed(t *testing.T) {
	db := newOrphanTestDB(t)
	repo := NewOrphanRetryRepo(db)
	ctx := context.Background()

	if err := repo.Enqueue(ctx, 2002, "TXN_DUP", 5000, "first", time.Minute); err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	if err := repo.Enqueue(ctx, 2002, "TXN_DUP", 5000, "second", time.Minute); err != nil {
		t.Fatalf("second enqueue should not crash (no unique constraint): %v", err)
	}

	var count int64
	if err := db.Model(&PaymentOrphanRetry{}).
		Where("payment_id = ? AND wx_txid = ?", 2002, "TXN_DUP").
		Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows for duplicate enqueue, got %d", count)
	}
}

// TestOrphanRetryRepo_Enqueue_ZeroDelay delay=0 时 next_retry_at ≈ now。
func TestOrphanRetryRepo_Enqueue_ZeroDelay(t *testing.T) {
	db := newOrphanTestDB(t)
	repo := NewOrphanRetryRepo(db)

	before := time.Now()
	if err := repo.Enqueue(context.Background(), 3003, "TXN_ZERO", 100, "now", 0); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	after := time.Now()

	var row PaymentOrphanRetry
	if err := db.First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.NextRetryAt.Before(before.Add(-time.Second)) || row.NextRetryAt.After(after.Add(time.Second)) {
		t.Fatalf("zero delay: next_retry_at should ~= now, got %v", row.NextRetryAt)
	}
}

// TestOrphanRetryRepo_Enqueue_SnowflakeIDsUnique 多次入队 ID 不重复。
func TestOrphanRetryRepo_Enqueue_SnowflakeIDsUnique(t *testing.T) {
	db := newOrphanTestDB(t)
	repo := NewOrphanRetryRepo(db)
	ctx := context.Background()

	const n = 20
	for i := 0; i < n; i++ {
		if err := repo.Enqueue(ctx, int64(4000+i), "TXN_SEQ", 1, "x", time.Second); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}

	var rows []PaymentOrphanRetry
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != n {
		t.Fatalf("expected %d rows, got %d", n, len(rows))
	}
	seen := make(map[int64]struct{}, n)
	for _, r := range rows {
		if _, dup := seen[r.ID]; dup {
			t.Fatalf("duplicate snowflake id: %d", r.ID)
		}
		seen[r.ID] = struct{}{}
	}
}
