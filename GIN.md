# Gin 用法速查（覆盖本项目）

风格对齐 `command.md`：**作用 | 入参 | 返回**。只列简易电商会用到的 API。

安装：`go get github.com/gin-gonic/gin`

---

## 1. 引擎与启动

### `gin.New()`
| | |
|--|--|
| **作用** | 创建空引擎（不含默认 Logger/Recovery） |
| **入参** | 无 |
| **返回** | `*gin.Engine` |

本项目用 `gin.New()` + 手动 `Use`，比 `gin.Default()` 更清晰。

### `gin.Default()`
| | |
|--|--|
| **作用** | `New` + 自带 `Logger`、`Recovery` |
| **入参** | 无 |
| **返回** | `*gin.Engine` |

### `(*Engine).Run(addr)`
| | |
|--|--|
| **作用** | 监听并阻塞服务（阶段 A 够用） |
| **入参** | `addr string`，如 `":8080"`；空则 `:8080` |
| **返回** | `error`（正常退出少见；失败才返回） |

阶段 D 建议改成标准库：

```go
srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r}
go srv.ListenAndServe()
// signal → srv.Shutdown(ctx)
```

### `gin.SetMode(mode)`
| | |
|--|--|
| **作用** | 运行模式（影响日志量） |
| **入参** | `gin.DebugMode` / `ReleaseMode` / `TestMode` |
| **返回** | 无 |

---

## 2. 中间件

### `(*Engine).Use(middleware...)`
| | |
|--|--|
| **作用** | 全局中间件，每个请求按顺序执行 |
| **入参** | 若干 `gin.HandlerFunc` |
| **返回** | `IRoutes`（可链式） |

本项目：`gin.Recovery()`、`gin.Logger()`、自写 `RequestID()`、分组上的 `JWTAuth`。

### `gin.Recovery()`
| | |
|--|--|
| **作用** | panic 转 500，避免进程崩 |
| **入参** | 无（工厂函数返回 `HandlerFunc`） |
| **返回** | `gin.HandlerFunc` |

### `gin.Logger()`
| | |
|--|--|
| **作用** | 打印方法、路径、状态、耗时 |
| **入参** | 无 |
| **返回** | `gin.HandlerFunc` |

### 自定义中间件形态

```go
func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 校验失败：response.Fail(...); c.Abort(); return
        // 成功：c.Set("userID", uid); c.Next()
    }
}
```

| 方法 | 作用 | 入参 | 返回 |
|------|------|------|------|
| `c.Next()` | 调用后续 handler/中间件 | 无 | 无 |
| `c.Abort()` | 中止后续处理（须先写响应） | 无 | 无 |
| `c.AbortWithStatus(code)` | 中止并写 HTTP 状态 | `code int` | 无 |

---

## 3. 路由与分组

### `(*Engine).GET/POST/PUT/PATCH/DELETE(path, handlers...)`
| | |
|--|--|
| **作用** | 注册 HTTP 方法 + 路径 |
| **入参** | `path string`；`handlers ...gin.HandlerFunc` |
| **返回** | `IRoutes` |

本项目映射：

| 方法 | 路径 | Handler |
|------|------|---------|
| GET | `/healthz` | `Healthz` |
| POST | `/api/v1/auth/register` | `AuthHandler.Register` |
| POST | `/api/v1/auth/login` | `AuthHandler.Login` |
| GET | `/api/v1/products` | `ProductHandler.List` |
| GET | `/api/v1/products/:id` | `ProductHandler.Get` |
| POST | `/api/v1/products` | `ProductHandler.Create` |
| POST | `/api/v1/cart/items` | `CartHandler.Add`（需 JWT） |
| POST | `/api/v1/orders` | `OrderHandler.Place`（需 JWT） |
| GET | `/api/v1/orders` | `OrderHandler.List`（需 JWT） |
| POST | `/api/v1/orders/:id/transition` | `OrderHandler.Transition`（需 JWT） |

### `(*Engine).Group(relativePath, handlers...)`
| | |
|--|--|
| **作用** | 路径前缀分组；可挂组级中间件 |
| **入参** | `relativePath`（如 `"/api/v1"`）；可选中间件 |
| **返回** | `*gin.RouterGroup` |

```go
v1 := r.Group("/api/v1")
authz := v1.Group("")
authz.Use(middleware.JWTAuth(secret))
authz.POST("/orders", orderH.Place)
```

---

## 4. `*gin.Context` —— Handler 核心

Handler 签名：`func(c *gin.Context)`。

### 读请求

