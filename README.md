# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 阶段对照：[TODO.md](./TODO.md)
- Gin 速查：[GIN.md](./GIN.md)

**进度：阶段 A–F 全部完成（路线图 3.1–3.5）。**  
**运行方式以 Docker Compose 为准**（不要求本机安装 Go / MySQL）。

## 快速开始

```bash
git clone <本仓库地址>
cd e_commerce_platform

# 若 8080 / 3306 / 6379 被本机进程占用，先释放：
sudo bash scripts/free_ports.sh

# 启动 app + MySQL + Redis
sudo docker compose -f deployments/docker-compose.yml up --build
```

另开终端：

```bash
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

```bash
BASE=http://127.0.0.1:8080

curl -i "$BASE/healthz"
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

| 阶段 | 路线图 | 目标 | 状态 |
|------|--------|------|------|
| A | 3.1 | healthz、商品 CRUD | ✅ |
| B | （业务） | 加购、事务下单、状态流转 | ✅ |
| C | 3.2 | JWT 鉴权 | ✅ |
| D | 3.3 | request_id、订单分页、优雅停机 | ✅ |
| E | 3.5 | 超卖压测 | ✅ |
| F | 3.4 + 部署 | 单测、Docker | ✅ |

## 订单状态

```text
pending → paid → shipped → done
       ↘ cancelled
```
