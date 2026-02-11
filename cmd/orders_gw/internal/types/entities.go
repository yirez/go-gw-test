package types

// OrderRecord represents a basic order record.
type OrderRecord struct {
	ID       int64  `gorm:"primaryKey;column:id" json:"id"`
	UserID   int64  `gorm:"column:user_id" json:"user_id"`
	Item     string `gorm:"column:item" json:"item"`
	Quantity int    `gorm:"column:quantity" json:"quantity"`
	Status   string `gorm:"column:status" json:"status"`
}
