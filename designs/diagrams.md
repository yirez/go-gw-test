# API Gateway Service Diagrams

## System Overview

```mermaid
flowchart LR
  client[Clients] -->|HTTP + Bearer token| api_gw[api_gw]
  api_gw -->|proxy| users_gw[users_gw]
  api_gw -->|proxy| orders_gw[orders_gw]
  client -->|login/token| auth_gw[auth_gw]

  api_gw -->|token data + rate limit| redis[(Redis)]
  users_gw -->|token validation + data| redis
  orders_gw -->|token validation + data| redis
  auth_gw -->|issue tokens| redis

  users_gw -->|data| postgres[(Postgres)]
  orders_gw -->|data| postgres
  auth_gw -->|auth data (optional)| postgres

  prometheus[Prometheus] -->|scrape /metrics| api_gw
  prometheus -->|scrape /metrics| users_gw
  prometheus -->|scrape /metrics| orders_gw
  prometheus -->|scrape /metrics| auth_gw

  config[config.hcl] --> api_gw
```

## Request Flow (Client -> Gateway -> Backend)

```mermaid
sequenceDiagram
  autonumber
  participant C as Client
  participant G as api_gw
  participant R as Redis
  participant U as users_gw

  C->>G: Request /api/v1/users/{id}\nAuthorization: Bearer <token>
  G->>R: Load token data + check rate limit
  R-->>G: token data (allowed_routes, expires_at, rate_limit)
  G->>G: Validate token + allowed_routes (mux matcher)
  G->>U: Proxy request
  U->>R: Validate token data (per-service interceptor)
  R-->>U: token data
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
  A->>A: Sign JWT (system clock key)\nTTL 1 hour
  A->>R: Store token data
  R-->>A: OK
  A-->>S: JWT token
```

