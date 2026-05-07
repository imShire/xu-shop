package nav_icon

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册金刚区图标模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端 - 公开
	r.GET("/c/nav-icons", h.ListActiveNavIcons)

	// 后台 - 金刚区图标管理
	r.GET("/admin/nav-icons", adminAuth("nav_icon.view"), h.AdminListNavIcons)
	r.POST("/admin/nav-icons", adminAuth("nav_icon.edit"), h.AdminCreateNavIcon)
	r.PUT("/admin/nav-icons/:id", adminAuth("nav_icon.edit"), h.AdminUpdateNavIcon)
	r.DELETE("/admin/nav-icons/:id", adminAuth("nav_icon.edit"), h.AdminDeleteNavIcon)
	r.PATCH("/admin/nav-icons/:id/toggle", adminAuth("nav_icon.edit"), h.AdminToggleNavIcon)
	r.PATCH("/admin/nav-icons/sort", adminAuth("nav_icon.edit"), h.AdminBulkSortNavIcons)
}
