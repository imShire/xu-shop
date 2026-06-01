package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitize_JSONTopLevel(t *testing.T) {
	in := []byte(`{"username":"alice","password":"p@ss","token":"abc"}`)
	out := Sanitize(in)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if m["username"] != "alice" {
		t.Errorf("username changed: %v", m["username"])
	}
	if m["password"] != "***" || m["token"] != "***" {
		t.Errorf("expected password/token masked, got: %v", m)
	}
}

func TestSanitize_JSONNested(t *testing.T) {
	in := []byte(`{"user":{"name":"a","old_password":"x"},"creds":[{"secret_key":"k"}]}`)
	out := Sanitize(in)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	user := m["user"].(map[string]any)
	if user["old_password"] != "***" {
		t.Errorf("nested old_password not masked: %v", user)
	}
	creds := m["creds"].([]any)
	first := creds[0].(map[string]any)
	if first["secret_key"] != "***" {
		t.Errorf("secret_key not masked: %v", first)
	}
}

func TestSanitize_FormFallback(t *testing.T) {
	in := []byte(`username=alice&password=secret123&keep=ok`)
	out := Sanitize(in)
	s := string(out)
	if !strings.Contains(s, "password=***") {
		t.Errorf("password not masked in form fallback: %s", s)
	}
	if !strings.Contains(s, "username=alice") {
		t.Errorf("non-sensitive field lost: %s", s)
	}
}

func TestSanitize_InvalidJSONTruncated(t *testing.T) {
	long := strings.Repeat("a", 4096)
	in := []byte(`{not json ` + long)
	out := Sanitize(in)
	// fallback 截断上限以 pkg/sensitive 内部常量为准；这里只断言显著小于输入。
	if len(out) >= len(in) {
		t.Errorf("fallback should truncate, got %d bytes (input %d)", len(out), len(in))
	}
}

func TestSanitize_Empty(t *testing.T) {
	if got := Sanitize(nil); got != nil {
		t.Errorf("nil should pass through, got: %v", got)
	}
}

func TestSanitize_PreservesNonSensitive(t *testing.T) {
	in := []byte(`{"order_no":"O1","items":[{"id":"1","qty":2}]}`)
	out := Sanitize(in)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid: %v", err)
	}
	if m["order_no"] != "O1" {
		t.Errorf("order_no lost: %v", m)
	}
}
