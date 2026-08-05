#!/usr/bin/env bash
set -uo pipefail

BASE="${BASE_URL:-http://127.0.0.1:8080}"
EMAIL="${EMAIL:-oversell-$(date +%s)@example.com}"
PASSWORD="${PASSWORD:-123456}"
STOCK="${STOCK:-30}"
QTY="${QTY:-1}"
N="${CONCURRENCY:-50}"
COMPOSE_FILE="${COMPOSE_FILE:-deployments/docker-compose.yml}"

echo "==> register/login"
curl -sS -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" >/dev/null || true

echo "==> promote to admin (create product requires RequireRole admin)"
sudo docker compose -f "$COMPOSE_FILE" exec -T mysql \
  mysql -utrainer -p'Train2026Lib!' training_lib -e \
  "UPDATE users SET role='admin' WHERE email='${EMAIL}';" >/dev/null

TOKEN="$(
  curl -sS -X POST "$BASE/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  | jq -r '.data.token'
)"
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "login failed, need jq and running server" >&2
  exit 1
fi

echo "==> create product stock=$STOCK"
PRODUCT_ID="$(
  curl -sS -X POST "$BASE/api/v1/products" \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"oversell-demo\",\"price\":100,\"stock\":$STOCK}" \
  | jq -r '.data.ID // .data.id'
)"
if [[ -z "$PRODUCT_ID" || "$PRODUCT_ID" == "null" ]]; then
  echo "create product failed (need admin token)" >&2
  exit 1
fi
echo "    PRODUCT_ID=$PRODUCT_ID  TOKEN ok"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "==> fire $N concurrent orders qty=$QTY"
for i in $(seq 1 "$N"); do
  (
    code="$(
      curl -sS -o "$TMPDIR/$i.body" -w "%{http_code}" \
        -X POST "$BASE/api/v1/orders" \
        -H "Authorization: Bearer $TOKEN" \
        -H 'Content-Type: application/json' \
        -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":$QTY}]}" \
        2>"$TMPDIR/$i.err" || echo 000
    )"
    echo "$code" >"$TMPDIR/$i.code"
  ) &
done
wait

ok=0
fail=0
for i in $(seq 1 "$N"); do
  code="$(cat "$TMPDIR/$i.code" 2>/dev/null || echo 000)"
  if [[ "$code" == "200" ]] && [[ "$(jq -r '.code // empty' "$TMPDIR/$i.body" 2>/dev/null)" == "0" ]]; then
    ok=$((ok + 1))
  else
    fail=$((fail + 1))
  fi
done

echo "==> ok=$ok fail=$fail total=$N stock=$STOCK"
echo "    check DB: SELECT id,stock FROM products WHERE id=$PRODUCT_ID;"
echo "    check DB: SELECT COALESCE(SUM(quantity),0) FROM order_items WHERE product_id=$PRODUCT_ID AND deleted_at IS NULL;"
