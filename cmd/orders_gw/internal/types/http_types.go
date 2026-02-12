package types

// OrderResponse represents an order response payload.
type OrderResponse struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Status string `json:"status"`
}

// OrdersResponse represents a list response payload.
type OrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
}

// OrderItemResponse represents an order item payload.
type OrderItemResponse struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// OrderItemsResponse represents a list of items for an order.
type OrderItemsResponse struct {
	Items []OrderItemResponse `json:"items"`
}
