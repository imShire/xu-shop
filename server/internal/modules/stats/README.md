# modules/stats

> 对应 PRD：`docs/prd/13-dashboard.md`
> 对应 arch：`docs/arch/13-dashboard.md`

## 实施阶段
阶段 5

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `StatsDaily`、`StatsProductDaily`、`StatsChannelDaily` |
| `dto.go` | |
| `repo.go` | 预聚合表查询 |
| `service.go` | 概览查询、商品销量、用户分析、渠道分析、聚合任务实现（被 jobs 调用）、CSV 导出（用 SafeWriter） |
| `handler.go` | |
| `router.go` | Admin |
| `service_test.go` | |

## API
- `/admin/stats/(overview|sales-trend|category-pie|products|users|channels)`
- `/admin/stats/products/export`

## 依赖
- `pkg/csv`（防注入导出）
