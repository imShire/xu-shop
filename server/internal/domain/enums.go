// Package domain 定义跨模块共享的枚举常量。
package domain

// 用户状态
const (
	UserStatusActive       = "active"
	UserStatusDisabled     = "disabled"
	UserStatusDeactivating = "deactivating"
	UserStatusDeactivated  = "deactivated"
)

// 管理员状态
const (
	AdminStatusActive   = "active"
	AdminStatusDisabled = "disabled"
)

// 订单状态
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunding = "refunding"
	OrderStatusRefunded  = "refunded"
)

// 支付状态
const (
	PaymentStatusPending = "pending"
	PaymentStatusSuccess = "success"
	PaymentStatusFailed  = "failed"
)

// 物流状态
const (
	ShipmentStatusPicked     = "picked"
	ShipmentStatusTransit    = "transit"
	ShipmentStatusDelivering = "delivering"
	ShipmentStatusDelivered  = "delivered"
	ShipmentStatusException  = "exception"
)
