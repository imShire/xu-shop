// Package domain 定义领域事件（asynq 任务 payload）。
package domain

import "time"

// OrderPaidEvent 订单支付成功事件。
type OrderPaidEvent struct {
	OrderID   int64     `json:"order_id"`
	PaymentID int64     `json:"payment_id"`
	PaidAt    time.Time `json:"paid_at"`
}

// OrderShippedEvent 订单发货事件。
type OrderShippedEvent struct {
	OrderID    int64     `json:"order_id"`
	ShipmentID int64     `json:"shipment_id"`
	ShippedAt  time.Time `json:"shipped_at"`
}

// OrderCompletedEvent 订单完成事件。
type OrderCompletedEvent struct {
	OrderID     int64     `json:"order_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// OrderCancelledEvent 订单取消事件。
type OrderCancelledEvent struct {
	OrderID     int64     `json:"order_id"`
	CancelledAt time.Time `json:"cancelled_at"`
	Reason      string    `json:"reason"`
}

// UserDeactivateEvent 用户注销定时任务事件。
type UserDeactivateEvent struct {
	UserID int64 `json:"user_id"`
}
