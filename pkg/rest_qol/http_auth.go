package rest_qol

import (
	"fmt"
	"net/http"
	"strings"
)

// BearerTokenFromRequest extracts the Bearer token from Authorization header.
func BearerTokenFromRequest(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("request is nil")
	}

	return BearerTokenFromHeader(r.Header.Get("Authorization"))
}

// BearerTokenFromHeader extracts the Bearer token from an Authorization header value.
func BearerTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", fmt.Errorf("empty bearer token")
	}

	return token, nil
}
