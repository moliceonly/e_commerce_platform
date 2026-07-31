# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。勾完一阶段再进下一阶段。

---

## 阶段 A · 3.1 空壳可跑

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ✅ 已有 | `handler.Healthz` | 探活，可先不动 |
| ✅ 已有 | `config.Load` | 读 env；可改 viper |
| ✅ 已有 | `response.OK` / `Fail` | 统一响应，直接用 |
| ☐ | `cmd/server/main.go` → `main` | 开 MySQL + GORM；组装 repo→svc→`handler.Deps` |
| ☐ | `repository.ProductRepo.AutoMigrate` | `AutoMigrate` 五张表 |
| ☐ | `handler.NewRouter` | `r.Group("/api/v1")`；挂商品路由（可先不鉴权） |
| ☐ | `handler.ProductHandler.Create/Get/List` | bind / path / query → 调 CatalogService |
| ☐ | `service.CatalogService.CreateProduct/GetProduct/ListProducts` | 校验入参 → ProductRepo |
| ☐ | `repository.ProductRepo.Create/Get/List` | GORM CRUD |

**验收：** `go run ./cmd/server` + `curl /healthz`；能创建/查商品。

---

## 阶段 B · 加购下单与状态流转

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `handler.CartHandler.Add` |（可暂时写死 userID=1）bind → CartService |
| ☐ | `service.CartService.Add` | 查商品存在 → CartRepo.Upsert |
| ☐ | `repository.CartRepo.Upsert` / `ListByUser` | 购物车读写 |
| ☐ | `handler.OrderHandler.Place` | bind items → OrderService.PlaceOrder |
| ☐ | `service.OrderService.PlaceOrder` | **事务**：扣库存 → 写订单行 →（可选）清购物车 |
| ☐ | `repository.ProductRepo.DeductStockTx` | 事务内扣库存（可先简版，阶段 E 再加固） |
| ☐ | `repository.OrderRepo.CreateWithItems` | 同事务写 `Order` + `OrderItem` |
| ☐ | `handler.OrderHandler.Transition` | bind status → OrderService.Transition |
| ☐ | `service.OrderService.Transition` | 状态机 pending→paid→shipped→done |
| ☐ | `repository.OrderRepo.GetByID` / `UpdateStatus` | 读单、条件更新状态 |
| ☐ | `handler.NewRouter` | 挂 `POST /cart/items`、`POST /orders`、`POST /orders/:id/transition` |

**验收：** 下单后库存减少；能推进订单状态。

---

## 阶段 C · 3.2 认证鉴权

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `auth.HashPassword` / `CheckPassword` | bcrypt |
| ☐ | `auth.SignToken` / `ParseToken` + `Claims` | JWT HS256 |
| ☐ | `repository.UserRepo.Create` / `FindByEmail` | 用户表 |
| ☐ | `service.AuthService.Register` / `Login` | 哈希入库 / 验密发 token |
| ☐ | `handler.AuthHandler.Register` / `Login` | bind + binding 标签校验 |
| ☐ | `middleware.JWTAuth` | Bearer → ParseToken → `c.Set(CtxUserID)` |
| ☐ | `handler.NewRouter` | 公开：`/auth/*`、商品；保护：cart/orders + `JWTAuth` |
| ☐ | `handler.CartHandler.Add` / `OrderHandler.*` | 用 `c.GetUint(middleware.CtxUserID)`，禁止信任 body 里的 user_id |
| ☐ | `service.OrderService.Transition` | **越权**：订单 `user_id` 必须等于当前用户 |

**验收：** 无 token 访问订单 401；用户 A 不能改用户 B 订单。

---

## 阶段 D · 3.3 横切

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `middleware.RequestID` | 读/生成 `X-Request-ID`，`c.Set` + 写响应头 |
| ☐ | `handler.NewRouter` | `r.Use(middleware.RequestID())` |
| ☐ | `handler.OrderHandler.List` | query `page`/`page_size`/`status` |
| ☐ | `service`（可加 `ListOrders`）+ `repository.OrderRepo.ListByUser` | 分页过滤 |
| ☐ | `handler.ProductHandler.List` | 补充分页参数 |
| ☐ | `cmd/server/main.go` → `main` | `http.Server` + `SIGINT/SIGTERM` + `Shutdown` |

**验收：** 响应带 `X-Request-ID`；订单列表可分页；`Ctrl+C` 优雅退出。

---

## 阶段 E · 3.5 超卖

| 状态 | 目录 / 符号 | 要做什么 |
|------|-------------|----------|
| ☐ | `scripts/oversell_demo.sh` | 填好 TOKEN / PRODUCT_ID 后并发打 |
| ☐ | `repository.ProductRepo.DeductStockTx` | `FOR UPDATE` 或 `UPDATE ... WHERE stock>=?` + `RowsAffected` |
| ☐ | `service.OrderService.PlaceOrder` | 确保全程同一 `tx`，失败回滚 |

**验收：** 修复前能超卖；修复后再跑脚本库存不错乱。

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

每做一个 API，通常一次打通三层：

```text
NewRouter 挂路由
  → Handler（Gin 读请求、写响应）
    → Service（业务规则）
      → Repository（SQL/GORM）
```

可选工具包（阶段 C）：`internal/auth/*`  
可选中间件（C/D）：`internal/middleware/*`
