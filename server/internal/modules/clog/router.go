package clog

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册前端日志上报路由。
//
//   - 路径：/internal/clog 与 /internal/clog/batch
//   - 不强制登录；若客户端携带 user/admin token 则解析出 id 写入 ctx（信任边界，CR Blocker #1）
//   - 通过 IP 限流（默认 60 条 / 分钟）
//   - 上层 main.go 应当在挂载 audit 中间件时通过路径前缀跳过 /internal/
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	rl := middleware.RateLimiter(rdb, "clog_submit", 60)
	userOpt := middleware.UserOptionalAuth(rdb, db, jwtCfg)
	adminOpt := middleware.AdminOptionalAuth(rdb, db, jwtCfg)
	g := r.Group("/internal/clog", rl, userOpt, adminOpt)
	g.POST("", h.Submit)
	g.POST("/batch", h.SubmitBatch)
}
