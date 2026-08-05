# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 阶段对照：[TODO.md](./TODO.md)
- Gin 速查：[GIN.md](./GIN.md)

**进度：阶段 A–G 完成；阶段 H 进行中（H1/H2/H3 已完成，见 [TODO.md](./TODO.md)）。**  
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

Compose 中 `APP_ENV=docker`，镜像内需有 `configs/config.docker.yaml`（已打进镜像；DSN/JWT 仍可由 compose 的 environment 覆盖）。

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

## 安全注意（H2）

- **密钥**：`JWT_SECRET` / DSN 用环境变量或本地 gitignore 的 `configs/config.*.yaml`，勿提交真实生产密钥。
- **CORS**：仅白名单 Origin（默认示例 `http://localhost:3000`）；生产勿 `*` + Credentials。
- **RBAC**：`POST /api/v1/products` 需 JWT 且 `role=admin`；注册默认 `user`。
- **Token**：登录返回短寿命 `token`（access）与长寿命 `refresh`；业务接口只用 access；过期后用 `/auth/refresh` 换发，不必再输密码。

## 冒烟 / 超卖

服务已用 Compose 跑在 `8080` 时，另开终端执行（需本机有 `curl`、`jq`）。  
本机 `go run` 时请先：`sudo docker compose -f deployments/docker-compose.yml up -d mysql redis`，并 `cp configs/config.local.yaml.example configs/config.local.yaml`。

```bash
BASE=http://127.0.0.1:8080
COMPOSE="sudo docker compose -f deployments/docker-compose.yml"

# ========== D/G：request_id ==========
curl -i "$BASE/healthz"
curl -i -H 'X-Request-ID: my-trace-1' "$BASE/healthz"

# ========== H2：CORS 预检（应 204，且带 Allow-*）==========
curl -i -X OPTIONS "$BASE/api/v1/products" \
  -H 'Origin: http://localhost:3000' \
  -H 'Access-Control-Request-Method: GET'

# ========== 注册 / 登录（拿到 access + refresh）==========
curl -s -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}' | jq .

LOGIN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}')
echo "$LOGIN" | jq .
export TOKEN=$(echo "$LOGIN" | jq -r '.data.token')
export REFRESH=$(echo "$LOGIN" | jq -r '.data.refresh')
echo "TOKEN=$TOKEN"
echo "REFRESH=$REFRESH"

# ========== H2：普通 user 建商品 → 403 ==========
curl -s -i -X POST "$BASE/api/v1/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}'

# ========== H2：提权 admin（改库后必须重新登录，JWT 里才有新 role）==========
$COMPOSE exec -T mysql \
  mysql -utrainer -p'Train2026Lib!' training_lib -e \
  "UPDATE users SET role='admin' WHERE email='demo@example.com';"

LOGIN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}')
export TOKEN=$(echo "$LOGIN" | jq -r '.data.token')
export REFRESH=$(echo "$LOGIN" | jq -r '.data.refresh')

# admin 建商品 → 200
export PRODUCT_ID=$(curl -s -X POST "$BASE/api/v1/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}' | jq -r '.data.ID // .data.id')
echo "PRODUCT_ID=$PRODUCT_ID"

# ========== H2：refresh 换发新 access ==========
export TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}" | jq -r '.data.token')
echo "NEW_TOKEN=$TOKEN"
# 用新 token 访问需登录接口（应成功）
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=5" | jq .

# ========== H3：Swagger / 头像上传 / job tick ==========
# 浏览器打开 http://127.0.0.1:8080/swagger/index.html
# 上传后文件落在宿主机 data/uploads/（compose 已挂卷）
# curl -s -X POST "$BASE/api/v1/me/avatar" \
#   -H "Authorization: Bearer $TOKEN" \
#   -F "file=@/path/to/avatar.png" | jq .
# 返回 url 形如 http://127.0.0.1:8080/static/{uid}_avatar.png ，浏览器可直接打开
# job 每分钟日志：job tick（$COMPOSE logs -f app）

# ========== G3：商品缓存 ==========
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .
$COMPOSE exec -T redis redis-cli GET "product:$PRODUCT_ID"
curl -s "$BASE/api/v1/products?page=1&page_size=20" | jq .

# ========== 鉴权：无 Token 下单 → 401 ==========
curl -s -i -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":1}]}"

# ========== 加购 / 下单 / 状态（会失效商品缓存）==========
curl -s -X POST "$BASE/api/v1/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"product_id\":$PRODUCT_ID,\"quantity\":2}" | jq .

curl -s -X POST "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"product_id\":$PRODUCT_ID,\"quantity\":1}]}" | jq .

$COMPOSE exec -T redis redis-cli GET "product:$PRODUCT_ID"
curl -s "$BASE/api/v1/products/$PRODUCT_ID" | jq .

export ORDER_ID=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=1" | jq -r '.data[0].ID // .data[0].id')
echo "ORDER_ID=$ORDER_ID"
curl -s -X POST "$BASE/api/v1/orders/$ORDER_ID/transition" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"paid"}' | jq .

curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=20&status=paid" | jq .

# ========== G3：登录失败限流（可选；会占用该邮箱计数）==========
# for i in 1 2 3 4 5 6; do
#   curl -s -X POST "$BASE/api/v1/auth/login" \
#     -H 'Content-Type: application/json' \
#     -d '{"email":"demo@example.com","password":"wrong"}'; echo
# done

# ========== 超卖压测（脚本内会提权 admin 再建商品）==========
BASE_URL=http://127.0.0.1:8080 ./scripts/oversell_demo.sh
```

