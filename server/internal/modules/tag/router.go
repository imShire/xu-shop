package tag

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册标签模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	r.GET("/admin/user-tags", adminAuth("tag.dict.view"), h.AdminListTags)
	r.POST("/admin/user-tags", adminAuth("tag.dict.edit"), h.AdminCreateTag)
	r.PUT("/admin/user-tags/:code", adminAuth("tag.dict.edit"), h.AdminUpdateTag)
	r.DELETE("/admin/user-tags/:code", adminAuth("tag.dict.edit"), h.AdminDeleteTag)

	r.GET("/admin/users/:user_id/tags", adminAuth("user.view"), h.AdminGetUserTags)
	r.POST("/admin/users/:user_id/tags", adminAuth("tag.user.edit"), h.AdminAddUserTag)
	r.DELETE("/admin/users/:user_id/tags/:tag_code", adminAuth("tag.user.edit"), h.AdminRemoveUserTag)

	r.POST("/admin/audience/preview", adminAuth("tag.audience.view"), h.AdminPreviewAudience)
}
