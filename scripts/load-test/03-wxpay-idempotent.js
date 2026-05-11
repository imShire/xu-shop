/**
 * 压测 03：微信支付回调 5 次重发幂等测试
 *
 * 验收条件（docs/arch/93-implementation-plan.md 阶段 4.5）：
 *   - 5 次相同 transaction_id 回调均返回 200
 *   - 订单 paid_at 仅有一个时间戳（状态只变更一次）
 *   - payment 表无重复记录
 *
 * 工作原理：
 *   微信支付回调通知因网络问题会重发，服务端必须对同一 transaction_id
 *   保证幂等处理（arch 红线第 4 条：所有写操作必须幂等）。
 *
 * 前置条件：
 *   - 服务启动时设置 WXPAY_MOCK_MODE=true（跳过微信签名验证）
 *   - ORDER_ID 是一笔待支付（status=pending_payment）的订单号
 *   - TRANSACTION_ID 是伪造的微信支付流水号（格式 TEST_TXN_xxx）
 *
 * 运行：
 *   k6 run \
 *     -e BASE_URL=http://localhost:8080 \
 *     -e ORDER_ID=<order_id> \
 *     -e TRANSACTION_ID=TEST_TXN_001 \
 *     03-wxpay-idempotent.js
 *
 * ⚠️  此测试需要服务端配置 WXPAY_MOCK_MODE=true 才能运行，
 *     生产环境不得开启此选项。
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const notifyAccepted = new Counter('notify_accepted');
const notifyRejected = new Counter('notify_rejected');

export const options = {
  vus: 1,
  iterations: 5, // 模拟微信支付回调 5 次重发

  thresholds: {
    // 所有 5 次回调必须被正常处理（200）
    notify_accepted: ['count==5'],
    // 无服务端错误
    http_req_failed: ['rate==0'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;
const ORDER_ID = __ENV.ORDER_ID || '';
const TRANSACTION_ID = __ENV.TRANSACTION_ID || `TEST_TXN_${Date.now()}`;

/**
 * 构造模拟微信支付通知 body
 *
 * 真实场景下 resource.ciphertext 是 AES-256-GCM 加密的 JSON；
 * 测试模式（WXPAY_MOCK_MODE=true）下服务端读取 __mock_data 字段直接解析。
 *
 * 参考：https://pay.weixin.qq.com/docs/merchant/apis/jsapi-payment/payment-notice.html
 */
function buildNotifyBody() {
  return JSON.stringify({
    id: TRANSACTION_ID,
    create_time: new Date().toISOString(),
    event_type: 'TRANSACTION.SUCCESS',
    resource_type: 'encrypt-resource',
    summary: '支付成功',
    resource: {
      algorithm: 'AEAD_AES_256_GCM',
      ciphertext: 'MOCK_CIPHERTEXT',
      nonce: 'MOCK_NONCE_12',
      associated_data: 'transaction',
      original_type: 'transaction',
      // 仅在 WXPAY_MOCK_MODE=true 时有效
      __mock_data: {
        transaction_id: TRANSACTION_ID,
        out_trade_no: ORDER_ID,
        trade_state: 'SUCCESS',
        trade_state_desc: '支付成功',
        bank_type: 'OTHERS',
        success_time: new Date().toISOString(),
        amount: {
          total: 100,       // 单位分（1 元），仅供测试
          payer_total: 100,
          currency: 'CNY',
          payer_currency: 'CNY',
        },
        payer: {
          openid: 'MOCK_OPENID_TEST_USER',
        },
      },
    },
  });
}

export default function () {
  // 5 次回调使用完全相同的 body（相同 transaction_id）
  const body = buildNotifyBody();

  const res = http.post(`${API_BASE}/notify/wxpay`, body, {
    headers: {
      'Content-Type': 'application/json',
      // 微信回调签名头（mock 模式下服务端跳过验证）
      'Wechatpay-Timestamp': String(Math.floor(Date.now() / 1000)),
      'Wechatpay-Nonce': `mock_nonce_${__ITER}`,
      'Wechatpay-Signature': 'mock_signature_for_test',
      'Wechatpay-Serial': 'mock_serial_no',
    },
  });

  const accepted = check(res, {
    'notify: returns 200 (idempotent)': (r) => r.status === 200,
    'notify: no server error': (r) => r.status < 500,
  });

  if (accepted && res.status === 200) {
    notifyAccepted.add(1);
  } else {
    notifyRejected.add(1);
    console.log(`Iteration ${__ITER}: unexpected status ${res.status}, body: ${res.body}`);
  }

  // 模拟微信重发间隔（实际约 15s / 15s / 30s / 3min / 10min）
  // 压测中压缩为 0.5s 快速验证幂等逻辑
  sleep(0.5);
}

export function handleSummary(data) {
  const accepted = data.metrics['notify_accepted']?.values?.count || 0;
  const rejected = data.metrics['notify_rejected']?.values?.count || 0;

  const pass = accepted === 5 && rejected === 0;

  const summary = {
    verdict: pass ? '✅ PASS' : '❌ FAIL',
    total_callbacks: 5,
    accepted,
    rejected,
    rules: [
      '5 次相同回调均返回 200',
      '（需手动验证）订单 paid_at 仅一个时间戳',
      '（需手动验证）payment 表无重复记录',
    ],
    manual_verify: [
      `SELECT paid_at FROM orders WHERE id = '${ORDER_ID}'`,
      `SELECT COUNT(*) FROM payments WHERE transaction_id = '${TRANSACTION_ID}'  -- 期望: 1`,
    ],
  };

  console.log('\n=== 03 微信支付回调幂等测试结果 ===');
  console.log(JSON.stringify(summary, null, 2));

  return {
    stdout: JSON.stringify(summary, null, 2) + '\n',
  };
}
