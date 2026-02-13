package usecase

import (
	"errors"
	"net/http"
	"strconv"

	"go-gw-test/cmd/orders_gw/internal/repo"
	"go-gw-test/cmd/orders_gw/internal/types"
	"go-gw-test/cmd/orders_gw/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrdersUseCaseImpl implements orders endpoints and helpers.
type OrdersUseCaseImpl struct {
	repo repo.OrdersRepo
}

// NewOrdersUseCase constructs an OrdersUseCase implementation.
func NewOrdersUseCase(ordersRepo repo.OrdersRepo) *OrdersUseCaseImpl {
	return &OrdersUseCaseImpl{
		repo: ordersRepo,
	}
}

// ListOrders returns all orders.
// @Summary List orders
// @Tags orders-gw
// @Produce json
// @Success 200 {object} types.OrdersResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [get]
func (u *OrdersUseCaseImpl) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orders, err := u.repo.ListOrders(ctx)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list orders"})
		return
	}

	resp := mapOrdersResponse(orders)
	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetOrder returns an order by ID.
// @Summary Get order
// @Tags orders-gw
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (u *OrdersUseCaseImpl) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idValue := mux.Vars(r)["id"]
	orderID, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		zap.L().Error("parse order id", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	order, err := u.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch order"})
		return
	}

	resp := types.OrderResponse{
		ID:     order.ID,
		UserID: order.UserID,
		Status: order.Status,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetOrderItems returns items for an order.
// @Summary Get order items
// @Tags orders-gw
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.OrderItemsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id}/items [get]
func (u *OrdersUseCaseImpl) GetOrderItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idValue := mux.Vars(r)["id"]
	orderID, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		zap.L().Error("parse order id", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	_, err = u.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch order"})
		return
	}

	items, err := u.repo.ListOrderItems(ctx, orderID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list order items"})
		return
	}

	resp := mapOrderItemsResponse(items)
	utils.WriteJSON(w, http.StatusOK, resp)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (u *OrdersUseCaseImpl) NotFound(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

// mapOrdersResponse maps entities into a list response.
func mapOrdersResponse(orders []types.OrderRecord) types.OrdersResponse {
	resp := types.OrdersResponse{Orders: make([]types.OrderResponse, 0, len(orders))}
	for _, order := range orders {
		resp.Orders = append(resp.Orders, types.OrderResponse{
			ID:     order.ID,
			UserID: order.UserID,
			Status: order.Status,
		})
	}

	return resp
}

// mapOrderItemsResponse maps item entities into a list response.
func mapOrderItemsResponse(items []types.OrderItem) types.OrderItemsResponse {
	resp := types.OrderItemsResponse{Items: make([]types.OrderItemResponse, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, types.OrderItemResponse{
			ID:        item.ID,
			OrderID:   item.OrderID,
			SKU:       item.SKU,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return resp
}
