# xu-shop 阶段 4.5 压测脚本

> 对应 `docs/arch/93-implementation-plan.md` 阶段 4.5 验收要求。
> 未通过压测，禁止进入阶段 5。

---

## 验收标准（PRD 15 A6）

| 编号 | 场景 | 判定条件 |
|------|------|----------|
| 01 | 100 VU 并发抢 10 件库存 SKU | 成功下单数 ≤ 10，无超卖 |
| 02 | 100 RPS 持续 5 分钟混合流量 | P95 ≤ 500 ms，错误率 < 0.5% |
| 03 | 微信支付回调 5 次重发 | 全部返回 200，订单状态仅变更一次 |
| 04 | 快递鸟限流（Token Bucket） | 无 5xx，超限请求返回 429 |
| 05 | Redis 抖动 30 秒降级策略 | 读接口可用率 > 70%，写敏感接口返回 503 |

---

## 前置条件

### 1. 安装 k6

```bash
# macOS
brew install k6

# Docker（无需安装）
docker run --rm -i grafana/k6 run - <script.js

# 官方安装文档
# https://k6.io/docs/get-started/installation/
```

### 2. 准备测试数据

在运行压测前，必须准备以下数据：

| 变量 | 说明 | 获取方式 |
|------|------|----------|
| `USER_TOKEN` | C 端用户 JWT Token | 调用 `/api/v1/c/auth/mp/login`（测试环境 mock 模式） |
| `ADMIN_TOKEN` | 管理员 JWT Token | 调用 `/api/v1/admin/auth/login` |
| `SKU_ID` | **库存恰好为 10** 的 SKU ID | 后台创建商品，设置 SKU 库存为 10 |
| `PRODUCT_IDS` | 逗号分隔的商品 ID 列表 | 后台已上架商品 |
| `ADDRESS_ID` | 测试用户的收货地址 ID | 调用 `/api/v1/c/addresses` 创建后获取 |
| `ORDER_ID` | 已创建（待支付）的订单 ID | 先手动下单一笔 |
| `TRANSACTION_ID` | 测试用的微信支付流水号 | 自定义字符串，格式：`TEST_TXN_<timestamp>` |
| `BASE_URL` | 后端服务地址 | 默认 `http://localhost:8080` |

### 3. 启动本地服务

```bash
# 启动依赖（PostgreSQL + Redis + MinIO）
cd deploy && docker compose -f docker-compose.dev.yml up -d

# 启动后端 API
cd server && make run-api

# 验证服务
curl http://localhost:8080/healthz
```

### 4. 测试环境特殊配置

脚本 03（微信支付回调）需要在测试环境禁用微信签名校验：

```bash
# server/.env.test
WXPAY_MOCK_MODE=true
WXPAY_SKIP_VERIFY=true
```

脚本 05（Redis 降级）需要运行时手动触发 Redis 故障：

```bash
# 方式 A：暂停 Docker 容器
docker pause xu-shop-redis-1

# 方式 B：令 Redis 休眠（不中断连接，更真实）
redis-cli DEBUG SLEEP 60

# 恢复
docker unpause xu-shop-redis-1
# 或等待 SLEEP 超时
```

---

## 运行方式

### 单独运行

```bash
cd scripts/load-test

# 测试 01：库存抢购
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e USER_TOKEN=<token> \
  -e SKU_ID=<sku_id> \
  -e ADDRESS_ID=<addr_id> \
  01-inventory-race.js

# 测试 02：混合流量
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e USER_TOKEN=<token> \
  -e ADDRESS_ID=<addr_id> \
  -e PRODUCT_IDS=1,2,3,4,5 \
  02-browse-order.js

# 测试 03：支付回调幂等
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e ORDER_ID=<order_id> \
  -e TRANSACTION_ID=TEST_TXN_001 \
  03-wxpay-idempotent.js

# 测试 04：快递限流
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e USER_TOKEN=<token> \
  -e ORDER_ID=<order_id> \
  04-express-ratelimit.js

# 测试 05：Redis 降级（先触发 Redis 故障，再运行）
docker pause xu-shop-redis-1
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e USER_TOKEN=<token> \
  -e ADDRESS_ID=<addr_id> \
  05-redis-degradation.js
```

### 批量运行（生成报告）

```bash
cd scripts/load-test
mkdir -p results

BASE_URL=http://localhost:8080 \
USER_TOKEN=<token> \
ADMIN_TOKEN=<admin_token> \
SKU_ID=<sku_id> \
ADDRESS_ID=<addr_id> \
PRODUCT_IDS=1,2,3,4,5 \
ORDER_ID=<order_id> \
TRANSACTION_ID=TEST_TXN_001 \
bash run-all.sh
```

报告输出到 `scripts/load-test/results/` 目录。

---

## 目录结构

```
scripts/load-test/
├── README.md                  # 本文件
├── 01-inventory-race.js       # 库存抢购并发测试（100 VU × 10 库存）
├── 02-browse-order.js         # 混合流量压测（100 RPS × 5 分钟）
├── 03-wxpay-idempotent.js     # 微信支付回调幂等测试（5 次重发）
├── 04-express-ratelimit.js    # 快递 Token Bucket 限流验证
├── 05-redis-degradation.js    # Redis 抖动降级策略验证
├── run-all.sh                 # 批量运行脚本
└── results/                   # 测试结果（gitignore）
```

---

## 常见问题

**Q: 测试 01 成功数超过 10？**
检查库存服务的 Redis Lua 脚本（`internal/pkg/inventory/lua_scripts.go`），确认 `DECRBY` 原子操作正确实现了库存扣减，且 DB 有二次校验。

**Q: 测试 02 P95 超出 500ms？**
先检查慢查询日志，重点关注商品列表和下单接口的 DB 查询是否走了索引。

**Q: 测试 03 回调返回 400/401？**
确认服务启动时设置了 `WXPAY_MOCK_MODE=true`，签名校验已旁路。

**Q: 测试 05 读接口也返回 503？**
确认中间件实现符合 `docs/arch/99-revisions.md` 规范：JWT 黑名单 Redis 故障时，**写敏感接口拒绝（503），读接口放过**。
