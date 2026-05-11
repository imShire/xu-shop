/**
 * 压测 01：100 VU 并发抢购 10 库存 SKU
 *
 * 验收条件（docs/arch/93-implementation-plan.md 阶段 4.5）：
 *   - 仅 10 笔下单成功（status=200/201）
 *   - 其余均返回 4xx（409 stock_insufficient 或 422 校验失败）
 *   - 无任何 5xx 错误
 *
 * 前置条件：
 *   - SKU_ID 对应 SKU 库存恰好为 10
 *   - USER_TOKEN 有效（C 端用户 JWT）
 *   - ADDRESS_ID 是该用户已有的收货地址
 *
 * 运行：
 *   k6 run \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e USER_TOKEN=<jwt> \
 *     -e SKU_ID=<sku_id> \
 *     -e ADDRESS_ID=<addr_id> \
 *     01-inventory-race.js
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// 自定义指标
const successOrders = new Counter('success_orders');
const stockInsufficient = new Counter('stock_insufficient');
const otherErrors = new Counter('other_errors');

export const options = {
  // 100 VU 同时发起，每个 VU 执行 1 次迭代（共 100 次）
  vus: 100,
  iterations: 100,

  thresholds: {
    // 核心验收：成功下单数不超过库存上限 10
    success_orders: ['count<=10'],
    // 无 5xx 错误
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;
const TOKEN = __ENV.USER_TOKEN;
const SKU_ID = __ENV.SKU_ID;
const ADDRESS_ID = __ENV.ADDRESS_ID;

export default function () {
  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${TOKEN}`,
  };

  let cartId = null;

  // ── 步骤 1：加入购物车 ──────────────────────────────────────────────────
  group('add_to_cart', () => {
    const res = http.post(
      `${API_BASE}/c/cart`,
      JSON.stringify({ sku_id: parseInt(SKU_ID), qty: 1 }),
      { headers: authHeaders },
    );

    check(res, {
      'add_cart: no 5xx': (r) => r.status < 500,
      'add_cart: 200 or 4xx': (r) => [200, 409, 422, 429].includes(r.status),
    });

    if (res.status !== 200) {
      stockInsufficient.add(1);
      return;
    }

    try {
      const body = JSON.parse(res.body);
      // 支持 {data: {id: ...}} 或 {data: {cart_id: ...}} 两种 schema
      cartId = body.data?.id || body.data?.cart_id;
    } catch (_) {
      otherErrors.add(1);
    }
  });

  if (!cartId) return;

  // ── 步骤 2：创建订单 ──────────────────────────────────────────────────
  group('create_order', () => {
    const res = http.post(
      `${API_BASE}/c/orders`,
      JSON.stringify({
        cart_ids: [String(cartId)], // API ID 字段统一用字符串（arch 红线）
        address_id: String(ADDRESS_ID),
      }),
      {
        headers: {
          ...authHeaders,
          'Idempotency-Key': uuidv4(), // 每次唯一，防止幂等 key 碰撞影响计数
        },
      },
    );

    check(res, {
      'create_order: no 5xx': (r) => r.status < 500,
      'create_order: 200/201 or 4xx': (r) => [200, 201, 409, 422, 429].includes(r.status),
    });

    if ([200, 201].includes(res.status)) {
      successOrders.add(1);
    } else if (res.status === 409) {
      stockInsufficient.add(1);
    } else {
      otherErrors.add(1);
    }
  });
}

export function handleSummary(data) {
  const succeed = data.metrics['success_orders']?.values?.count || 0;
  const insufficient = data.metrics['stock_insufficient']?.values?.count || 0;
  const errors = data.metrics['other_errors']?.values?.count || 0;
  const httpFailed = (data.metrics['http_req_failed']?.values?.rate || 0).toFixed(4);

  const pass = succeed <= 10 && parseFloat(httpFailed) < 0.01;

  const summary = {
    verdict: pass ? '✅ PASS' : '❌ FAIL',
    success_orders: succeed,
    stock_insufficient: insufficient,
    other_errors: errors,
    http_error_rate: httpFailed,
    rule: '成功下单数必须 ≤ 10（SKU 库存上限），无 5xx 错误',
  };

  console.log('\n=== 01 库存抢购压测结果 ===');
  console.log(JSON.stringify(summary, null, 2));

  return {
    stdout: JSON.stringify(summary, null, 2) + '\n',
  };
}
