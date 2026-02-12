package types

// UserProfileResponse represents a user response payload.
type UserProfileResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UsersResponse represents a list response payload.
type UsersResponse struct {
	Users []UserProfileResponse `json:"users"`
}

// ContactInfoResponse represents a stored contact-info payload.
type ContactInfoResponse struct {
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	AddressLine1 string `json:"address_line1"`
	City         string `json:"city"`
	Country      string `json:"country"`
}
