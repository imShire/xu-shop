// Package jwt 封装 golang-jwt/jwt/v5，提供签发和解析工具。
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 JWT 载荷结构。
type Claims struct {
	Sub   int64    `json:"sub"`
	Typ   string   `json:"typ"` // user / admin
	JTI   string   `json:"jti"`
	Roles []string `json:"roles,omitempty"`
	Perms []string `json:"perms,omitempty"`
	jwt.RegisteredClaims
}

// Config 是 JWT 配置。
type Config struct {
	Secret        string
	UserExpiry    time.Duration // C 端 access token 有效期，默认 30 天
	AdminExpiry   time.Duration // 后台 access token 有效期，默认 8 小时
	RefreshExpiry time.Duration // refresh token 有效期（C 端 90 天，admin 7 天）
}

// Sign 签发 JWT。
func Sign(cfg Config, claims Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.Secret))
}

// Parse 解析并验证 JWT，返回 Claims。
func Parse(cfg Config, tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt: unexpected signing method")
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	return c, nil
}
