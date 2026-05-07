# modules/inventory

> 对应 PRD：`docs/prd/04-inventory.md`
> 对应 arch：`docs/arch/04-inventory.md`（v1.1 含 #13 手动调整一致性、#24 reconciler 冷启动判断）

## 实施阶段
阶段 2

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `InventoryLog`、`LowStockAlert` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 锁定 / 释放 / 扣减（依赖 pkg/stock 的 Lua 脚本）、手动 in/out/set 三模式、批量 CSV 导入 |
| `handler.go` | |
| `router.go` | 后台 |
| `reconciler.go` | 启动 + 每日 04:00 一次性比对 Redis 与 DB |
| `service_test.go` | 100 并发抢 10 件断言无超卖 |

## API
- Admin：`/admin/skus/:id/adjust`、`/admin/skus/import`、`/admin/inventory/(logs|alerts)`

## 依赖
- `pkg/stock`：三个 Lua 脚本（lock/unlock/deduct）
