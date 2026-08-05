package model

import (
	"time"

	"gorm.io/gorm"
)

// User 会员（3.2 注册登录）。
type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;size:128"`
	PasswordHash string `gorm:"size:255"`
	Role         string `gorm:"size:32;default:user"` // user / admin
}

// Product 商品（含库存；可选 Version 做乐观锁）。
type Product struct {
	gorm.Model
	Name    string `gorm:"size:128"`
	Price   int64  // 分
	Stock   int
	Version int `gorm:"default:1"` // 乐观锁用
}

// CartItem 购物车行。
type CartItem struct {
	gorm.Model
	UserID    uint `gorm:"index"`
	ProductID uint `gorm:"index"`
	Quantity  int
}

// OrderStatus 订单状态流转。
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderShipped   OrderStatus = "shipped"
	OrderDone      OrderStatus = "done"
	OrderCancelled OrderStatus = "cancelled"
)

// Order 订单头。
type Order struct {
	gorm.Model
	UserID  uint        `gorm:"index"`
	Status  OrderStatus `gorm:"size:32;index"`
	Total   int64       // 分
	OrderAt *time.Time
	PaidAt  *time.Time
	ShipAt  *time.Time
	DoneAt  *time.Time
	Items   []OrderItem `gorm:"foreignKey:OrderID"`
}

// OrderItem 订单行。
type OrderItem struct {
	gorm.Model
	OrderID   uint `gorm:"index"`
	ProductID uint
	Quantity  int
	Price     int64 // 下单时单价快照（分）
}
