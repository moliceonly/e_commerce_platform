package handler

import (
	"net/http"
	"strconv"

	"e_commerce_platform/internal/model"
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
	var req struct {
		Name  string `json:"name" binding:"required"`
		Price int64  `json:"price" binding:"gte=0"`
		Stock int    `json:"stock" binding:"gte=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	p, err := h.Svc.CreateProduct(c.Request.Context(), req.Name, req.Price, req.Stock)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40003, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Get(c *gin.Context) {
	// TODO(3.1): path :id → Svc.GetProduct
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	product, err := h.Svc.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, 40401, err.Error())
		return
	}
	response.OK(c, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	// TODO(3.1 / 3.3): query page/page_size → Svc.ListProducts
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, err := h.Svc.ListProducts(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.OK(c, list)
}

type CartHandler struct{ Svc *service.CartService }

func (h *CartHandler) Add(c *gin.Context) {
	// TODO(3.2): 从 context 取 userID，bind product_id/quantity → Svc.Add
	userid := uint(1)
	// if err != nil {
	// 	response.Fail(c, http.StatusBadRequest, 40001, err.Error())
	// 	return
	// }

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	if err := h.Svc.Add(c.Request.Context(), uint(userid), req.ProductID, req.Quantity); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

type OrderHandler struct{ Svc *service.OrderService }

func (h *OrderHandler) Place(c *gin.Context) {
	// TODO(3.1/B): bind items → Svc.PlaceOrder（事务扣库存）
	userID := uint(1)
	var req struct {
		Items []struct {
			ProductID uint `json:"product_id" binding:"required"`
			Quantity  int  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	items := make([]service.OrderLine, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.OrderLine{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	o, err := h.Svc.PlaceOrder(c.Request.Context(), userID, items)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.OK(c, o)
}

func (h *OrderHandler) List(c *gin.Context) {
	// TODO(3.3): 分页 + status 过滤；只查当前用户
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO: ListOrders")
}

func (h *OrderHandler) Transition(c *gin.Context) {
	// TODO(3.2): 校验订单归属 + 状态机 → Svc.Transition
	userID := uint(1)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "Invalid id")
		return
	}

	var req struct {
		Status model.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusForbidden, 40001, err.Error())
		return
	}

	if err := h.Svc.Transition(c.Request.Context(), userID, uint(id), req.Status); err != nil {
		response.Fail(c, http.StatusForbidden, 40301, err.Error())
		return
	}

	response.OK(c, gin.H{"ok": true})
}
