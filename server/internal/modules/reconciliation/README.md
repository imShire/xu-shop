# reconciliation 模块（A8 日终对账差异中枢）

对应阶段：阶段 6 / 日终对账。

## 文件清单
- `entity.go`：`Diff` 实体（表 `reconciliation_diff`），含 Job/RefType/Severity/Status 常量
- `repo.go`：仓储接口与 GORM 实现，`Upsert` 采用 `ON CONFLICT DO UPDATE` 按业务日唯一键去重
- `service.go`：`RecordDiff` 供 cron 作业写入；`Acknowledge` / `Resolve` 走单一函数变更状态
- `dto.go`：响应 DTO，ID 字段统一字符串
- `handler.go` / `router.go`：admin 三接口（列表 / ack / resolve），权限点 `system.reconciliation.view` 与 `system.reconciliation.handle`
- `service_test.go` / `handler_test.go`：单元测试

## 表结构（详见 `migrations/20260531000003_reconciliation_diff_v2.sql`）

旧 `reconciliation_diff`（仅支付场景）已迁移为 `legacy_reconciliation_diff`，
新表通用化支持三类对账作业，并以 `(job, biz_date, ref_type, ref_id, field)` 唯一键去重。

## 三个对账作业（写入方）

| 作业 | 文件 | 默认 cron | 说明 |
| --- | --- | --- | --- |
| 支付对账 | `internal/jobs/reconciliation/payment.go` | `0 30 1 * * *` | 拉远端微信查单与本地 payment_record 对账 |
| 库存对账 | `internal/jobs/reconciliation/inventory.go` | `0 0 2 * * *` | sku.locked_stock vs 未发货订单合计 |
| 佣金对账 | `internal/jobs/reconciliation/commission.go` | `0 30 2 * * *` | 佣金记录 amount_cents vs 重算值 |

环境变量覆盖：`RECONCILE_PAYMENT_CRON` / `RECONCILE_INVENTORY_CRON` / `RECONCILE_COMMISSION_CRON`。

## 约束

- 对账作业 **不自动修复业务数据**（金额不一致已在支付回调走原路退款；其余仅记录差异并依靠 ack→resolve 闭环）
- 三方调用走既有 `pkg/wxpay`，不要新建客户端
- 严重程度：金额差 → `critical`；状态/数量差 → `warn`；其他 → `info`
