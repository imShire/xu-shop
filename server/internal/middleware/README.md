# internal/middleware

> 对应文档：`docs/arch/01-account.md`、`docs/arch/10-admin.md`、`docs/arch/15-nonfunctional.md`、`docs/arch/90-api-conventions.md`

## 文件清单

| 文件 | 用途 |
| --- | --- |
| `recovery.go` | panic 捕获 + 5xx 响应 |
| `logging.go` | 结构化访问日志（zap） + request_id 注入 |
| `tracing.go` | OpenTelemetry span |
| `cors.go` | CORS（按域名白名单） |
| `ratelimit.go` | Redis 令牌桶（IP / user_id / admin_id 维度） |
| `auth_user.go` | C 端 JWT 校验（含黑名单分级降级 + status 缓存） |
| `auth_admin.go` | 后台 JWT + 权限点校验 |
| `csrf.go` | 仅 cookie 模式启用（H5、后台） |
| `audit.go` | 后台写操作审计日志（依赖路由 meta） |
| `request_id.go` | 透传 / 生成 X-Request-Id |

## 实施提示

- auth 相关务必看 `docs/arch/01-account.md` 的"分级降级"和"H5 cookie"小节
- CSRF 仅 H5 + 后台启用，小程序 Bearer 模式不挂载该中间件
