# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。

**进度：阶段 A–G 已完成；当前加练阶段 H（Part 03 知识点补齐）。**

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

## 阶段 H · Part 03 知识点补齐 ← 当前

对照 [`go-web-roadmap.html`](../go-web-roadmap.html) Part 03 知识要点，在 **不推倒 A–G** 的前提下补齐缺口。  
骨架已生成（函数多返回 `TODO(H)` / 501 / 空实现），按表自己填。建议顺序：**H1 → H2 → H3 → H4 → H5**。

### H1 · 3.1 配置与错误码 ✅

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ✅ | `configs/*.yaml` | local / staging / prod（example + 本地 gitignore 副本） |
| ✅ | `internal/config/viper.go` → `LoadYAML` | viper 读 yaml，env 覆盖 |
| ✅ | `cmd/server/main.go` | 按 `APP_ENV` 调用 `LoadYAML` |
| ✅ | `internal/errcode/errcode.go` | 统一业务错误码；handler/middleware 已改用 |

**验收：** `APP_ENV=local|staging` + `JWT_SECRET=...` 能覆盖文件值。

### H2 · 3.2 安全增强（RBAC / CORS / Refresh）✅

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ✅ | `internal/middleware/cors.go` | Origin 白名单 + OPTIONS 204 |
| ✅ | `internal/middleware/rbac.go` → `RequireRole` | 校验 role；403 |
| ✅ | `middleware.JWTAuth` | `c.Set(CtxRole, claims.Role)` |
| ✅ | `internal/auth` + `AuthService` | access + refresh；`/auth/refresh` 换发 |
| ✅ | `handler` / `router` | refresh 路由；`POST /products` 需 admin |
| ✅ | README 冒烟 / 安全说明 | 见 README「冒烟」与「安全注意」 |

**验收：** user 建商品 403；提权 admin 后成功；refresh 可换新 access（命令见 README）。

### H3 · 3.3 上传 / 异步 / OpenAPI

| 状态 | 去哪里实现　　　　　　　　　　　　　　　　　　　　 | 要做什么　　　　　　　　　　　　　　　　　　　　　　　　　　　 |
| ------| ----------------------------------------------------| ----------------------------------------------------------------|
| ✅　　| `internal/service` UploadService + `handler` Avatar | multipart 存 `data/uploads/`，返回可访问 URL　　　　　　　　　 |
| ✅　　| `router`　　　　　　　　　　　　　　　　　　　　　 | `POST /api/v1/me/avatar`（需登录）；`GET /static/*`　　　　　　 |
| ✅　　| `internal/job/job.go`　　　　　　　　　　　　　　　| ticker 扫超时 pending → `UpdateStatus` 取消；每轮打 job tick　 |
| ✅　　| `cmd/server/main.go`　　　　　　　　　　　　　　　 | `job.Start`；`defer job.Stop`　　　　　　　　　　　　　　　　　 |
| ✅　　| swag + `/swagger/*`　　　　　　　　　　　　　　　　| 注解生成 `docs/`；`GET /swagger/index.html`　　　　　　　　　　|

**验收：** 上传头像后能用返回 URL 访问文件；进程日志出现 job tick；Swagger/OpenAPI 能打开或被 raw 查看。

Compose：`data/uploads` 已挂载到容器 `/app/data/uploads`（重启不丢文件）。

### H4 · 3.4 Lint / Metrics / pprof

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ☐ | `.golangci.yml` | 启用 govet / errcheck / staticcheck 等；本地 `golangci-lint run` 清告警 |
| ☐ | `Makefile` → `lint` | 封装 lint 命令 |
| ☐ | `internal/metrics/metrics.go` | `go get` prometheus client；请求计数/延迟直方图；`GET /metrics` |
| ☐ | `internal/observability/pprof.go` | 非 prod 挂载 `/debug/pprof/*` |
| ☐ | （可选）集成测试 | testcontainers 或 compose profile 已有则补 1 条关键路径 |

**验收：** `make lint` 通过；`/metrics` 有数据；本机 `go tool pprof` 能连上。

### H5 · 3.5 CI + Nginx

| 状态 | 去哪里实现 | 要做什么 |
|------|------------|----------|
| ☐ | `.github/workflows/ci.yml` | test → lint → docker build（可不 push） |
| ☐ | `deployments/nginx.conf.example` | 反代 `app:8080`；转发 `X-Request-ID` |
| ☐ | `deployments/docker-compose.yml` | （可选）增加 `nginx` 服务 profile |
| ☐ | `docs/runbook.md` | 排障：日志 → metrics → 慢 SQL / Redis → pprof |

**验收：** PR/push 能跑 CI 绿；浏览器/ curl 经 Nginx 访问 healthz 成功。

### 建议实现顺序

```text
H1 Viper 多环境配置 + errcode
  → H2 CORS + CtxRole + RequireRole + Refresh
    → H3 头像上传 + job + OpenAPI
      → H4 lint + metrics + pprof
        → H5 CI + Nginx + runbook
```

---

## 调用链

```text
NewRouter（RequestID → AccessLog → CORS）
  → Handler（含 upload / refresh / swagger）
    → Service（接口 Store + Cache；日志带 request_id）
      → Repository / Redis
  (+ job 后台；/metrics /debug/pprof)
```
