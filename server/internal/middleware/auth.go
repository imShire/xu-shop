package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

const (
	// userStatusCacheTTL user/admin 状态 Redis 缓存时长。
	userStatusCacheTTL  = 60 * time.Second
	ctxKeyUserID        = "user_id"
	ctxKeyAdminID       = "admin_id"
	ctxKeySensitive     = "sensitive"
)

// UserAuth 校验用户 JWT，注入 user_id。
// 若 sensitive=true 且 Redis 故障，则返回 503。
func UserAuth(rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			// 尝试从 HttpOnly Cookie 取
			if cookie, err := c.Cookie("access_token"); err == nil {
				token = cookie
			}
		}
		if token == "" {
			Fail(c, errs.ErrUnauth)
			return
		}

		claims, err := pkgjwt.Parse(jwtCfg, token)
		if err != nil || claims.Typ != "user" {
			Fail(c, errs.ErrUnauth)
			return
		}

		// 检查 JWT 黑名单
		blacklisted, redisErr := isBlacklisted(c.Request.Context(), rdb, claims.JTI)
		if redisErr != nil {
			sensitive, _ := c.Get(ctxKeySensitive)
			if sensitive == true {
				Fail(c, errs.ErrServiceDegraded)
				return
			}
			// 读接口 Redis 故障放过
		} else if blacklisted {
			Fail(c, errs.ErrUnauth)
			return
		}

		// 查 user.status（60s Redis 缓存）
		status, statusErr := getUserStatus(c.Request.Context(), rdb, db, claims.Sub)
		if statusErr != nil {
			sensitive, _ := c.Get(ctxKeySensitive)
			if sensitive == true {
				Fail(c, errs.ErrServiceDegraded)
				return
			}
		} else if status != "active" {
			Fail(c, errs.ErrAccountLocked)
			return
		}

		c.Set(ctxKeyUserID, claims.Sub)
		c.Next()
	}
}

// UserOptionalAuth 可选用户认证：token 有效则注入 user_id，无 token 或无效时直接放行。
func UserOptionalAuth(rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			if cookie, err := c.Cookie("access_token"); err == nil {
				token = cookie
			}
		}
		if token == "" {
			c.Next()
			return
		}

		claims, err := pkgjwt.Parse(jwtCfg, token)
		if err != nil || claims.Typ != "user" {
			c.Next()
			return
		}

		blacklisted, _ := isBlacklisted(c.Request.Context(), rdb, claims.JTI)
		if blacklisted {
			c.Next()
			return
		}

		status, statusErr := getUserStatus(c.Request.Context(), rdb, db, claims.Sub)
		if statusErr == nil && status == "active" {
			c.Set(ctxKeyUserID, claims.Sub)
		}
		c.Next()
	}
}
func AdminAuth(rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config, perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			Fail(c, errs.ErrUnauth)
			return
		}

		claims, err := pkgjwt.Parse(jwtCfg, token)
		if err != nil || claims.Typ != "admin" {
			Fail(c, errs.ErrUnauth)
			return
		}

		// 检查黑名单
		blacklisted, redisErr := isBlacklisted(c.Request.Context(), rdb, claims.JTI)
		if redisErr != nil {
			sensitive, _ := c.Get(ctxKeySensitive)
			if sensitive == true {
				Fail(c, errs.ErrServiceDegraded)
				return
			}
		} else if blacklisted {
			Fail(c, errs.ErrUnauth)
			return
		}

		// 查 admin.status（60s 缓存）
		status, statusErr := getAdminStatus(c.Request.Context(), rdb, db, claims.Sub)
		if statusErr != nil {
			sensitive, _ := c.Get(ctxKeySensitive)
			if sensitive == true {
				Fail(c, errs.ErrServiceDegraded)
				return
			}
		} else if status != "active" {
			Fail(c, errs.ErrForbidden)
			return
		}

		// 检查权限点
		if len(perms) > 0 {
			// super_admin 旁路：跳过权限点检查
			for _, r := range claims.Roles {
				if r == "super_admin" {
					c.Set(ctxKeyAdminID, claims.Sub)
					c.Next()
					return
				}
			}
			permSet := make(map[string]struct{}, len(claims.Perms))
			for _, p := range claims.Perms {
				permSet[p] = struct{}{}
			}
			for _, required := range perms {
				if _, ok := permSet[required]; !ok {
					Fail(c, errs.ErrForbidden)
					return
				}
			}
		}

		c.Set(ctxKeyAdminID, claims.Sub)
		c.Next()
	}
}

