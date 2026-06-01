// Package reconciliation 提供三类日终对账作业：支付 / 库存 / 佣金。
// 作业入口注册到 asynq mux（任务名见常量），由 scheduler 按 cron 触发。
package reconciliation

// asynq 任务名。
const (
	TaskPayment    = "reconciliation:payment"
	TaskInventory  = "reconciliation:inventory"
	TaskCommission = "reconciliation:commission"
)

// 默认 cron 表达式（5 字段，UTC+本地视 scheduler 配置而定）。
const (
	DefaultPaymentCron    = "30 1 * * *" // 每日 01:30
	DefaultInventoryCron  = "0 2 * * *"  // 每日 02:00
	DefaultCommissionCron = "30 2 * * *" // 每日 02:30
)
