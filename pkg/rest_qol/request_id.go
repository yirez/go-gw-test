package rest_qol

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const headerXRequestID = "X-Request-Id"

// EnsureRequestID sets request/response request-id headers and returns the effective ID.
func EnsureRequestID(w http.ResponseWriter, r *http.Request) string {
	requestID := r.Header.Get(headerXRequestID)
	if requestID == "" {
		requestID = newRequestID()
		r.Header.Set(headerXRequestID, requestID)
	}

	w.Header().Set(headerXRequestID, requestID)
	return requestID
}

func newRequestID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
