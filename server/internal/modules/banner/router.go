package banner

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册横幅模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端 - 公开
	r.GET("/c/banners", h.ListActiveBanners)

	// 后台 - 横幅管理
	r.GET("/admin/banners", adminAuth("banner.view"), h.AdminListBanners)
	r.POST("/admin/banners", adminAuth("banner.edit"), h.AdminCreateBanner)
	r.PUT("/admin/banners/:id", adminAuth("banner.edit"), h.AdminUpdateBanner)
	r.DELETE("/admin/banners/:id", adminAuth("banner.edit"), h.AdminDeleteBanner)
	r.PATCH("/admin/banners/:id/toggle", adminAuth("banner.edit"), h.AdminToggleBanner)
	r.PATCH("/admin/banners/sort", adminAuth("banner.edit"), h.AdminBulkSortBanners)
}
