# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。勾完一阶段再进下一阶段。

**当前进度：阶段 A–E（含路线图 3.1 / 3.2 / 3.3 / 3.5）已完成；最后阶段 F（3.4 测试 + Docker）。**

---

## 阶段 A · 3.1 空壳可跑 ✅

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ✅ | `handler.Healthz` | 探活 |
| ✅ | `config.Load` | 读 env |
| ✅ | `response.OK` / `Fail` | 统一响应 |
| ✅ | `cmd/server/main.go` → `main` | MySQL + GORM；组装 repo→svc→`handler.Deps` |
| ✅ | `repository.ProductRepo.AutoMigrate` | 五张表 |
| ✅ | `handler.NewRouter` | `/api/v1` + 商品路由 |
| ✅ | `handler.ProductHandler.Create/Get/List` | bind / path / query → CatalogService |
| ✅ | `service.CatalogService.CreateProduct/GetProduct/ListProducts` | → ProductRepo |
| ✅ | `repository.ProductRepo.Create/Get/List` | GORM CRUD |

**验收：** `go run ./cmd/server` + `curl /healthz`；能创建/查商品。

---

## 阶段 B · 加购下单与状态流转 ✅

| 状态 | 目录 / 符号　　　　　　　　　　　　　　　　　　 | 要做什么　　　　　　　　　　　　　　　　　　　|
| ------| -------------------------------------------------| -----------------------------------------------|
| ✅　　| `handler.CartHandler.Add`　　　　　　　　　　　 | 暂写死 userID=1；bind → CartService　　　　　 |
| ✅　　| `service.CartService.Add`　　　　　　　　　　　 | `Products.Get` → `CartRepo.Upsert`　　　　　　|
| ✅　　| `repository.CartRepo.Upsert` / `ListByUser`　　 | 购物车读写　　　　　　　　　　　　　　　　　　|
| ✅　　| `handler.OrderHandler.Place`　　　　　　　　　　| bind items → PlaceOrder　　　　　　　　　　　 |
| ✅　　| `service.OrderService.PlaceOrder`　　　　　　　 | 事务：扣库存 → 写订单行（阶段 D 已加清车）　 |
| ✅　　| `repository.ProductRepo.DeductStockTx`　　　　　| 事务内 `FOR UPDATE` 扣库存　　　　　　　　　　|
| ✅　　| `repository.OrderRepo.CreateWithItems`　　　　　| 同事务写 Order + OrderItem　　　　　　　　　　|
| ✅　　| `handler.OrderHandler.Transition`　　　　　　　 | bind status → Transition　　　　　　　　　　　|
| ✅　　| `service.OrderService.Transition`　　　　　　　 | 状态机 pending→paid→shipped→done　　　　　　　|
| ✅　　| `repository.OrderRepo.GetByID` / `UpdateStatus` | 读单、条件更新状态　　　　　　　　　　　　　　|
| ✅　　| `handler.NewRouter`　　　　　　　　　　　　　　 | 挂 cart / orders / transition　　　　　　　　 |

**验收：** 下单后库存减少；能推进订单状态。  
**说明：** userID 已在阶段 C 改为 JWT context（`c.GetUint("userID")`）。

---

## 阶段 C · 3.2 认证鉴权 ✅

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ✅ | `auth.HashPassword` / `CheckPassword` | bcrypt |
| ✅ | `auth.SignToken` / `ParseToken` + `Claims` | JWT HS256 |
| ✅ | `repository.UserRepo.Create` / `FindByEmail` | 用户表 |
| ✅ | `service.AuthService.Register` / `Login` | 哈希入库 / 验密发 token |
| ✅ | `handler.AuthHandler.Register` / `Login` | bind + binding |
| ✅ | `middleware.JWTAuth` | Bearer → ParseToken → `c.Set(CtxUserID)` |
| ✅ | `handler.NewRouter` | 公开 auth/商品；保护 cart/orders + `JWTAuth` |
| ✅ | `handler.CartHandler.Add` / `OrderHandler.*` | `c.GetUint("userID")` |
| ✅ | `service.OrderService.Transition` | 校验订单归属当前用户 |

