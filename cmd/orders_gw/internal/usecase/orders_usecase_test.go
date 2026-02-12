package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gw-test/cmd/orders_gw/internal/types"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type fakeOrdersRepo struct {
	orders   []types.OrderRecord
	listErr  error
	order    types.OrderRecord
	orderErr error
	items    []types.OrderItem
	itemsErr error
}

// ListOrders returns configured fake order list.
func (f *fakeOrdersRepo) ListOrders(ctx context.Context) ([]types.OrderRecord, error) {
	return f.orders, f.listErr
}

// FindOrderByID returns configured fake order lookup.
func (f *fakeOrdersRepo) FindOrderByID(ctx context.Context, orderID int64) (types.OrderRecord, error) {
	return f.order, f.orderErr
}

// ListOrderItems returns configured fake item list.
func (f *fakeOrdersRepo) ListOrderItems(ctx context.Context, orderID int64) ([]types.OrderItem, error) {
	return f.items, f.itemsErr
}

// SeedIfEmpty is a no-op fake.
func (f *fakeOrdersRepo) SeedIfEmpty(ctx context.Context) error {
	return nil
}

// TestOrdersUseCaseListOrdersSuccess verifies list response shape.
func TestOrdersUseCaseListOrdersSuccess(t *testing.T) {
	u := NewOrdersUseCase(&fakeOrdersRepo{
		orders: []types.OrderRecord{{ID: 1, UserID: 1, Status: "processing"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	rr := httptest.NewRecorder()
	u.ListOrders(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp types.OrdersResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Orders) != 1 {
		t.Fatalf("expected one order, got %d", len(resp.Orders))
	}
}

// TestOrdersUseCaseGetOrderInvalidID verifies bad path id handling.
func TestOrdersUseCaseGetOrderInvalidID(t *testing.T) {
	u := NewOrdersUseCase(&fakeOrdersRepo{})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/orders/not-int", nil), map[string]string{"id": "not-int"})
	rr := httptest.NewRecorder()

	u.GetOrder(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

// TestOrdersUseCaseGetOrderItemsNotFound verifies 404 when order is missing.
func TestOrdersUseCaseGetOrderItemsNotFound(t *testing.T) {
	u := NewOrdersUseCase(&fakeOrdersRepo{orderErr: gorm.ErrRecordNotFound})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/orders/1/items", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetOrderItems(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

// TestOrdersUseCaseGetOrderItemsSuccess verifies items are returned when order exists.
func TestOrdersUseCaseGetOrderItemsSuccess(t *testing.T) {
	u := NewOrdersUseCase(&fakeOrdersRepo{
		order: types.OrderRecord{ID: 1, UserID: 1, Status: "processing"},
		items: []types.OrderItem{
			{ID: 1, OrderID: 1, SKU: "sku1", Name: "Item1", Quantity: 1, UnitPrice: 9.99},
		},
	})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/orders/1/items", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetOrderItems(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp types.OrderItemsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(resp.Items))
	}
}

// TestOrdersUseCaseGetOrderItemsListError verifies item list failures map to 500.
func TestOrdersUseCaseGetOrderItemsListError(t *testing.T) {
	u := NewOrdersUseCase(&fakeOrdersRepo{
		order:    types.OrderRecord{ID: 1},
		itemsErr: errors.New("db fail"),
	})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/orders/1/items", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetOrderItems(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
