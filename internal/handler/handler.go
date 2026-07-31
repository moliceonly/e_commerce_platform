package handler

import (
	"net/http"

	"e_commerce_platform/internal/response"
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
)

// Healthz 探活（阶段 A 先跑通这个）。
func Healthz(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

type AuthHandler struct{ Svc *service.AuthService }

func (h *AuthHandler) Register(c *gin.Context) {
	// TODO(3.2): bind email/password → Svc.Register → OK
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: Register")
}

func (h *AuthHandler) Login(c *gin.Context) {
	// TODO(3.2): bind → Svc.Login → 返回 token
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: Login")
}

type ProductHandler struct{ Svc *service.CatalogService }

func (h *ProductHandler) Create(c *gin.Context) {
	// TODO(3.1): bind name/price/stock → Svc.CreateProduct
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: CreateProduct")
}

func (h *ProductHandler) Get(c *gin.Context) {
	// TODO(3.1): path :id → Svc.GetProduct
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: GetProduct")
}

func (h *ProductHandler) List(c *gin.Context) {
	// TODO(3.1 / 3.3): query page/page_size → Svc.ListProducts
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: ListProducts")
}

type CartHandler struct{ Svc *service.CartService }

func (h *CartHandler) Add(c *gin.Context) {
	// TODO(3.2): 从 context 取 userID，bind product_id/quantity → Svc.Add
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: CartAdd")
}

type OrderHandler struct{ Svc *service.OrderService }

func (h *OrderHandler) Place(c *gin.Context) {
	// TODO(3.1/B): bind items → Svc.PlaceOrder（事务扣库存）
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: PlaceOrder")
}

func (h *OrderHandler) List(c *gin.Context) {
	// TODO(3.3): 分页 + status 过滤；只查当前用户
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: ListOrders")
}

func (h *OrderHandler) Transition(c *gin.Context) {
	// TODO(3.2): 校验订单归属 + 状态机 → Svc.Transition
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: Transition")
}
