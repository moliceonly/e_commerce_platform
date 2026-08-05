package handler

import (
	"e_commerce_platform/internal/middleware"
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Deps 路由装配依赖（handler → service，禁止 handler 直连 DB）。
type Deps struct {
	JWTSecret string
	Auth      *service.AuthService
	Catalog   *service.CatalogService
	Cart      *service.CartService
	Order     *service.OrderService
	Upload    *service.UploadService
}

// NewRouter 注册路由分组。先保证 /healthz；其余自己挂。
func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.CORS([]string{"http://localhost:3000"}))
	// TODO(H4): r.Use(metrics.Middleware()) ; metrics.Register(r)
	// TODO(H4): observability.MountPprof(r, cfg.AppEnv != "prod")

	r.GET("/healthz", Healthz)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	prodH := &ProductHandler{Svc: d.Catalog}
	cartH := &CartHandler{Svc: d.Cart}
	orderH := &OrderHandler{Svc: d.Order}
	authH := &AuthHandler{Svc: d.Auth}
	uploadH := &UploadHandler{Svc: d.Upload}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", prodH.List)
		v1.GET("/products/:id", prodH.Get)
		v1.POST("/products", middleware.JWTAuth(d.JWTSecret), middleware.RequireRole("admin"), prodH.Create)
		v1.POST("/auth/register", authH.Register)
		v1.POST("/auth/login", authH.Login)
		v1.POST("/auth/refresh", authH.Refresh)
	}

	authz := v1.Group("")
	authz.Use(middleware.JWTAuth(d.JWTSecret))
	{
		authz.POST("/cart/items", cartH.Add)
		authz.POST("/orders", orderH.Place)
		authz.POST("/orders/:id/transition", orderH.Transition)
		authz.GET("/orders", orderH.List)
		authz.POST("/me/avatar", uploadH.Avatar)
	}
	r.Static("/static", "data/uploads")

	return r
}
