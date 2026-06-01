package reconciliation

import (
	"testing"

	rec "github.com/xushop/xu-shop/internal/modules/reconciliation"
)

func TestExpectedCommission(t *testing.T) {
	cases := []struct {
		base int64
		rate float64
		want int64
	}{
		{10000, 0.10, 1000},
		{9999, 0.10, 999},
		{0, 0.1, 0},
		{1000, 0, 0},
		{12345, 0.075, 925}, // floor(925.875)
	}
	for _, c := range cases {
		got := expectedCommission(c.base, c.rate)
		if got != c.want {
			t.Errorf("expectedCommission(%d, %v)=%d want %d", c.base, c.rate, got, c.want)
		}
	}
}

func TestDiffCommission_Match(t *testing.T) {
	if d := diffCommission(CommissionRow{ID: 1, BaseAmountCents: 10000, Rate: 0.1, AmountCents: 1000}); len(d) != 0 {
		t.Fatalf("expect no diff, got %+v", d)
	}
}

func TestDiffCommission_SmallDiffWarn(t *testing.T) {
	// expected=1000, actual=1050 → delta=50 → warn
	diffs := diffCommission(CommissionRow{ID: 1, BaseAmountCents: 10000, Rate: 0.1, AmountCents: 1050})
	if len(diffs) != 1 {
		t.Fatalf("expect 1 diff, got %d", len(diffs))
	}
	if diffs[0].Severity != rec.SeverityWarn {
		t.Errorf("expect warn, got %s", diffs[0].Severity)
	}
	if diffs[0].DiffCents != 50 {
		t.Errorf("expect delta=50, got %d", diffs[0].DiffCents)
	}
}

func TestDiffCommission_LargeDiffCritical(t *testing.T) {
	// expected=1000, actual=1200 → delta=200 → critical (> 100)
	diffs := diffCommission(CommissionRow{ID: 2, BaseAmountCents: 10000, Rate: 0.1, AmountCents: 1200})
	if len(diffs) != 1 || diffs[0].Severity != rec.SeverityCritical {
		t.Fatalf("expect 1 critical diff, got %+v", diffs)
	}
}

func TestDiffCommission_NegativeDelta(t *testing.T) {
	// expected=1000, actual=500 → delta=-500 → critical
	diffs := diffCommission(CommissionRow{ID: 3, BaseAmountCents: 10000, Rate: 0.1, AmountCents: 500})
	if len(diffs) != 1 || diffs[0].DiffCents != -500 || diffs[0].Severity != rec.SeverityCritical {
		t.Fatalf("unexpected: %+v", diffs)
	}
}
