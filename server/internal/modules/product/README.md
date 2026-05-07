# modules/product

> 对应 PRD：`docs/prd/03-product.md`
> 对应 arch：`docs/arch/03-product.md`（v1.1 含 #4 富文本严格校验、#21 ETag、#22 STS、#26 双产物）

## 实施阶段
阶段 2

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `Category`、`Product`、`ProductSpec`、`ProductSpecValue`、`SKU`、`UserFavorite`、`UserViewHistory` |
| `dto.go` | 商品列表、详情（C/Admin）、批量改价/改库存请求 |
| `repo.go` | 含 pg_trgm 搜索查询 |
| `service.go` | 商品 CUD、复制、上下架、SKU 笛卡尔积生成、批量、富文本校验、双产物生成 |
| `handler.go` | |
| `router.go` | C 端 + 后台 |
| `event.go` | ProductOnsale / ProductOffsale（影响购物车失效检查） |
| `richtext.go` | bluemonday 净化 + 图片白名单正则 |
| `service_test.go` | |

## API
- C：`/c/categories`、`/c/products[/:id]`、`/c/favorites/:product_id`、`/c/view-history`
- Admin：`/admin/categories`、`/admin/products`、`/admin/products/:id/(copy|onsale|offsale)`、`/admin/skus/batch-price`、`/admin/skus/batch-stock`、`/admin/upload/(sts|confirm)`

## 依赖
- `pkg/oss`（含 STS 签发）
- bluemonday（XSS 净化）
