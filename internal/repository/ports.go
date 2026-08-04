package repository

import (
	"context"

	"e_commerce_platform/internal/model"

	"gorm.io/gorm"
)

// 数据端口（阶段 G）：service 依赖这些接口，便于单测 mock。
// 具体实现仍是本包的 *ProductRepo / *OrderRepo / *CartRepo / *UserRepo。

type ProductStore interface {
	Create(ctx context.Context, p *model.Product) error
	Get(ctx context.Context, id uint) (*model.Product, error)
	List(ctx context.Context, offset, limit int) ([]model.Product, error)
	DeductStockTx(ctx context.Context, tx *gorm.DB, productID uint, qty int) error
}

type OrderStore interface {
	CreateWithItems(ctx context.Context, tx *gorm.DB, order *model.Order, items []model.OrderItem) error
	GetByID(ctx context.Context, id uint) (*model.Order, error)
	ListByUser(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Order, error)
	UpdateStatus(ctx context.Context, orderID uint, from, to model.OrderStatus) error
}

type CartStore interface {
	Upsert(ctx context.Context, userID, productID uint, qty int) error
	ListByUser(ctx context.Context, userID uint) ([]model.CartItem, error)
	ClearByUser(ctx context.Context, tx *gorm.DB, userID uint, productIDs []uint) error
}

type UserStore interface {
	Create(ctx context.Context, u *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}
