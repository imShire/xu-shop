# internal/config

> 对应文档：`docs/arch/00-overview.md`、`docs/arch/15-nonfunctional.md`、`server/env.example`

## 文件清单

| 文件 | 内容 |
| --- | --- |
| `config.go` | 顶层 `Config` struct + `Load()`，使用 viper / envconfig 从环境变量加载 |
| `secret.go` | 敏感配置 AES-GCM 加解密（`APP_SECRET_KEY`） |

## 实施提示

- 12-factor，不读文件，只读 env
- Load 失败必须 fatal，禁止空配置启动
- 启动时打印一次配置（敏感字段脱敏）
