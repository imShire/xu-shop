# xu-shop

**单商家私域电商全栈开源项目**

微信小程序 + H5 + Vue3 后台管理 + Go 后端，开箱即用的私域电商解决方案

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vue.js)](https://vuejs.org)
[![Taro](https://img.shields.io/badge/Taro-4.x-00C8B5)](https://taro.zone)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?logo=postgresql)](https://postgresql.org)

</div>

---

## 项目简介

xu-shop 是一套面向**单商家私域电商**场景的完整技术方案，覆盖从用户下单到仓库发货的完整业务闭环。

流量入口在微信生态（小程序 / 公众号 H5 / 企微），运营和履约在 Web 后台完成。项目使用 **Go + PostgreSQL** 构建高性能后端，**Taro** 实现"一套代码跑小程序和 H5"，**Vue 3 + Element Plus** 搭建功能完善的运营后台。

> **本项目作为完整的生产级参考实现开放源码**，包含详细的 PRD 文档（16 份）、架构设计文档（17 份）和分阶段实施计划，适合用于学习、参考或作为二次开发的起点。

---

## 技术栈

| 层次 | 技术选型 |
|---|---|
| **后端** | Go 1.22 · Gin · GORM · PostgreSQL 15 · Redis 7 · asynq |
| **C 端跨端** | Taro 4 · React 18 · TypeScript · NutUI-React-Taro |
| **运营后台** | Vue 3.4 · Element Plus 2.7 · Vite 5 · Pinia |
| **基础设施** | Docker Compose · MinIO · Nginx · OpenTelemetry |
| **三方对接** | 微信小程序登录 · 微信支付 V3 · 快递鸟电子面单 · 阿里云 OSS |

---

## 功能模块

### C 端（小程序 / H5）

- **账号**：微信一键登录（小程序 code2session + 公众号 OAuth2）、手机号绑定
- **商品**：分类浏览、SKU 多规格、商品收藏、浏览历史
- **购物车**：实时库存校验、选中结算
- **订单**：下单 → 超时自动关单（asynq）→ 支付 → 发货 → 确认收货 → 评价
- **支付**：微信 JSAPI 支付（小程序）+ 微信 H5 支付
- **物流**：快递鸟实时轨迹查询、订阅消息推送
- **消息**：小程序订阅消息 + 公众号模板消息

### 运营后台（Vue3）

- **商品管理**：SPU / SKU 编辑、分类管理、图片上传（OSS）、上下架
- **库存管理**：调拨记录、低库存预警、Redis Lua 原子扣减防超卖
- **订单管理**：批量发货、打印电子面单、手工操作退款
- **用户管理**：用户列表、标签、企微渠道码
- **数据看板**：GMV、订单量、用户增长（预聚合）
- **系统设置**：店铺信息、管理员账号、角色权限（RBAC）

---

## 架构亮点

- **防超卖**：Redis Lua 脚本原子扣减库存，支持 100 并发不超卖
- **ID 防溢出**：所有 API ID 字段统一 `string` 类型，避免 JS BigInt 精度丢失
- **幂等设计**：三方回调 + 客户端 `Idempotency-Key` 双重幂等保护
- **状态机**：订单、支付、退款状态机单一入口，禁止散点 UPDATE
- **安全加固**：JWT 黑名单、HttpOnly Cookie、富文本白名单过滤、SQL 参数化

---

## 快速开始

### 环境要求

- Go 1.22+
- Node 20+ · pnpm 9+
- Docker & Docker Compose
- 微信开发者工具（调试小程序）

### 启动本地依赖

```bash
make deps-up        # 启动 PostgreSQL · Redis · MinIO · asynqmon
```

### 后端

```bash
cd server
cp env.example .env   # 填写微信支付、OSS 等配置
make migrate-up       # 执行数据库迁移
make run-api          # http://localhost:8080
make run-worker       # asynq worker
```

### 运营后台

```bash
cd admin
pnpm install
pnpm dev              # http://localhost:5273
```

### C 端

```bash
cd client
pnpm install
pnpm dev:h5           # H5 预览 http://localhost:10086
pnpm dev:weapp        # 小程序，用微信开发者工具打开 client/dist/weapp
```

---

---

## 参与贡献

欢迎提交 Issue 和 Pull Request！贡献前请先阅读架构文档，保持与现有设计一致。

---

## 许可证

本项目基于 **GNU Affero General Public License v3.0 (AGPL-3.0)** 开源。

- ✅ 允许：学习、研究、个人使用、修改代码
- ✅ 允许：基于本项目二次开发，但**必须以相同协议开源所有修改**
- ❌ 禁止：将本项目（或其修改版本）作为商业服务提供，除非同样开源
- ❌ 禁止：在不开源的商业产品中使用本项目代码

详见 [LICENSE](LICENSE) 文件。
