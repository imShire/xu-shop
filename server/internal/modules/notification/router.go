package notification

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册通知模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.GET("/admin/notifications", adminAuth("notif.view"), h.AdminListNotifications)
	r.GET("/admin/notification-templates", adminAuth("notif.config"), h.AdminListTemplates)
	r.PUT("/admin/notification-templates/:code", adminAuth("notif.config"), h.AdminUpdateTemplate)
	r.POST("/admin/notification-templates/:code/test", adminAuth("notif.config"), h.AdminTestSend)
}
