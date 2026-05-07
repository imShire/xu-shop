package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册系统设置路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perms ...string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perms...)
	}

	admin := r.Group("/admin")
	admin.GET("/settings/upload", adminAuth("system.upload.view"), h.GetUploadSettings)
	admin.PUT("/settings/upload", adminAuth("system.upload.edit"), h.UpdateUploadSettings)
	admin.POST("/settings/upload/test", adminAuth("system.upload.edit"), h.TestUploadSettings)
	admin.POST("/settings/upload/probe", adminAuth("system.upload.edit"), h.ProbeUploadSettings)
	admin.GET("/settings/:group", adminAuth("system.setting.view"), h.GetSettings)
	admin.PUT("/settings/:group", adminAuth("system.setting.edit"), h.UpdateSettings)
}
