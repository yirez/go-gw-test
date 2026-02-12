package repo

import (
	"net/http/httptest"
	"testing"

	"go-gw-test/cmd/api_gw/internal/types"
)

// TestGatewayRepoMatchRoutePrefersSpecific verifies the longest matching route wins.
func TestGatewayRepoMatchRoutePrefersSpecific(t *testing.T) {
	gwRepo := NewGatewayRepo()
	routes, err := gwRepo.BuildRouteEntries([]types.EndpointConfig{
		{GwEndpoint: "/api/v1/users/*", LiveEndpoint: "http://users:8087", LiveTimeoutSec: 10},
		{GwEndpoint: "/api/v1/users/1/*", LiveEndpoint: "http://users:8087", LiveTimeoutSec: 10},
	})
	if err != nil {
		t.Fatalf("build route entries: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/users/1/contact", nil)
	entry, ok := gwRepo.MatchRoute(routes, req)
	if !ok {
		t.Fatalf("expected route match")
	}
	if entry.Config.GwEndpoint != "/api/v1/users/1/*" {
		t.Fatalf("expected most specific route, got %s", entry.Config.GwEndpoint)
	}
}

// TestGatewayRepoAllowedRouteWildcard verifies wildcard route checks.
func TestGatewayRepoAllowedRouteWildcard(t *testing.T) {
	gwRepo := NewGatewayRepo()
	req := httptest.NewRequest("GET", "/api/v1/orders/1/items", nil)

	ok := gwRepo.IsAllowedRoute([]string{"/api/v1/orders/*"}, req)
	if !ok {
		t.Fatalf("expected wildcard route to match")
	}
}

// TestGatewayRepoIsRoleAllowed verifies endpoint role authorization behavior.
func TestGatewayRepoIsRoleAllowed(t *testing.T) {
	gwRepo := NewGatewayRepo()
	if !gwRepo.IsRoleAllowed([]string{"user_all", "user_users"}, "user_users") {
		t.Fatalf("expected role to be allowed")
	}
	if gwRepo.IsRoleAllowed([]string{"user_users"}, "user_orders") {
		t.Fatalf("expected role to be denied")
	}
	if !gwRepo.IsRoleAllowed(nil, "user_any") {
		t.Fatalf("expected empty role list to allow all roles")
	}
}
