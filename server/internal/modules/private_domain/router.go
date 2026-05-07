package private_domain

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册私域模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// 渠道码管理
	r.GET("/admin/channel-codes", adminAuth("channel.view"), h.AdminListChannelCodes)
	r.POST("/admin/channel-codes", adminAuth("channel.create"), h.AdminCreateChannelCode)
	r.PUT("/admin/channel-codes/:id", adminAuth("channel.edit"), h.AdminUpdateChannelCode)
	r.DELETE("/admin/channel-codes/:id", adminAuth("channel.delete"), h.AdminDeleteChannelCode)
	r.GET("/admin/channel-codes/:id/stats", adminAuth("channel.view"), h.AdminGetChannelCodeStats)

	// 标签管理
	r.GET("/admin/tags", adminAuth("tag.view"), h.AdminListTags)
	r.POST("/admin/tags", adminAuth("tag.create"), h.AdminCreateTag)
	r.PUT("/admin/tags/:id", adminAuth("tag.edit"), h.AdminUpdateTag)
	r.DELETE("/admin/tags/:id", adminAuth("tag.delete"), h.AdminDeleteTag)

	// 用户标签
	r.GET("/admin/users/:id/tags", adminAuth("tag.view"), h.AdminGetUserTags)
	r.POST("/admin/users/:id/tags", adminAuth("tag.create"), h.AdminAddUserTag)
	r.DELETE("/admin/users/:id/tags/:tag_id", adminAuth("tag.create"), h.AdminRemoveUserTag)

	// C 端
	r.POST("/c/share/visit", userAuth, h.RecordShareVisit)
	r.POST("/c/products/:id/poster", userAuth, h.GeneratePoster)
	r.GET("/c/share/resolve", h.ResolveShareScene)

	// 企微回调
	r.POST("/notify/qywx", h.HandleQYWXCallback)
}
