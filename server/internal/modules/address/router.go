package address

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册地址模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)

	// C 端地址路由
	addresses := r.Group("/c/addresses", userAuth)
	{
		addresses.GET("", h.List)
		addresses.GET("/:id", h.Get)
		addresses.POST("", h.Create)
		addresses.PUT("/:id", h.Update)
		addresses.DELETE("/:id", h.Delete)
		addresses.POST("/:id/default", h.SetDefault)
		addresses.POST("/decrypt-wx", h.DecryptWxAddress)
	}

	// 开放行政区划接口（无鉴权）
	r.GET("/open/regions", h.ListRegions)
}
