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
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}
	sensitive := middleware.MarkSensitive()

	// ===== C 端 =====
	c := r.Group("/c/aftersales")
	c.Use(userAuth)
	{
		c.POST("", h.UserApply)
		c.GET("", h.UserList)
		c.GET("/:id", h.UserGet)
		c.POST("/:id/cancel", h.UserCancel)
		c.POST("/:id/express", h.UserFillExpress)
		c.POST("/:id/messages", h.UserAppendMessage)
	}

	// ===== Admin =====
	admin := r.Group("/admin/aftersales")
	{
		admin.GET("", adminAuth("aftersale.view"), h.AdminList)
		admin.GET("/:id", adminAuth("aftersale.view"), h.AdminGet)
		admin.POST("/:id/agree", sensitive, adminAuth("aftersale.process"), h.AdminAgree)
		admin.POST("/:id/reject", sensitive, adminAuth("aftersale.process"), h.AdminReject)
		admin.POST("/:id/confirm-received", sensitive, adminAuth("aftersale.process"), h.AdminConfirmReceived)
		admin.POST("/:id/close", sensitive, adminAuth("aftersale.process"), h.AdminClose)
		admin.POST("/:id/messages", adminAuth("aftersale.process"), h.AdminAppendMessage)
	}

	// ===== Legacy cancel-request（v1.3 之前的旧路径迁移到 /admin/orders）=====
	legacy := r.Group("/admin/orders")
	{
		legacy.GET("/cancel-requests", adminAuth("aftersale.view"), h.AdminLegacyList)
		legacy.POST("/:id/cancel-request/approve",
			sensitive, adminAuth("aftersale.process"), h.AdminLegacyApprove)
		legacy.POST("/:id/cancel-request/reject",
			sensitive, adminAuth("aftersale.process"), h.AdminLegacyReject)
	}
}
