# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 阶段对照：[TODO.md](./TODO.md)
- Gin 速查：[GIN.md](./GIN.md)

**进度：阶段 A–F 全部完成（路线图 3.1–3.5）。**

## 阶段与路线图编号对应

本仓库用 **A–F** 做实现顺序；括号内是路线图章节。**B 没有单独的 3.x 编号**。

| 本仓库阶段 | 路线图 | 目标 | 状态 |
|------------|--------|------|------|
| A | **3.1** | `go run` + `/healthz` + env + `/api/v1` 商品 CRUD | ✅ |
| B | （业务） | 加购、事务下单扣库存、订单状态流转 | ✅ |
| C | **3.2** | 注册登录 JWT、鉴权中间件、越权校验 | ✅ |
| D | **3.3** | request_id、订单分页、优雅停机 | ✅ |
| E | **3.5** | 超卖压测与加固 | ✅ |
| F | **3.4** + 部署 | 单测、Dockerfile / compose | ✅ |

> 实现顺序 A→B→C→D→E→F；路线图编号 3.1→3.2→3.3→3.4→3.5。E/F 与 3.4/3.5 交叉：先超卖压测，再测试与 Docker。

## 目录

```text
cmd/server/           # 入口：配置、DB、依赖组装、优雅停机
internal/
  config/             # env
  model/              # User / Product / Cart / Order
  repository/         # GORM（含 DeductStockTx）
  service/            # 业务
  handler/            # HTTP + 路由 + 测试
  middleware/         # JWTAuth、RequestID
  auth/               # bcrypt / JWT + 测试
  response/           # {code,message,data}
scripts/              # oversell_demo.sh
deployments/          # Dockerfile、docker-compose.yml
```

依赖：**handler → service → repository**。

公开：`/healthz`、`/api/v1/auth/*`、商品。  
需登录：`/cart/*`、`/orders*`（`Authorization: Bearer <token>`）。

## 本地启动

```bash
cd train_hub/e_commerce_platform
export GOPROXY=https://goproxy.cn,direct
export MYSQL_DSN='trainer:Train2026Lib!@tcp(127.0.0.1:3306)/training_lib?charset=utf8mb4&parseTime=True&loc=Local'
go mod tidy
go run ./cmd/server
```

MySQL：`trainer` / `Train2026Lib!` / `training_lib`。停机：`Ctrl+C`。

## Docker Compose

需已安装 Docker 与 Compose（Ubuntu 可用 `docker-compose-v2`）。访问 Docker Hub / `proxy.golang.org` 超时时，可配 registry 镜像，Dockerfile 已设 `GOPROXY=https://goproxy.cn,direct`。

```bash
cd train_hub/e_commerce_platform
sudo docker compose -f deployments/docker-compose.yml up --build
```

默认端口映射（避免与本机 MySQL / `go run` 冲突）：

| 服务 | 宿主机 | 容器内 |
|------|--------|--------|
| app  | **8081** | 8080 |
| mysql | **3307** | 3306 |
| redis | **6380** | 6379 |

```bash
curl -i http://127.0.0.1:8081/healthz
BASE_URL=http://127.0.0.1:8081 ./scripts/oversell_demo.sh

# 停掉
sudo docker compose -f deployments/docker-compose.yml down
# 连数据卷清空：down -v
```

本机连 compose 里的 MySQL：

```bash
mysql -u trainer -p -h 127.0.0.1 -P 3307 training_lib
```

改 Go 代码后需重新构建 app：`up --build`。

## 测试

```bash
go test ./internal/auth/ ./internal/handler/ ./internal/service/ -count=1
```

`handler` / `service` 测试依赖本机 MySQL（连不上会 Skip）。

## 清空练习数据（本机库）

```bash
mysql -u trainer -p -h 127.0.0.1 training_lib -e "
SET FOREIGN_KEY_CHECKS=0;
TRUNCATE TABLE order_items;
TRUNCATE TABLE orders;
TRUNCATE TABLE cart_items;
TRUNCATE TABLE products;
TRUNCATE TABLE users;
SET FOREIGN_KEY_CHECKS=1;
"
```

## 冒烟测试（本地 :8080）

```bash
BASE=http://127.0.0.1:8080

curl -i "$BASE/healthz"
curl -i -H 'X-Request-ID: my-trace-1' "$BASE/healthz"

curl -s -X POST "$BASE/api/v1/products" \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}'

curl -s -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}'
export TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}' \
  | jq -r '.data.token')

curl -s -X POST "$BASE/api/v1/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":2}'
curl -s -X POST "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'
export ORDER_ID=1
curl -s -X POST "$BASE/api/v1/orders/$ORDER_ID/transition" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"paid"}'
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=20&status=paid"
```

## 超卖压测

```bash
./scripts/oversell_demo.sh
# Docker 下：BASE_URL=http://127.0.0.1:8081 ./scripts/oversell_demo.sh
```

期望：`stock + sold == 初始库存`，且不超卖。

## 订单状态

```text
pending → paid → shipped → done
       ↘ cancelled
```
