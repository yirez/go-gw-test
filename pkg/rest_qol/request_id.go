package rest_qol

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

const headerXRequestID = "X-Request-Id"

// EnsureRequestID sets request/response request-id headers and returns the effective ID.
// Prefix is only applied when the incoming request does not already provide X-Request-Id.
func EnsureRequestID(w http.ResponseWriter, r *http.Request, fallbackPrefix string) string {
	requestID := r.Header.Get(headerXRequestID)
	if requestID == "" {
		requestID = newRequestID(fallbackPrefix)
		r.Header.Set(headerXRequestID, requestID)
	}

	w.Header().Set(headerXRequestID, requestID)
	return requestID
}

func newRequestID(prefix string) string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	if prefix == "" {
		return hex.EncodeToString(buf)
	}

	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(buf))
}
