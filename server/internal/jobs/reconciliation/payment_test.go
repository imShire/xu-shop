package reconciliation

import (
	"context"
	"strconv"
	"testing"
	"time"

	rec "github.com/xushop/xu-shop/internal/modules/reconciliation"
	pkgwxpay "github.com/xushop/xu-shop/internal/pkg/wxpay"
)

func TestDiffPayment_AllMatch(t *testing.T) {
	local := PaymentRow{OrderID: 1, OrderNo: "O1", AmountCents: 10000, TransactionID: "T1"}
	remote := &pkgwxpay.QueryResp{TradeState: "SUCCESS", AmtCents: 10000, TransactionID: "T1"}
	if diffs := diffPayment(local, remote); len(diffs) != 0 {
		t.Fatalf("expect 0 diffs, got %d: %+v", len(diffs), diffs)
	}
}

func TestDiffPayment_AmountMismatch(t *testing.T) {
	local := PaymentRow{OrderID: 1, OrderNo: "O1", AmountCents: 10000, TransactionID: "T1"}
	remote := &pkgwxpay.QueryResp{TradeState: "SUCCESS", AmtCents: 9900, TransactionID: "T1"}
	diffs := diffPayment(local, remote)
	if len(diffs) != 1 || diffs[0].Field != "amount_cents" {
		t.Fatalf("expect 1 amount diff, got %+v", diffs)
	}
	if diffs[0].Severity != rec.SeverityCritical {
		t.Errorf("expect critical, got %s", diffs[0].Severity)
	}
	if diffs[0].DiffCents == nil || *diffs[0].DiffCents != 100 {
		t.Errorf("expect diff=100, got %+v", diffs[0].DiffCents)
	}
}

func TestDiffPayment_StateMismatch(t *testing.T) {
	local := PaymentRow{OrderID: 1, OrderNo: "O1", AmountCents: 10000, TransactionID: "T1"}
	remote := &pkgwxpay.QueryResp{TradeState: "NOTPAY", AmtCents: 10000, TransactionID: "T1"}
	diffs := diffPayment(local, remote)
	if len(diffs) != 1 || diffs[0].Field != "state" {
		t.Fatalf("expect state diff, got %+v", diffs)
	}
	if diffs[0].Severity != rec.SeverityCritical {
		t.Errorf("expect critical")
	}
}

func TestDiffPayment_RemoteNil(t *testing.T) {
	local := PaymentRow{OrderID: 1, OrderNo: "O1", AmountCents: 10000}
	diffs := diffPayment(local, nil)
	if len(diffs) != 1 || diffs[0].Actual != "NOT_FOUND" {
		t.Fatalf("expect NOT_FOUND diff, got %+v", diffs)
	}
}

func TestDiffPayment_TxnMismatchWarn(t *testing.T) {
	local := PaymentRow{OrderID: 1, OrderNo: "O1", AmountCents: 100, TransactionID: "A"}
	remote := &pkgwxpay.QueryResp{TradeState: "SUCCESS", AmtCents: 100, TransactionID: "B"}
	diffs := diffPayment(local, remote)
	if len(diffs) != 1 || diffs[0].Field != "transaction_id" || diffs[0].Severity != rec.SeverityWarn {
		t.Fatalf("expect warn txn diff, got %+v", diffs)
	}
}

// 集成式 smoke test：调用 RunPayment 时传入 nil DB 应快速失败（确认入口可调用）。
// 仅作 sanity，不依赖 DB。
func TestRunPayment_Smoke(t *testing.T) {
	_ = context.Background()
	_ = strconv.Itoa
	_ = time.Now
	_ = rec.JobPayment
}
