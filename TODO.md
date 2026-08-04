# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。

**进度：阶段 A–G 已完成。**

---

## 阶段 A · 3.1 空壳可跑 ✅

| 状态 | 目录 / 符号　　　　　　　　　　　　　　　| 内容　　　　　　　　　　　　　　　　　　　 |
| ------| ------------------------------------------| --------------------------------------------|
| ✅　　| `handler.Healthz`　　　　　　　　　　　　| 探活　　　　　　　　　　　　　　　　　　　 |
| ✅　　| `config.Load`　　　　　　　　　　　　　　| 读 env　　　　　　　　　　　　　　　　　　 |
| ✅　　| `response.OK` / `Fail`　　　　　　　　　 | 统一响应　　　　　　　　　　　　　　　　　 |
| ✅　　| `cmd/server/main.go`　　　　　　　　　　 | MySQL + GORM；组装 repo→svc→`handler.Deps` |
| ✅　　| `repository.ProductRepo.AutoMigrate`　　 | 五张表　　　　　　　　　　　　　　　　　　 |
| ✅　　| `handler.NewRouter`　　　　　　　　　　　| `/api/v1` + 商品路由　　　　　　　　　　　 |
| ✅　　| `handler.ProductHandler.Create/Get/List` | → CatalogService　　　　　　　　　　　　　 |
| ✅　　| `service.CatalogService.*`　　　　　　　 | → ProductRepo　　　　　　　　　　　　　　　|
| ✅　　| `repository.ProductRepo.Create/Get/List` | GORM CRUD　　　　　　　　　　　　　　　　　|

---

## 阶段 B · 加购下单与状态流转 ✅

| 状态 | 目录 / 符号　　　　　　　　　　　　　　　　　　　| 内容　　　　　　　　　　　|
| ------| --------------------------------------------------| ---------------------------|
| ✅　　| `CartHandler` / `CartService` / `CartRepo`　　　 | 加购 Upsert　　　　　　　 |
| ✅　　| `OrderHandler.Place` / `OrderService.PlaceOrder` | 事务下单 + 扣库存　　　　 |
| ✅　　| `ProductRepo.DeductStockTx`　　　　　　　　　　　| `FOR UPDATE`　　　　　　　|
| ✅　　| `OrderRepo.CreateWithItems`　　　　　　　　　　　| 同事务写订单行　　　　　　|
| ✅　　| `Transition` + 状态机　　　　　　　　　　　　　　| pending→paid→shipped→done |

---

## 阶段 C · 3.2 认证鉴权 ✅

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `auth` bcrypt / JWT | Hash、Check、Sign、Parse |
| ✅ | `UserRepo` / `AuthService` / `AuthHandler` | 注册登录 |
| ✅ | `middleware.JWTAuth` | Bearer → `userID` |
| ✅ | 保护 cart / orders | 无 token 401；订单归属校验 |

---

## 阶段 D · 3.3 横切 ✅

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `middleware.RequestID` | `X-Request-ID` |
| ✅ | `OrderHandler.List` / `ListOrders` | 分页 + status |
| ✅ | `main` 优雅停机 | `http.Server` + `Shutdown` |
| ✅ | `CartRepo.ClearByUser` | 下单同事务按商品清车 |

---

## 阶段 E · 3.5 超卖 ✅

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `scripts/oversell_demo.sh` | 注册登录 + 建商品 + 并发下单 |
| ✅ | `DeductStockTx` + `PlaceOrder` | 压测：`stock + sold == 初始库存` |

---

## 阶段 F · 测试与 Docker ✅

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `internal/auth/auth_test.go` | 哈希、JWT、错 secret / 过期 |
| ✅ | `internal/handler/handler_test.go` | 全路由 httptest（含鉴权链路） |
| ✅ | `internal/service/service_test.go` | PlaceOrder 成功 / 库存不足回滚 |
| ✅ | `deployments/Dockerfile` | 多阶段构建 + `GOPROXY` |
| ✅ | `deployments/docker-compose.yml` | app + mysql + redis；`test` profile |
| ✅ | `scripts/free_ports.sh` | 释放 8080/3306/6379 等本机占用 |
| ✅ | `README.md` | **仅 Docker 启动说明** |

```bash
sudo bash scripts/free_ports.sh
sudo docker compose -f deployments/docker-compose.yml up --build
curl -i http://127.0.0.1:8080/healthz
sudo docker compose -f deployments/docker-compose.yml --profile test run --rm test
```

---

## 阶段 G · 补强（Redis + 接口 mock + 结构化日志）✅

### G1 · 结构化日志 + 全链路 request_id ✅

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ✅ | `internal/applog/applog.go` | `Setup` / `WithRequestID` / `RequestIDFrom` / `FromContext`（`log/slog`） |
| ✅ | `internal/middleware/middleware.go` → `RequestID` | `applog.WithRequestID` 写入 `c.Request.Context()` |
| ✅ | `internal/middleware/middleware.go` → `AccessLog` | method/path/status/latency + request_id |
| ✅ | `internal/handler/router.go` | `AccessLog` 在 `RequestID` 之后；可去掉 `gin.Logger` |
| ✅ | `cmd/server/main.go` | `applog.Setup(cfg.AppEnv)` |
| ✅ | `internal/service/service.go`（关键路径） | Create/Get/PlaceOrder 等 applog；repo 层不打 |

**验收：** `curl -i /healthz` 响应有 `X-Request-ID`；日志带相同 `request_id`。

### G2 · repository 接口 + service mock 单测 ✅

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ✅ | `internal/repository/ports.go` | `ProductStore` / `OrderStore` / `CartStore` / `UserStore` |
| ✅ | `internal/service/service.go` | 字段为接口类型 |
| ✅ | `internal/service/catalog_mock_test.go` | mock 内存仓；Get/Create |
| ✅ | `internal/service/order_mock_test.go` | PlaceOrder 成功 / 库存不足；Transition |

**验收：** `go test ./internal/service/ -run Mock -count=1` 不连 MySQL 也能过。

### G3 · Redis 缓存 ✅

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ✅ | 依赖 | `github.com/redis/go-redis/v9` |
| ✅ | `internal/cache/redis.go` | `NewRedis` / `Get` / `Set` / `Del` / `DelByPrefix` / `Ping` |
| ✅ | `cmd/server/main.go` | `cache.NewRedis`，注入各 Service `Cache` |
| ✅ | `CatalogService.GetProduct` | `product:{id}` 读缓存；未命中查库回填 |
| ✅ | `CatalogService.CreateProduct` / `ListProducts` | 写后清列表；列表短 TTL |
| ✅ | `CartService.Add` / `List` | 写后 Del；列表可读缓存 |
| ✅ | `OrderService.PlaceOrder` | 成功后失效 product / cart / 列表页 |
| ✅ | `AuthService.Login` | 失败次数限流（`login:fail:{email}`） |

**验收：** Compose 起 redis 后，连续两次 `GET /products/:id` 第二次可命中缓存；下单后详情/列表缓存失效；连续错误密码可触发限流。

---

## 调用链

```text
NewRouter（RequestID → AccessLog）
  → Handler
    → Service（接口 Store + 可选 Cache；日志带 request_id）
      → Repository 实现 / Redis
```
