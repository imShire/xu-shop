# modules/shipping

> 对应 PRD：`docs/prd/08-shipping.md`
> 对应 arch：`docs/arch/08-shipping.md`（v1.1 含 #16 token bucket）

## 实施阶段
阶段 4

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `entity.go` | `SenderAddress`、`Carrier`、`Shipment`、`ShipmentTrack` |
| `dto.go` | |
| `repo.go` | |
| `service.go` | 单单发货、批量发货任务、改单、合并 PDF、Webhook 处理、兜底拉取、状态映射 |
| `handler.go` | |
| `router.go` | Admin / Notify / C（轨迹查询） |
| `event.go` | ShipmentCreated / ShipmentDelivered |
| `service_test.go` | 单单/批量、限流排队 |

## API
- Admin：`/admin/sender-addresses`、`/admin/carriers`、`/admin/orders/:id/ship`、`/admin/orders/batch-ship[/:task_id[/pdf]]`、`/admin/shipments[/:id]`
- C：`/c/orders/:id/tracks`
- Notify：`/notify/express`

## 依赖
- `pkg/kdniao`（含 token bucket）
- `pkg/oss`（面单 PDF 存储）
- `pdfcpu`（合并）
