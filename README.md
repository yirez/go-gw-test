# go-gw-test

Basic API gateway implementation in Go with `api_gw`, `auth_gw`, `users_gw`, and `orders_gw`.

## Services
- `auth_gw` (`:8084`): user/service auth, JWT mint + validate.
- `api_gw` (`:8085`): protected reverse proxy with token validation and Redis-backed per-token rate limiting.
- `orders_gw` (`:8086`): read-only sample orders/items API. Simulates an unprotected live endpoint.
- `users_gw` (`:8087`): read-only sample users/contact API. Simulates an unprotected live endpoint

## Requirements Mapping
- Proxy service and path routing: `api_gw` routes by `endpoint_configuration` and proxies all methods on `/api/v1/*`.
- API protection: bearer token required; `api_gw` validates token on each request through `auth_gw /auth/validate`.
- Redis token data: token metadata stored under `token:{api_key}` with `rate_limit`, `expires_at`, `allowed_routes`.
- Rate limiting: per token + endpoint + second, key format `rl:{api_key}:{endpoint}:{unix_second}`.
- Token expiration: both JWT expiry and Redis metadata expiry are enforced; Redis key TTL is aligned to token expiry.
- Env config: all services read `config.yml` via `configuration_manager`.
- Concurrency/error/logging: middleware + repo-level checks and structured logging with zap.
- Route matching model: exact match or prefix wildcard (`/*`) with boundary-safe checks and most-specific-prefix wins.

## Token Model
`auth_gw` issues JWTs with:
- `jti`: UUID (used as `api_key` by `api_gw`)
- `role`
- `exp`

`api_gw` uses `api_key` (UUID) for Redis metadata/rate-limit keys.

## Running Locally
### Infra only
```powershell
cd compose
docker compose up
```

### Full build stack (all services + infra)
```powershell
cd build
docker compose --env-file .env.dev up --build
```

Use pre env showcase:
```powershell
cd build
docker compose --env-file .env.pre up --build
```

## Ports (host)
- `api_gw`: `8085`
- `auth_gw`: `8084`
- `orders_gw`: `8086`
- `users_gw`: `8087`
- `postgres`: `5435`
- `redis`: `6389`
- `prometheus`: `9090`

## Postman test cases
Json formatted postman requests are under `postman_tests` dir

## Tests
Run all tests:
```powershell
go test ./...
```

Current unit tests cover critical paths across gateways:
- auth middleware behavior
- route matching and role checks
- Redis token metadata read/write
- rate limiter increment semantics
- `auth_gw` login/validate/auth-middleware behavior
- `users_gw` list/get/contact handlers (success + error cases)
- `orders_gw` list/get/items handlers (success + error cases)

## k6 Rate-Limit Test
The script `tests/k6/rate_limit_per_service_per_token.js` validates:
- Per-service rate limiting for the same token (`/api/v1/users` and `/api/v1/orders` use separate counters).
- Per-token isolation for the same service (token A can be rate limited while token B is still allowed in the same second).

Run:
```powershell
k6 run tests/k6/rate_limit_per_service_per_token.js
```

Optional overrides:
```powershell
$env:API_GW_BASE_URL="http://localhost:8085"
$env:AUTH_GW_BASE_URL="http://localhost:8084"
$env:AUTH_USERNAME="user_all"
$env:AUTH_PASSWORD="123"
$env:BURST_REQUESTS="8"
k6 run tests/k6/rate_limit_per_service_per_token.js
```

## Swagger
Each service exposes Swagger UI at `/swagger/index.html`:
- `api_gw`: `http://localhost:8085/swagger/index.html`
- `auth_gw`: `http://localhost:8084/swagger/index.html`
- `orders_gw`: `http://localhost:8086/swagger/index.html`
- `users_gw`: `http://localhost:8087/swagger/index.html`

Generate docs manually:
```powershell
go install github.com/swaggo/swag/cmd/swag@v1.16.6
go generate ./cmd/api_gw
go generate ./cmd/auth_gw
go generate ./cmd/orders_gw
go generate ./cmd/users_gw
```

`build/Dockerfile` also runs `go generate ./cmd/${SERVICE}` during image builds so Swagger docs are always refreshed in compose builds.
