package service

import (
	"context"
	"fmt"
	"time"

	"e_commerce_platform/internal/applog"
	"e_commerce_platform/internal/auth"
	"e_commerce_platform/internal/cache"
	"e_commerce_platform/internal/model"
	"e_commerce_platform/internal/repository"

	"gorm.io/gorm"
)

// CatalogService 商品。
type CatalogService struct {
	Products repository.ProductStore // 接口：便于 mock
	Cache    cache.Cache             // 可选；阶段 G 接入 Redis
}

func (s *CatalogService) CreateProduct(ctx context.Context, name string, price int64, stock int) (*model.Product, error) {
	product := model.Product{
		Name:  name,
		Price: price,
		Stock: stock,
	}
	if err := s.Products.Create(ctx, &product); err != nil {
		applog.FromContext(ctx).Error("create product failed", "err", err.Error())
		return nil, err
	}
	// TODO(G): 写库成功后删列表缓存 s.Cache.Del(ctx, cache.ProductsPageKey(...))
	applog.FromContext(ctx).Info("create product ok", "id", product.ID, "name", name)
	return &product, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, id uint) (*model.Product, error) {
	// TODO(G): 先查 Redis cache.ProductKey(id)；未命中再 Products.Get，回填 Set
	applog.FromContext(ctx).Info("get product", "id", id)
	return s.Products.Get(ctx, id)
}

func (s *CatalogService) ListProducts(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	// TODO(G): 可选短缓存 cache.ProductsPageKey(page, pageSize)
	offset := (page - 1) * pageSize
	return s.Products.List(ctx, offset, pageSize)
}

// AuthService 注册登录。
type AuthService struct {
	Users     repository.UserStore
	JWTSecret string
	Cache     cache.Cache // TODO(G): 登录失败计数 / token 黑名单（可选）
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	// TODO(3.2): auth.HashPassword + Users.Create
	hashPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Email:        email,
		PasswordHash: hashPassword,
		Role:         "user",
	}
	if _, err := s.Users.FindByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("email already used")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err := s.Users.Create(ctx, &user); err != nil {
		return nil, err
	}
	applog.FromContext(ctx).Info("register ok", "user_id", user.ID, "email", email)
	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, err error) {
	// TODO(3.2): FindByEmail + CheckPassword + SignToken
	user, err := s.Users.FindByEmail(ctx, email)
	if err != nil {
		applog.FromContext(ctx).Warn("login failed", "email", email)
		return "", fmt.Errorf("invalid email or password")
	}
	if !auth.CheckPassword(user.PasswordHash, password) {
		applog.FromContext(ctx).Warn("login failed", "email", email)
		return "", fmt.Errorf("invalid email or password")
	}
	applog.FromContext(ctx).Info("login ok", "user_id", user.ID)
	return auth.SignToken(s.JWTSecret, user.ID, user.Role, 24*time.Hour)
}

// CartService 加购。
type CartService struct {
	Carts    repository.CartStore
	Products repository.ProductStore
	Cache    cache.Cache // TODO(G): 加购后刷新/删除 cart:{userID}
}

func (s *CartService) Add(ctx context.Context, userID, productID uint, qty int) error {

	if _, err := s.Products.Get(ctx, productID); err != nil {
		return fmt.Errorf("product not exists")
	}

	if err := s.Carts.Upsert(ctx, userID, productID, qty); err != nil {
		return err
	}
	// TODO(G): s.Cache.Del(ctx, cache.CartKey(userID))
	applog.FromContext(ctx).Info("cart add ok", "user_id", userID, "product_id", productID, "qty", qty)
	return nil

}

// OrderService 下单与状态流转。
type OrderService struct {
	DB       *gorm.DB
	Products repository.ProductStore
	Orders   repository.OrderStore
	Carts    repository.CartStore
	Cache    cache.Cache // TODO(G): 下单成功后失效 product:{id} / cart:{userID}
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
		var productIDs = make([]uint, 0, len(items))
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
			productIDs = append(productIDs, item.ProductID)
		}

		order = &model.Order{
			UserID: userID,
			Status: model.OrderPending,
			Total:  total,
		}
		if err := s.Orders.CreateWithItems(ctx, tx, order, orderItems); err != nil {
			return err
		}
		return s.Carts.ClearByUser(ctx, tx, userID, productIDs)

	})

	if err != nil {
		applog.FromContext(ctx).Error("place order failed", "user_id", userID, "err", err.Error())
		return nil, err
	}
	// TODO(G): 事务成功后失效缓存
	// for _, id := range productIDs { _ = s.Cache.Del(ctx, cache.ProductKey(id)) }
	// _ = s.Cache.Del(ctx, cache.CartKey(userID))
	applog.FromContext(ctx).Info("place order ok", "order_id", order.ID, "user_id", userID)
	return order, nil

}

func (s *OrderService) Transition(ctx context.Context, userID, orderID uint, to model.OrderStatus) error {
	// TODO: 校验归属防越权；pending→paid→shipped→done
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		applog.FromContext(ctx).Warn("transition forbidden", "user_id", userID, "order_id", orderID)
		return fmt.Errorf("User forbidden")
	}

	from, ok := order.Status, false
	for _, nextSta := range allowed[from] {
		if nextSta == to {
			ok = true
		}
	}

	if !ok {
		applog.FromContext(ctx).Warn("invalid transition", "order_id", orderID, "from", from, "to", to)
		return fmt.Errorf("invalid transition %s -> %s", from, to)
	}

	if err := s.Orders.UpdateStatus(ctx, orderID, from, to); err != nil {
		return err
	}
	applog.FromContext(ctx).Info("transition ok", "order_id", orderID, "from", from, "to", to)
	return nil
}

// ListOrders 当前用户订单分页列表（可选 status 过滤）。
func (s *OrderService) ListOrders(ctx context.Context, userID uint, status string, page, pageSize int) ([]model.Order, error) {
	// TODO(3.3): offset := (page-1)*pageSize → Orders.ListByUser
	offset := (page - 1) * pageSize
	q, err := s.Orders.ListByUser(ctx, userID, status, offset, pageSize)
	if err != nil {
		applog.FromContext(ctx).Error("list orders failed", "user_id", userID, "err", err.Error())
		return nil, err
	}
	return q, nil
}