| API | 作用 | 入参 | 返回 |
|-----|------|------|------|
| `c.Param(key)` | 路径参数 `:id` | `key string` | `string` |
| `c.Query(key)` | query `?page=1` | `key string` | `string`（无则 `""`） |
| `c.DefaultQuery(key, def)` | query，缺省用默认 | `key, def string` | `string` |
| `c.GetHeader(key)` | 请求头 | `key string` | `string` |
| `c.ShouldBindJSON(obj)` | 解析 JSON body 到结构体 | `obj any`（指针） | `error`；失败勿再写成功响应 |
| `c.ShouldBindQuery(obj)` | 解析 query 到结构体 | `obj any` | `error` |
| `c.ShouldBindUri(obj)` | 解析 path 到结构体 | `obj any` | `error` |
| `c.Request` | 标准 `*http.Request` | — | 用于 `c.Request.Context()` 传给 service |
| `c.ClientIP()` | 客户端 IP | 无 | `string` |

**本项目 bind 示例：**

```go
var req struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}
if err := c.ShouldBindJSON(&req); err != nil {
    response.Fail(c, 400, 40001, err.Error())
    return
}
```

常用 `binding` 标签：`required`、`email`、`min=N`、`max=N`、`gte=0`、`oneof=pending paid`。

### 写响应

| API | 作用 | 入参 | 返回 |
|-----|------|------|------|
| `c.JSON(code, obj)` | JSON 响应 | HTTP 状态码；任意可序列化对象 | 无 |
| `c.String(code, format, args...)` | 纯文本 | 状态码 + 格式化串 | 无 |
| `c.Status(code)` | 只写状态码 | `code int` | 无 |
| `c.Header(k, v)` | 设置响应头 | key, value | 无 |
| `c.File(filepath)` | 回文件（可选上传/下载） | 路径 | 无 |
| `c.Redirect(code, loc)` | 重定向 | 状态码、URL | 无 |

本项目统一走 `response.OK` / `response.Fail`（内部即 `c.JSON`）。

### Context 存取（鉴权传 userID）

| API | 作用 | 入参 | 返回 |
|-----|------|------|------|
| `c.Set(key, val)` | 中间件写入 | `key string`, `val any` | 无 |
| `c.Get(key)` | 读取 | `key string` | `(any, bool)` |
| `c.GetUint(key)` | 读 uint | `key string` | `uint`（无则 0） |
| `c.GetString(key)` | 读 string | `key string` | `string` |

```go
uid := c.GetUint(middleware.CtxUserID) // JWTAuth 里 Set 过
```

### 其它

| API | 作用 | 入参 | 返回 |
|-----|------|------|------|
| `c.Writer` | `http.ResponseWriter` | — | 设头、流式写 |
| `c.Copy()` | 拷贝 Context（给 goroutine 用） | 无 | `*Context`；**禁止**在新 goroutine 直接用原 `c` |

---

## 5. Handler 写法模板（本项目）

```go
func (h *OrderHandler) Place(c *gin.Context) {
    uid := c.GetUint(middleware.CtxUserID)

    var req struct {
        Items []struct {
            ProductID uint `json:"product_id" binding:"required"`
            Quantity  int  `json:"quantity" binding:"required,min=1"`
        } `json:"items" binding:"required,min=1"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, 40001, err.Error())
        return
    }

    // 转成 service 入参…
    o, err := h.Svc.PlaceOrder(c.Request.Context(), uid, items)
    if err != nil {
        response.Fail(c, http.StatusConflict, 40901, err.Error())
        return
    }
    response.OK(c, o)
}
```

要点：
1. **只在 handler 碰 Gin**；service 收 `context.Context` + 普通类型。
2. bind 失败立刻 `return`。
3. 业务错误用 `response.Fail`；成功用 `response.OK`。

---

## 6. 单测里怎么挂 Gin

```go
r := gin.New()
r.POST("/api/v1/orders", orderH.Place)
w := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", body)
r.ServeHTTP(w, req)
// 断言 w.Code、w.Body
```

| API | 作用 | 入参 | 返回 |
|-----|------|------|------|
| `httptest.NewRequest(...)` | 构造请求 | method, url, body | `*http.Request, error` |
| `httptest.NewRecorder()` | 假响应 | 无 | `*ResponseRecorder` |
| `r.ServeHTTP(w, req)` | 走完整中间件+路由 | `ResponseWriter`, `*Request` | 无 |

---

## 7. 与本项目文件对应

| 你写的代码 | 主要用到的 Gin API |
|------------|-------------------|
| `handler/router.go` | `New`, `Use`, `Group`, `GET`/`POST` |
| `handler/handler.go` | `ShouldBindJSON`, `Param`, `Query`/`DefaultQuery`, `GetUint`, `Request.Context` |
| `middleware/*.go` | `HandlerFunc`, `GetHeader`, `Set`, `Next`, `Abort` |
| `response/*.go` | `c.JSON` |
| `cmd/server/main.go` | `Run` 或 `http.Server{Handler: engine}` |

官方文档：https://gin-gonic.com/docs/
