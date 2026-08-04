package handler

import (
	"net/http"
	"strconv"

	"e_commerce_platform/internal/errcode"
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
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	user, err := h.Svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, gin.H{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	// TODO(3.2): bind → Svc.Login → 返回 token
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	token, refresh, err := h.Svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, gin.H{
		"token":   token,
		"refresh": refresh,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	token, err := h.Svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, errcode.ErrUnauthorized, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token})
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
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}

	p, err := h.Svc.CreateProduct(c.Request.Context(), req.Name, req.Price, req.Stock)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Get(c *gin.Context) {
	// TODO(3.1): path :id → Svc.GetProduct
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	product, err := h.Svc.GetProduct(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, errcode.ErrNotFound, err.Error())
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
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, list)
}

type CartHandler struct{ Svc *service.CartService }

func (h *CartHandler) Add(c *gin.Context) {
	// TODO(3.2): 从 context 取 userID，bind product_id/quantity → Svc.Add
	userId := c.GetUint("userID")

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	if err := h.Svc.Add(c.Request.Context(), userId, req.ProductID, req.Quantity); err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

type OrderHandler struct{ Svc *service.OrderService }

func (h *OrderHandler) Place(c *gin.Context) {
	// TODO(3.1/B): bind items → Svc.PlaceOrder（事务扣库存）
	userId := c.GetUint("userID")
	var req struct {
		Items []struct {
			ProductID uint `json:"product_id" binding:"required"`
			Quantity  int  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}

	items := make([]service.OrderLine, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.OrderLine{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	o, err := h.Svc.PlaceOrder(c.Request.Context(), userId, items)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, o)
}

func (h *OrderHandler) List(c *gin.Context) {
	// TODO(3.3): 分页 + status 过滤；只查当前用户
	userId := c.GetUint("userID")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, err := h.Svc.ListOrders(c.Request.Context(), userId, status, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusNotImplemented, errcode.ErrNotImplemented, "order not found")
		return
	}

	response.OK(c, list)
}

func (h *OrderHandler) Transition(c *gin.Context) {
	// TODO(3.2): 校验订单归属 + 状态机 → Svc.Transition
	userId := c.GetUint("userID")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, "Invalid id")
		return
	}

	var req struct {
		Status model.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusForbidden, errcode.ErrForbidden, err.Error())
		return
	}

	if err := h.Svc.Transition(c.Request.Context(), userId, uint(id), req.Status); err != nil {
		response.Fail(c, http.StatusForbidden, errcode.ErrForbidden, err.Error())
		return
	}

	response.OK(c, gin.H{"ok": true})
}
