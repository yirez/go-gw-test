package types

// UserProfile represents a basic user profile record.
type UserProfile struct {
	ID    int64  `gorm:"primaryKey;column:id" json:"id"`
	Name  string `gorm:"column:name" json:"name"`
	Email string `gorm:"column:email" json:"email"`
	Phone string `gorm:"column:phone" json:"phone"`
}

// UserContactInfo stores basic contact details for a user.
type UserContactInfo struct {
	UserID       int64  `gorm:"primaryKey;column:user_id" json:"user_id"`
	Email        string `gorm:"column:email" json:"email"`
	Phone        string `gorm:"column:phone" json:"phone"`
	AddressLine1 string `gorm:"column:address_line1" json:"address_line1"`
	City         string `gorm:"column:city" json:"city"`
	Country      string `gorm:"column:country" json:"country"`
}
