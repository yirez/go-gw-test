package types

// OrderResponse represents an order response payload.
type OrderResponse struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

// OrdersResponse represents a list response payload.
type OrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
}
