# API Gateway Service: Requirements and Core Structure

This document captures the core requirements and a proposed structure for the API Gateway/Proxy service before implementation.

## Scope and Goals

- Build a production-ready API gateway in Go 1.21+.
- Provide API key validation and token-based rate limiting.
- Proxy requests to backend services using path-based routing.
- Store token data and rate limit state in Redis.
- Provide clean structure, logging, and tests.

## Functional Requirements

### 1) Proxy Service

- Forward HTTP requests to configurable backend services.
- Support path-based routing rules (e.g., `/api/v1/users/*` -> user-service).
- Preserve and forward request/response headers correctly.
- Use middleware for cross-cutting concerns (auth, rate limit, logging, etc.).
- Provide a proxy endpoint: `/api/v1/*` and proxy all HTTP methods.

### 2) API Protection (Token Validation)

- Tokens supplied via HTTP headers (example: `Authorization: Bearer <token>`).
- On each request, validate token data fetched from Redis:
  - `api_key` (token identifier)
  - `rate_limit`
  - `expires_at` (RFC3339)
  - `allowed_routes` (path patterns)
- Reject requests for invalid tokens, expired tokens, or disallowed routes.
- Be explicit about how tokens are created/provided (static seed or helper script) in `README.md`.

### 3) Rate Limiting

- Enforce rate limits per token.
- Redis-backed counter is required (satisfies synchronization needs).
- If a more robust distributed strategy is implemented (e.g., sliding window, token bucket), highlight it.

### 4) Operational Requirements

- Environment-based configuration.
- Concurrent request handling.
- Proper error handling and logging.

## Non-Functional Requirements

- Go 1.21+.
- Redis as the single source of truth for token data and rate limit counters.
- GORM for database access in services that persist data (all except `api_gw`), including connection pool configuration (max open/idle, lifetime, idle time).
- Unit tests covering critical paths: auth, rate limit, routing, proxying.
- Dockerfile.
- Basic documentation in `README.md` or godocs.

## Token Data Model (Redis)

```
{
  "api_key": "xxx-xxx-xxx",
  "rate_limit": 100,
  "expires_at": "2024-12-31T23:59:59Z",
  "allowed_routes": ["/api/v1/users/{id}", "/api/v1/users/{id}/contact", "/api/v1/orders/{id}", "/api/v1/orders/{id}/details"]
}
```

## Proposed Core Structure (Monorepo)

- `cmd/gateway/`
  - `main.go` (service entrypoint)
  - `internal/gateway/`
    - `handler/` (HTTP handlers)
    - `middleware/` (auth, rate limit, logging, recovery)
    - `router/` (mux routes, path-based routing config)
    - `usecase/` (business rules; token validation, routing decisions)
    - `repo/` (Redis access; token store and rate limiter)
    - `proxy/` (reverse proxy logic)
    - `config/` (env config parsing)
    - `model/` (token model, config DTOs)
- `cmd/users_gw/`
  - `main.go`
  - `internal/users_gw/` (same architecture as gateway)
- `cmd/orders_gw/`
  - `main.go`
  - `internal/orders_gw/` (same architecture as gateway)
- `cmd/auth_gw/`
  - `main.go`
  - `internal/auth_gw/` (auth service: login, token issuing, validation helpers)
- `pkg/` (shared libs across services, if any)
- `api/` (proto or API schema if needed later)
- `compose/` (local docker compose for redis/postgres/prometheus)
- `designs/` (this document and future design notes)

Notes:
- Use usecase-repo structure inside `internal/gateway/`.
- Gorilla mux is the HTTP router.
- Comments are required on all functions.

## Design Considerations (Pre-Coding)

- **Routing rules**: dynamic routing from `config.yml`; rules map path patterns to backend services.
- **Gateway endpoint**: `gw_endpoint` in `config.yml` defines the public entrypoint for clients.
- **Configuration**: all services load shared settings (env/port/db) from `config.yml` via `configuration_manager`.
- **Allowed routes**: use gorilla mux path template semantics for token `allowed_routes` matching.
- **Rate limiting**: use Redis atomic operations (e.g., INCR with TTL) or Lua script for fixed window.
- **Error mapping**: consistent HTTP responses (401 invalid token, 403 disallowed, 429 rate limit).
- **Observability**: structured logs with request id; metrics endpoint for Prometheus scraping.
- **Auth service**: `auth_gw` issues and validates tokens. JWT signing key derived from system clock; TTL = 1 hour. `api_gw` delegates token validation to `auth_gw`.
- **Service-to-service tokens**: each service (except `auth_gw`) generates a token for itself; `api_gw` is the primary auth gatekeeper for incoming requests.
- **Config and logging**: all apps call `configuration_manager` `InitStandardConfigs` to load env/port, initialize a zap logger, and (when configured) return a GORM Postgres connection before starting the REST server.
- **Concurrency**: ensure per-request token fetch and rate limit checks are safe and efficient.

## Open Questions / Clarifications

- None at this time.
