package service

import (
	"context"
	"fmt"
	"testing"

	"e_commerce_platform/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// —— mock OrderStore / CartStore（与 catalog_mock_test 同包，复用 mockProductStore）——

type mockOrderStore struct {
	byID   map[uint]*model.Order
	nextID uint
}

func newMockOrderStore(seed ...*model.Order) *mockOrderStore {
	m := &mockOrderStore{byID: make(map[uint]*model.Order), nextID: 760001}
	for _, o := range seed {
		cp := *o
		m.byID[cp.ID] = &cp
		if cp.ID >= m.nextID {
			m.nextID = cp.ID + 1
		}
	}
	return m
}

func (m *mockOrderStore) CreateWithItems(ctx context.Context, tx *gorm.DB, order *model.Order, items []model.OrderItem) error {
	if order.ID == 0 {
		order.ID = m.nextID
		m.nextID++
	}
	order.Items = items
	cp := *order
	cp.Items = append([]model.OrderItem(nil), items...)
	m.byID[order.ID] = &cp
	return nil
}

func (m *mockOrderStore) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	o, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockOrderStore) ListByUser(ctx context.Context, userID uint, status string, offset, limit int) ([]model.Order, error) {
	var out []model.Order
	for _, o := range m.byID {
		if o.UserID != userID {
			continue
		}
		if status != "" && o.Status != model.OrderStatus(status) {
			continue
		}
		out = append(out, *o)
	}
	if offset >= len(out) {
		return []model.Order{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func (m *mockOrderStore) UpdateStatus(ctx context.Context, orderID uint, from, to model.OrderStatus) error {
	o, ok := m.byID[orderID]
	if !ok || o.Status != from {
		return fmt.Errorf("order not found or status not %s", from)
	}
	o.Status = to
	return nil
}

type mockCartStore struct {
	cleared []cartClearCall
}

type cartClearCall struct {
	userID     uint
	productIDs []uint
}

func (m *mockCartStore) Upsert(ctx context.Context, userID, productID uint, qty int) error {
	return nil
}

func (m *mockCartStore) ListByUser(ctx context.Context, userID uint) ([]model.CartItem, error) {
	return nil, nil
}

func (m *mockCartStore) ClearByUser(ctx context.Context, tx *gorm.DB, userID uint, productIDs []uint) error {
	ids := append([]uint(nil), productIDs...)
	m.cleared = append(m.cleared, cartClearCall{userID: userID, productIDs: ids})
	return nil
}

// 仅用于驱动 gorm.Transaction；业务数据全在 mock 里。
func openMockTxDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:order_mock?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestOrderService_PlaceOrder_ok_withMock(t *testing.T) {
	const (
		buyerID    uint = 520318
		keyboardID uint = 884201
	)
	keyboard := &model.Product{
		Name:  "Keychron K2 无线机械键盘",
		Price: 79900,
		Stock: 36,
	}
	keyboard.ID = keyboardID

	products := newMockProductStore(keyboard)
	orders := newMockOrderStore()
	carts := &mockCartStore{}
	svc := &OrderService{
		DB:       openMockTxDB(t),
		Products: products,
		Orders:   orders,
		Carts:    carts,
	}

	o, err := svc.PlaceOrder(context.Background(), buyerID, []OrderLine{
		{ProductID: keyboardID, Quantity: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.UserID != buyerID || o.Status != model.OrderPending {
		t.Fatalf("order=%+v", o)
	}
	if o.Total != 79900*2 {
		t.Fatalf("total=%d want %d", o.Total, 79900*2)
	}
	if o.ID == 0 {
		t.Fatal("order id not assigned")
	}

	left, err := products.Get(context.Background(), keyboardID)
	if err != nil {
		t.Fatal(err)
	}
	if left.Stock != 34 {
		t.Fatalf("stock=%d want 34", left.Stock)
	}
	if len(carts.cleared) != 1 || carts.cleared[0].userID != buyerID {
		t.Fatalf("cart clear=%+v", carts.cleared)
	}
}

func TestOrderService_PlaceOrder_stockNotEnough_withMock(t *testing.T) {
	const (
		buyerID uint = 520318
		teaID   uint = 330088
	)
	tea := &model.Product{
		Name:  "三顿半超即溶咖啡 24 颗装",
		Price: 11800,
		Stock: 3,
	}
	tea.ID = teaID

	products := newMockProductStore(tea)
	svc := &OrderService{
		DB:       openMockTxDB(t),
		Products: products,
		Orders:   newMockOrderStore(),
		Carts:    &mockCartStore{},
	}

	_, err := svc.PlaceOrder(context.Background(), buyerID, []OrderLine{
		{ProductID: teaID, Quantity: 8},
	})
	if err == nil {
		t.Fatal("expected stock not enough")
	}

	left, err := products.Get(context.Background(), teaID)
	if err != nil {
		t.Fatal(err)
	}
	if left.Stock != 3 {
		t.Fatalf("stock should stay 3, got %d", left.Stock)
	}
}

func TestOrderService_Transition_withMock(t *testing.T) {
	const (
		ownerID uint = 520318
		otherID uint = 520999
		orderID uint = 760045
	)
	pending := &model.Order{
		UserID: ownerID,
		Status: model.OrderPending,
		Total:  159800,
	}
	pending.ID = orderID

	orders := newMockOrderStore(pending)
	svc := &OrderService{
		DB:       openMockTxDB(t),
		Products: newMockProductStore(),
		Orders:   orders,
		Carts:    &mockCartStore{},
	}
	ctx := context.Background()

	if err := svc.Transition(ctx, ownerID, orderID, model.OrderPaid); err != nil {
		t.Fatal(err)
	}
	got, _ := orders.GetByID(ctx, orderID)
	if got.Status != model.OrderPaid {
		t.Fatalf("status=%s", got.Status)
	}

	if err := svc.Transition(ctx, otherID, orderID, model.OrderShipped); err == nil {
		t.Fatal("expected forbidden")
	}

	if err := svc.Transition(ctx, ownerID, orderID, model.OrderDone); err == nil {
		t.Fatal("expected invalid transition paid -> done")
	}
}
