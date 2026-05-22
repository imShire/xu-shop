package distribution

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册分销 / 分享溯源 / 提现模块的全部路由（C 端 + 后台 + 微信回调）。
//
// 调用方在 cmd/api/main.go 中：
//
//	dSvc := distribution.NewService(...)
//	dH := distribution.NewHandler(dSvc, transferClient)
//	distribution.RegisterRoutes(v1, dH, rdb, db, jwtCfg, root)
//
// 短链 /s/:token 与微信支付商家转账回调 /notify/wxpay/transfer 在根路由组挂载。
func RegisterRoutes(
	v1 *gin.RouterGroup,
	h *Handler,
	rdb *redis.Client,
	db *gorm.DB,
	jwtCfg pkgjwt.Config,
	root *gin.Engine,
) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// ---- C 端 ----
	c := v1.Group("/c")
	{
		// 公开埋点（必须能匿名上报）
		c.POST("/share/track", h.TrackShare)

		auth := c.Group("", userAuth)
		auth.POST("/share/links", h.CreateShareLink)
		auth.POST("/distributor/apply", h.ApplyDistributor)
		auth.GET("/distributor/me", h.GetMyProfile)
		auth.GET("/distributor/commissions", h.GetMyCommissions)
		auth.GET("/distributor/withdraws", h.GetMyWithdraws)
		auth.POST("/distributor/withdraw/sms", h.SendWithdrawSms)
		auth.POST("/distributor/withdraw", h.RequestWithdraw)
	}

	// ---- 后台 ----
	a := v1.Group("/admin")
	{
		a.GET("/distributors", adminAuth("distributor.view"), h.AdminListDistributors)
		a.POST("/distributors/:id/approve", adminAuth("distributor.audit"), h.AdminApproveDistributor)
		a.POST("/distributors/:id/reject", adminAuth("distributor.audit"), h.AdminRejectDistributor)
		a.POST("/distributors/:id/ban", adminAuth("distributor.audit"), h.AdminBanDistributor)
		a.POST("/distributors/:id/level", adminAuth("distributor.edit"), h.AdminAdjustLevel)
		a.POST("/distributors/:id/rate", adminAuth("distributor.edit"), h.AdminAdjustRate)

		a.GET("/share-links", adminAuth("share.view"), h.AdminListShareLinks)

		a.GET("/commissions", adminAuth("commission.view"), h.AdminListCommissions)
		a.POST("/commissions/:id/audit", adminAuth("commission.audit"), h.AdminAuditCommission)
		a.POST("/commissions/:id/release", adminAuth("commission.audit"), h.AdminAuditCommissionRelease)
		a.POST("/commissions/:id/cancel", adminAuth("commission.audit"), h.AdminAuditCommissionCancel)

		a.GET("/settlements", adminAuth("commission.view"), h.AdminListSettlements)

		a.GET("/withdraws", adminAuth("withdraw.view"), h.AdminListWithdraws)
		a.POST("/withdraws/:id/retry", adminAuth("withdraw.audit"), h.AdminRetryWithdraw)

		a.GET("/distribution/funnel", adminAuth("distribution.view"), h.AdminFunnelReport)
	}

	// ---- 根路由：短链 + 微信支付商家转账回调 ----
	root.GET("/s/:token", h.ResolveShortToken)
	root.POST("/notify/wxpay/transfer", h.NotifyWxTransfer)
}
