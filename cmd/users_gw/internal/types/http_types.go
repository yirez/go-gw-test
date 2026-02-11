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
