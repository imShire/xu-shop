package stats

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册数据看板路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.GET("/admin/stats/overview", adminAuth("stats.view"), h.AdminOverview)
	r.GET("/admin/stats/sales-trend", adminAuth("stats.view"), h.AdminSalesTrend)
	r.GET("/admin/stats/category-pie", adminAuth("stats.view"), h.AdminCategoryPie)
	r.GET("/admin/stats/products", adminAuth("stats.view"), h.AdminTopProducts)
	r.GET("/admin/stats/users", adminAuth("stats.view"), h.AdminUserStats)
	r.GET("/admin/stats/channels", adminAuth("stats.view"), h.AdminChannelStats)
	r.GET("/admin/stats/products/export", adminAuth("stats.export"), h.AdminExportProducts)
	r.GET("/admin/stats/workbench", adminAuth("stats.view"), h.AdminWorkbench)
}
