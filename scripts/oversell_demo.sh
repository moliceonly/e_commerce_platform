#!/usr/bin/env bash
# 超卖并发演示（阶段 E）
# 用法：
#   export TOKEN=... PRODUCT_ID=1 QTY=1 CONCURRENCY=50
#   ./scripts/oversell_demo.sh
set -euo pipefail
BASE="${BASE_URL:-http://127.0.0.1:8080}"
TOKEN="${TOKEN:?set TOKEN from login}"
PRODUCT_ID="${PRODUCT_ID:?}"
QTY="${QTY:-1}"
N="${CONCURRENCY:-30}"

echo "firing $N concurrent place-order against product=$PRODUCT_ID"
for i in $(seq 1 "$N"); do
  curl -sS -X POST "$BASE/api/v1/orders" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":$QTY}]}" &
done
wait
echo "done — check product stock vs sum(order_items.quantity)"