### H1 配置冒烟（本机 go run）

```bash
sudo docker compose -f deployments/docker-compose.yml up -d mysql redis
cp -n configs/config.local.yaml.example configs/config.local.yaml

# 默认读 configs/config.local.yaml
APP_ENV=local go run ./cmd/server

# 另开终端：用环境变量覆盖 JWT（进程日志 env=staging）
# APP_ENV=staging JWT_SECRET=smoke-override go run ./cmd/server
```

## 目录

```text
cmd/server/           # 入口（由 Dockerfile 编译进镜像）
configs/              # Viper 多环境 yaml（*.example 可提交）
internal/             # 业务（applog / cache / middleware / errcode 等）
docs/                 # openapi / runbook
scripts/              # free_ports.sh、oversell_demo.sh
deployments/          # Dockerfile、docker-compose.yml、nginx 示例
```

依赖：**handler → service → repository**（Cache 可选）。  
公开：`/healthz`、`/api/v1/auth/*`、商品查询。  
需登录：`/cart/*`、`/orders*`。  
需 **admin**：`POST /api/v1/products`。

## 阶段一览

| 阶段 | 路线图　　　| 目标　　　　　　　　　　　　　 | 状态 |
| ------| -------------| --------------------------------| ------|
| A　　| 3.1　　　　 | healthz、商品 CRUD　　　　　　 | ✅　　|
| B　　| （业务）　　| 加购、事务下单、状态流转　　　 | ✅　　|
| C　　| 3.2　　　　 | JWT 鉴权　　　　　　　　　　　 | ✅　　|
| D　　| 3.3　　　　 | request_id、订单分页、优雅停机 | ✅　　|
| E　　| 3.5　　　　 | 超卖压测　　　　　　　　　　　 | ✅　　|
| F　　| 3.4 + 部署　| 单测、Docker　　　　　　　　　 | ✅　　|
| G　　| 加练　　　　| slog、接口 mock、Redis 缓存　　| ✅　　|
| H　　| Part03 补齐 | H1～H3 完成；H4～H5 进行中　　　| ◐　　|

## 订单状态

```text
pending → paid → shipped → done
       ↘ cancelled
```
