package order

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册订单模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}
	rl10 := middleware.RateLimiter(rdb, "order_create", 10)
	idem := middleware.Idempotency(rdb)

	// C 端订单路由
	c := r.Group("/c/orders", userAuth)
	{
		c.POST("", rl10, idem, h.Create)
		c.GET("", h.List)
		c.GET("/:id", h.Get)
		c.POST("/:id/cancel", h.Cancel)
		c.POST("/:id/cancel-request/withdraw", h.WithdrawCancel)
		c.POST("/:id/confirm", h.Confirm)
		c.POST("/:id/buy-again", h.BuyAgain)
	}

	// 后台订单路由
	r.GET("/admin/orders", adminAuth("order.view"), h.AdminList)
	r.GET("/admin/orders/:id", adminAuth("order.view"), h.AdminGet)
	r.POST("/admin/orders/export", adminAuth("order.export"), h.AdminExport)
	r.POST("/admin/orders/:id/cancel", adminAuth("order.cancel"), h.AdminCancel)
	r.POST("/admin/orders/:id/remarks", adminAuth("order.remark"), h.AdminAddRemark)
	r.GET("/admin/orders/:id/remarks", adminAuth("order.view"), h.AdminListRemarks)

	// 运费模板
	r.GET("/admin/freight-templates", adminAuth("system.setting.view"), h.AdminListFreight)
	r.POST("/admin/freight-templates", adminAuth("system.setting.edit"), h.AdminCreateFreight)
	r.PUT("/admin/freight-templates/:id", adminAuth("system.setting.edit"), h.AdminUpdateFreight)
	r.DELETE("/admin/freight-templates/:id", adminAuth("system.setting.edit"), h.AdminDeleteFreight)
}
