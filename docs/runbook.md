# 排障手册骨架（阶段 H · 3.5）

## 症状 → 排查顺序

1. **看日志**：`request_id` / AccessLog / service Warn|Error
2. **看指标**：`GET /metrics`（实现 H4 后）QPS、延迟、5xx
3. **看依赖**：MySQL 连接、慢 SQL；Redis `PING`、关键 key 是否存在
4. **看运行时**：非 prod 下 `/debug/pprof`（heap / goroutine）

## 常见模拟故障

| 现象 | 可能原因 | 动手方向 |
|------|----------|----------|
| 超卖 | 无行锁 / 事务外扣库存 | `DeductStockTx` + 压测脚本 |
| 读到旧库存 | 缓存未失效 | PlaceOrder 后 Del product/list |
| 登录总失败 | 限流 key | `login:fail:{email}` TTL |
| 接口全 401 | JWT secret 不一致 | compose 与本地 env |

## TODO(H5)

- [ ] 补一条你真实排过的问题与结论
- [ ] 写清 compose 日志查看命令：`docker compose logs -f app`
