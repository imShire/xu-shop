package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const idempotencyTTL = 24 * time.Hour

// Idempotency 幂等中间件，基于 Idempotency-Key 请求头。
// 相同 key 24h 内返回 409 Conflict（防止重复提交）。
func Idempotency(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		redisKey := "idem:" + key
		set, err := rdb.SetNX(c.Request.Context(), redisKey, "1", idempotencyTTL).Result()
		if err != nil {
			// Redis 故障时放过，不阻断主流程
			c.Next()
			return
		}
		if !set {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"code":       10005,
				"message":    "重复请求，Idempotency-Key 已使用",
				"data":       nil,
				"request_id": c.GetString("request_id"),
			})
			return
		}
		c.Next()
	}
}
