# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Install / tidy dependencies
go mod tidy

# Build
go build ./cmd/server

# Run locally (requires .env)
go run ./cmd/server

# Run all tests
go test ./...

# Run a single test
go test ./internal/service/... -run TestLogin

# Lint (requires golangci-lint)
golangci-lint run ./...

# Start postgres + redis for local dev
docker-compose up postgres redis

# Apply DB migrations
migrate -path migrations -database "$DATABASE_URL" up

# Build Docker image
docker build -t minimule-backend:latest .

# Apply all K8s manifests
kubectl apply -f k8s/
```

## Architecture

### Layer order

```
HTTP → Middleware chain → GraphQL schema → RootResolver → Service → Queries (SQL) → PostgreSQL
                                                        ↘ Redis (cache / rate-limit)
```

### Directory layout

- `cmd/server/main.go` — entry point: wires config → db → redis → jwt → queries → services → resolver → HTTP
- `graph/schema/schema.graphqls` — single GraphQL schema file loaded at runtime
- `graph/model/` — domain structs (`user.go`, `product.go`, `cart.go`, `order.go`, `models.go`)
- `graph/resolver/` — `RootResolver` + per-domain resolver types; no code generation
- `internal/config/` — env-var loading via godotenv; panics on missing required vars
- `internal/database/` — pgxpool wrapper + `queries/` (raw SQL per domain)
- `internal/cache/` — Redis client with typed helpers (SetJSON/GetJSON, rate-limit counter)
- `internal/auth/` — JWT signing/validation + bcrypt password helpers
- `internal/middleware/` — CORS, Bearer JWT extraction (fail-open), fixed-window rate limiting
- `internal/service/` — business logic layer between resolvers and queries
- `migrations/` — golang-migrate SQL files (`000001_init.up.sql` / `down.sql`)
- `k8s/` — Kubernetes manifests (namespace, configmap, secret, deployment, service, ingress, hpa)

### GraphQL library

Uses `github.com/graph-gophers/graphql-go` (reflection-based, no code generation). Resolver method names must match schema field names exactly in PascalCase. The schema is loaded from the `.graphqls` file at startup — a missing or mismatched method causes a panic.

### Auth flow

Middleware validates the `Authorization: Bearer <token>` header and injects `*AuthClaims{UserID, Email, Role}` into the request context. Fail-open: invalid/missing token produces an unauthenticated context, not a 403. Resolvers call `requireClaims(ctx)` or `requireAdmin(ctx)` to enforce access.

`login` returns a single long-lived JWT (7-day TTL by default via `JWT_REFRESH_TTL`). No separate refresh-token endpoint — designed for React Native mobile clients.

### Database

- Driver: `github.com/jackc/pgx/v5` with `pgxpool`
- Postgres connects with retry/backoff on startup (up to 10 attempts)
- All SQL is in `internal/database/queries/` — no ORM
- `ErrNotFound` / `ErrDuplicate` sentinels are mapped to service-layer errors in `internal/service/errors.go`
- Order creation runs in a transaction (`CreateOrder` in `queries/order.go`)

### Environment variables

Required: `DATABASE_URL`, `JWT_SECRET`

Optional (with defaults): `PORT` (8080), `ENV` (development), `JWT_ACCESS_TTL` (15m), `JWT_REFRESH_TTL` (168h), `RATE_LIMIT_REQUESTS` (100), `RATE_LIMIT_WINDOW` (1m), `DB_MAX_CONNS` (25), `DB_MIN_CONNS` (5), `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`, `CORS_ORIGINS`

Copy `.env.example` to `.env` for local development. Never commit `.env`.

### Kubernetes

HPA scales `minimule-api` between 2–10 replicas on CPU >60% or memory >75%. Rolling update with `maxUnavailable: 0` ensures zero-downtime deploys. TLS is managed by cert-manager via the `letsencrypt-prod` ClusterIssuer. The `k8s/secret.yaml` is an example only — use Sealed Secrets or External Secrets Operator in production.
