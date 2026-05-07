# modules/order

> 对应 PRD：`docs/prd/06-order.md`
> 对应 arch：`docs/arch/06-order.md`（v1.1 含 #11 idem 时间窗口）

## 实施阶段
阶段 3

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `Order`、`OrderItem`、`OrderLog`、`OrderRemark`、`FreightTemplate` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 创建订单、状态机 Transition、运费计算、超时关单、自动确认、再次购买、导出 |
| `handler.go` | |
| `router.go` | C + Admin |
| `event.go` | OrderCreated / OrderPaid / OrderShipped / OrderCancelled / OrderCompleted |
| `state_machine.go` | 13 条迁移规则 + 守卫 |
| `freight.go` | 运费规则 |
| `service_test.go` | 状态机全覆盖 |

## API
- C：`/c/orders` POST GET、`/c/orders/:id`、`/c/orders/:id/cancel|confirm|buy-again`、`/c/orders/:id/cancel-request[/withdraw]`
- Admin：`/admin/orders[/:id]`、`/admin/orders/export`、`/admin/orders/:id/(cancel|remarks)`、`/admin/freight-templates`

## 依赖
- 调用 `inventory.LockStock` / `inventory.ReleaseStock`
- 发出事件供 notification / stats / private_domain 订阅
