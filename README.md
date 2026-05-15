# miniMule Backend

GraphQL API backend for **miniMule** — an art & craft marketplace where independent artists sell sticker designs, custom artwork, and handmade goods. Built for a React Native mobile app.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22+ |
| API | GraphQL (`graph-gophers/graphql-go`) |
| Database | PostgreSQL 15 (pgx v5 / pgxpool) |
| Cache & Rate Limiting | Redis |
| Auth | JWT (HS256), bcrypt |
| Container | Docker (multi-stage Alpine build) |
| Orchestration | Kubernetes + HPA |

## Features

- **Auth** — register, login, JWT-based auth (single long-lived token for mobile)
- **Products** — listing, filtering, categories, images, variants, reviews, ratings
- **Cart** — add/update/remove items, subtotal calculation
- **Orders** — checkout, order history, shipping methods, shipment tracking
- **Social feed** — posts, likes, comments, votes, showcase tab
- **Search** — product search with history
- **Promotions** — promo code validation
- **Notifications** — in-app notifications
- **User profiles** — addresses, payment methods, PDPA consent, preferences
- **Rate limiting** — fixed-window per IP via Redis
- **Admin** — product creation (artist/admin roles)

## Getting Started

### Prerequisites

- Go 1.22+
- Docker Desktop (for Postgres + Redis)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI

### Local setup

```bash
# 1. Clone and install dependencies
git clone https://github.com/jakka/minimule-backend
cd minimule-backend
go mod tidy

# 2. Configure environment
cp .env.example .env
# Edit .env and set DATABASE_URL and JWT_SECRET

# 3. Start Postgres and Redis
docker-compose up -d postgres redis

# 4. Run database migrations
migrate -path migrations -database "$DATABASE_URL" up

# 5. (Optional) Load seed data for development
psql "$DATABASE_URL" -f migrations/seed.sql

# 6. Start the server
go run ./cmd/server
# → http://localhost:8080/graphql
# → http://localhost:8080/playground  (GraphQL explorer)
```

### Run tests

```bash
# Integration tests (requires server running + seed data)
go test ./tests/... -v -timeout 60s

# Unit tests
go test ./internal/...
```

## Project Structure

```
cmd/server/          Entry point — wires all dependencies
graph/
  schema/            schema.graphqls — single GraphQL schema file
  model/             Domain structs (User, Product, Cart, Order, Post, …)
  resolver/          RootResolver + per-domain resolvers
internal/
  auth/              JWT signing/validation, bcrypt helpers
  cache/             Redis client (SetJSON/GetJSON, rate-limit counter)
  config/            Env-var loading via godotenv
  database/
    postgres.go      pgxpool connect with retry/backoff
    queries/         Raw SQL per domain (no ORM)
  middleware/        CORS, Bearer JWT extraction, rate limiting
  service/           Business logic layer
migrations/          golang-migrate SQL files (up + down)
k8s/                 Kubernetes manifests (deployment, service, ingress, HPA)
reports/             Test reports and deployment guides
```

## API

The single endpoint is `POST /graphql`. A GraphQL playground is available at `/playground` when `ENV=development`.

Example — login:
```graphql
mutation {
  login(email: "customer@minimule.com", password: "password123")
}
```

Example — browse products:
```graphql
{
  products(limit: 10) {
    id name basePrice
    images { url }
    category { name }
  }
}
```

Example — add to cart (requires `Authorization: Bearer <token>`):
```graphql
mutation {
  addToCart(productId: "...", quantity: 1) {
    id status subtotal
    items { quantity unitPrice product { name } }
  }
}
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | Postgres connection string |
| `JWT_SECRET` | ✅ | — | HMAC-SHA256 signing key |
| `REDIS_ADDR` | — | — | Redis host:port |
| `PORT` | — | `8080` | HTTP listen port |
| `ENV` | — | `development` | Set `production` to disable playground |
| `JWT_REFRESH_TTL` | — | `168h` | Token lifetime (7 days for mobile) |
| `RATE_LIMIT_REQUESTS` | — | `100` | Max requests per window |

See `.env.example` for the full list.

## Deployment

This service is a persistent Go process with a connection pool — **not compatible with Vercel serverless**. Deploy to Railway or Fly.io:

```bash
# Railway
railway login
railway up
```

See [`reports/railway-deploy.md`](reports/railway-deploy.md) for the full step-by-step guide including migrations and environment variable setup.

For Kubernetes (production), apply the manifests in `k8s/`. The HPA scales between 2–10 replicas on CPU > 60% or memory > 75%.

## Test Results

34/34 integration tests passing. See [`reports/integration-test-report.md`](reports/integration-test-report.md).
