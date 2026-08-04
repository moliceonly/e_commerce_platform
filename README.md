# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 阶段对照：[TODO.md](./TODO.md)
- Gin 速查：[GIN.md](./GIN.md)

**进度：阶段 A–G 完成**（含 Redis 缓存、接口 mock、结构化日志）。  
**运行方式以 Docker Compose 为准**（不要求本机安装 Go / MySQL）。

## 快速开始

```bash
# 克隆仓库
git clone git@github.com:moliceonly/e_commerce_platform.git
cd e_commerce_platform

# 若 8080 / 3306 / 6379 被本机进程占用，先释放
sudo bash scripts/free_ports.sh

# 构建并启动 app + MySQL + Redis（前台日志；Ctrl+C 可停）
sudo docker compose -f deployments/docker-compose.yml up --build
```

另开终端：

```bash
# 探活：应返回 200，响应头带 X-Request-ID
curl -i http://127.0.0.1:8080/healthz
```

| 服务 | 地址 |
|------|------|
| HTTP | http://127.0.0.1:8080 |
| MySQL | `127.0.0.1:3306`，库 `training_lib`，用户 `trainer`，密码 `Train2026Lib!` |
| Redis | `127.0.0.1:6379` |

停止：

```bash
sudo docker compose -f deployments/docker-compose.yml down
# 同时删除 MySQL/Redis 数据卷（库与缓存全部清空）：
# sudo docker compose -f deployments/docker-compose.yml down -v
```

改代码后重新构建 app：

```bash
sudo docker compose -f deployments/docker-compose.yml up --build -d app
```

国内拉镜像慢时：

- **Compose / Dockerfile 已走 DaoCloud 前缀**（`docker.m.daocloud.io/library/...`），一般可直接 `up --build`。
- 系统级加速可复制示例并重启 Docker：

```bash
sudo cp deployments/daemon.json.example /etc/docker/daemon.json
sudo systemctl daemon-reload
sudo systemctl restart docker
# 验证：sudo docker info | grep -A5 'Registry Mirrors'
```

镜像构建已使用 `GOPROXY=https://goproxy.cn,direct`。

## 清空业务数据（保留容器）

Compose 仍在跑时，可用下面命令**只清表数据**，不必 `down -v`：

```bash
# 清空 MySQL 全部业务表（注意外键顺序）
sudo docker compose -f deployments/docker-compose.yml exec -T mysql \
  mysql -utrainer -p'Train2026Lib!' training_lib -e "
SET FOREIGN_KEY_CHECKS=0;
TRUNCATE TABLE order_items;
TRUNCATE TABLE orders;
TRUNCATE TABLE cart_items;
TRUNCATE TABLE products;
TRUNCATE TABLE users;
SET FOREIGN_KEY_CHECKS=1;
"

# 清空 Redis 当前库（商品缓存、登录失败计数等）
sudo docker compose -f deployments/docker-compose.yml exec -T redis redis-cli FLUSHDB
```

或一次性销毁卷后重建（更彻底）：

```bash
sudo docker compose -f deployments/docker-compose.yml down -v
sudo docker compose -f deployments/docker-compose.yml up --build
```

## 测试（容器内）

连 Compose 里的 MySQL，不依赖本机 Go：

```bash
sudo docker compose -f deployments/docker-compose.yml --profile test run --rm test
```

仅 mock 单测（不连 MySQL，可在有 Go 的环境）：

```bash
go test ./internal/service/ -run Mock -count=1
```

## 冒烟 / 超卖

服务已用 Compose 跑在 `8080` 时，另开终端执行（需本机有 `curl`、`jq`）：

