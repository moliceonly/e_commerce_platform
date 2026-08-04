package service

import (
	"context"
	"fmt"
	"testing"

	"e_commerce_platform/internal/model"

	"gorm.io/gorm"
)

// mockProductStore 内存假仓库，不连 MySQL。
type mockProductStore struct {
	byID    map[uint]*model.Product
	nextID  uint
	createErr error
}

func newMockProductStore(seed ...*model.Product) *mockProductStore {
	m := &mockProductStore{
		byID:   make(map[uint]*model.Product),
		nextID: 10001,
	}
	for _, p := range seed {
		cp := *p
		m.byID[cp.ID] = &cp
		if cp.ID >= m.nextID {
			m.nextID = cp.ID + 1
		}
	}
	return m
}

func (m *mockProductStore) Create(ctx context.Context, p *model.Product) error {
	if m.createErr != nil {
		return m.createErr
	}
	if p.ID == 0 {
		p.ID = m.nextID
		m.nextID++
	}
	cp := *p
	m.byID[p.ID] = &cp
	return nil
}

func (m *mockProductStore) Get(ctx context.Context, id uint) (*model.Product, error) {
	p, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockProductStore) List(ctx context.Context, offset, limit int) ([]model.Product, error) {
	out := make([]model.Product, 0, len(m.byID))
	for _, p := range m.byID {
		out = append(out, *p)
	}
	if offset >= len(out) {
		return []model.Product{}, nil
	}
	end := offset + limit
	if end > len(out) || limit <= 0 {
		end = len(out)
	}
	return out[offset:end], nil
}

func (m *mockProductStore) DeductStockTx(ctx context.Context, tx *gorm.DB, productID uint, qty int) error {
	p, ok := m.byID[productID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if p.Stock < qty {
		return fmt.Errorf("stock not enough")
	}
	p.Stock -= qty
	return nil
}

func TestCatalogService_GetProduct_withMock(t *testing.T) {
	const productID uint = 884201
	seed := &model.Product{
		Name:  "Keychron K2 无线机械键盘",
		Price: 79900, // 分：¥799.00
		Stock: 36,
	}
	seed.ID = productID

	mock := newMockProductStore(seed)
	svc := &CatalogService{Products: mock}
	ctx := context.Background()

	got, err := svc.GetProduct(ctx, productID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Keychron K2 无线机械键盘" {
		t.Fatalf("name=%q", got.Name)
	}
	if got.Price != 79900 || got.Stock != 36 {
		t.Fatalf("price=%d stock=%d", got.Price, got.Stock)
	}

	_, err = svc.GetProduct(ctx, 999999)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("want ErrRecordNotFound, got %v", err)
	}
}

func TestCatalogService_CreateProduct_withMock(t *testing.T) {
	mock := newMockProductStore()
	svc := &CatalogService{Products: mock}
	ctx := context.Background()

	created, err := svc.CreateProduct(ctx, "罗技 MX Master 3S 鼠标", 69900, 120)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("expected assigned product id")
	}
	if created.Name != "罗技 MX Master 3S 鼠标" || created.Price != 69900 || created.Stock != 120 {
		t.Fatalf("unexpected product: %+v", created)
	}

	stored, err := mock.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != created.Name {
		t.Fatalf("store name=%q want %q", stored.Name, created.Name)
	}

	mock.createErr = fmt.Errorf("duplicate sku")
	if _, err := svc.CreateProduct(ctx, "失败商品", 100, 1); err == nil {
		t.Fatal("expected create error")
	}
}
