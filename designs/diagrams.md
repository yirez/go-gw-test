# API Gateway Service Diagrams

## System Overview (Core Traffic)

```mermaid
flowchart LR
  client[Clients] -->|login/token| auth_gw[auth_gw]
  client -->|HTTP + Bearer token| api_gw[api_gw]
  api_gw -->|proxy| users_gw[users_gw]
  api_gw -->|proxy| orders_gw[orders_gw]

  api_gw -->|token data + rate limit| redis[(Redis)]
  users_gw -->|service data| redis
  orders_gw -->|service data| redis
  auth_gw -->|issue tokens| redis

  users_gw -->|data| postgres[(Postgres)]
  orders_gw -->|data| postgres
  auth_gw -->|auth data optional| postgres
```

## Configuration and Observability

```mermaid
flowchart LR
  config[config.hcl] --> api_gw[api_gw]
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
  A->>R: Load token data
  R-->>A: token data (allowed_routes, expires_at, rate_limit)
  A-->>G: token data
  G->>G: Check rate limit + allowed_routes (mux matcher)
  G->>U: Proxy request
  U-->>G: Response
  G-->>C: Response
```

## Auth Token Issuance (Service-to-Service Example)

```mermaid
sequenceDiagram
  autonumber
  participant S as Service (users_gw)
  participant A as auth_gw
  participant R as Redis

  S->>A: Request service token (self identity)
  A->>A: Sign JWT (system clock key) - TTL 1 hour
  A->>R: Store token data
  R-->>A: OK
  A-->>S: JWT token
```

## Auth Token Issuance (Client Example)

```mermaid
sequenceDiagram
  autonumber
  participant C as Client
  participant A as auth_gw
  participant R as Redis

  C->>A: Login (credentials)
  A->>A: Sign JWT (system clock key) - TTL 1 hour
  A->>R: Store token data
  R-->>A: OK
  A-->>C: JWT token
```
