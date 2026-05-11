/**
 * 压测 05：Redis 抖动 30 秒降级策略验证
 *
 * 验收条件（docs/arch/99-revisions.md / 阶段 4.5）：
 *   - Redis 不可用时：读接口（商品列表）可用率 > 70%
 *   - Redis 不可用时：写敏感接口（下单/加购）返回 503（Service Unavailable）
 *   - JWT 黑名单 Redis 故障 → 写敏感接口拒绝（503），读接口放过
 *
 * ⚠️  此测试需在运行前手动触发 Redis 抖动！
 *
 * 触发方式（二选一）：
 *   # 方式 A：暂停 Docker 容器（更真实）
 *   docker pause xu-shop-redis-1
 *   # 恢复
 *   docker unpause xu-shop-redis-1
 *
 *   # 方式 B：DEBUG SLEEP（不中断连接池，测试超时处理）
 *   redis-cli -h localhost -p 6379 DEBUG SLEEP 60
 *
 * 运行时序：
 *   1. 先触发 Redis 故障
 *   2. 立即运行本脚本（脚本运行 35s，覆盖 30s 抖动窗口）
 *   3. 脚本结束后恢复 Redis
 *
 * 运行：
 *   k6 run \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e USER_TOKEN=<jwt> \
 *     -e ADDRESS_ID=<addr_id> \
 *     05-redis-degradation.js
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// 读接口成功率（Redis 故障时应仍能提供服务）
const readSuccessRate = new Rate('read_success_during_redis_fault');
// 写敏感接口被正确拒绝的比率（503 = 符合降级预期）
const writeBlockedRate = new Rate('write_blocked_503_during_redis_fault');

export const options = {
  vus: 10,
  // 35 秒覆盖 30 秒 Redis 抖动窗口 + 5 秒缓冲
  duration: '35s',

  thresholds: {
    // 读接口（商品列表）在 Redis 故障时可用率 > 70%
    read_success_during_redis_fault: ['rate>0.7'],
    // 写敏感接口（下单）在 Redis 故障时 > 50% 返回 503
    write_blocked_503_during_redis_fault: ['rate>0.5'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;
const TOKEN = __ENV.USER_TOKEN;
const ADDRESS_ID = __ENV.ADDRESS_ID || '1';

export default function () {
  // ── 测试 A：读接口（无需 Redis 的商品列表）─────────────────────────
  // 商品列表走 DB 查询，Redis 故障时应有降级逻辑继续提供服务
  group('read_product_list', () => {
    const res = http.get(`${API_BASE}/c/products?page=1&page_size=10`, {
      headers: { 'Content-Type': 'application/json' },
      tags: { type: 'read' },
    });

    const ok = check(res, {
      'read: accessible during Redis fault': (r) => r.status === 200,
    });

    readSuccessRate.add(ok ? 1 : 0);
  });

  sleep(0.1);

  // ── 测试 B：读接口（商品详情，可能有 Redis 缓存）────────────────────
  group('read_product_detail', () => {
    const res = http.get(`${API_BASE}/c/products/1`, {
      headers: { 'Content-Type': 'application/json' },
      tags: { type: 'read_cached' },
    });

    // 200 = 缓存命中或 DB 降级；404 = 商品不存在（也算接口可用）
    check(res, {
      'read_detail: accessible or not found': (r) => [200, 404].includes(r.status),
    });
  });

  sleep(0.1);

  // ── 测试 C：写敏感接口（下单，需 Redis JWT 黑名单校验）──────────────
  // 根据 arch/99-revisions.md：JWT 黑名单 Redis 故障 → 写敏感接口拒绝（503），读放过
  group('write_create_order', () => {
    const res = http.post(
      `${API_BASE}/c/orders`,
      JSON.stringify({
        cart_ids: ['1'],
        address_id: ADDRESS_ID,
      }),
      {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${TOKEN}`,
          'Idempotency-Key': uuidv4(),
        },
        tags: { type: 'write_sensitive' },
      },
    );

    // 降级预期：Redis 故障时应返回 503（写敏感接口不允许降级放过）
    // 注：如果 JWT 验证本身依赖 Redis 且直接返回 401，也视为正确拒绝
    const isBlocked = [503, 401].includes(res.status);

    check(res, {
      'write: gracefully rejected during Redis fault (503/401)': () => isBlocked,
      'write: no panic (no 5xx other than 503)': (r) =>
        r.status < 500 || r.status === 503,
    });

    writeBlockedRate.add(isBlocked ? 1 : 0);
  });

  sleep(0.2);
}

export function handleSummary(data) {
  const readRate = data.metrics['read_success_during_redis_fault']?.values?.rate || 0;
  const writeBlockRate = data.metrics['write_blocked_503_during_redis_fault']?.values?.rate || 0;
  const totalReqs = data.metrics['http_reqs']?.values?.count || 0;

  const pass = readRate > 0.7 && writeBlockRate > 0.5;

  const summary = {
    verdict: pass ? '✅ PASS' : '❌ FAIL',
    total_requests: totalReqs,
    read_success_rate: readRate.toFixed(3),
    write_blocked_rate: writeBlockRate.toFixed(3),
    thresholds: {
      'read success > 70%': readRate > 0.7 ? 'PASS' : 'FAIL',
      'write blocked (503) > 50%': writeBlockRate > 0.5 ? 'PASS' : 'FAIL',
    },
    arch_ref: 'docs/arch/99-revisions.md: JWT 黑名单 Redis 故障 → 写敏感接口拒绝（503），读放过',
    troubleshoot: readRate <= 0.7
      ? '读接口成功率不足：检查商品列表是否有 Redis 硬依赖，应实现 DB fallback'
      : writeBlockRate <= 0.5
        ? '写接口未被正确拒绝：检查 middleware/auth.go 中 Redis 故障时是否返回 503'
        : null,
  };

  console.log('\n=== 05 Redis 抖动降级策略测试结果 ===');
  console.log(JSON.stringify(summary, null, 2));

  return {
    stdout: JSON.stringify(summary, null, 2) + '\n',
  };
}
