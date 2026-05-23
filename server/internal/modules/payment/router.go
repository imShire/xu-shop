package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册支付模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// C 端支付路由
	// MarkSensitive() 必须在 userAuth 之前，否则 auth 中间件读取 sensitive flag 时仍为 false
	r.POST("/c/pay/wxpay/prepay", middleware.MarkSensitive(), userAuth, h.Prepay)
	r.GET("/c/orders/:id/pay-status", userAuth, h.QueryPayStatus)

	// 微信回调（无 auth）
	r.POST("/notify/wxpay", h.WxpayNotify)
	r.POST("/notify/wxpay/refund", h.WxpayRefundNotify)

	// 后台支付路由
	r.POST("/admin/orders/:id/refund", middleware.MarkSensitive(), adminAuth("refund.create"), h.AdminApplyRefund)
	r.GET("/admin/payments", adminAuth("payment.view"), h.AdminListPayments)
	r.GET("/admin/refunds", adminAuth("payment.view"), h.AdminListRefunds)
	r.GET("/admin/reconciliation", adminAuth("reconcile.view"), h.AdminListDiffs)
	r.POST("/admin/reconciliation/:id/resolve", adminAuth("reconcile.view"), h.AdminResolveDiff)
	r.GET("/admin/audit-logs", adminAuth("system.audit.view"), h.AdminListAuditLogs)
}
