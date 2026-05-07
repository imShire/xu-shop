# modules/payment

> 对应 PRD：`docs/prd/07-payment.md`（v1.1 部分退款放开）
> 对应 arch：`docs/arch/07-payment.md`（v1.1 含 #1 金额不一致退款、#7 scene 校验、#9 部分退款、#12 限速、#18 prepay 复用）

## 实施阶段
阶段 4

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `Payment`、`Refund`、`ReconciliationDiff` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | Prepay（scene 三元组校验、复用 prepay）、HandleNotify（金额不符自动退）、ApplyRefund（部分退款支持）、HandleRefundNotify、ActiveQuery、Reconcile |
| `handler.go` | |
| `router.go` | C / Admin / Notify |
| `service_test.go` | 重复回调幂等、金额不符自动退、部分退款累计 |

## API
- C：`/c/pay/wxpay/prepay`、`/c/orders/:id/pay-status`
- Admin：`/admin/orders/:id/refund`、`/admin/payments`、`/admin/refunds`、`/admin/reconciliation`
- Notify：`/notify/wxpay`、`/notify/wxpay/refund`

## 依赖
- `pkg/wxpay`
