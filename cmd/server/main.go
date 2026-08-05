// @title           e_commerce_platform API
// @version         0.1.0
// @description     电商训练项目 API（阶段 H · OpenAPI / swag）
// @host            127.0.0.1:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description     填写 `Bearer <access_token>`
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
	"e_commerce_platform/internal/job"
	"e_commerce_platform/internal/repository"
	"e_commerce_platform/internal/service"

	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "e_commerce_platform/docs"
)

// 入口：读配置 → 组装依赖 → 起 HTTP。
// 按 README 阶段 A→F 自己补：MySQL/GORM、Redis、优雅停机等。
func main() {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	cfg, err := config.LoadYAML(env)
	if err != nil {
		log.Fatal(err)
	}

	applog.Setup(cfg.AppEnv)

	mysqlDsn := cfg.MySQLDSN
	if mysqlDsn == "" {
		log.Fatal(errors.New("mysqlDSN empty"))
	}
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

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
	uploadSvc := &service.UploadService{Dir: "data/uploads", BaseURL: "http://127.0.0.1:8080/static"}

	r := handler.NewRouter(handler.Deps{
		JWTSecret:   cfg.JWTSecret,
		Catalog:     catalogSvc,
		Cart:        cartSvc,
		Order:       orderSvc,
		Auth:        authSvc,
		Upload:      uploadSvc,
		EnablePprof: cfg.AppEnv != "prod",
	})

	log.Printf("listening on %s env=%s", cfg.HTTPAddr, cfg.AppEnv)

	jobs := &job.Runner{
		Timeout: 30 * time.Minute,
		Tick:    time.Minute,
		Orders:  &orderRepo,
		DB:      db,
	}
	jobs.Start(context.Background())
	defer jobs.Stop()

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
