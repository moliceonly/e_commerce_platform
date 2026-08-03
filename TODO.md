# 阶段实现对照表（Todo）

依赖方向：**handler → service → repository**。

**进度：阶段 A–F 全部完成（路线图 3.1–3.5 / Part 03）。**

---

## 阶段 A · 3.1 空壳可跑 ✅

| 状态 | 目录 / 符号 | 内容 |
|------|-------------|------|
| ✅ | `handler.Healthz` | 探活 |
| ✅ | `config.Load` | 读 env |
| ✅ | `response.OK` / `Fail` | 统一响应 |
| ✅ | `cmd/server/main.go` | MySQL + GORM；组装 repo→svc→`handler.Deps` |
| ✅ | `repository.ProductRepo.AutoMigrate` | 五张表 |
| ✅ | `handler.NewRouter` | `/api/v1` + 商品路由 |
| ✅ | `handler.ProductHandler.Create/Get/List` | → CatalogService |
| ✅ | `service.CatalogService.*` | → ProductRepo |
| ✅ | `repository.ProductRepo.Create/Get/List` | GORM CRUD |

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
| ✅ | `deployments/docker-compose.yml` | app + mysql + redis；宿主机 8081 / 3307 / 6380 |
| ✅ | `README.md` | 本地 / Compose / 测试说明 |

```bash
go test ./internal/auth/ ./internal/handler/ ./internal/service/ -count=1
sudo docker compose -f deployments/docker-compose.yml up --build
curl -i http://127.0.0.1:8081/healthz
```

---

## 调用链

```text
NewRouter
  → Handler
    → Service
      → Repository
```
