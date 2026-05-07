# internal/server

> 对应文档：`docs/arch/00-overview.md`、`docs/arch/90-api-conventions.md`

## 文件清单

| 文件 | 内容 |
| --- | --- |
| `server.go` | gin.Engine 装配、监听启停 |
| `routes.go` | 注册全部模块路由（`/api/v1/c/*`、`/api/v1/admin/*`、`/api/v1/notify/*`、`/api/v1/open/*`、`/healthz`、`/metrics`） |
| `middleware_chain.go` | 中间件挂载顺序（recovery → logging → tracing → cors → ratelimit → auth → audit） |
| `openapi_gen.go` | （oapi-codegen 生成产物，禁止手改） |

## 实施提示

- 路由注册由各模块的 `router.go` 暴露 `RegisterRoutes(rg *gin.RouterGroup, deps)` 完成；本目录只管聚合
- 反代真实 IP：必须 `engine.SetTrustedProxies(...)`，详见 arch 90「反代与真实 IP」
