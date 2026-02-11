package types

// UserRecord represents a user credential record.
type UserRecord struct {
	ID           int64  `gorm:"primaryKey;column:id"`
	Username     string `gorm:"uniqueIndex;column:username"`
	PasswordHash string `gorm:"column:password_hash"`
	Role         string `gorm:"column:role"`
}

// ServiceRecord represents a service credential record.
type ServiceRecord struct {
	ID         int64  `gorm:"primaryKey;column:id"`
	SecretHash string `gorm:"column:secret_hash"`
	Role       string `gorm:"column:role"`
}
