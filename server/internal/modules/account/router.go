package account

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册账号模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perms ...string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perms...)
	}
	sensitive := middleware.MarkSensitive()
	rateLimitCaptcha := middleware.RateLimiter(rdb, "admin_captcha", 30)
	rateLimitLogin := middleware.RateLimiter(rdb, "admin_login", 5)

	// C 端路由
	c := r.Group("/c")
	{
		auth := c.Group("/auth")
		auth.POST("/mp/login", h.MpLogin)
		auth.GET("/h5/code", h.H5GetOAuthURL)
		auth.GET("/h5/callback", h.H5Callback)
		// sensitive 必须在 userAuth 之前，否则 auth 中间件读取 sensitive flag 时仍为 false
		auth.POST("/bind-phone", sensitive, userAuth, h.BindPhone)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", sensitive, userAuth, h.Logout)
		auth.POST("/sms/send", h.SendSmsCode)
		auth.POST("/phone-register", h.PhoneRegister)
		auth.POST("/phone-login", h.PhoneLogin)
		auth.POST("/reset-password", h.ResetPassword)

		// GET /c/me 可选认证：已登录返回用户信息，未登录返回 null（200）
		optionalAuth := middleware.UserOptionalAuth(rdb, db, jwtCfg)
		c.GET("/me", optionalAuth, h.GetMe)

		me := c.Group("/me", userAuth)
		me.PUT("", h.UpdateMe)
		// 注销属敏感操作；将 sensitive 放在 group 已注入的 userAuth 之前需单独绑定路由
		c.POST("/me/deactivate", sensitive, userAuth, h.RequestDeactivate)
		me.POST("/deactivate/cancel", h.CancelDeactivate)
		me.GET("/balance", h.GetMyBalance)
	}

	// Admin 路由
	admin := r.Group("/admin")
	{
		authGrp := admin.Group("/auth")
		authGrp.POST("/captcha", rateLimitCaptcha, h.AdminGetCaptcha)
		authGrp.POST("/login", rateLimitLogin, h.AdminLogin)
		authGrp.POST("/logout", adminAuth(), h.AdminLogout)
		admin.GET("/me", adminAuth(), h.AdminGetMe)

		admins := admin.Group("/admins")
		admins.GET("", adminAuth("system.admin.view"), h.ListAdmins)
		admins.POST("", adminAuth("system.admin.create"), h.CreateAdmin)
		admins.PUT("/:id", adminAuth("system.admin.edit"), h.UpdateAdmin)
		// sensitive 必须在 adminAuth 之前
		admins.POST("/:id/disable", sensitive, adminAuth("system.admin.disable"), h.DisableAdmin)
		admins.POST("/:id/enable", sensitive, adminAuth("system.admin.enable"), h.EnableAdmin)
		// sensitive 必须在 adminAuth 之前
		admins.POST("/:id/reset-pwd", sensitive, adminAuth("system.admin.reset_pwd"), h.ResetAdminPwd)

		roles := admin.Group("/roles")
		roles.GET("", adminAuth("system.role.view"), h.ListRoles)
		roles.POST("", adminAuth("system.role.create"), h.CreateRole)
		roles.PUT("/:id", adminAuth("system.role.edit"), h.UpdateRole)
		roles.DELETE("/:id", adminAuth("system.role.delete"), h.DeleteRole)

		admin.GET("/permissions", adminAuth("system.role.view"), h.ListPermissions)

		users := admin.Group("/users")
		users.GET("", adminAuth("user.view"), h.AdminListUsers)
		users.POST("", adminAuth("user.create"), h.AdminCreateUser)
		users.GET("/:id", adminAuth("user.view"), h.AdminGetUser)
		// sensitive 必须在 adminAuth 之前
		users.POST("/:id/disable", sensitive, adminAuth("user.disable"), h.AdminDisableUser)
		users.POST("/:id/enable", sensitive, adminAuth("user.enable"), h.AdminEnableUser)
		// 充值动钱属强敏感写操作，sensitive 必须在 adminAuth 之前
		users.POST("/:id/recharge", sensitive, adminAuth("user.recharge"), h.AdminRechargeBalance)
		users.GET("/:id/balance-logs", adminAuth("user.view"), h.AdminListBalanceLogs)
	}
}
