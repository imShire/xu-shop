# tag 模块

> 对应文档：`docs/prd/17-tag-recall.md` / `docs/arch/17-tag-recall.md`
> 实施阶段：阶段 5.3
> 依赖：account、order、product

用户标签字典 + 关系 + 自动计算引擎。**与 recall 拆为两个独立包**：标签是数据资产，召回是动作。

## 文件清单

```
tag/
├── README.md
├── entity.go        # UserTag / UserTagRelation / UserTagSnapshot
├── repo.go
├── service.go       # AddManual / RemoveManual / QueryUserTags / PreviewAudience
├── handler.go       # /admin/user-tags /admin/users/{id}/tags /admin/audience/preview
├── dto.go
├── router.go
├── compute/
│   ├── rfm.go              # tag:recompute-rfm
│   ├── lifecycle.go
│   ├── category_pref.go
│   ├── price_band.go
│   ├── member_level.go
│   ├── source_channel.go
│   └── incremental.go      # tag:incremental（事件触发，5s 去重窗口）
├── snapshot.go             # tag:snapshot-monthly
└── event.go                # 订阅 OrderPaid 触发 incremental
```

## asynq jobs

| code | cron |
| --- | --- |
| `tag:recompute-rfm` | 每日 02:00 |
| `tag:recompute-lifecycle` | 每日 02:10 |
| `tag:recompute-category-pref` | 每日 02:20 |
| `tag:recompute-price-band` | 每日 02:30 |
| `tag:recompute-member-level` | 每日 02:40 |
| `tag:recompute-source-channel` | 每日 02:45 |
| `tag:incremental` | OrderPaid 事件 |
| `tag:snapshot-monthly` | 每月 1 日 04:00 |

## 红线
1. `user_tag` 是字典；`user_tag_relation` 是关系（user_id+tag_code 唯一）。
2. 标签 `source='auto'` 由 job 维护，admin 不可手动改；`manual` 才能 admin 增删。
3. 全量重算单次必须能在 30min 内完成 100w 用户（分批 + 并发，参见 arch）。
4. 增量重算 5s 去重窗口（同 user_id 在 5s 内多次触发只执行一次）。
