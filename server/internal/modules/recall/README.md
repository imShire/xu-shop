# recall 模块

> 对应文档：`docs/prd/17-tag-recall.md` / `docs/arch/17-tag-recall.md`（同一份）
> 实施阶段：阶段 5.3（在 tag 之后，依赖 marketing/notification）
> 依赖：tag、marketing（grant_coupon 动作）、notification（订阅消息动作）

召回活动引擎。三种触发器（cron / event / immediate）+ 节流器 + 动作执行器。

## 文件清单

```
recall/
├── README.md
├── entity.go         # RecallCampaign / RecallLog
├── repo.go
├── service.go        # CRUD / Online / Pause / Close / Funnel / LogQuery
├── handler.go        # /admin/recall-campaigns/* /admin/recall-logs
├── dto.go
├── router.go
├── engine/
│   ├── cron.go             # recall:cron-scan
│   ├── event.go            # recall:event-handler（订阅 cart_idle/payment_failed/...）
│   ├── immediate.go
│   ├── audience.go         # 调 tag.PreviewAudience 取人群
│   ├── throttle.go         # Redis 单活动单用户 / 全局每日 / 活动每日 / 活动总量
│   └── execute.go          # recall:execute（grant_coupon/grant_point/notify/sms）
└── attribution.go    # 7 天归因窗口：订阅 OrderPaid 检查最近召回日志
```

## asynq jobs

| code | cron / 触发 |
| --- | --- |
| `recall:cron-scan` | 每 10min 扫描 cron 类活动 |
| `recall:event-handler` | 订阅业务事件流 |
| `recall:execute` | 由前两者按用户粒度 enqueue |

## 红线
1. 所有动作必须先过 `engine.throttle`，写 `recall_log` 后再执行（避免重复触达）。
2. 节流键：`(campaign_id, user_id, day)` 唯一索引保证单日单活动单人最多一次。
3. 暂停活动 5min 内必须停止所有新触达（worker 出队前再次检查 status）。
4. 归因窗口默认 7 天，按活动可配；多活动归因取**最近一次**（last-touch）。
