package service

import (
	"context"
	"testing"

	"e_commerce_platform/internal/model"

	"gorm.io/gorm"
)

// 阶段 G 骨架：mock ProductStore，不连真实 MySQL。
// 实现步骤：
//  1. 填满 mockProductStore 各方法（可用字段记录调用、返回预设数据）
//  2. 去掉 t.Skip，断言 GetProduct / CreateProduct 行为

type mockProductStore struct {
	// TODO(G): 按需加字段，如 getResult *model.Product, getErr error, created []*model.Product
}

func (m *mockProductStore) Create(ctx context.Context, p *model.Product) error {
	// TODO(G)
	return nil
}

func (m *mockProductStore) Get(ctx context.Context, id uint) (*model.Product, error) {
	// TODO(G)
	return nil, nil
}

func (m *mockProductStore) List(ctx context.Context, offset, limit int) ([]model.Product, error) {
	// TODO(G)
	return nil, nil
}

func (m *mockProductStore) DeductStockTx(ctx context.Context, tx *gorm.DB, productID uint, qty int) error {
	// TODO(G)
	return nil
}

func TestCatalogService_GetProduct_withMock(t *testing.T) {
	t.Skip("TODO(G): 用 mockProductStore 测 CatalogService.GetProduct，不连 MySQL")

	// 示例骨架：
	// mock := &mockProductStore{ ... }
	// svc := &CatalogService{Products: mock}
	// p, err := svc.GetProduct(context.Background(), 1)
	// if err != nil { t.Fatal(err) }
	// ...
	_ = context.Background
}