// MarkSensitive 标记路由为敏感接口（Redis 故障时返回 503）。
func MarkSensitive() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ctxKeySensitive, true)
		c.Next()
	}
}

// Fail 调用 server 包的 Fail，但 auth 中间件直接写响应。
func Fail(c *gin.Context, err *errs.AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, gin.H{
		"code":       err.Code,
		"message":    err.Message,
		"data":       nil,
		"request_id": c.GetString("request_id"),
	})
}

func extractBearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// isBlacklisted 检查 JTI 是否在 Redis 黑名单中。
func isBlacklisted(ctx context.Context, rdb *redis.Client, jti string) (bool, error) {
	result, err := rdb.Exists(ctx, fmt.Sprintf("jwt:bl:%s", jti)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// getUserStatus 查询 user 状态，优先读 Redis 缓存（60s）。
func getUserStatus(ctx context.Context, rdb *redis.Client, db *gorm.DB, userID int64) (string, error) {
	key := fmt.Sprintf("user:status:%d", userID)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}
	if err != redis.Nil {
		return "", err
	}

	var status string
	if err := db.WithContext(ctx).
		Table("user").
		Select("status").
		Where("id = ?", userID).
		Scan(&status).Error; err != nil {
		return "", err
	}
	if status == "" {
		return "", errs.ErrNotFound
	}

	_ = rdb.SetEx(ctx, key, status, userStatusCacheTTL)
	return status, nil
}

// getAdminStatus 查询 admin 状态，优先读 Redis 缓存（60s）。
func getAdminStatus(ctx context.Context, rdb *redis.Client, db *gorm.DB, adminID int64) (string, error) {
	key := fmt.Sprintf("admin:status:%d", adminID)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}
	if err != redis.Nil {
		return "", err
	}

	var status string
	if err := db.WithContext(ctx).
		Table("admin").
		Select("status").
		Where("id = ? AND deleted_at IS NULL", adminID).
		Scan(&status).Error; err != nil {
		return "", err
	}
	if status == "" {
		return "", errs.ErrNotFound
	}

	_ = rdb.SetEx(ctx, key, status, userStatusCacheTTL)
	return status, nil
}

// AddToBlacklist 将 JTI 加入 Redis 黑名单，到期时间与 token 保持一致。
func AddToBlacklist(ctx context.Context, rdb *redis.Client, jti string, ttl time.Duration) error {
	return rdb.SetEx(ctx, fmt.Sprintf("jwt:bl:%s", jti), "1", ttl).Err()
}

// InvalidateUserStatusCache 清除 user 状态缓存（admin 修改用户状态后调用）。
func InvalidateUserStatusCache(ctx context.Context, rdb *redis.Client, userID int64) error {
	return rdb.Del(ctx, fmt.Sprintf("user:status:%d", userID)).Err()
}

// InvalidateAdminStatusCache 清除 admin 状态缓存。
func InvalidateAdminStatusCache(ctx context.Context, rdb *redis.Client, adminID int64) error {
	return rdb.Del(ctx, fmt.Sprintf("admin:status:%d", adminID)).Err()
}

// RequireHTTPS 强制 HTTPS，非 HTTPS 时返回 400。
func RequireHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil && c.GetHeader("X-Forwarded-Proto") != "https" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    10001,
				"message": "请使用 HTTPS",
			})
			return
		}
		c.Next()
	}
}
