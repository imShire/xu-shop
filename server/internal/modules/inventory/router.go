package inventory

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册库存模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.POST("/admin/skus/:id/adjust", adminAuth("inventory.adjust"), h.Adjust)
	r.GET("/admin/inventory/logs", adminAuth("inventory.view"), h.ListLogs)
	r.GET("/admin/inventory/alerts", adminAuth("inventory.view"), h.ListAlerts)
	r.POST("/admin/inventory/alerts/:id/read", adminAuth("inventory.view"), h.MarkAlertRead)
	r.POST("/admin/inventory/alerts/read-all", adminAuth("inventory.view"), h.MarkAllAlertsRead)
}
