# internal/domain

> 对应文档：`docs/arch/00-overview.md`

## 职责

跨模块共享的纯类型，**不含业务逻辑**。

## 文件清单

| 文件 | 内容 |
| --- | --- |
| `enums.go` | 订单状态、支付状态、库存变更类型、通知渠道等 string 枚举 |
| `events.go` | 领域事件类型定义（OrderCreated、OrderPaid、OrderShipped、OrderCancelled、RefundSuccess、LowStockAlert 等） |

## 实施提示

- 禁止 import 业务模块，避免循环
- 事件用 struct 表达，发布订阅由 `pkg/eventbus` 或 Redis Stream + asynq 实现
