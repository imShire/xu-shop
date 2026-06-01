package reconciliation

import (
	"testing"

	rec "github.com/xushop/xu-shop/internal/modules/reconciliation"
)

func TestDiffInventory_Match(t *testing.T) {
	if d := diffInventory(SKURow{SKUID: 1, LockedStock: 5}, 5); len(d) != 0 {
		t.Fatalf("expect no diff, got %+v", d)
	}
}

func TestDiffInventory_Mismatch(t *testing.T) {
	diffs := diffInventory(SKURow{SKUID: 1, LockedStock: 5}, 3)
	if len(diffs) != 1 {
		t.Fatalf("expect 1 diff, got %d", len(diffs))
	}
	d := diffs[0]
	if d.Field != "locked_stock_balance" || d.Expected != 3 || d.Actual != 5 {
		t.Errorf("unexpected diff: %+v", d)
	}
	if d.Severity != rec.SeverityWarn {
		t.Errorf("expect warn, got %s", d.Severity)
	}
}

func TestDiffInventory_Overshoot(t *testing.T) {
	diffs := diffInventory(SKURow{SKUID: 9, LockedStock: 0}, 10)
	if len(diffs) != 1 || diffs[0].Expected != 10 || diffs[0].Actual != 0 {
		t.Fatalf("unexpected: %+v", diffs)
	}
}
