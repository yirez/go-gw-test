# API Gateway Service Diagrams

## System Overview (Core Traffic)

```mermaid
flowchart LR
  client[Clients] -->|login/token| auth_gw[auth_gw]
  client -->|HTTP + Bearer token| api_gw[api_gw]
  api_gw -->|proxy| users_gw[users_gw]
  api_gw -->|proxy| orders_gw[orders_gw]

  api_gw -->|token metadata + rate limit counters| redis[(Redis)]
  api_gw -->|service auth + token validate| auth_gw
  auth_gw -->|users/services credentials| postgres[(Postgres)]

  users_gw -->|data| postgres
  orders_gw -->|data| postgres
```

## Configuration and Observability

```mermaid
flowchart LR
  config[config.yml] --> api_gw[api_gw]
  config --> users_gw[users_gw]
  config --> orders_gw[orders_gw]
  config --> auth_gw[auth_gw]

  prometheus[Prometheus] -->|scrape /metrics| api_gw
  prometheus -->|scrape /metrics| users_gw
  prometheus -->|scrape /metrics| orders_gw
  prometheus -->|scrape /metrics| auth_gw
```

## Request Flow (Client -> Gateway -> Backend)

```mermaid
sequenceDiagram
  autonumber
  participant C as Client
  participant A as auth_gw
  participant G as api_gw
  participant R as Redis
  participant U as users_gw

  C->>A: Login
  A-->>C: JWT token
  C->>G: Request /api/v1/users/{id}\nAuthorization: Bearer <token>
  G->>A: Validate token (JWT)
  A-->>G: api_key + role + expires_at
  G->>R: Read token:{api_key}
  alt metadata missing
    G->>G: Build allowed_routes/rate_limit from role + endpoint config
    G->>R: Write token:{api_key} with token-aligned expiry
  end
  G->>R: Increment rl:{api_key}:{endpoint}:{unix_second}
  G->>G: Check route permission and per-second limit
  G->>U: Proxy request
  U-->>G: Response
  G-->>C: Response
```

## Auth Token Issuance (Service-to-Service Example)

```mermaid
sequenceDiagram
  autonumber
  participant G as api_gw
  participant A as auth_gw

  G->>A: Request service token (service_id + secret)
  A->>A: Sign JWT (system clock key) - TTL 1 hour
  A-->>G: Service JWT token
```

## Auth Token Issuance (Client Example)

```mermaid
sequenceDiagram
  autonumber
  participant C as Client
  participant A as auth_gw

  C->>A: Login (credentials)
  A->>A: Sign JWT (system clock key) - TTL 1 hour
  A-->>C: JWT token
```
