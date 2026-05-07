# internal/jobs

> 对应文档：`docs/arch/06-order.md`、`docs/arch/07-payment.md`、`docs/arch/08-shipping.md`、`docs/arch/11-notification.md`、`docs/arch/13-dashboard.md`

## 任务清单

| 文件 | 任务名 | 类型 | 触发 |
| --- | --- | --- | --- |
| `order_close.go` | `order:close` | 延迟 15 分钟 | order 创建后 |
| `order_auto_confirm.go` | `order:auto-confirm` | 延迟 7 天 | shipment delivered 后 |
| `pay_reconcile.go` | `payment:reconcile` | 周期 每日 00:30 | Periodic |
| `pay_active_query.go` | `payment:active-query` | 周期 每 2 分钟 | Periodic（限速 5 QPS） |
| `express_subscribe.go` | `express:subscribe` | 即时 | shipment 创建后 |
| `express_pull.go` | `express:pull-stale` | 周期 每 2 小时 | Periodic |
| `notification_send.go` | `notification:send` | 即时 + 重试 | 业务事件触发 |
| `inventory_deduct_sync.go` | `inventory:deduct-sync` | 即时 | 支付成功后 |
| `inventory_release_sync.go` | `inventory:release-sync` | 即时 | 关单 / 取消后 |
| `stats_aggregate_daily.go` | `stats:aggregate-daily` | 周期 每日 01:00 | Periodic |
| `stats_aggregate_hourly.go` | `stats:aggregate-hourly` | 周期 每小时 | Periodic（当日增量） |
| `user_deactivate.go` | `user:deactivate-finalize` | 周期 每日 03:00 | Periodic |
| `register.go` | - | - | `Register(mux, deps)` 聚合注册 |

## 实施提示

- 所有任务 handler 必须幂等
- 重试策略统一：`asynq.MaxRetry(3)` + 指数退避
- 周期任务用 `asynq.Scheduler` 注册
