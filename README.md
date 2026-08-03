# e_commerce_platform

对齐 `go-web-roadmap.html` **Part 03**：简易电商。

- 各阶段对照：见 [TODO.md](./TODO.md)
- Gin API 速查：见 [GIN.md](./GIN.md)

**进度：阶段 A、B 已完成 → 下一步阶段 C（JWT）。**

## 阶段总览

| 阶段 | 目标 | 状态 |
|------|------|------|
| A · 3.1 | `go run` + `/healthz` + env + `/api/v1` 商品 CRUD | ✅ |
| B | 加购、事务下单扣库存、订单状态流转 | ✅ |
| C · 3.2 | 注册登录 JWT、鉴权中间件、越权校验 | ☐ |
| D · 3.3 | request_id、订单分页、优雅停机 | ☐ |
| E · 3.5 | 超卖压测与加固 | ☐（DeductStockTx 已有 FOR UPDATE） |
| F | 单测、Docker | ☐ |

## 目录

```text
cmd/server/           # 入口：配置、DB、组装依赖、起 HTTP
internal/
  config/             # env 配置
  model/              # User / Product / Cart / Order
  repository/         # DB 访问（含 DeductStockTx）
  service/            # 业务编排
  handler/            # Gin HTTP + 路由
  middleware/         # RequestID / JWTAuth（阶段 C/D）
  auth/               # bcrypt / JWT（阶段 C）
  response/           # {code,message,data}
scripts/              # 超卖脚本
deployments/          # Docker
```

依赖方向：**handler → service → repository**（handler 禁止直连 DB）。

## 启动

```bash
cd train_hub/e_commerce_platform
export GOPROXY=https://goproxy.cn,direct   # 或你能用的代理
export MYSQL_DSN='trainer:Train2026Lib!@tcp(127.0.0.1:3306)/training_lib?charset=utf8mb4&parseTime=True&loc=Local'
go mod tidy
go run ./cmd/server
```

MySQL 可复用 Part02：`trainer` / `Train2026Lib!` / `training_lib`。

## 冒烟测试（A + B）

```bash
# A · 探活与商品
curl http://127.0.0.1:8080/healthz
curl -X POST http://127.0.0.1:8080/api/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"键盘","price":9900,"stock":10}'
curl 'http://127.0.0.1:8080/api/v1/products?page=1&page_size=20'
curl http://127.0.0.1:8080/api/v1/products/1

# B · 加购 / 下单 / 流转（userID 暂写死为 1）
curl -X POST http://127.0.0.1:8080/api/v1/cart/items \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":1}'
curl -X POST http://127.0.0.1:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"product_id":1,"quantity":1}]}'
# 将 :id 换成上一步返回的订单 ID
curl -X POST http://127.0.0.1:8080/api/v1/orders/1/transition \
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

事务内 `SELECT ... FOR UPDATE`（`DeductStockTx` 已实现），或 `UPDATE stock=stock-? WHERE id=? AND stock>=?` 检查 `RowsAffected`。阶段 E 用 `scripts/oversell_demo.sh` 压测确认。
