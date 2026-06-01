package sensitive

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitize_JSONMasksPassword(t *testing.T) {
	in := []byte(`{"username":"alice","password":"p@ss"}`)
	out := Sanitize(in)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("not json: %v / %s", err, out)
	}
	if m["password"] != "***" {
		t.Errorf("password not masked: %v", m)
	}
	if m["username"] != "alice" {
		t.Errorf("username changed: %v", m)
	}
}

func TestSanitize_LoginResponseTokens(t *testing.T) {
	in := []byte(`{"code":0,"data":{"access_token":"AAA","refresh_token":"BBB","user":{"id":"1"}}}`)
	out := Sanitize(in)
	var m map[string]any
	_ = json.Unmarshal(out, &m)
	data := m["data"].(map[string]any)
	if data["access_token"] != "***" || data["refresh_token"] != "***" {
		t.Errorf("tokens not masked: %v", data)
	}
	user := data["user"].(map[string]any)
	if user["id"] != "1" {
		t.Errorf("non-sensitive nested field altered: %v", user)
	}
}

func TestSanitize_NestedSecret(t *testing.T) {
	in := []byte(`{"cfg":{"mch_id":"123456","aes_key":"xxxxx","name":"ok"}}`)
	out := Sanitize(in)
	if !strings.Contains(string(out), `"mch_id":"***"`) {
		t.Errorf("mch_id not masked: %s", out)
	}
	if !strings.Contains(string(out), `"aes_key":"***"`) {
		t.Errorf("aes_key not masked: %s", out)
	}
	if !strings.Contains(string(out), `"name":"ok"`) {
		t.Errorf("name should be kept: %s", out)
	}
}

func TestSanitize_FormFallback(t *testing.T) {
	in := []byte(`username=bob&password=qwerty&age=18`)
	out := Sanitize(in)
	s := string(out)
	if !strings.Contains(s, "password=***") {
		t.Errorf("form password not masked: %s", s)
	}
	if !strings.Contains(s, "age=18") {
		t.Errorf("non-sensitive form field altered: %s", s)
	}
}

func TestSanitize_NonJSONLongTruncated(t *testing.T) {
	huge := strings.Repeat("x", fallbackMaxKeep+1000) + "password=zzz"
	out := Sanitize([]byte(huge))
	if len(out) > fallbackMaxKeep {
		t.Errorf("non-json long input not truncated: %d", len(out))
	}
}

func TestSanitize_Empty(t *testing.T) {
	if got := Sanitize(nil); got != nil {
		t.Errorf("nil input should pass through")
	}
	if got := Sanitize([]byte{}); len(got) != 0 {
		t.Errorf("empty input should pass through, got %q", got)
	}
}
