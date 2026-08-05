package handler

import "e_commerce_platform/internal/model"

// 请求/响应 DTO：供 handler bind 与 swag 注解引用。

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateProductRequest struct {
	Name  string `json:"name" binding:"required"`
	Price int64  `json:"price" binding:"gte=0"`
	Stock int    `json:"stock" binding:"gte=0"`
}

type AddCartItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type PlaceOrderItem struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type PlaceOrderRequest struct {
	Items []PlaceOrderItem `json:"items" binding:"required,min=1"`
}

type TransitionRequest struct {
	Status model.OrderStatus `json:"status" binding:"required"`
}

type RegisterData struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type LoginData struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

type RefreshData struct {
	Token string `json:"token"`
}

type AvatarData struct {
	URL string `json:"url"`
}

type OKData struct {
	OK bool `json:"ok"`
}
