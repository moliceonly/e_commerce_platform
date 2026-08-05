package handler

import (
	"net/http"
	"strconv"

	"e_commerce_platform/internal/errcode"
	"e_commerce_platform/internal/middleware"
	"e_commerce_platform/internal/response"
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
)

// Healthz 探活（阶段 A 先跑通这个）。
// @Summary      探活
// @Tags         system
// @Produce      json
// @Success      200 {object} response.Body
// @Router       /healthz [get]
func Healthz(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

type AuthHandler struct{ Svc *service.AuthService }

// Register 用户注册。
// @Summary      注册
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "注册信息"
// @Success      200 {object} response.Body
// @Failure      400 {object} response.Body
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	user, err := h.Svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, RegisterData{
		ID:    user.ID,
		Email: user.Email,
	})
}

// Login 登录，返回 access + refresh token。
// @Summary      登录
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "登录信息"
// @Success      200 {object} response.Body
// @Failure      400 {object} response.Body
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	token, refresh, err := h.Svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, LoginData{
		Token:   token,
		Refresh: refresh,
	})
}

// Refresh 用 refresh token 换新的 access token。
// @Summary      刷新 access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "refresh_token"
// @Success      200 {object} response.Body
// @Failure      401 {object} response.Body
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	token, err := h.Svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, errcode.ErrUnauthorized, err.Error())
		return
	}
	response.OK(c, RefreshData{Token: token})
}

type ProductHandler struct{ Svc *service.CatalogService }

// Create 创建商品（需 admin）。
// @Summary      创建商品
// @Tags         products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body CreateProductRequest true "商品"
// @Success      200 {object} response.Body
// @Failure      400 {object} response.Body
// @Failure      403 {object} response.Body
// @Router       /api/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
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

// Get 商品详情。
// @Summary      商品详情
// @Tags         products
// @Produce      json
// @Param        id path int true "商品 ID"
// @Success      200 {object} response.Body
// @Failure      404 {object} response.Body
// @Router       /api/v1/products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
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

// List 商品列表。
// @Summary      商品列表
// @Tags         products
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} response.Body
// @Router       /api/v1/products [get]
func (h *ProductHandler) List(c *gin.Context) {
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

// Add 加入购物车。
// @Summary      加入购物车
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body AddCartItemRequest true "购物车项"
// @Success      200 {object} response.Body
// @Failure      401 {object} response.Body
// @Router       /api/v1/cart/items [post]
func (h *CartHandler) Add(c *gin.Context) {
	userId := c.GetUint("userID")

	var req AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, err.Error())
		return
	}
	if err := h.Svc.Add(c.Request.Context(), userId, req.ProductID, req.Quantity); err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, err.Error())
		return
	}
	response.OK(c, OKData{OK: true})
}

type OrderHandler struct{ Svc *service.OrderService }

// Place 下单（事务扣库存）。
// @Summary      下单
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body PlaceOrderRequest true "订单行"
// @Success      200 {object} response.Body
// @Failure      401 {object} response.Body
// @Router       /api/v1/orders [post]
func (h *OrderHandler) Place(c *gin.Context) {
	userId := c.GetUint("userID")
	var req PlaceOrderRequest
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

// List 当前用户订单列表。
// @Summary      订单列表
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "状态过滤 pending/paid/shipped/done/cancelled"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} response.Body
// @Failure      401 {object} response.Body
// @Router       /api/v1/orders [get]
func (h *OrderHandler) List(c *gin.Context) {
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

// Transition 订单状态流转。
// @Summary      订单状态流转
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path int true "订单 ID"
// @Param        body body TransitionRequest true "目标状态"
// @Success      200 {object} response.Body
// @Failure      403 {object} response.Body
// @Router       /api/v1/orders/{id}/transition [post]
func (h *OrderHandler) Transition(c *gin.Context) {
	userId := c.GetUint("userID")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, "Invalid id")
		return
	}

	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusForbidden, errcode.ErrForbidden, err.Error())
		return
	}

	if err := h.Svc.Transition(c.Request.Context(), userId, uint(id), req.Status); err != nil {
		response.Fail(c, http.StatusForbidden, errcode.ErrForbidden, err.Error())
		return
	}

	response.OK(c, OKData{OK: true})
}

// UploadHandler 用户资料上传（阶段 H · 3.3）。
type UploadHandler struct{ Svc *service.UploadService }

// Avatar POST multipart field name = "file"
// @Summary      上传头像
// @Tags         user
// @Security     BearerAuth
// @Accept       mpfd
// @Produce      json
// @Param        file formData file true "头像文件"
// @Success      200 {object} response.Body
// @Failure      400 {object} response.Body
// @Router       /api/v1/me/avatar [post]
func (h *UploadHandler) Avatar(c *gin.Context) {
	uid := c.GetUint(middleware.CtxUserID)
	fh, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.ErrBadRequest, "invalid file")
		return
	}

	url, err := h.Svc.SaveAvatar(c.Request.Context(), uid, fh)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.ErrInternal, "save file failed")
		return
	}
	response.OK(c, AvatarData{URL: url})
}
