package handler

import (
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
)

// Deps 路由装配依赖（handler → service，禁止 handler 直连 DB）。
type Deps struct {
	JWTSecret string
	Auth      *service.AuthService
	Catalog   *service.CatalogService
	Cart      *service.CartService
	Order     *service.OrderService
}

// NewRouter 注册路由分组。先保证 /healthz；其余自己挂。
func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	// TODO(3.3): r.Use(middleware.RequestID())

	r.GET("/healthz", Healthz)
	prodH := &ProductHandler{Svc: d.Catalog}
	cartH := &CartHandler{Svc: d.Cart}
	orderH := &OrderHandler{Svc: d.Order}

	_ = d // 装配 Auth/Catalog/Cart/Order handler 后挂到下面分组

	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", prodH.List)
		v1.GET("/products/:id", prodH.Get)
		v1.POST("/products", prodH.Create)
		v1.POST("/cart/items", cartH.Add)
		v1.POST("/orders", orderH.Place)
		v1.POST("/orders/:id/transition", orderH.Transition)
	}
	// TODO(3.2): POST /auth/register, /auth/login
	// TODO(3.2): 需登录组 + middleware.JWTAuth(d.JWTSecret)
	//   POST /cart/items, POST /orders, GET /orders, POST /orders/:id/transition

	return r
}
