package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"e_commerce_platform/internal/model"
	"e_commerce_platform/internal/repository"
	"e_commerce_platform/internal/response"
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "trainer:Train2026Lib!@tcp(127.0.0.1:3306)/training_lib?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("mysql unavailable:", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Product{}, &model.CartItem{}, &model.Order{}, &model.OrderItem{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := openTestDB(t)
	secret := "handler-test-secret"
	productRepo := &repository.ProductRepo{DB: db}
	userRepo := &repository.UserRepo{DB: db}
	cartRepo := &repository.CartRepo{DB: db}
	orderRepo := &repository.OrderRepo{DB: db}
	return NewRouter(Deps{
		JWTSecret: secret,
		Catalog:   &service.CatalogService{Products: productRepo},
		Auth:      &service.AuthService{Users: userRepo, JWTSecret: secret},
		Cart:      &service.CartService{Carts: cartRepo, Products: productRepo},
		Order:     &service.OrderService{DB: db, Products: productRepo, Carts: cartRepo, Orders: orderRepo},
	})
}

func doJSON(t *testing.T, r http.Handler, method, path, token string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	var body *bytes.Buffer
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewBuffer(b)
	} else {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeOK(t *testing.T, w *httptest.ResponseRecorder) response.Body {
	t.Helper()
	var body response.Body
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, w.Body.String())
	}
	return body
}

func TestHealthz(t *testing.T) {
	r := NewRouter(Deps{JWTSecret: "test"})
	w := doJSON(t, r, http.MethodGet, "/healthz", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID")
	}
	body := decodeOK(t, w)
	if body.Code != 0 {
		t.Fatalf("code=%d msg=%s", body.Code, body.Message)
	}
}

func TestAllRoutes(t *testing.T) {
	r := newTestRouter(t)
	email := fmt.Sprintf("h-%d@example.com", time.Now().UnixNano())
	pass := "123456"

	// POST /api/v1/auth/register
	w := doJSON(t, r, http.MethodPost, "/api/v1/auth/register", "", gin.H{
		"email": email, "password": pass,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register status=%d body=%s", w.Code, w.Body.String())
	}
	reg := decodeOK(t, w)
	if reg.Code != 0 {
		t.Fatalf("register code=%d msg=%s", reg.Code, reg.Message)
	}

	// POST /api/v1/auth/login（普通 user）
	w = doJSON(t, r, http.MethodPost, "/api/v1/auth/login", "", gin.H{
		"email": email, "password": pass,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", w.Code, w.Body.String())
	}
	login := decodeOK(t, w)
	m, _ := login.Data.(map[string]any)
	token, _ := m["token"].(string)
	refresh, _ := m["refresh"].(string)
	if token == "" {
		t.Fatalf("empty token body=%s", w.Body.String())
	}

	// user 创建商品 → 403
	w = doJSON(t, r, http.MethodPost, "/api/v1/products", token, gin.H{
		"name": "handler-ut", "price": 9900, "stock": 20,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("user create want 403 got %d body=%s", w.Code, w.Body.String())
	}

	// 提权 admin 后重新登录（JWT 内 role 才会变）
	db := openTestDB(t)
	if err := db.Model(&model.User{}).Where("email = ?", email).Update("role", "admin").Error; err != nil {
		t.Fatal(err)
	}
	w = doJSON(t, r, http.MethodPost, "/api/v1/auth/login", "", gin.H{
		"email": email, "password": pass,
	})
	login = decodeOK(t, w)
	m, _ = login.Data.(map[string]any)
	token, _ = m["token"].(string)
	if token == "" {
		t.Fatal("empty admin token")
	}

	// refresh 换发 access
	if refresh != "" {
		w = doJSON(t, r, http.MethodPost, "/api/v1/auth/refresh", "", gin.H{
			"refresh_token": refresh,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("refresh status=%d body=%s", w.Code, w.Body.String())
		}
	}

	// POST /api/v1/products（admin）
	w = doJSON(t, r, http.MethodPost, "/api/v1/products", token, gin.H{
		"name": "handler-ut", "price": 9900, "stock": 20,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create product status=%d body=%s", w.Code, w.Body.String())
	}
	created := decodeOK(t, w)
	pm, _ := created.Data.(map[string]any)
	var productID uint
	switch id := pm["ID"].(type) {
	case float64:
		productID = uint(id)
	default:
		if v, ok := pm["id"].(float64); ok {
			productID = uint(v)
		}
	}
	if productID == 0 {
		t.Fatalf("bad product id body=%s", w.Body.String())
	}

	// GET /api/v1/products/:id
	w = doJSON(t, r, http.MethodGet, fmt.Sprintf("/api/v1/products/%d", productID), "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get product status=%d body=%s", w.Code, w.Body.String())
	}
	if decodeOK(t, w).Code != 0 {
		t.Fatal("get product business code != 0")
	}

	// GET /api/v1/products?page=&page_size=
	w = doJSON(t, r, http.MethodGet, "/api/v1/products?page=1&page_size=20", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list products status=%d body=%s", w.Code, w.Body.String())
	}

	// POST /api/v1/orders without token → 401
	w = doJSON(t, r, http.MethodPost, "/api/v1/orders", "", gin.H{
		"items": []gin.H{{"product_id": productID, "quantity": 1}},
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d body=%s", w.Code, w.Body.String())
	}

	// POST /api/v1/cart/items
	w = doJSON(t, r, http.MethodPost, "/api/v1/cart/items", token, gin.H{
		"product_id": productID, "quantity": 2,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("cart add status=%d body=%s", w.Code, w.Body.String())
	}

	// POST /api/v1/orders
	w = doJSON(t, r, http.MethodPost, "/api/v1/orders", token, gin.H{
		"items": []gin.H{{"product_id": productID, "quantity": 1}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("place order status=%d body=%s", w.Code, w.Body.String())
	}
	orderBody := decodeOK(t, w)
	om, _ := orderBody.Data.(map[string]any)
	var orderID uint
	switch id := om["ID"].(type) {
	case float64:
		orderID = uint(id)
	default:
		if v, ok := om["id"].(float64); ok {
			orderID = uint(v)
		}
	}
	if orderID == 0 {
		t.Fatalf("bad order id body=%s", w.Body.String())
	}

	// GET /api/v1/orders
	w = doJSON(t, r, http.MethodGet, "/api/v1/orders?page=1&page_size=20", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list orders status=%d body=%s", w.Code, w.Body.String())
	}

	// GET /api/v1/orders?status=pending
	w = doJSON(t, r, http.MethodGet, "/api/v1/orders?page=1&page_size=20&status=pending", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list orders status filter=%d body=%s", w.Code, w.Body.String())
	}

	// POST /api/v1/orders/:id/transition
	w = doJSON(t, r, http.MethodPost, fmt.Sprintf("/api/v1/orders/%d/transition", orderID), token, gin.H{
		"status": "paid",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("transition status=%d body=%s", w.Code, w.Body.String())
	}
}
