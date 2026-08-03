# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。勾完一阶段再进下一阶段。

**当前进度：阶段 A、B、C、D（含路线图 3.1 / 3.2 / 3.3）已完成；下一阶段 E（3.5 超卖压测）。**

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

## 阶段 E · 3.5 超卖 ← 当前

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `scripts/oversell_demo.sh` | 填好 TOKEN / PRODUCT_ID 后并发打 |
| ⚠️ | `repository.ProductRepo.DeductStockTx` | 已有 `FOR UPDATE`；阶段 E 用脚本压测确认 |
| ⚠️ | `service.OrderService.PlaceOrder` | 已用同一 `tx`；压测验证不超卖 |
| ☐ | （可选）对照实验 | 临时去掉锁/条件更新，观察超卖后再改回 |

**验收：** 并发下单后库存与订单数量一致、不超卖。  
**做法：** 先造低库存商品（如 stock=5），`CONCURRENCY=50` 打 `oversell_demo.sh`，再查：

```sql
SELECT id, stock FROM products WHERE id = ?;
SELECT COALESCE(SUM(quantity),0) FROM order_items WHERE product_id = ?;
-- 期望：成功扣减合计 ≤ 初始库存，且 stock + 已售 = 初始库存
```

---

## 阶段 F · 测试与 Docker

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `internal/service/*_test.go`（自建） | PlaceOrder / DeductStock 单测 |
| ☐ | `internal/handler/*_test.go`（自建） | `httptest.NewRecorder` + Gin 路由 |
| ☐ | `deployments/Dockerfile` / `docker-compose.yml` | 能 `compose up` 起服务 |
| ☐ | `README.md` | 补启动步骤与超卖复现说明 |

---

## 建议实现顺序（按调用链）

```text
NewRouter 挂路由
  → Handler（Gin 读请求、写响应）
    → Service（业务规则）
      → Repository（SQL/GORM）
```

下一阶段重点：跑 `scripts/oversell_demo.sh` 压测，核对库存与订单行不超卖（阶段 E · 3.5）。