```bash
# API 根地址
BASE=http://127.0.0.1:8080

# —— D/G：request_id + 访问日志 ——
# 探活：看 HTTP 状态与 X-Request-ID 响应头；app 日志应有 http + request_id
curl -i "$BASE/healthz"

# 自定义请求 ID：响应头应回显 my-trace-1，日志同字段一致
curl -i -H 'X-Request-ID: my-trace-1' "$BASE/healthz"

# —— 商品 + G3 缓存 ——
# 创建商品：记下返回里的 id
export PRODUCT_ID=$(curl -s -X POST "$BASE/api/v1/products" \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}' | jq -r '.data.ID // .data.id')
echo "PRODUCT_ID=$PRODUCT_ID"

# 第一次 GET：未命中 → 查库并回填 Redis（key: product:{id}）
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .

# 第二次 GET：应走缓存（app 仍打 get product 日志；可用 redis-cli 确认 key）
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .
sudo docker compose -f deployments/docker-compose.yml exec -T redis \
  redis-cli GET "product:$PRODUCT_ID"

# 商品分页列表（短 TTL 缓存 products:page:…）
curl -s "$BASE/api/v1/products?page=1&page_size=20" | jq .

# —— 鉴权 ——
# 无 Token 下单：应返回 401
curl -s -i -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":1}]}"

# 注册用户（邮箱已存在会失败，可换邮箱）
curl -s -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}'

# 登录：从 JSON 取出 data.token
export TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}' \
  | jq -r '.data.token')
echo "TOKEN=$TOKEN"

# —— G3：登录失败限流（可选单独验证；会占用该邮箱失败计数）——
# 连续用错误密码登录约 5 次后，应返回 too many login tries
for i in 1 2 3 4 5 6; do
  curl -s -X POST "$BASE/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{"email":"demo@example.com","password":"wrong"}'
  echo
done

# —— 加购 / 下单（会失效 product / cart / 列表缓存）——
curl -s -X POST "$BASE/api/v1/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"product_id\":$PRODUCT_ID,\"quantity\":2}"

curl -s -X POST "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":1}]}"

# 下单后详情缓存应已删除（GET 可能为空）
sudo docker compose -f deployments/docker-compose.yml exec -T redis \
  redis-cli GET "product:$PRODUCT_ID"

# 再 GET 一次：重新回填；库存应比创建时少 1
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .

# 订单状态 pending → paid（把 ORDER_ID 换成真实 id）
export ORDER_ID=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=1" | jq -r '.data[0].ID // .data[0].id')
echo "ORDER_ID=$ORDER_ID"
curl -s -X POST "$BASE/api/v1/orders/$ORDER_ID/transition" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"paid"}'

curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=20&status=paid" | jq .

# 超卖压测：自动注册/建商品/并发下单，再按打印的 SQL 查库核对库存
BASE_URL=http://127.0.0.1:8080 ./scripts/oversell_demo.sh
```

## 目录

```text
cmd/server/           # 入口（由 Dockerfile 编译进镜像）
internal/             # 业务代码（applog / cache / service mock 等）
scripts/              # free_ports.sh、oversell_demo.sh
deployments/          # Dockerfile、docker-compose.yml
```

依赖：**handler → service → repository**（Cache 可选）。  
公开：`/healthz`、`/api/v1/auth/*`、商品。  
需登录：`/cart/*`、`/orders*`（`Authorization: Bearer <token>`）。

## 阶段一览

| 阶段 | 路线图　　 | 目标　　　　　　　　　　　　　 | 状态 |
| ------| ------------| --------------------------------| ------|
| A　　| 3.1　　　　| healthz、商品 CRUD　　　　　　 | ✅　　|
| B　　| （业务）　 | 加购、事务下单、状态流转　　　 | ✅　　|
| C　　| 3.2　　　　| JWT 鉴权　　　　　　　　　　　 | ✅　　|
| D　　| 3.3　　　　| request_id、订单分页、优雅停机 | ✅　　|
| E　　| 3.5　　　　| 超卖压测　　　　　　　　　　　 | ✅　　|
| F　　| 3.4 + 部署 | 单测、Docker　　　　　　　　　 | ✅　　|
| G　　| 加练　　　 | slog、接口 mock、Redis 缓存　　 | ✅　　|

## 订单状态

```text
pending → paid → shipped → done
       ↘ cancelled
```
