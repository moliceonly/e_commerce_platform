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
	product := model.Product{
		Name:  name,
		Price: price,
		Stock: stock,
	}
	return &product, s.Products.Create(ctx, &product)
}

func (s *CatalogService) GetProduct(ctx context.Context, id uint) (*model.Product, error) {
	return s.Products.Get(ctx, id)
}

func (s *CatalogService) ListProducts(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	offset := (page - 1) * pageSize
	return s.Products.List(ctx, offset, pageSize)
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

	if _, err := s.Products.Get(ctx, productID); err != nil {
		return fmt.Errorf("product not exists")
	}

	return s.Carts.Upsert(ctx, userID, productID, qty)

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

var allowed = map[model.OrderStatus][]model.OrderStatus{
	model.OrderPending: {model.OrderPaid, model.OrderCancelled},
	model.OrderPaid:    {model.OrderShipped},
	model.OrderShipped: {model.OrderDone},
}

// PlaceOrder 事务下单 + 扣库存（防超卖核心编排）。
func (s *OrderService) PlaceOrder(ctx context.Context, userID uint, items []OrderLine) (*model.Order, error) {
	// TODO(阶段 B/E):
	//  1. 开事务
	//  2. 每行 Products.DeductStockTx
	//  3. Orders.CreateWithItems
	//  4. 清空购物车（如有）
	var order *model.Order

	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var total int64
		var orderItems []model.OrderItem

		for _, item := range items {

			product, err := s.Products.Get(ctx, item.ProductID)
			if err != nil {
				return err
			}

			if err := s.Products.DeductStockTx(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return err
			}
			total += product.Price * int64(item.Quantity)
			orderItems = append(orderItems, model.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})
		}

		order = &model.Order{
			UserID: userID,
			Status: model.OrderPending,
			Total:  total,
		}

		return s.Orders.CreateWithItems(ctx, tx, order, orderItems)

	})

	if err != nil {
		return nil, err
	}
	return order, nil

}

func (s *OrderService) Transition(ctx context.Context, userID, orderID uint, to model.OrderStatus) error {
	// TODO: 校验归属防越权；pending→paid→shipped→done
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return fmt.Errorf("User forbidden")
	}

	from, ok := order.Status, false
	for _, nextSta := range allowed[from] {
		if nextSta == to {
			ok = true
		}
	}
	if !ok {
		return fmt.Errorf("invalid transition %s -> %s", from, to)
	}
	return s.Orders.UpdateStatus(ctx, orderID, from, to)
}
