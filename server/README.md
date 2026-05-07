# server — Go 后端

> 对应文档：`docs/arch/00-overview.md`、`docs/arch/90-api-conventions.md`、`docs/arch/91-db-schema.md`、`docs/arch/92-deployment.md`、`docs/arch/93-implementation-plan.md`、`docs/arch/99-revisions.md`

## 目录与文档对照

| 目录 | 职责 | 对应文档 |
| --- | --- | --- |
| `cmd/api/` | HTTP API 入口（Gin） | arch 00 |
| `cmd/worker/` | asynq 异步 worker 入口 | arch 00 / 11 |
| `cmd/cli/` | 运维 CLI（建超管、seed） | arch 92 |
| `internal/config/` | 12-factor 环境变量加载 | arch 00、15 |
| `internal/bootstrap/` | DB / Redis / wxpay / kdniao / oss 等单例装配 | arch 00 |
| `internal/server/` | gin 引擎、中间件注册、路由总入口 | arch 00、90 |
| `internal/middleware/` | recovery / cors / auth / ratelimit / audit / logging | arch 01、10、15、90 |
| `internal/modules/<feature>/` | 业务模块（package by feature） | arch 01-14 各对应 |
| `internal/jobs/` | asynq 任务定义 | arch 06、07、08、11、13 |
| `internal/pkg/<vendor>/` | 三方封装与通用工具 | arch 各处 |
| `internal/domain/` | 跨模块共享枚举与领域事件 | arch 00 |
| `migrations/` | goose SQL 迁移 | arch 91 |
| `api/openapi.yaml` | API 契约（唯一真相源） | arch 90 |
| `deploy/` | Dockerfile、compose、生产部署 | arch 92 |

## 实施阶段（来自 93 文档）

阶段 0 ~ 6 顺序推进。每阶段完成后跑对应验收，再进入下一阶段。

| 阶段 | 涉及目录 | 关键产出 |
| --- | --- | --- |
| 0 | 全局 | go.mod 启动、`/healthz`、`/metrics`、CI 跑通 |
| 1 | account / address + middleware + pkg(jwt/errs/logger/snowflake/lock/validator) | 登录、地址 CRUD、超管 CLI |
| 2 | product / inventory + pkg(oss) | 商品、Redis Lua 库存 |
| 3 | cart / order + jobs(order_close/auto_confirm) | 下单 + 超时关单 |
| 4 | payment / shipping / aftersale + pkg(wxpay/kdniao) + jobs(pay_reconcile/express_*) | 支付、电子面单 |
| 4.5 | - | 压测验收（必须） |
| 5 | notification / private_domain / stats + pkg(wxsubscribe/qywx/poster) | 通知、私域、看板 |
| 6 | reserved + 监控告警 + 上线 | 预留表、生产化 |

## 模块内文件约定

每个 `internal/modules/<feature>/` 默认：

```
<feature>/
├── README.md       # 该模块 PRD + arch 引用 + 实施提示
├── entity.go       # GORM model + 状态枚举
├── repo.go         # 数据访问层（gorm queries）
├── service.go      # 领域服务（事务、状态机、不变式）
├── handler.go      # HTTP handler（参数绑定 + 调 service + 序列化）
├── dto.go          # 请求 / 响应结构
├── router.go       # gin 路由注册（含权限点）
├── event.go        # 领域事件定义与订阅（如需）
└── *_test.go       # 单元 + 集成测试
```

## pkg 约定

每个 `internal/pkg/<vendor>/`：
- `client.go`：单例构造 + 配置注入
- `<api>.go`：按业务能力拆方法
- `mock.go`：测试用桩（实现同一接口）
- `README.md`：第三方接口要点 + 当前用到的能力清单

## 启动命令

详见 `Makefile`：`make run-api`、`make run-worker`、`make migrate-up`、`make migrate-down`、`make gen-openapi`、`make lint`、`make test`。
