package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// NewRequestID returns a random request identifier.
func NewRequestID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
