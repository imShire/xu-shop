# modules/cart

> 对应 PRD：`docs/prd/05-cart.md`
> 对应 arch：`docs/arch/05-cart.md`

## 实施阶段
阶段 3

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `CartItem` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 加入合并、上限校验、precheck（结算前校验：可用性/价格/库存） |
| `handler.go` | |
| `router.go` | 仅 C 端 |
| `service_test.go` | |

## API
- `/c/cart` GET POST、`/c/cart/:id` PUT DELETE、`/c/cart/batch-delete`、`/c/cart/clean-invalid`、`/c/cart/count`、`/c/cart/precheck`
