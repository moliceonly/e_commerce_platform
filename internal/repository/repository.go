package repository

import (
	"context"
	"fmt"

	"e_commerce_platform/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductRepo 商品。
type ProductRepo struct{ DB *gorm.DB }

func (r *ProductRepo) AutoMigrate() error {
	// TODO(3.1): AutoMigrate User / Product / CartItem / Order / OrderItem
	return r.DB.AutoMigrate(&model.User{}, &model.Product{}, &model.CartItem{}, &model.Order{}, &model.OrderItem{})
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	return r.DB.WithContext(ctx).Create(p).Error
}

func (r *ProductRepo) Get(ctx context.Context, id uint) (*model.Product, error) {
	var product model.Product
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepo) List(ctx context.Context, offset, limit int) ([]model.Product, error) {
	var productList []model.Product
	err := r.DB.WithContext(ctx).Offset(offset).Limit(limit).Find(&productList).Error
	return productList, err
}

// DeductStockTx 事务内扣库存（阶段 E 防超卖）。
func (r *ProductRepo) DeductStockTx(ctx context.Context, tx *gorm.DB, productID uint, qty int) error {

	var product model.Product

	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&product, productID).Error; err != nil {
		return err
	}

	if product.Stock < qty {
		return fmt.Errorf("stock not enough")
	}

	upd := tx.WithContext(ctx).Model(&product).Update("stock", product.Stock-qty)

	return upd.Error

}

// OrderRepo 订单。
type OrderRepo struct{ DB *gorm.DB }

func (r *OrderRepo) CreateWithItems(ctx context.Context, tx *gorm.DB, order *model.Order, items []model.OrderItem) error {
	order.Items = items
	return tx.WithContext(ctx).Create(order).Error
}

func (r *OrderRepo) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&order).Error
	return &order, err
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Order, error) {

	var orders []model.Order

	q := r.DB.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}

	err := q.Offset(offset).Limit(limit).Find(&orders).Error
	return orders, err

}

func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID uint, from, to model.OrderStatus) error {

	if from == "" || to == "" {
		return fmt.Errorf("Invalid status")
	}

	upd := r.DB.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status = ?", orderID, from).Update("status", to)
	if upd.Error != nil {
		return upd.Error
	}

	if upd.RowsAffected == 0 {
		return fmt.Errorf("order not found or status not %s", from)
	}

	return nil
}

// UserRepo 用户。
type UserRepo struct{ DB *gorm.DB }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.DB.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	q := r.DB.WithContext(ctx).Find(&user, email)
	return &user, q.Error
}

// CartRepo 购物车。
type CartRepo struct{ DB *gorm.DB }

func (r *CartRepo) Upsert(ctx context.Context, userID, productID uint, qty int) error {

	var cartItem model.CartItem
	q := r.DB.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem)

	if q.Error == gorm.ErrRecordNotFound {
		return r.DB.WithContext(ctx).Create(&model.CartItem{
			UserID:    userID,
			ProductID: productID,
			Quantity:  qty,
		}).Error
	} else if q.Error != nil {
		return q.Error
	}

	return r.DB.WithContext(ctx).Model(&cartItem).Update("quantity", cartItem.Quantity+qty).Error
}

func (r *CartRepo) ListByUser(ctx context.Context, userID uint) ([]model.CartItem, error) {

	var cartItem []model.CartItem

	q := r.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&cartItem)
	if q.Error != nil {
		return nil, q.Error
	}
	return cartItem, nil
}
