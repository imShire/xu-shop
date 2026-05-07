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
		auth.POST("/bind-phone", userAuth, sensitive, h.BindPhone)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", userAuth, sensitive, h.Logout)
		auth.POST("/sms/send", h.SendSmsCode)
		auth.POST("/phone-register", h.PhoneRegister)
		auth.POST("/phone-login", h.PhoneLogin)
		auth.POST("/reset-password", h.ResetPassword)

		// GET /c/me 可选认证：已登录返回用户信息，未登录返回 null（200）
		optionalAuth := middleware.UserOptionalAuth(rdb, db, jwtCfg)
		c.GET("/me", optionalAuth, h.GetMe)

		me := c.Group("/me", userAuth)
		me.PUT("", h.UpdateMe)
		me.POST("/deactivate", sensitive, h.RequestDeactivate)
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
		admins.POST("/:id/disable", adminAuth("system.admin.disable"), h.DisableAdmin)
		admins.POST("/:id/enable", adminAuth("system.admin.enable"), h.EnableAdmin)
		admins.POST("/:id/reset-pwd", adminAuth("system.admin.reset_pwd"), sensitive, h.ResetAdminPwd)

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
		users.POST("/:id/disable", adminAuth("user.disable"), h.AdminDisableUser)
		users.POST("/:id/enable", adminAuth("user.enable"), h.AdminEnableUser)
		users.POST("/:id/recharge", adminAuth("user.recharge"), h.AdminRechargeBalance)
		users.GET("/:id/balance-logs", adminAuth("user.view"), h.AdminListBalanceLogs)
	}
}
