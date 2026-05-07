# modules/private_domain

> 对应 PRD：`docs/prd/12-private-domain.md`
> 对应 arch：`docs/arch/12-private-domain.md`（v1.1 含 #14 short code）

## 实施阶段
阶段 5

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `ChannelCode`、`CustomerTag`、`UserTag`、`ShareAttribution`、`ShareShortCode` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 渠道码生成（企微 add_contact_way）、客户标签同步、分享归因、短码 UPSERT、海报渲染调度 |
| `handler.go` | |
| `router.go` | C / Admin / Notify（企微回调） |
| `service_test.go` | |

## API
- Admin：`/admin/channel-codes`、`/admin/channel-codes/stats`、`/admin/tags`、`/admin/users/:id/tags`
- C：`/c/share/visit`、`/c/share/resolve`（短码反查）、`/c/products/:id/poster`
- Notify：`/notify/qywx`

## 依赖
- `pkg/qywx`、`pkg/poster`、`pkg/oss`
