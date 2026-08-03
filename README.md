# e_commerce_platform

对齐 [`go-web-roadmap.html`](../go-web-roadmap.html) **Part 03**：简易电商。

- 各阶段要改哪个函数：见 [TODO.md](./TODO.md)
- Gin API 速查：见 [GIN.md](./GIN.md)

**进度：阶段 A / B / C（含路线图 3.1、3.2）已完成 → 下一步阶段 D（3.3）。**

## 阶段与路线图编号对应

本仓库用 **A–F** 做实现顺序；括号内是路线图章节。**B 没有单独的 3.x 编号**，是接在 3.1 之后的业务能力，再进入 3.2 鉴权。

| 本仓库阶段 | 路线图 | 目标 | 状态 |
|------------|--------|------|------|
| A | **3.1** | `go run` + `/healthz` + env + `/api/v1` 商品 CRUD | ✅ |
| B | （业务，无单独章节） | 加购、事务下单扣库存、订单状态流转 | ✅ |
| C | **3.2** | 注册登录 JWT、鉴权中间件、越权校验 | ✅ |
| D | **3.3** | request_id、订单分页、优雅停机 | ☐ |
| E | **3.5** | 超卖压测与加固 | ☐（`DeductStockTx` 已有 FOR UPDATE） |
| F | **3.4** + 部署 | 单测、Dockerfile / compose | ☐ |

> 说明：实现顺序是 A→B→C→D→E→F；路线图编号是 3.1→3.2→3.3→3.4→3.5。表中 E/F 与 3.4/3.5 **刻意交叉**：先压测超卖（3.5），再补测试与 Docker（3.4）。若你严格按路线图读，3.4 对应阶段 F，3.5 对应阶段 E。

## 目录

```text
cmd/server/           # 入口：配置、DB、组装依赖、起 HTTP
internal/
  config/             # env 配置
  model/              # User / Product / Cart / Order
  repository/         # DB 访问（含 DeductStockTx）
  service/            # 业务编排
  handler/            # Gin HTTP + 路由
  middleware/         # JWTAuth ✅；RequestID 待 D
  auth/               # bcrypt / JWT ✅
  response/           # {code,message,data}
scripts/              # 超卖脚本
deployments/          # Docker
```

依赖方向：**handler → service → repository**（handler 禁止直连 DB）。

公开路由：`/healthz`、`/api/v1/auth/*`、商品 CRUD。  
需登录：`/cart/*`、`/orders*`（`Authorization: Bearer <token>`）。

## 启动

```bash
cd train_hub/e_commerce_platform
export GOPROXY=https://goproxy.cn,direct   # 或你能用的代理
export MYSQL_DSN='trainer:Train2026Lib!@tcp(127.0.0.1:3306)/training_lib?charset=utf8mb4&parseTime=True&loc=Local'
# 可选：export JWT_SECRET=dev-secret-change-me
go mod tidy
go run ./cmd/server
```

MySQL 可复用 Part02：`trainer` / `Train2026Lib!` / `training_lib`。

## 清空练习数据（可选）

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

## 冒烟测试（A + B + C）

```bash
# —— A · 探活与商品 ——
curl http://127.0.0.1:8080/healthz
curl -X POST http://127.0.0.1:8080/api/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}'
curl 'http://127.0.0.1:8080/api/v1/products?page=1&page_size=20'
curl http://127.0.0.1:8080/api/v1/products/1

# —— C · 无 token 下单应 401 ——
curl -s -X POST http://127.0.0.1:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'

# —— C · 注册 / 登录 ——
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}'
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"123456"}'
# 将 TOKEN 换成登录返回的 data.token
export TOKEN='粘贴token'

# —— B+C · 加购 / 下单 / 流转（需 Bearer）——
curl -s -X POST http://127.0.0.1:8080/api/v1/cart/items \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":1}'
curl -s -X POST http://127.0.0.1:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'
# 将下面路径中的 1 换成订单 ID
curl -s -X POST http://127.0.0.1:8080/api/v1/orders/1/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"paid"}'
```

查库订单状态：

```bash
mysql -u trainer -p -h 127.0.0.1 training_lib \
  -e "SELECT id, user_id, status, total FROM orders ORDER BY id DESC;"
```

## 订单状态约定

```text
pending → paid → shipped → done
       ↘ cancelled
```

## 超卖修复方向

事务内 `SELECT ... FOR UPDATE`（`DeductStockTx` 已实现），或 `UPDATE stock=stock-? WHERE id=? AND stock>=?` 检查 `RowsAffected`。阶段 E（路线图 3.5）用 `scripts/oversell_demo.sh` 压测确认。
