package types

// UserProfile represents a basic user profile record.
type UserProfile struct {
	ID    int64  `gorm:"primaryKey;column:id" json:"id"`
	Name  string `gorm:"column:name" json:"name"`
	Email string `gorm:"column:email" json:"email"`
	Phone string `gorm:"column:phone" json:"phone"`
}
