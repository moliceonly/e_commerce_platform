# e_commerce_platform

对齐 `go-web-roadmap.html` **Part 03**：简易电商。**空骨架，业务你自己实现。**

## 你做什么

按阶段把各层 `TODO` 填满；依赖方向：**handler → service → repository**（handler 禁止直连 DB）。

| 阶段 | 目标 |
|------|------|
| A · 3.1 | `go run` + `/healthz` + env 配置 + `/api/v1` 分组 |
| B | 模型 Migrate、商品、加购、事务下单扣库存、订单状态流转 |
| C · 3.2 | 注册登录 JWT、鉴权中间件、越权校验、binding |
| D · 3.3 | request_id、分页、优雅停机 |
| E · 3.5 | `scripts/oversell_demo.sh` 复现超卖 → 修 `DeductStockTx` |
| F | 单测、Dockerfile / compose |

## 目录

```text
cmd/server/           # 入口（目前只挂 /healthz）
internal/
  config/             # env 配置（可改 viper）
  model/              # 表结构已给字段约定，可改
  repository/         # 全部 TODO
  service/            # 全部 TODO
  handler/            # 除 Healthz 外全部 TODO；路由表在 router.go 注释里
  middleware/         # RequestID / JWTAuth TODO
  auth/               # bcrypt / JWT TODO
  response/           # 统一 {code,message,data}（可直接用）
scripts/              # 超卖脚本骨架
deployments/          # Docker 骨架
```

## 启动（阶段 A）

```bash
cd train_hub/e_commerce_platform
export GOPROXY=https://goproxy.cn,direct   # 或你能用的代理
go mod tidy
go run ./cmd/server
# curl http://127.0.0.1:8080/healthz
```

MySQL 可复用 Part02：`trainer` / `Train2026Lib!` / `training_lib`。DSN 建议用 `mysql.Config.FormatDSN()`。

## 订单状态约定

```text
pending → paid → shipped → done
       ↘ cancelled
```

## 超卖修复方向

事务内 `SELECT ... FOR UPDATE`，或 `UPDATE stock=stock-? WHERE id=? AND stock>=?` 检查 `RowsAffected`。
