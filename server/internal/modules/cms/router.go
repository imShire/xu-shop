package cms

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册 CMS 模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端（无需认证）
	r.GET("/c/articles", h.ListPublished)
	r.GET("/c/articles/:id", h.GetPublished)

	// 后台 - 文章管理
	r.GET("/admin/articles", adminAuth("article.view"), h.AdminList)
	r.GET("/admin/articles/:id", adminAuth("article.view"), h.AdminGet)
	r.POST("/admin/articles", adminAuth("article.edit"), h.AdminCreate)
	r.PUT("/admin/articles/:id", adminAuth("article.edit"), h.AdminUpdate)
	r.DELETE("/admin/articles/:id", adminAuth("article.edit"), h.AdminDelete)
}
