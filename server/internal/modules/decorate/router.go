package decorate

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册页面装修模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端（无需认证）
	r.GET("/c/page-config", h.GetActivePage)

	// 后台 - 页面装修
	r.GET("/admin/decorate/versions", adminAuth("decorate.view"), h.AdminListVersions)
	r.POST("/admin/decorate/save", adminAuth("decorate.edit"), h.AdminSave)
	r.POST("/admin/decorate/activate/:id", adminAuth("decorate.edit"), h.AdminActivate)
}