**验收：** 无 token 访问订单 401；登录后带 Bearer 可下单。

---

## 阶段 D · 3.3 横切 ✅

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ✅ | `middleware.RequestID` | 读/生成 `X-Request-ID`，`c.Set` + 写响应头 |
| ✅ | `handler.NewRouter` | `r.Use(middleware.RequestID())` |
| ✅ | `handler.OrderHandler.List` | query `page`/`page_size`/`status` |
| ✅ | `service.OrderService.ListOrders` + `repository.OrderRepo.ListByUser` | 分页过滤 |
| ✅ | `handler.ProductHandler.List` | 分页已接 |
| ✅ | `cmd/server/main.go` → `main` | `http.Server` + `SIGINT/SIGTERM` + `Shutdown` |
| ✅ | `PlaceOrder` + `CartRepo.ClearByUser` | 同事务按本次 `product_id IN` 清车 |

**验收：** 响应带 `X-Request-ID`；订单列表可分页；`Ctrl+C` 优雅退出。

---

## 阶段 E · 3.5 超卖 ✅

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ✅ | `scripts/oversell_demo.sh` | 注册登录 + 建商品 + 并发下单 |
| ✅ | `repository.ProductRepo.DeductStockTx` | `FOR UPDATE`；压测确认不超卖 |
| ✅ | `service.OrderService.PlaceOrder` | 同 `tx`；库存与销量一致 |

**验收：** `stock + sold == 初始库存`，无负数库存。

---

## 阶段 F · 测试与 Docker ← 当前（最后阶段）

对应路线图 **3.4**（测试）+ 部署（Dockerfile / compose，与路线图 3.5 后半重叠）。  
`deployments/Dockerfile`、`docker-compose.yml` **骨架已有**，本阶段目标是跑通并补测试。

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `internal/service/*_test.go`（自建） | 至少测 `PlaceOrder` 或 `DeductStockTx`（成功扣减 / 库存不足失败） |
| ☐ | `internal/handler/*_test.go`（自建） | `httptest`：`GET /healthz`；可选 Register/Login |
| ☐ | `deployments/Dockerfile` | 多阶段构建已有；确认 `go build ./cmd/server` 镜像能起 |
| ☐ | `deployments/docker-compose.yml` | `compose up --build` → app + mysql(+redis) |
| ☐ | `README.md` | 补 `go test` / `compose` 命令（骨架说明已写） |

### F1 · 单测怎么下手

```bash
# 建议先从最薄的测起
go test ./internal/auth/ -v          # Hash/Check、Sign/Parse（可不连库）
# 再写 service / handler 测试文件后：
go test ./internal/... -count=1
```

- **auth**：纯函数，最适合先写。  
- **handler**：`r := NewRouter(...)`，`httptest.NewRecorder` + `http.NewRequest`，断言状态码与 JSON。  
- **service 下单**：需要 DB——可用本机 `training_lib` 的测试库名，或在测试里 `gorm` 连 MySQL；测「库存够下单成功 / 不够返回 error」。

参考 Part02：`httptest` 写法（`training_golang` 里 httppractice）。

### F2 · Docker 怎么验收

```bash
cd ~/train_hub/e_commerce_platform
docker compose -f deployments/docker-compose.yml up --build
# 另开终端
curl -i http://127.0.0.1:8080/healthz
```

compose 里 app 的 `MYSQL_DSN` 指向服务名 `mysql`；本机端口 `8080` / `3306`。  
停：`Ctrl+C` 或 `docker compose -f deployments/docker-compose.yml down`。

**验收：** `go test ./internal/...` 通过；`compose up` 后 `/healthz` 200。

---

## 建议实现顺序（阶段 F）

```text
1. auth 单测（无 DB）
2. handler /healthz httptest
3. （可选）service PlaceOrder 连测试库
4. docker compose up --build + curl healthz
```
