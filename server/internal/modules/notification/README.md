# modules/notification

> 对应 PRD：`docs/prd/11-notification.md`
> 对应 arch：`docs/arch/11-notification.md`（v1.1 含 #20 cooldown）

## 实施阶段
阶段 5

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `NotificationTemplate`、`NotificationTask` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | Dispatch（事件 → 任务）、Send（worker 调用）、cooldown 管理 |
| `handler.go` | |
| `router.go` | Admin |
| `subscriber.go` | 订阅领域事件（OrderPaid 等） |
| `service_test.go` | |

## API
- Admin：`/admin/notifications`、`/admin/notification-templates[/:code]`、`/admin/notification-templates/:code/test`

## 依赖
- `pkg/wxsubscribe`、`pkg/qywx`（机器人）
- asynq job：`notification:send`
