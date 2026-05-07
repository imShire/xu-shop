package cart

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册购物车模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	rl60 := middleware.RateLimiter(rdb, "cart_add", 60)

	g := r.Group("/c/cart", userAuth)
	{
		g.GET("", h.List)
		g.POST("", rl60, h.Add)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
		g.POST("/batch-delete", h.BatchDelete)
		g.POST("/clean-invalid", h.CleanInvalid)
		g.GET("/count", h.Count)
		g.POST("/precheck", h.Precheck)
	}
}
