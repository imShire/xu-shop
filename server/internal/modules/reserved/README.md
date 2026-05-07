# modules/reserved

> 对应 PRD：`docs/prd/14-reserved.md`
> 对应 arch：`docs/arch/14-reserved.md`

## 实施阶段
阶段 6

## 范围
仅含 GORM 模型 + 迁移文件。**无 service / handler / router / 路由注册**。

二期开启分销 / 拼团 / 优惠券 / 积分时基于这里扩展。

## 文件清单
| 文件 | 内容 |
| --- | --- |
| `model.go` | `Distributor`、`CommissionRecord`、`GroupBuyActivity`、`GroupBuyOrder`、`Coupon`、`UserCoupon`、`PointsLog` GORM struct |
| `model_test.go` | 写读 smoke test |
