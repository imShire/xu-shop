package shipping

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/middleware"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
)

// RegisterRoutes 注册发货模块路由。
func RegisterRoutes(r *gin.RouterGroup, h *Handler, rdb *redis.Client, db *gorm.DB, jwtCfg pkgjwt.Config) {
	userAuth := middleware.UserAuth(rdb, db, jwtCfg)
	adminAuth := func(perm string) gin.HandlerFunc {
		return middleware.AdminAuth(rdb, db, jwtCfg, perm)
	}

	// 发货地址管理
	addrPerm := adminAuth("system.setting.edit")
	r.GET("/admin/sender-addresses", adminAuth("system.setting.view"), h.ListSenderAddresses)
	r.POST("/admin/sender-addresses", addrPerm, h.CreateSenderAddress)
	r.PUT("/admin/sender-addresses/:id", addrPerm, h.UpdateSenderAddress)
	r.DELETE("/admin/sender-addresses/:id", addrPerm, h.DeleteSenderAddress)
	r.POST("/admin/sender-addresses/:id/default", addrPerm, h.SetDefaultSenderAddress)

	// 快递公司管理
	r.GET("/admin/carriers", adminAuth("system.setting.view"), h.ListCarriers)
	r.PUT("/admin/carriers/:code", adminAuth("system.setting.edit"), h.UpdateCarrier)

	// 发货操作
	r.POST("/admin/orders/:id/ship", adminAuth("shipment.ship"), h.Ship)
	r.POST("/admin/orders/batch-ship", adminAuth("shipment.batch_ship"), h.BatchShip)
	r.GET("/admin/orders/batch-ship/:task_id", adminAuth("shipment.ship"), h.GetBatchShipStatus)
	r.GET("/admin/orders/batch-ship/:task_id/pdf", adminAuth("shipment.ship"), h.GetBatchShipPDF)

	// 发货单列表 & 修改
	r.GET("/admin/shipments", adminAuth("shipment.view"), h.ListShipments)
	r.PUT("/admin/shipments/:id", adminAuth("shipment.update"), h.UpdateShipment)

	// C 端轨迹
	r.GET("/c/orders/:id/tracks", userAuth, h.ListTracks)

	// 快递鸟 Webhook（无 auth）
	r.POST("/notify/express", h.ExpressPush)
}
