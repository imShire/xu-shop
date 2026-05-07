package aftersale

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册售后模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.GET("/admin/aftersales", adminAuth("aftersale.view"), h.AdminListAftersales)
	r.POST("/admin/aftersales/:order_id/approve", adminAuth("aftersale.process"), middleware.MarkSensitive(), h.AdminApproveCancel)
	r.POST("/admin/aftersales/:order_id/reject", adminAuth("aftersale.process"), h.AdminRejectCancel)
}
