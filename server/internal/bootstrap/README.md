# internal/bootstrap

> 对应文档：`docs/arch/00-overview.md`

## 职责

集中装配所有依赖单例：DB、Redis、asynq client/server、wxpay、wxlogin、kdniao、qywx、oss、logger、tracer、prometheus。

## 文件清单

| 文件 | 内容 |
| --- | --- |
| `app.go` | `App` struct（持有所有单例） + `NewApp(cfg)` 构造函数 |
| `db.go` | gorm + pg 初始化、连接池参数 |
| `redis.go` | go-redis 客户端 |
| `asynq.go` | asynq.Client + Server 配置 |
| `vendors.go` | wxpay / kdniao / qywx / oss 客户端构造（依赖 system_setting，可热更新） |
| `observability.go` | logger / tracer / metrics |

## 实施提示

- 不要在这里写业务逻辑，只做"创建 + 注入"
- 三方客户端构造失败仅日志告警，不阻断启动（除非生产）
