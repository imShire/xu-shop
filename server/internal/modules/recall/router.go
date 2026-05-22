package recall

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册召回模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.GET("/admin/recall/campaigns", adminAuth("recall.view"), h.AdminListCampaigns)
	r.GET("/admin/recall/campaigns/:id", adminAuth("recall.view"), h.AdminGetCampaign)
	r.POST("/admin/recall/campaigns", adminAuth("recall.edit"), h.AdminCreateCampaign)
	r.PUT("/admin/recall/campaigns/:id", adminAuth("recall.edit"), h.AdminUpdateCampaign)
	r.POST("/admin/recall/campaigns/:id/transition", adminAuth("recall.edit"), h.AdminTransition)

	r.GET("/admin/recall/campaigns/:id/funnel", adminAuth("recall.view"), h.AdminFunnel)
	r.GET("/admin/recall/campaigns/:id/logs", adminAuth("recall.view"), h.AdminListLogs)
}
