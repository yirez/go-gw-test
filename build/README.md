# Build Stack

Run all services and infrastructure from the `build` folder:

```powershell
cd build
docker compose --env-file .env.dev up --build
```

Switch environment showcase (`dev` vs `pre`):

```powershell
cd build
docker compose --env-file .env.pre up --build
```

Config folders:
- `build/configs/dev`
- `build/configs/pre`

Endpoints:
- `api_gw`: `http://localhost:8085`
- `auth_gw`: `http://localhost:8084`
- `orders_gw`: `http://localhost:8086`
- `users_gw`: `http://localhost:8087`
- `postgres`: `localhost:5435`
- `redis`: `localhost:6389`
- `prometheus`: `http://localhost:9090`

Credential seeding and source of truth:
- On fresh Postgres volume init, `build/init.sql` seeds auth credentials into `auth.user_records` and `auth.service_records`.
- Postgres records are treated as the source of truth for login/service credentials in this project.
- Seeded test users: `user_all/123`, `user_users/123`, `user_orders/123`.
- Seeded test services: id `1`/secret `123`, id `2`/secret `123`, id `3`/secret `123`.

Metrics endpoints:
- `api_gw`: `http://localhost:8085/metrics`
- `auth_gw`: `http://localhost:8084/metrics`
- `orders_gw`: `http://localhost:8086/metrics`
- `users_gw`: `http://localhost:8087/metrics`

Swagger UI endpoints:
- `api_gw`: `http://localhost:8085/swagger/index.html`
- `auth_gw`: `http://localhost:8084/swagger/index.html`
- `orders_gw`: `http://localhost:8086/swagger/index.html`
- `users_gw`: `http://localhost:8087/swagger/index.html`
