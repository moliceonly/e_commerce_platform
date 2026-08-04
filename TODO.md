# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。

**进度：阶段 A–F 已完成；当前加练阶段 G（Redis / 接口 mock / 结构化日志）。**

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

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `CartHandler` / `CartService` / `CartRepo` | 加购 Upsert |
| ✅ | `OrderHandler.Place` / `OrderService.PlaceOrder` | 事务下单 + 扣库存 |
| ✅ | `ProductRepo.DeductStockTx` | `FOR UPDATE` |
| ✅ | `OrderRepo.CreateWithItems` | 同事务写订单行 |
| ✅ | `Transition` + 状态机 | pending→paid→shipped→done |

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

## 阶段 G · 补强（Redis + 接口 mock + 结构化日志）← 当前

骨架已生成，按表在对应文件填 TODO(G)。建议顺序：**G1 日志 → G2 接口 mock → G3 Redis**。

### G1 · 结构化日志 + 全链路 request_id

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ☐ | `internal/applog/applog.go` | 实现 `Setup` / `WithRequestID` / `RequestIDFrom` / `FromContext`（建议 `log/slog`） |
| ☐ | `internal/middleware/middleware.go` → `RequestID` | `applog.WithRequestID` 写入 `c.Request.Context()` |
| ☐ | `internal/middleware/middleware.go` → `AccessLog` | 请求结束后打 method/path/status/latency + request_id |
| ☐ | `internal/handler/router.go` | 确认 `AccessLog` 在 `RequestID` 之后；可去掉 `gin.Logger` |
| ☐ | `cmd/server/main.go` | 启动时调用 `applog.Setup(cfg.AppEnv)` |
| ☐ | `internal/service/service.go`（关键路径） | `GetProduct` / `PlaceOrder` 等业务函数各 1～2 条 applog；repo 层不打 |

**验收：** `curl -i /healthz` 响应有 `X-Request-ID`；容器/终端日志每行带相同 `request_id`。

### G2 · repository 接口 + service mock 单测

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ☐ | `internal/repository/ports.go` | 已有 `ProductStore` / `OrderStore` / `CartStore` / `UserStore`（一般不用改） |
| ☐ | `internal/service/service.go` | 字段已改为接口类型；确认编译通过 |
| ☐ | `internal/service/catalog_mock_test.go` | 补全 `mockProductStore`；去掉 `t.Skip`；测 `GetProduct`/`CreateProduct` |
| ☐ | （可选）同目录再加 `order_mock_test.go` | mock `ProductStore`+`OrderStore`+`CartStore` 测库存不足等 |

**验收：** `go test ./internal/service/ -run Mock -count=1` 不连 MySQL 也能过。

### G3 · Redis 缓存（不止下单读库）

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ☐ | 依赖 | `go get github.com/redis/go-redis/v9` |
| ☐ | `internal/cache/redis.go` | 实现 `NewRedis` / `Get` / `Set` / `Del` / `Ping` |
| ☐ | `cmd/server/main.go` | `cache.NewRedis(cfg.RedisAddr)`，注入各 Service 的 `Cache` 字段 |
| ☐ | `CatalogService.GetProduct` | 读 `product:{id}` 缓存；未命中查库并回填 |
| ☐ | `CatalogService.CreateProduct` / `ListProducts` | 写后删列表缓存；列表可选短 TTL |
| ☐ | `CartService.Add` | 写后 `Del(cart:{userID})` |
| ☐ | `OrderService.PlaceOrder` | 成功后失效相关 `product:{id}` 与 `cart:{userID}` |
| ☐ | （可选）`AuthService.Login` | 失败次数限流或 token 黑名单 |

**验收：** Compose 起 redis 后，连续两次 `GET /products/:id`，第二次可在日志/逻辑上体现命中缓存；改库存/下单后缓存被删掉。

### 建议实现顺序

```text
G1 applog + RequestID 进 context + AccessLog
  → G2 mockProductStore 单测
    → G3 Redis 客户端 + 商品/购物车/下单失效
```

---

## 调用链

```text
NewRouter（RequestID → AccessLog）
  → Handler
    → Service（接口 Store + 可选 Cache；日志带 request_id）
      → Repository 实现 / Redis
```
