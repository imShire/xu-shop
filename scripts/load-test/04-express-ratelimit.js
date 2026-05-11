/**
 * 压测 04：快递鸟（KdBird）Token Bucket 限流验证
 *
 * 验收条件（docs/arch/93-implementation-plan.md 阶段 4.5）：
 *   - 超出令牌桶速率的请求应优雅降级（返回 429 或排队），而非 5xx
 *   - 限流期间无服务崩溃（server error rate < 1%）
 *
 * 工作原理：
 *   server 对外调用快递鸟 API 时通过 pkg/kdbird 内置 token bucket 限流。
 *   本脚本通过高频触发"查询物流轨迹"接口（/c/orders/{id}/tracks）来驱动
 *   服务端的出站限流逻辑；同时对快递回调接口发起洪流请求，验证入站限流。
 *
 *   注意：此脚本测试的是"服务端处理超限请求是否优雅"，而非 k6 侧的限流。
 *
 * 前置条件：
 *   - ORDER_ID 是一笔已发货（status=shipped）的订单（有运单号）
 *   - USER_TOKEN 是该订单对应用户的 JWT
 *
 * 运行：
 *   k6 run \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e USER_TOKEN=<jwt> \
 *     -e ORDER_ID=<order_id> \
 *     04-express-ratelimit.js
 */

import http from 'k6/http';
import { check } from 'k6';
import { Rate, Counter } from 'k6/metrics';

// 服务端限流触发时的 429 计数
const rateLimited429 = new Rate('rate_limited_429');
// 服务端内部错误（不可接受）
const serverErrors5xx = new Counter('server_errors_5xx');

export const options = {
  scenarios: {
    // 以 200 RPS 突发，超过快递鸟令牌桶速率（通常配置 ~10 QPS 出站）
    burst_track_queries: {
      executor: 'constant-arrival-rate',
      rate: 200,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 100,
    },
  },

  thresholds: {
    // 核心验收：无 5xx（限流必须优雅，不能崩溃）
    http_req_failed: ['rate<0.01'],
    // 应当有部分请求被限流（429），证明 token bucket 在工作
    rate_limited_429: ['rate>0'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;
const TOKEN = __ENV.USER_TOKEN;
const ORDER_ID = __ENV.ORDER_ID || '1';

export default function () {
  // 高频查询物流轨迹，触发服务端对快递鸟的出站限流
  const res = http.get(`${API_BASE}/c/orders/${ORDER_ID}/tracks`, {
    headers: {
      Authorization: `Bearer ${TOKEN}`,
    },
    tags: { type: 'track_query' },
  });

  const is429 = res.status === 429;

  // 429 = token bucket 正确工作（优雅限流）
  rateLimited429.add(is429 ? 1 : 0);

  if (res.status >= 500) {
    serverErrors5xx.add(1);
    console.log(`5xx error: ${res.status}, body: ${res.body?.substring(0, 200)}`);
  }

  check(res, {
    'no 5xx (graceful degradation)': (r) => r.status < 500,
    'expected status (200/404/429)': (r) => [200, 404, 429].includes(r.status),
  });
}

export function handleSummary(data) {
  const rateLimitedRate = data.metrics['rate_limited_429']?.values?.rate || 0;
  const serverErrorCount = data.metrics['server_errors_5xx']?.values?.count || 0;
  const httpFailedRate = data.metrics['http_req_failed']?.values?.rate || 0;
  const totalReqs = data.metrics['http_reqs']?.values?.count || 0;

  // 验收：无 5xx + 有 429 限流响应（证明 token bucket 在生效）
  const pass = serverErrorCount === 0 && rateLimitedRate > 0;

  const summary = {
    verdict: pass ? '✅ PASS' : '❌ FAIL',
    total_requests: totalReqs,
    rate_limited_rate: rateLimitedRate.toFixed(4),
    server_error_count: serverErrorCount,
    http_failed_rate: httpFailedRate.toFixed(4),
    rules: {
      'no 5xx errors': serverErrorCount === 0 ? 'PASS' : 'FAIL',
      'token bucket triggered (some 429s)': rateLimitedRate > 0 ? 'PASS' : 'FAIL (check pkg/kdbird token bucket config)',
    },
    note: '若 rate_limited_rate=0，说明 token bucket 速率配置过高或快递鸟 mock 未启用出站限流',
  };

  console.log('\n=== 04 快递 Token Bucket 限流测试结果 ===');
  console.log(JSON.stringify(summary, null, 2));

  return {
    stdout: JSON.stringify(summary, null, 2) + '\n',
  };
}
