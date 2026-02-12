# go-gw-test

Basic API gateway implementation in Go with `api_gw`, `auth_gw`, `users_gw`, and `orders_gw`.

## Services
- `auth_gw` (`:8084`): user/service auth, JWT mint + validate.
- `api_gw` (`:8085`): protected reverse proxy with token validation and Redis-backed per-token rate limiting.
- `orders_gw` (`:8086`): read-only sample orders/items API.
- `users_gw` (`:8087`): read-only sample users/contact API.

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
