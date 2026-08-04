package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	if s.Cache != nil {
		if b, err := json.Marshal(&product); err == nil {
			if redisErr := s.Cache.Set(ctx, cache.ProductKey(product.ID), string(b), 30*time.Second); redisErr != nil {
				applog.FromContext(ctx).Warn("rediscreate product failed", "product_id", product.ID, "redis_err", redisErr)
			}
		}
		// 新建商品后所有分页列表都可能过期
		if redisErr := s.Cache.DelByPrefix(ctx, cache.ProductsPagePrefix); redisErr != nil {
			applog.FromContext(ctx).Warn("redis clear product list failed", "redis_err", redisErr)
		}
	}
	applog.FromContext(ctx).Info("create product ok", "id", product.ID, "name", name)
	return &product, nil

}

func (s *CatalogService) GetProduct(ctx context.Context, id uint) (*model.Product, error) {
	// TODO(G): 先查 Redis cache.ProductKey(id)；未命中再 Products.Get，回填 Set
	applog.FromContext(ctx).Info("get product", "id", id)
	redisKey := cache.ProductKey(id)

	if s.Cache != nil {
		if raw, err := s.Cache.Get(ctx, redisKey); err == nil && raw != "" {
			var p model.Product
			if json.Unmarshal([]byte(raw), &p) == nil {
				return &p, nil
			}
		}
	}

	product, err := s.Products.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.Cache != nil {
		if b, err := json.Marshal(product); err == nil {
			if redisErr := s.Cache.Set(ctx, redisKey, string(b), 30*time.Second); redisErr != nil {
				applog.FromContext(ctx).Warn("redisset product failed", "product_id", product.ID, "redis_err", redisErr)
			}
		}
	}
	return product, nil
}

func (s *CatalogService) ListProducts(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	// TODO(G): 可选短缓存 cache.ProductsPageKey(page, pageSize)
	offset := (page - 1) * pageSize
	redisKey := cache.ProductsPageKey(page, pageSize)

	if s.Cache != nil {
		if raw, err := s.Cache.Get(ctx, redisKey); err == nil && raw != "" {
			var l []model.Product
			if json.Unmarshal([]byte(raw), &l) == nil {
				return l, nil
			}
		}
	}

	//未命中查库
	listProducts, err := s.Products.List(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	//查中则写缓存
	if s.Cache != nil {
		if b, err := json.Marshal(listProducts); err == nil {
			if redisErr := s.Cache.Set(ctx, redisKey, string(b), 30*time.Second); redisErr != nil {
				applog.FromContext(ctx).Warn("redisset product list failed", "product_list_key", redisKey, "redis_err", redisErr)
			}
		}
	}

	return listProducts, nil
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
	failKey := cache.FailKey(email)
	if s.Cache != nil {
		if raw, err := s.Cache.Get(ctx, failKey); err == nil && raw != "" {
			if redisCount, err := strconv.Atoi(raw); err == nil && redisCount >= cache.MaxFail {
				return "", fmt.Errorf("too many login tries, try 60 seconds later")
			}
		}
	}

	user, err := s.Users.FindByEmail(ctx, email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, password) {
		if s.Cache != nil {
			count := 1
			if raw, _ := s.Cache.Get(ctx, failKey); raw != "" {
				if n, parseErr := strconv.Atoi(raw); parseErr == nil {
					count = n + 1
				}
			}
			if redisErr := s.Cache.Set(ctx, failKey, strconv.Itoa(count), 60*time.Second); redisErr != nil {
				applog.FromContext(ctx).Warn("redis count login failed", "fail_key", failKey, "redis_err", redisErr)
			}
		}
		applog.FromContext(ctx).Warn("login failed", "email", email)
		return "", fmt.Errorf("invalid email or password")
	}

	if s.Cache != nil {
		if redisErr := s.Cache.Del(ctx, failKey); redisErr != nil {
			applog.FromContext(ctx).Warn("redis delete tries failed", "fail_key", failKey, "redis_err", redisErr)
		}
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

	redisKey := cache.CartKey(userID)
	if s.Cache != nil {
		if redisErr := s.Cache.Del(ctx, redisKey); redisErr != nil {
			applog.FromContext(ctx).Warn("redis delete cart item failed", "user_id", userID, "product_id", productID, "quantity", qty, "redis_err", redisErr)
		}
	}

	applog.FromContext(ctx).Info("cart add ok", "user_id", userID, "product_id", productID, "qty", qty)
	return nil

}

func (s *CartService) List(ctx context.Context, userID uint) ([]model.CartItem, error) {
	// 先查询缓存
	redisKey := cache.CartKey(userID)
	if s.Cache != nil {
		if raw, err := s.Cache.Get(ctx, redisKey); err == nil && raw != "" {
			var cartItems []model.CartItem
			if err := json.Unmarshal([]byte(raw), &cartItems); err == nil {
				return cartItems, nil
			}
		}
	}

	// 未命中查库
	cartItems, err := s.Carts.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 若存在库则写入缓存
	if s.Cache != nil {
		if b, err := json.Marshal(cartItems); b != nil && err == nil {
			if redisErr := s.Cache.Set(ctx, redisKey, string(b), 30*time.Second); redisErr != nil {
				applog.FromContext(ctx).Warn("redisset cart item fail", "cart_items_key", redisKey, "redis_err", redisErr)
			}
		}
	}

	return cartItems, nil
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
	var productIDs = make([]uint, 0, len(items))

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

		_ = s.Carts.ClearByUser(ctx, tx, userID, productIDs)

		return nil

	})

	if err != nil {
		applog.FromContext(ctx).Error("place order failed", "user_id", userID, "err", err.Error())
		return nil, err
	}

	// TODO(G): 事务成功后失效缓存
	if s.Cache != nil {
		for _, id := range productIDs {
			if redisErr := s.Cache.Del(ctx, cache.ProductKey(id)); redisErr != nil {
				applog.FromContext(ctx).Warn("redisdelete product id failed", "product_id", id, "redis_err", redisErr)
			}
		}
		if redisErr := s.Cache.Del(ctx, cache.CartKey(userID)); redisErr != nil {
			applog.FromContext(ctx).Warn("redisdelete cart id failed", "user_id", userID, "redis_err", redisErr)
		}
		if redisErr := s.Cache.DelByPrefix(ctx, cache.ProductsPagePrefix); redisErr != nil {
			applog.FromContext(ctx).Warn("redis clear product list fail", "redis_err", redisErr)
		}
	}

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
		applog.FromContext(ctx).Warn("transition forbidden", "user_id", userID, "order_belong_user_id", order.UserID)
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
