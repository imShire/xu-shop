package marketing

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	"github.com/xushop/xu-shop/internal/modules/marketing/coupon"
	"github.com/xushop/xu-shop/internal/modules/marketing/point"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册营销聚合模块路由（c 端 + 后台）。
//
// 调用方在 cmd/api/main.go 中：
//
//	mkSvc := marketing.NewService(db)
//	marketing.RegisterRoutes(v1, mkSvc, rdb, db, jwtCfg)
func RegisterRoutes(rg *gin.RouterGroup, svc *Service, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	couponH := coupon.NewHandler(svc.Coupon)
	pointH := point.NewHandler(svc.Point)

	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// ---- C 端 ----
	c := rg.Group("/c")
	{
		// 领券中心（公开）
		c.GET("/coupons/list", couponH.CPublicList)

		// 需登录
		auth := c.Group("", userAuth)
		auth.POST("/coupons/claim", couponH.CClaim)
		auth.POST("/coupons/redeem", couponH.CClaimByCode)
		auth.GET("/coupons/my", couponH.CMyList)

		auth.GET("/points/balance", pointH.CBalance)
		auth.GET("/points/history", pointH.CHistory)
	}

	// ---- 后台 ----
	// 优惠券模板
	rg.GET("/admin/coupons/templates", adminAuth("marketing.coupon.view"), couponH.AdminListTemplates)
	rg.GET("/admin/coupons/templates/:id", adminAuth("marketing.coupon.view"), couponH.AdminGetTemplate)

	// 积分调整工单
	rg.POST("/admin/points/tickets", adminAuth("marketing.point.edit"), pointH.AdminCreateTicket)
	rg.POST("/admin/points/tickets/:id/approve", adminAuth("marketing.point.approve"), pointH.AdminApproveTicket)
	rg.POST("/admin/points/tickets/:id/reject", adminAuth("marketing.point.approve"), pointH.AdminRejectTicket)
}
