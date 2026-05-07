# modules/account

> 对应 PRD：`docs/prd/01-account.md`
> 对应 arch：`docs/arch/01-account.md`（v1.1 含 #2/#3/#5/#8/#29 修复）

## 实施阶段
阶段 1（脚手架后第一个模块）

## 文件清单（统一模板）

| 文件 | 内容 |
| --- | --- |
| `entity.go` | `User` / `Admin` / `Role` / `Permission` / `LoginLog` GORM 模型 |
| `dto.go` | 请求/响应结构（MpLoginReq、AdminLoginReq、MeResp、RoleDTO 等） |
| `repo.go` | 数据访问（FindUserByOpenid、UpsertUserByUnionid、LockAdmin 等） |
| `service.go` | 登录逻辑、JWT 签发、密码 hash、状态机、注销脱敏、强密码校验 |
| `handler.go` | gin handler |
| `router.go` | `RegisterCRoutes`、`RegisterAdminRoutes` |
| `event.go` | 发出事件：UserDeactivated、AdminDisabled |
| `permission_seed.go` | 权限点常量 + 默认角色映射（供 CLI seed） |
| `service_test.go` | 单元测试 |

## 关键交付（按 arch 01）

- C 端：`/c/auth/mp/login`、`/c/auth/h5/code`、`/c/auth/h5/callback`（Set-Cookie HttpOnly）、`/c/auth/bind-phone`、`/c/auth/refresh`、`/c/auth/logout`、`/c/me`、`/c/me/deactivate[/cancel]`
- 后台：`/admin/auth/captcha`、`/admin/auth/login`、`/admin/auth/logout`、`/admin/me`、`/admin/admins[/:id]`、`/admin/roles[/:id]`、`/admin/permissions`
- 中间件：`UserAuth`、`AdminAuth(perms ...)`（在 `internal/middleware/`）

## 依赖
- `pkg/jwt`、`pkg/wxlogin`、`pkg/lock`、`pkg/errs`、`pkg/validator`
- Redis：黑名单、status 缓存、登录失败计数、图形验证码
