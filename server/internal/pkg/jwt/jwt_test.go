package jwt_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

var testCfg = pkgjwt.Config{
	Secret:      "test-secret-key-for-unit-testing",
	UserExpiry:  time.Hour,
	AdminExpiry: 8 * time.Hour,
}

func TestSignAndParse(t *testing.T) {
	claims := pkgjwt.Claims{
		Sub: 12345,
		Typ: "user",
		JTI: "test-jti-001",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(testCfg.UserExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := pkgjwt.Sign(testCfg, claims)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	parsed, err := pkgjwt.Parse(testCfg, token)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if parsed.Sub != 12345 {
		t.Errorf("expected sub=12345, got %d", parsed.Sub)
	}
	if parsed.Typ != "user" {
		t.Errorf("expected typ=user, got %s", parsed.Typ)
	}
	if parsed.JTI != "test-jti-001" {
		t.Errorf("expected jti=test-jti-001, got %s", parsed.JTI)
	}
}

func TestExpiredToken(t *testing.T) {
	claims := pkgjwt.Claims{
		Sub: 99999,
		Typ: "user",
		JTI: "expired-jti",
		RegisteredClaims: jwt.RegisteredClaims{
			// 已过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		},
	}

	token, err := pkgjwt.Sign(testCfg, claims)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}

	_, err = pkgjwt.Parse(testCfg, token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}
