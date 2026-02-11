package types

// LoginRequest captures user login payload.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse captures user login response.
type LoginResponse struct {
	Token string `json:"token"`
}

// ServiceTokenRequest captures service-to-service login payload.
type ServiceTokenRequest struct {
	ServiceID string `json:"service_id"`
	Secret    string `json:"secret"`
}

// ServiceTokenResponse captures service token response.
type ServiceTokenResponse struct {
	Token string `json:"token"`
}

// ValidateRequest captures token validation payload.
type ValidateRequest struct {
	Token string `json:"token"`
}

// ValidateResponse captures token metadata for gateway checks.
type ValidateResponse struct {
	APIKey    string `json:"api_key"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}
