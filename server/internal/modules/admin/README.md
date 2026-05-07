# modules/admin

> 对应 PRD：`docs/prd/10-admin.md`
> 对应 arch：`docs/arch/10-admin.md`（v1.1 含 #19 audit 实名快照）

## 实施阶段
阶段 1（基础部分） + 阶段 4-5（补全各模块入口）

## 范围
本模块只放后台**通用**子能力：
- 操作日志查看 `audit_log`
- 系统参数 `system_setting`（CRUD + 敏感字段加密）

> 业务的"管理后台 API"（商品/订单/履约/...）放在各自模块的 `router.go`。

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `AuditLog`、`SystemSetting` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 参数读写（敏感 AES）、热更新事件 |
| `handler.go` | |
| `router.go` | 后台 |
| `service_test.go` | |

## API
- `/admin/audit-logs`
- `/admin/settings/:group`、`/admin/settings/secret/:key/reveal`
