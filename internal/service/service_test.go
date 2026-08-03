package service

import (
	"context"
	"os"
	"testing"

	"e_commerce_platform/internal/model"
	"e_commerce_platform/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "trainer:Train2026Lib!@tcp(127.0.0.1:3306)/training_lib?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("mysql unavailable:", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Product{}, &model.CartItem{}, &model.Order{}, &model.OrderItem{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPlaceOrder_okAndStock(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	productRepo := &repository.ProductRepo{DB: db}
	orderRepo := &repository.OrderRepo{DB: db}
	cartRepo := &repository.CartRepo{DB: db}
	svc := &OrderService{DB: db, Products: productRepo, Orders: orderRepo, Carts: cartRepo}

	p := &model.Product{Name: "ut-place", Price: 100, Stock: 10}
	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	o, err := svc.PlaceOrder(ctx, 1, []OrderLine{{ProductID: p.ID, Quantity: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if o == nil || o.Total != 300 || o.Status != model.OrderPending {
		t.Fatalf("bad order: %+v", o)
	}

	got, err := productRepo.Get(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stock != 7 {
		t.Fatalf("stock want 7 got %d", got.Stock)
	}
}

func TestPlaceOrder_stockNotEnough(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	productRepo := &repository.ProductRepo{DB: db}
	svc := &OrderService{
		DB:       db,
		Products: productRepo,
		Orders:   &repository.OrderRepo{DB: db},
		Carts:    &repository.CartRepo{DB: db},
	}

	p := &model.Product{Name: "ut-short", Price: 100, Stock: 1}
	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.PlaceOrder(ctx, 1, []OrderLine{{ProductID: p.ID, Quantity: 5}}); err == nil {
		t.Fatal("expected stock not enough")
	}

	got, err := productRepo.Get(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stock != 1 {
		t.Fatalf("stock should rollback to 1, got %d", got.Stock)
	}
}
