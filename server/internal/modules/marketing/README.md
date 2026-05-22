# marketing 模块

> 对应文档：`docs/prd/16-membership.md` / `docs/arch/16-membership.md`
> 实施阶段：阶段 5.2（见 `docs/arch/93-implementation-plan.md`）
> 依赖：account、order、notification

会员体系合一包，对外暴露 `marketing.Service`，内部按子目录拆分领域。

## 子目录布局

```
marketing/
├── README.md
├── service.go                  # 对外门面（Quote/Lock/Consume/Release/Refund/...）
├── event.go                    # 订阅 OrderPaid/OrderCancelled/OrderRefunded
├── coupon/
│   ├── entity.go               # CouponTemplate / UserCoupon / CouponRedeemCode / CouponGrantTask
│   ├── repo.go
│   ├── service.go              # 状态机入口：CouponService.Transition(uc, event)
│   ├── handler.go              # /c/coupons/*  /admin/coupon-templates/*
│   ├── dto.go
│   └── router.go
├── point/
│   ├── entity.go               # PointAccount / PointTransaction / PointRule / PointAdjustTicket
│   ├── repo.go
│   ├── service.go              # FIFO 入账 / 过期 / 冻结 / 冲销
│   ├── handler.go              # /c/me/points/* /admin/point-*
│   ├── dto.go
│   └── router.go
├── member/
│   ├── entity.go               # MemberLevel
│   ├── repo.go
│   ├── service.go              # 等级重算 / 升级触发
│   ├── handler.go              # /c/me/level /admin/member-levels/*
│   ├── dto.go
│   └── router.go
└── shared/
    ├── quote.go                # 结算计算（券折扣 + 积分抵扣 + 会员积分预估）
    └── idem.go                 # 通用 order_id 幂等键
```

## 状态机（唯一入口）

| 实体 | 函数 | 事件 |
| --- | --- | --- |
| `UserCoupon` | `coupon.CouponService.Transition(uc, event)` | `lock` / `consume` / `release_full_refund` / `release_cancel` / `expire` |
| `PointAdjustTicket` | `point.PointService.TicketTransition(t, event)` | `approve` / `reject` |

## asynq jobs（task type）

| code | cron / trigger |
| --- | --- |
| `coupon:grant` | admin 创建发放任务 enqueue |
| `coupon:expire-scan` | 每日 03:00 |
| `coupon:birthday-cron` | 每日 09:00 |
| `coupon:expire-warning` | 每日 10:00 |
| `point:earn-grant` | 订单确认收货 T+N |
| `point:expire-scan` | 每日 03:30 |
| `point:expire-warning` | 每日 10:00 |
| `point:rollback` | OrderRefunded 事件 |
| `member:level-recompute` | 每日 03:00 |

## 红线
1. 所有面向 order 的接口都用 `order_id` 做幂等键。
2. 金额一律 `*_cents int64`；折扣率用 `decimal(5,4)`。
3. 状态机变更走单一函数，禁止散点 UPDATE。
4. 积分流水 append-only，账户余额仅由流水汇总更新（事务内）。
