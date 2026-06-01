package reconciliation

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册对账差异 admin 路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perms ...string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perms...)
	}
	g := r.Group("/admin/reconciliation/diff")
	g.GET("", adminAuth("system.reconciliation.view"), h.ListDiffs)
	g.POST("/:id/ack", adminAuth("system.reconciliation.handle"), h.AckDiff)
	g.POST("/:id/resolve", adminAuth("system.reconciliation.handle"), h.ResolveDiff)
}
