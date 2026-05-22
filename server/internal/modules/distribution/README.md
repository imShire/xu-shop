# distribution 模块

> 对应文档：`docs/prd/18-distribution.md` / `docs/arch/18-distribution.md`
> 实施阶段：阶段 5.4（最后一个 P1 模块）
> 依赖：order、payment（必须先实装 `pkg/wxpay.Transfer`，见 `arch/07-payment.md` v1.2）、notification、tag（来源渠道）

分享溯源 + 一级分销 + 微信商家转账提现。

## 子目录布局

```
distribution/
├── README.md
├── service.go                  # 对外门面（OnOrderPaid / OnRefund / ResolveOnOrder）
├── event.go                    # 订阅 OrderPaid/OrderRefunded
├── antifraud.go                # 防刷规则集中
├── share/
│   ├── entity.go               # ShareLink / ShareClick / ShareAttribution
│   ├── repo.go
│   ├── service.go              # CreateLink / ResolveTrace / 短链 redirect
│   ├── handler.go              # /c/share/links /c/share/poster /s/{token}
│   ├── dto.go
│   └── router.go
├── distributor/
│   ├── entity.go               # Distributor / DistributorRelation
│   ├── repo.go
│   ├── service.go              # Apply / Approve / Disable / SetRate / BindOnOrder
│   ├── handler.go              # /c/distributors/* /admin/distributors/*
│   ├── dto.go
│   └── router.go
├── commission/
│   ├── entity.go               # CommissionRecord / CommissionSettlement
│   ├── repo.go
│   ├── service.go              # Generate / Transition (8 events) / Settle
│   ├── handler.go              # /c/distributors/me/commissions /admin/commissions/*
│   ├── dto.go
│   └── router.go
└── withdraw/
    ├── entity.go               # WithdrawOrder
    ├── repo.go
    ├── service.go              # Apply (双因子) / Transition / Reconcile
    ├── handler.go              # /c/distributors/me/withdraws /admin/withdraws/* /notify/wxpay/transfer
    ├── dto.go
    └── router.go
```

## 状态机（唯一入口）

| 实体 | 函数 | 事件 |
| --- | --- | --- |
| `CommissionRecord` | `commission.Service.Transition(c, event, reason)` | `mark_suspect` / `unsuspect` / `pass_freeze` / `partial_refund` / `order_full_refund` / `manual_cancel` / `withdraw` / `withdraw_failed` |
| `WithdrawOrder` | `withdraw.Service.Transition(w, event)` | `submit` / `accepted` / `success` / `fail` / `cancel` |

## asynq jobs

| code | cron / 触发 |
| --- | --- |
| `commission:freeze-pass` | 每小时（freeze_until 到点） |
| `commission:level-recompute` | 每周一 03:00 |
| `commission:fraud-scan` | 每小时（防刷规则巡检） |
| `withdraw:requery` | 每 5min（兜底查 wxpay 转账状态） |
| `withdraw:reconcile` | 每日 04:00（与微信账单对账） |
| `share:click-flush` | 每分钟（Redis pipeline 累计落库） |

## 红线（v1.2 安全要求）
1. 提现走二次确认：sms_code 5min 有效，5 次/天，失败 5 次锁 1h。
2. 商家转账金额 > 2000 元强制实名校验。
3. 防刷规则集中在 `antifraud.go`：fp 30 天 5 次 / 同手机尾 4 / 单订单 5x 均值 / 24h 同收件人 3 单 / 退款率 70%。
4. 佣金生成走 `(order_id, distributor_user_id)` 唯一索引保证幂等。
5. 短链 token 6-10 位 base62，30 天默认 TTL，过期 404 不暴露详情。
6. trace_id 走 cookie + storage 7 天，归因取**首次**到访的 share_link。
