#!/usr/bin/env bash
# scripts/load-test/run-all.sh
#
# 批量执行阶段 4.5 全部压测，结果输出到 results/ 目录。
# 未达标的脚本会输出 ❌ FAIL，整体以非零退出码返回。
#
# 用法：
#   BASE_URL=http://localhost:8080 \
#   USER_TOKEN=<jwt> \
#   ADMIN_TOKEN=<admin_jwt> \
#   SKU_ID=<sku_id> \
#   ADDRESS_ID=<addr_id> \
#   PRODUCT_IDS=1,2,3,4,5 \
#   ORDER_ID=<order_id> \
#   TRANSACTION_ID=TEST_TXN_001 \
#   bash scripts/load-test/run-all.sh

set -euo pipefail

# ── 切换到脚本所在目录 ─────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# ── 必填环境变量校验 ───────────────────────────────────────────────────
: "${USER_TOKEN:?需要设置 USER_TOKEN（C 端用户 JWT）}"
: "${SKU_ID:?需要设置 SKU_ID（库存恰好为 10 的 SKU）}"
: "${ADDRESS_ID:?需要设置 ADDRESS_ID（用户收货地址 ID）}"
: "${ORDER_ID:?需要设置 ORDER_ID（已创建的待支付订单 ID）}"

# ── 可选变量默认值 ─────────────────────────────────────────────────────
BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-${USER_TOKEN}}"
PRODUCT_IDS="${PRODUCT_IDS:-1,2,3,4,5}"
TRANSACTION_ID="${TRANSACTION_ID:-TEST_TXN_$(date +%s)}"

# ── 创建结果目录 ───────────────────────────────────────────────────────
mkdir -p results

# ── 检查 k6 是否可用 ───────────────────────────────────────────────────
if ! command -v k6 &>/dev/null; then
  echo "❌ k6 未安装。请参考 README.md 安装 k6："
  echo "   macOS: brew install k6"
  echo "   官方文档: https://k6.io/docs/get-started/installation/"
  exit 1
fi

echo "k6 版本：$(k6 version)"
echo "BASE_URL：${BASE_URL}"
echo "开始时间：$(date '+%Y-%m-%d %H:%M:%S')"
echo "=================================================="

OVERALL_PASS=true

# ── 辅助函数 ──────────────────────────────────────────────────────────
run_test() {
  local num="$1"
  local name="$2"
  local script="$3"
  shift 3
  local extra_args=("$@")

  echo ""
  echo "=== ${num}. ${name} ==="

  local output_file="results/${num}.json"

  if k6 run \
    -e BASE_URL="$BASE_URL" \
    "${extra_args[@]}" \
    --out "json=${output_file}" \
    --summary-trend-stats "med,p(90),p(95),p(99)" \
    "$script"; then
    echo "→ ${num} 完成，结果：${output_file}"
  else
    echo "→ ❌ ${num} 失败"
    OVERALL_PASS=false
  fi
}

# ── 01：库存抢购（100 VU × 10 库存）──────────────────────────────────
run_test "01" "库存抢购压测（100 VU × 10 库存 SKU）" "01-inventory-race.js" \
  -e "USER_TOKEN=${USER_TOKEN}" \
  -e "SKU_ID=${SKU_ID}" \
  -e "ADDRESS_ID=${ADDRESS_ID}"

# ── 02：混合流量（100 RPS × 5 分钟）──────────────────────────────────
run_test "02" "混合流量压测（100 RPS × 5 分钟）" "02-browse-order.js" \
  -e "USER_TOKEN=${USER_TOKEN}" \
  -e "ADDRESS_ID=${ADDRESS_ID}" \
  -e "PRODUCT_IDS=${PRODUCT_IDS}"

# ── 03：微信支付回调幂等（5 次重发）──────────────────────────────────
echo ""
echo "=== 03. 微信支付回调幂等测试 ==="
echo "⚠️  请确认服务已启动且设置了 WXPAY_MOCK_MODE=true"
echo "⚠️  订单 ${ORDER_ID} 必须处于 pending_payment 状态"

run_test "03" "微信支付回调幂等（5 次重发）" "03-wxpay-idempotent.js" \
  -e "ORDER_ID=${ORDER_ID}" \
  -e "TRANSACTION_ID=${TRANSACTION_ID}"

# ── 04：快递 Token Bucket 限流 ────────────────────────────────────────
run_test "04" "快递 Token Bucket 限流验证" "04-express-ratelimit.js" \
  -e "USER_TOKEN=${USER_TOKEN}" \
  -e "ORDER_ID=${ORDER_ID}"

# ── 05：Redis 抖动降级策略 ────────────────────────────────────────────
echo ""
echo "=== 05. Redis 抖动降级策略验证 ==="
echo "⚠️  此测试需要先手动触发 Redis 故障！"
echo ""
echo "   方式 A（推荐）：docker pause \$(docker ps --filter name=redis -q)"
echo "   方式 B：redis-cli DEBUG SLEEP 60"
echo ""
read -r -p "Redis 故障已触发？继续运行测试 05？[y/N] " confirm
if [[ "${confirm}" =~ ^[Yy]$ ]]; then
  run_test "05" "Redis 抖动降级策略（35 秒）" "05-redis-degradation.js" \
    -e "USER_TOKEN=${USER_TOKEN}" \
    -e "ADDRESS_ID=${ADDRESS_ID}"

  echo ""
  echo "→ 请手动恢复 Redis（如使用 docker pause，执行 docker unpause <container>）"
else
  echo "→ 跳过测试 05"
fi

# ── 汇总 ─────────────────────────────────────────────────────────────
echo ""
echo "=================================================="
echo "压测完成时间：$(date '+%Y-%m-%d %H:%M:%S')"
echo "测试报告目录：${SCRIPT_DIR}/results/"
echo ""

if [ "$OVERALL_PASS" = true ]; then
  echo "✅ 全部测试通过，可进入阶段 5"
  exit 0
else
  echo "❌ 部分测试未通过，请修复后重跑，未达标不得进入阶段 5"
  exit 1
fi
