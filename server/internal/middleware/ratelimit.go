package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	redis_rate "github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"github.com/xushop/xu-shop/internal/pkg/errs"
)

// RateLimiter 返回基于 Redis token bucket 的限流中间件。
//
//   - key: 限流 key 标识（如 "admin_login" / "c_auth" / "order_post" / "global"）
//   - rate: 每分钟允许请求数
func RateLimiter(rdb *redis.Client, keyPrefix string, perMinute int) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(rdb)

	return func(c *gin.Context) {
		// 限流 key 由 前缀 + IP（或 user_id）构成
		key := keyPrefix + ":" + c.ClientIP()
		if uid, exists := c.Get("user_id"); exists {
			key = keyPrefix + ":uid:" + toString(uid)
		}

		result, err := limiter.Allow(c.Request.Context(), key, redis_rate.PerMinute(perMinute))
		if err != nil {
			// Redis 不可用时放过，避免全局雪崩
			c.Next()
			return
		}
		if result.Allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":       errs.ErrRateLimit.Code,
				"message":    errs.ErrRateLimit.Message,
				"request_id": c.GetString("request_id"),
				"data":       nil,
			})
			return
		}
		c.Next()
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
