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

Swagger UI endpoints:
- `api_gw`: `http://localhost:8085/swagger/index.html`
- `auth_gw`: `http://localhost:8084/swagger/index.html`
- `orders_gw`: `http://localhost:8086/swagger/index.html`
- `users_gw`: `http://localhost:8087/swagger/index.html`
