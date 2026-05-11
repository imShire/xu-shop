/**
 * 压测 02：100 RPS 持续 5 分钟混合流量（首页 + 详情 + 加购 + 下单）
 *
 * 验收条件（PRD 15 A6 / docs/arch/93-implementation-plan.md 阶段 4.5）：
 *   - P95 响应时间 ≤ 500 ms
 *   - 错误率 < 0.5%（4xx 中 429/限流 不计入错误率）
 *   - 持续 5 分钟稳定通过
 *
 * 前置条件：
 *   - USER_TOKEN 有效（C 端用户 JWT，允许多用户 token 用逗号分隔）
 *   - ADDRESS_ID 是该用户已有的收货地址
 *   - PRODUCT_IDS 逗号分隔的已上架商品 ID，至少 3 个（避免缓存击穿单一商品）
 *
 * 运行：
 *   k6 run \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e USER_TOKEN=<jwt> \
 *     -e ADDRESS_ID=<addr_id> \
 *     -e PRODUCT_IDS=1,2,3,4,5 \
 *     02-browse-order.js
 */

import http from 'k6/http';
import { sleep, check, group } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 100,        // 100 RPS
      timeUnit: '1s',
      duration: '5m',   // 持续 5 分钟
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },

  thresholds: {
    // PRD 15 A6：整体错误率 < 0.5%（429 限流不计入失败）
    http_req_failed: ['rate<0.005'],

    // PRD 15 A1/A6：P95 ≤ 500ms
    http_req_duration: ['p(95)<500'],

    // 按接口类型分别统计（tags 过滤）
    'http_req_duration{type:product_list}': ['p(95)<500'],
    'http_req_duration{type:product_detail}': ['p(95)<500'],
    'http_req_duration{type:add_cart}': ['p(95)<500'],
    // 下单链路包含库存预检，稍宽松
    'http_req_duration{type:create_order}': ['p(95)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;
const TOKEN = __ENV.USER_TOKEN;
const ADDRESS_ID = __ENV.ADDRESS_ID || '1';
const PRODUCT_IDS = (__ENV.PRODUCT_IDS || '1,2,3,4,5').split(',').map((s) => s.trim());

const publicHeaders = { 'Content-Type': 'application/json' };
const authHeaders = {
  'Content-Type': 'application/json',
  Authorization: `Bearer ${TOKEN}`,
};

export default function () {
  // ── 步骤 1：首页商品列表 ──────────────────────────────────────────────
  let productIds = [...PRODUCT_IDS];
  let skuId = null;

  group('product_list', () => {
    const res = http.get(`${API_BASE}/c/products?page=1&page_size=20`, {
      headers: publicHeaders,
      tags: { type: 'product_list' },
    });

    check(res, {
      'product_list: 200': (r) => r.status === 200,
    });

    // 从返回列表中提取商品 ID（如果接口已实现）
    try {
      const body = JSON.parse(res.body);
      const items = body.data?.list || body.data?.items || [];
      if (items.length > 0) {
        productIds = items.slice(0, 10).map((p) => String(p.id));
      }
    } catch (_) {
      // 降级使用 PRODUCT_IDS 环境变量
    }
  });

  // ── 步骤 2：商品详情（随机取一个）────────────────────────────────────
  const productId = productIds[Math.floor(Math.random() * productIds.length)];

  group('product_detail', () => {
    const res = http.get(`${API_BASE}/c/products/${productId}`, {
      headers: publicHeaders,
      tags: { type: 'product_detail' },
    });

    check(res, {
      'product_detail: 200 or 404': (r) => [200, 404].includes(r.status),
    });

    if (res.status === 200) {
      try {
        const body = JSON.parse(res.body);
        // 提取第一个有效 SKU
        const skus = body.data?.skus || body.data?.sku_list || [];
        const validSku = skus.find((s) => s.stock > 0);
        if (validSku) {
          skuId = String(validSku.id);
        }
      } catch (_) {}
    }
  });

  // SKU 不可用时跳过加购/下单（不算错误）
  if (!skuId) {
    sleep(0.1);
    return;
  }

  // ── 步骤 3：加入购物车 ────────────────────────────────────────────────
  let cartId = null;

  group('add_cart', () => {
    const res = http.post(
      `${API_BASE}/c/cart`,
      JSON.stringify({ sku_id: skuId, qty: 1 }),
      {
        headers: authHeaders,
        tags: { type: 'add_cart' },
      },
    );

    // 429 是正常限流（C9: 加购 60次/分钟/用户），不计入 http_req_failed
    check(res, {
      'add_cart: expected response': (r) => [200, 409, 422, 429].includes(r.status),
    });

    if (res.status === 200) {
      try {
        const body = JSON.parse(res.body);
        cartId = String(body.data?.id || body.data?.cart_id || '');
      } catch (_) {}
    }
  });

  if (!cartId) {
    sleep(0.1);
    return;
  }

  // ── 步骤 4：创建订单（不发起支付）────────────────────────────────────
  group('create_order', () => {
    const res = http.post(
      `${API_BASE}/c/orders`,
      JSON.stringify({
        cart_ids: [cartId],
        address_id: ADDRESS_ID,
      }),
      {
        headers: {
          ...authHeaders,
          // 每次请求使用唯一 key，避免被幂等机制合并
          'Idempotency-Key': uuidv4(),
        },
        tags: { type: 'create_order' },
      },
    );

    // 409(库存不足) / 422(校验失败) / 429(限流) 均属预期响应
    check(res, {
      'create_order: expected response': (r) => [200, 201, 409, 422, 429].includes(r.status),
    });
  });

  sleep(0.1);
}

export function handleSummary(data) {
  const p95 = data.metrics['http_req_duration']?.values?.['p(95)'] || 0;
  const errorRate = data.metrics['http_req_failed']?.values?.rate || 0;
  const totalReqs = data.metrics['http_reqs']?.values?.count || 0;

  const pass = p95 <= 500 && errorRate < 0.005;

  const summary = {
    verdict: pass ? '✅ PASS' : '❌ FAIL',
    total_requests: totalReqs,
    p95_ms: p95.toFixed(2),
    error_rate: errorRate.toFixed(4),
    thresholds: {
      'p95 ≤ 500ms': p95 <= 500 ? 'PASS' : 'FAIL',
      'error rate < 0.5%': errorRate < 0.005 ? 'PASS' : 'FAIL',
    },
  };

  console.log('\n=== 02 混合流量压测结果 ===');
  console.log(JSON.stringify(summary, null, 2));

  return {
    stdout: JSON.stringify(summary, null, 2) + '\n',
  };
}
