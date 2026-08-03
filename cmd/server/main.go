package main

import (
	"errors"
	"log"

	"e_commerce_platform/internal/config"
	"e_commerce_platform/internal/handler"
	"e_commerce_platform/internal/repository"
	"e_commerce_platform/internal/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 入口：读配置 → 组装依赖 → 起 HTTP。
// 按 README 阶段 A→F 自己补：MySQL/GORM、Redis、优雅停机等。
func main() {
	cfg := config.Load()

	// TODO(3.1): 用 cfg.MySQLDSN 打开 gorm，AutoMigrate
	mysqlDsn := cfg.MySQLDSN
	if mysqlDsn == "" {
		log.Fatal(errors.New("mysqlDSN empty"))
	}
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	cartRepo := repository.CartRepo{DB: db}
	orderRepo := repository.OrderRepo{DB: db}
	productRepo := repository.ProductRepo{DB: db}
	if err := productRepo.AutoMigrate(); err != nil {
		log.Fatal(err)
	}
	catalogSvc := &service.CatalogService{Products: &productRepo}
	cartSvc := &service.CartService{Carts: &cartRepo, Products: &productRepo}
	orderSvc := &service.OrderService{DB: db, Products: &productRepo, Carts: &cartRepo, Orders: &orderRepo}

	// TODO(3.1): 组装 repository → service → handler.Deps
	r := handler.NewRouter(handler.Deps{
		JWTSecret: cfg.JWTSecret,
		Catalog:   catalogSvc,
		Cart:      cartSvc,
		Order:     orderSvc,
		// Auth / Catalog / Cart / Order 自己注入
	})

	log.Printf("listening on %s env=%s", cfg.HTTPAddr, cfg.AppEnv)
	// TODO(3.3): 换成 http.Server + ListenAndServe + SIGTERM 优雅停机
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
