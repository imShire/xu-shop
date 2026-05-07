# modules/aftersale

> 对应 PRD：`docs/prd/09-aftersale.md`
> 对应 arch：`docs/arch/09-aftersale.md`

## 实施阶段
阶段 4

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `dto.go` | |
| `repo.go` | 复用 order 表 + cancel_request_* 字段 |
| `service.go` | 用户申请取消、撤回、客服同意（→ payment.ApplyRefund）、客服拒绝、客服直接退款 |
| `handler.go` | |
| `router.go` | C + Admin |
| `service_test.go` | |

## API
- C：`/c/orders/:id/cancel-request[/withdraw]`、`/c/orders/:id/cancel`
- Admin：`/admin/aftersales`、`/admin/aftersales/:order_id/(approve|reject)`

## 依赖
- `modules/payment`（退款）、`modules/order`（状态切换）
