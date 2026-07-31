package service

import (
	"context"
	"fmt"

	"e_commerce_platform/internal/model"
	"e_commerce_platform/internal/repository"

	"gorm.io/gorm"
)

// CatalogService 商品。
type CatalogService struct {
	Products *repository.ProductRepo
}

func (s *CatalogService) CreateProduct(ctx context.Context, name string, price int64, stock int) (*model.Product, error) {
	_ = ctx
	_ = name
	_ = price
	_ = stock
	return nil, fmt.Errorf("TODO: CreateProduct")
}

func (s *CatalogService) GetProduct(ctx context.Context, id uint) (*model.Product, error) {
	_ = ctx
	_ = id
	return nil, fmt.Errorf("TODO: GetProduct")
}

func (s *CatalogService) ListProducts(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	_ = ctx
	_ = page
	_ = pageSize
	return nil, fmt.Errorf("TODO: ListProducts")
}

// AuthService 注册登录。
type AuthService struct {
	Users     *repository.UserRepo
	JWTSecret string
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	_ = ctx
	_ = email
	_ = password
	// TODO(3.2): auth.HashPassword + Users.Create
	return nil, fmt.Errorf("TODO: Register")
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, err error) {
	_ = ctx
	_ = email
	_ = password
	// TODO(3.2): FindByEmail + CheckPassword + SignToken
	return "", fmt.Errorf("TODO: Login")
}

// CartService 加购。
type CartService struct {
	Carts    *repository.CartRepo
	Products *repository.ProductRepo
}

func (s *CartService) Add(ctx context.Context, userID, productID uint, qty int) error {
	_ = ctx
	_ = userID
	_ = productID
	_ = qty
	return fmt.Errorf("TODO: Cart Add")
}

// OrderService 下单与状态流转。
type OrderService struct {
	DB       *gorm.DB
	Products *repository.ProductRepo
	Orders   *repository.OrderRepo
	Carts    *repository.CartRepo
}

// OrderLine 下单行入参。
type OrderLine struct {
	ProductID uint
	Quantity  int
}

// PlaceOrder 事务下单 + 扣库存（防超卖核心编排）。
func (s *OrderService) PlaceOrder(ctx context.Context, userID uint, items []OrderLine) (*model.Order, error) {
	_ = ctx
	_ = userID
	_ = items
	// TODO(阶段 B/E):
	//  1. 开事务
	//  2. 每行 Products.DeductStockTx
	//  3. Orders.CreateWithItems
	//  4. 清空购物车（如有）
	return nil, fmt.Errorf("TODO: PlaceOrder")
}

func (s *OrderService) Transition(ctx context.Context, userID, orderID uint, to model.OrderStatus) error {
	_ = ctx
	_ = userID
	_ = orderID
	_ = to
	// TODO: 校验归属防越权；pending→paid→shipped→done
	return fmt.Errorf("TODO: Transition")
}
