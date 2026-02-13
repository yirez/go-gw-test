# API Gateway Service: Requirements and Core Structure

This document captures the core requirements and a proposed structure for the API Gateway/Proxy service before implementation.

## Scope and Goals

- Build a production-ready API gateway in Go 1.21+.
- Provide API key validation and token-based rate limiting.
- Proxy requests to backend services using path-based routing.
- Store token data and rate limit state in Redis.
- Provide clean structure, logging, and tests.

## Current Implementation Snapshot

- `api_gw` proxies `/api/v1/*` to live services based on `endpoint_configuration`.
- Token validation is delegated to `auth_gw` on every request.
- Token metadata is read from Redis in `api_gw`; if missing, it is initialized from role + endpoint config and stored with token-aligned expiry.
- Rate limiting is fixed-window per token + endpoint + second in Redis.
- `users_gw` and `orders_gw` are read-only sample backends with seeded data.

## Functional Requirements

### 1) Proxy Service

- Forward HTTP requests to configurable backend services.
- Support path-based routing rules (e.g., `/api/v1/users/*` -> user-service).
- Preserve and forward request/response headers correctly.
- Use middleware for cross-cutting concerns (auth, rate limit, logging, etc.).
- Provide a proxy endpoint: `/api/v1/*` and proxy all HTTP methods.

### 2) API Protection (Token Validation)

- Tokens supplied via HTTP headers (example: `Authorization: Bearer <token>`).
- On each request, validate token via `auth_gw`, then enforce token metadata from Redis:
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
  "api_key": "550e8400-e29b-41d4-a716-446655440000",
  "rate_limit": 100,
  "expires_at": "2024-12-31T23:59:59Z",
  "allowed_routes": ["/api/v1/users/*", "/api/v1/orders/*"]
}
```

## Implemented Structure (Monorepo)

- `cmd/api_gw`: gateway/proxy service.
- `cmd/auth_gw`: login/service token/validate service.
- `cmd/users_gw`: read-only users and contact-info service.
- `cmd/orders_gw`: read-only orders and order-items service.
- `pkg/configuration_manager`: shared config/db/redis/bootstrap utilities.
- `pkg/rest_qol`: shared HTTP helpers (server run, auth header parsing, request-id, access logging, operational routes, metrics middleware).
- `build`: full compose stack and per-env configs (`dev`, `pre`).
- `compose`: infra-only compose stack.
- `designs`: requirement and architecture notes.

## Design Considerations (Pre-Coding)

- **Routing rules**: dynamic routing from `config.yml`; rules map path patterns to backend services.
- **Gateway endpoint**: `gw_endpoint` in `config.yml` defines the public entrypoint for clients.
- **Configuration**: all services load shared settings (env/port/db) from `config.yml` via `configuration_manager`.
- **Allowed routes**: use exact or prefix wildcard patterns (e.g., `/api/v1/users/*`) for token `allowed_routes` matching.
- **Route matching**: choose the most specific matching prefix when multiple patterns match.
- **Rate limiting**: use Redis atomic operations (e.g., INCR with TTL) or Lua script for fixed window.
- **Error mapping**: consistent HTTP responses (401 invalid token, 403 disallowed, 429 rate limit).
- **Observability**: structured logs with request id; metrics endpoint for Prometheus scraping.
- **Auth service**: `auth_gw` issues and validates tokens. JWT signing key is currently derived from startup time and TTL = 1 hour. `api_gw` delegates token validation to `auth_gw`.
- **API key shape**: `auth_gw` sets JWT `jti` as a UUID and `api_gw` uses it as `api_key` in Redis (`token:{api_key}`).
- **Service-to-service tokens**: `api_gw` fetches a service token from `auth_gw` and uses it to call `/auth/validate`.
- **Token metadata bootstrap**: when token metadata is missing in Redis, `api_gw` computes allowed routes and max route rate from role/access config and writes `token:{api_key}` with aligned expiry.
- **Config and logging**: all apps call `configuration_manager` `InitStandardConfigs` to load env/port, initialize a zap logger, and (when configured) return a GORM Postgres connection before starting the REST server.
- **Concurrency**: ensure per-request token fetch and rate limit checks are safe and efficient.

## Open Questions / Clarifications

- None at this time.
