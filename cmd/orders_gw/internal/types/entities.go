package types

// OrderRecord represents a basic order record.
type OrderRecord struct {
	ID     int64  `gorm:"primaryKey;column:id" json:"id"`
	UserID int64  `gorm:"column:user_id" json:"user_id"`
	Status string `gorm:"column:status" json:"status"`
}

// OrderItem represents an item within an order.
type OrderItem struct {
	ID        int64   `gorm:"primaryKey;column:id" json:"id"`
	OrderID   int64   `gorm:"column:order_id;index" json:"order_id"`
	SKU       string  `gorm:"column:sku" json:"sku"`
	Name      string  `gorm:"column:name" json:"name"`
	Quantity  int     `gorm:"column:quantity" json:"quantity"`
	UnitPrice float64 `gorm:"column:unit_price" json:"unit_price"`
}
