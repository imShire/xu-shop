# internal/pkg

> 跨模块复用的工具库 + 三方厂商封装。每个 pkg 提供 `client.go` + 接口 + `mock.go`。

## 工具类（无外部依赖）

| pkg | 用途 | 阶段 |
| --- | --- | --- |
| `errs` | 业务错误码定义 + HTTP 状态映射（参考 `docs/arch/90-api-conventions.md` 错误码表） | 0 |
| `logger` | zap 封装，注入 request_id / trace_id | 0 |
| `pagination` | 通用分页参数解析 + 响应包 | 0 |
| `validator` | go-playground/validator 注册中文错误信息 | 0 |
| `csv` | `SafeWriter` — CSV 公式注入防御（v1.1 #27） | 5 |
| `snowflake` | ID 生成器，worker_id 通过 `INSTANCE_ID` 或 Redis 抢号（v1.1 #28） | 0 |
| `lock` | Redis 分布式锁（基于 redsync 或自实现 SETNX + Lua 释放） | 1 |
| `jwt` | JWT 编解码 + jti 黑名单封装 | 1 |
| `tracer` | OpenTelemetry 初始化 | 0 |

## 三方厂商封装（依赖外部 API）

| pkg | 用途 | 阶段 | 关键文件 |
| --- | --- | --- | --- |
| `wxpay` | 微信支付 V3（基于官方 wechatpay-go） | 4 | `client.go`、`prepay.go`、`refund.go`、`notify.go`、`query.go`、`bill.go`、`mock.go` |
| `wxlogin` | 小程序 code2Session、公众号 oauth2、加密数据解密 | 1 | `client.go`、`session.go`、`oauth.go`、`crypto.go`、`mock.go` |
| `wxsubscribe` | 订阅消息 / 模板消息 + access_token 缓存 | 5 | `client.go`、`token.go`、`send.go`、`mock.go` |
| `qywx` | 企微：通讯录、客户联系（渠道码）、群机器人 webhook、回调加解密 | 5 | `client.go`、`contact.go`、`tag.go`、`robot.go`、`crypto.go`、`mock.go` |
| `kdniao` | 快递鸟：电子面单 / 轨迹订阅 / 即时查询 / 推送解析；内置 token bucket（v1.1 #16） | 4 | `client.go`、`waybill.go`、`track.go`、`subscribe.go`、`push.go`、`ratelimit.go`、`mock.go` |
| `oss` | 阿里云 OSS：上传、签名 URL、STS 临时凭证（v1.1 #22） | 2 | `client.go`、`upload.go`、`sts.go`、`mock.go` |
| `poster` | 海报渲染（gg / freetype），合成商品图 + 小程序码 | 5 | `renderer.go`、`templates.go` |
| `stock` | Redis Lua 库存原子脚本（lock / unlock / deduct / set） | 2 | `client.go`、`lua/lock.lua`、`lua/unlock.lua`、`lua/deduct.lua`、`lua/set.lua` |

## 实施约定
- 三方 pkg 必须实现接口，业务模块只通过接口注入，便于 mock。
- 失败重试与指数退避在 pkg 内实现，业务层只关心结果。
- 所有外部超时默认 5 秒，可通过配置覆盖。
