package product

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册商品模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	userOptionalAuth := middleware.UserOptionalAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端 - 公开
	r.GET("/c/categories", h.ListCategories)
	r.GET("/c/products", h.ListProducts)
	r.GET("/c/products/:id", userOptionalAuth, h.GetProduct)

	// C 端 - 需要用户登录
	r.POST("/c/favorites/:product_id", userAuth, h.AddFavorite)
	r.DELETE("/c/favorites/:product_id", userAuth, h.RemoveFavorite)
	r.GET("/c/favorites", userAuth, h.ListFavorites)
	r.GET("/c/view-history", userAuth, h.GetViewHistory)
	r.DELETE("/c/view-history", userAuth, h.ClearViewHistory)

	// 后台 - 分类管理
	r.GET("/admin/categories", adminAuth("category.view"), h.AdminListCategories)
	r.POST("/admin/categories", adminAuth("category.create"), h.AdminCreateCategory)
	r.PUT("/admin/categories/:id", adminAuth("category.edit"), h.AdminUpdateCategory)
	r.DELETE("/admin/categories/:id", adminAuth("category.delete"), h.AdminDeleteCategory)

	// 后台 - 商品管理
	r.GET("/admin/products", adminAuth("product.view"), h.AdminListProducts)
	r.POST("/admin/products", adminAuth("product.create"), h.AdminCreateProduct)
	r.GET("/admin/products/:id", adminAuth("product.view"), h.AdminGetProduct)
	r.PUT("/admin/products/:id", adminAuth("product.edit"), h.AdminUpdateProduct)
	r.DELETE("/admin/products/:id", adminAuth("product.delete"), h.AdminDeleteProduct)
	r.POST("/admin/products/:id/copy", adminAuth("product.create"), h.AdminCopyProduct)
	r.POST("/admin/products/:id/onsale", adminAuth("product.edit"), h.AdminOnSale)
	r.POST("/admin/products/:id/offsale", adminAuth("product.edit"), h.AdminOffSale)
	r.POST("/admin/products/batch-status", adminAuth("product.edit"), h.AdminBatchStatus)
	r.POST("/admin/skus/batch-price", adminAuth("product.edit"), h.AdminBatchPrice)
	r.POST("/admin/upload/image", adminAuth("product.create"), h.AdminUploadImage)
}
