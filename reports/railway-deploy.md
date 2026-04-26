# Railway Deployment Guide — miniMule Backend

> **Why Railway and not Vercel?**  
> This service uses `pgxpool` (a persistent connection pool) and Redis. Vercel runs serverless functions that terminate after each request — a persistent connection pool can't survive there. Railway runs your Docker container as a long-lived process, which is the correct model.

---

## Prerequisites

- [Railway CLI](https://docs.railway.app/develop/cli): `npm install -g @railway/cli`
- Railway account (free tier is fine for development)
- A PostgreSQL addon and Redis addon in your Railway project

---

## One-time Setup

### 1. Login and create project

```bash
railway login
railway init          # creates a new project, link it to this repo
```

### 2. Add PostgreSQL and Redis plugins

In the Railway dashboard → your project → **New** → **Database** → add:
- **PostgreSQL**
- **Redis**

Railway auto-injects `DATABASE_URL` and `REDIS_URL` into your service's environment.

### 3. Set required environment variables

In Railway dashboard → your service → **Variables**, set:

| Variable | Value |
|---|---|
| `JWT_SECRET` | a random 32+ char string |
| `ENV` | `production` |
| `PORT` | `8080` |

Railway injects `DATABASE_URL` and `REDIS_URL` automatically from the linked plugins.

Optional (Railway provides `REDIS_URL` but the app reads individual vars):

| Variable | Value |
|---|---|
| `REDIS_ADDR` | from `$REDIS_URL` (host:port) |
| `REDIS_PASSWORD` | from `$REDIS_URL` |

> **Tip:** Add a startup script or update `config/config.go` to parse `REDIS_URL` directly if you prefer using Railway's injected variable.

### 4. Run migrations

After first deploy:

```bash
railway run migrate -path migrations -database "$DATABASE_URL" up
```

Or connect directly and run the migration SQL:

```bash
railway connect postgres
\i migrations/000001_init.up.sql
\i migrations/seed.sql   # optional: seed data for testing
```

---

## Deploy

```bash
# From repo root
railway up
```

Railway detects `Dockerfile` (or `railway.toml` `builder = "DOCKERFILE"`) and builds automatically.

### Check logs

```bash
railway logs
```

### Open the service

```bash
railway open
```

The GraphQL playground will be available at `https://<your-service>.railway.app/playground` (only in `ENV=development`; disable in production by setting `ENV=production`).

---

## Environment Variables Reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | Postgres connection string (injected by Railway) |
| `JWT_SECRET` | ✅ | — | HMAC-SHA256 signing key |
| `REDIS_ADDR` | — | — | Redis host:port |
| `REDIS_PASSWORD` | — | — | Redis auth password |
| `REDIS_DB` | — | `0` | Redis DB index |
| `PORT` | — | `8080` | HTTP listen port |
| `ENV` | — | `development` | Set to `production` to disable playground |
| `JWT_ACCESS_TTL` | — | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | — | `168h` | Long-lived token lifetime (mobile) |
| `RATE_LIMIT_REQUESTS` | — | `100` | Max requests per window |
| `RATE_LIMIT_WINDOW` | — | `1m` | Rate-limit window duration |
| `DB_MAX_CONNS` | — | `25` | pgxpool max connections |
| `DB_MIN_CONNS` | — | `5` | pgxpool min connections |
| `CORS_ORIGINS` | — | `*` | Comma-separated allowed origins |

---

## Alternative: Fly.io

If you prefer Fly.io, create `fly.toml`:

```toml
app = "minimule-backend"
primary_region = "sin"   # Singapore

[build]

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  memory = "512mb"
  cpu_kind = "shared"
  cpus = 1
```

Then:
```bash
fly launch --no-deploy
fly postgres create --name minimule-db
fly postgres attach minimule-db
fly secrets set JWT_SECRET=<your-secret>
fly deploy
fly postgres connect -a minimule-db  # run migrations
```
