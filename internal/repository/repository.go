package repository

import (
	"context"
	"fmt"

	"e_commerce_platform/internal/model"

	"gorm.io/gorm"
)

// ProductRepo 商品。
type ProductRepo struct{ DB *gorm.DB }

func (r *ProductRepo) AutoMigrate() error {
	// TODO(3.1): AutoMigrate User / Product / CartItem / Order / OrderItem
	return fmt.Errorf("TODO: AutoMigrate")
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	_ = ctx
	_ = p
	return fmt.Errorf("TODO: Product.Create")
}

func (r *ProductRepo) Get(ctx context.Context, id uint) (*model.Product, error) {
	_ = ctx
	_ = id
	return nil, fmt.Errorf("TODO: Product.Get")
}

func (r *ProductRepo) List(ctx context.Context, offset, limit int) ([]model.Product, error) {
	_ = ctx
	_ = offset
	_ = limit
	return nil, fmt.Errorf("TODO: Product.List")
}

// DeductStockTx 事务内扣库存（阶段 E 防超卖）。
func (r *ProductRepo) DeductStockTx(ctx context.Context, tx *gorm.DB, productID uint, qty int) error {
	_ = ctx
	_ = tx
	_ = productID
	_ = qty
	// TODO: FOR UPDATE 或 UPDATE ... WHERE stock>=? 检查 RowsAffected
	return fmt.Errorf("TODO: DeductStockTx")
}

// OrderRepo 订单。
type OrderRepo struct{ DB *gorm.DB }

func (r *OrderRepo) CreateWithItems(ctx context.Context, tx *gorm.DB, order *model.Order, items []model.OrderItem) error {
	_ = ctx
	_ = tx
	_ = order
	_ = items
	return fmt.Errorf("TODO: CreateWithItems")
}

func (r *OrderRepo) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	_ = ctx
	_ = id
	return nil, fmt.Errorf("TODO: Order.GetByID")
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Order, error) {
	_ = ctx
	_ = userID
	_ = status
	_ = offset
	_ = limit
	return nil, fmt.Errorf("TODO: Order.ListByUser")
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID uint, from, to model.OrderStatus) error {
	_ = ctx
	_ = orderID
	_ = from
	_ = to
	return fmt.Errorf("TODO: UpdateStatus")
}

// UserRepo 用户。
type UserRepo struct{ DB *gorm.DB }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	_ = ctx
	_ = u
	return fmt.Errorf("TODO: User.Create")
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	_ = ctx
	_ = email
	return nil, fmt.Errorf("TODO: User.FindByEmail")
}

// CartRepo 购物车。
type CartRepo struct{ DB *gorm.DB }

func (r *CartRepo) Upsert(ctx context.Context, userID, productID uint, qty int) error {
	_ = ctx
	_ = userID
	_ = productID
	_ = qty
	return fmt.Errorf("TODO: Cart.Upsert")
}

func (r *CartRepo) ListByUser(ctx context.Context, userID uint) ([]model.CartItem, error) {
	_ = ctx
	_ = userID
	return nil, fmt.Errorf("TODO: Cart.ListByUser")
}
