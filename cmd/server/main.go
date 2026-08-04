package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e_commerce_platform/internal/applog"
	"e_commerce_platform/internal/cache"
	"e_commerce_platform/internal/config"
	"e_commerce_platform/internal/handler"
	"e_commerce_platform/internal/repository"
	"e_commerce_platform/internal/service"

	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 入口：读配置 → 组装依赖 → 起 HTTP。
// 按 README 阶段 A→F 自己补：MySQL/GORM、Redis、优雅停机等。
func main() {
	cfg := config.Load()

	// TODO(G): applog.Setup(cfg.AppEnv)
	applog.Setup(cfg.AppEnv)

	mysqlDsn := cfg.MySQLDSN
	if mysqlDsn == "" {
		log.Fatal(errors.New("mysqlDSN empty"))
	}
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// TODO(G): rdb, err := cache.NewRedis(cfg.RedisAddr)；失败则 log.Fatal 或降级 nil
	rdb, err := cache.NewRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}

	cartRepo := repository.CartRepo{DB: db}
	orderRepo := repository.OrderRepo{DB: db}
	productRepo := repository.ProductRepo{DB: db}
	userRepo := repository.UserRepo{DB: db}
	if err := productRepo.AutoMigrate(); err != nil {
		log.Fatal(err)
	}
	catalogSvc := &service.CatalogService{Products: &productRepo, Cache: rdb}
	cartSvc := &service.CartService{Carts: &cartRepo, Products: &productRepo, Cache: rdb}
	orderSvc := &service.OrderService{DB: db, Products: &productRepo, Carts: &cartRepo, Orders: &orderRepo, Cache: rdb}
	authSvc := &service.AuthService{Users: &userRepo, JWTSecret: cfg.JWTSecret, Cache: rdb}

	r := handler.NewRouter(handler.Deps{
		JWTSecret: cfg.JWTSecret,
		Catalog:   catalogSvc,
		Cart:      cartSvc,
		Order:     orderSvc,
		Auth:      authSvc,
	})

	log.Printf("listening on %s env=%s", cfg.HTTPAddr, cfg.AppEnv)
	// TODO(3.3): 换成 http.Server + ListenAndServe + SIGTERM 优雅停机
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r}
	go srv.ListenAndServe()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

}
