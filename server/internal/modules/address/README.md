# modules/address

> 对应 PRD：`docs/prd/02-address.md`
> 对应 arch：`docs/arch/02-address.md`（v1.1 含 #17 region 国统局编码）

## 实施阶段
阶段 1

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `Address`、`Region` |
| `dto.go` | 请求/响应 |
| `repo.go` | CRUD + region 树 |
| `service.go` | 默认地址唯一性、上限 20、微信 chooseAddress 解密匹配 |
| `handler.go` | |
| `router.go` | C 端 + open（region 字典） |
| `service_test.go` | |

## API
- `/c/addresses[/:id]` CRUD + `/default` + `/decrypt-wx`
- `/open/regions`（公开，CDN 缓存 24h）
