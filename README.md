# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 阶段对照：[TODO.md](./TODO.md)
- Gin 速查：[GIN.md](./GIN.md)

**进度：阶段 A–F 完成；加练 G（Redis / 接口 mock / 结构化日志）见 [TODO.md](./TODO.md)。**  
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
# 清空库数据：down -v
```

改代码后重新构建 app：

```bash
sudo docker compose -f deployments/docker-compose.yml up --build -d app
```

国内拉镜像慢时，可为 Docker 配置 registry 镜像；镜像构建已使用 `GOPROXY=https://goproxy.cn,direct`。

## 测试（容器内）

连 Compose 里的 MySQL，不依赖本机 Go：

```bash
sudo docker compose -f deployments/docker-compose.yml --profile test run --rm test
```

## 冒烟 / 超卖

服务已用 Compose 跑在 `8080` 时，另开终端执行（需本机有 `curl`、`jq`）：

```bash
# API 根地址
BASE=http://127.0.0.1:8080

# 探活：看 HTTP 状态与 X-Request-ID 响应头
curl -i "$BASE/healthz"

# 自定义请求 ID：响应头应回显 my-trace-1
curl -i -H 'X-Request-ID: my-trace-1' "$BASE/healthz"

# 创建商品：记下返回里的 id（下面假定为 1）
curl -s -X POST "$BASE/api/v1/products" \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}'

# 商品分页列表
curl -s "$BASE/api/v1/products?page=1&page_size=20"

# 无 Token 下单：应返回 401
curl -s -i -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'

# 注册用户（邮箱已存在会失败，可换邮箱）
curl -s -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}'

# 登录：从 JSON 取出 data.token 写入环境变量 TOKEN
export TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}' \
  | jq -r '.data.token')
echo "TOKEN=$TOKEN"

# 加购：购物车写入商品 1，数量 2（需 Bearer）
curl -s -X POST "$BASE/api/v1/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":2}'

# 下单：扣库存并生成订单；记下返回的订单 id
curl -s -X POST "$BASE/api/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'

# 把下面的 1 换成真实订单 id 后，推进状态 pending → paid
export ORDER_ID=1
curl -s -X POST "$BASE/api/v1/orders/$ORDER_ID/transition" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"paid"}'

# 订单列表（分页 + 按 status 过滤）
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/orders?page=1&page_size=20&status=paid"

# 超卖压测脚本：自动注册/建商品/并发下单，再按打印的 SQL 查库核对库存
BASE_URL=http://127.0.0.1:8080 ./scripts/oversell_demo.sh
```

## 目录

```text
cmd/server/           # 入口（由 Dockerfile 编译进镜像）
internal/             # 业务代码
scripts/              # free_ports.sh、oversell_demo.sh
deployments/          # Dockerfile、docker-compose.yml
```

依赖：**handler → service → repository**。  
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

## 订单状态

```text
pending → paid → shipped → done
       ↘ cancelled
```
