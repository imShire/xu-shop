package middleware

import "testing"

// TestToString 覆盖 H-002：限流 key 拼接必须支持 int64（user_id 实际类型）
// 否则按用户限流会退化为全局共享桶。
func TestToString(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string", "abc", "abc"},
		{"int64", int64(1234567890), "1234567890"},
		{"int64_neg", int64(-7), "-7"},
		{"int", 42, "42"},
		{"int32", int32(99), "99"},
		{"uint64", uint64(18446744073709551610), "18446744073709551610"},
		{"bool_fallback", true, "true"},
		{"float_fallback", 3.14, "3.14"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := toString(c.in)
			if got != c.want {
				t.Fatalf("toString(%v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
