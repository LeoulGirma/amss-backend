# AMSS Backend

AMSS (Aircraft Maintenance Scheduling System) is the backend for planning, tracking, and auditing aircraft maintenance work. It exposes REST and gRPC APIs for managing organizations, users, aircraft, maintenance programs, tasks, parts inventory, compliance, imports, webhooks, and reports. Background workers handle async jobs such as program task generation, CSV import processing, and webhook delivery via an outbox pattern.

## What it does
- Auth and RBAC using JWT access/refresh tokens.
- Maintenance planning: programs generate tasks, tasks track execution state.
- Parts inventory: definitions, items, and task reservations.
- Compliance tracking and audit logs for traceability.
- CSV imports for aircraft, parts, and programs.
- Webhook notifications via outbox + delivery retries.
- Reports endpoints for operational summaries.

## Architecture at a glance
- amss-server: REST + gRPC APIs, middleware (auth, idempotency, rate limit), OpenAPI docs, metrics.
- amss-worker: outbox publisher, webhook dispatcher, import processor, program task generator.
- Postgres: primary system of record.
- Redis: rate limiting, streams for async jobs.
- Optional: OTLP tracing and Prometheus scraping.
- Diagram sources: docs/architecture-diagrams.md.

## Requirements
- Go 1.22+
- Docker + Docker Compose
- Goose (for migrations) or run migrations via the smoke test script.

## Quick start (fastest)
Run the smoke test script. It starts Postgres/Redis, runs migrations and seed data, starts the server, and hits a few endpoints.

```
powershell -ExecutionPolicy Bypass -File scripts/smoke_test.ps1
```

Use `-KeepRunning` to leave the services running afterward.

## Run with Docker Compose (local dev)
1. Edit `docker-compose.yml` and replace `JWT_PRIVATE_KEY_PEM` and `JWT_PUBLIC_KEY_PEM` with real RSA keys.
2. Start services:
   ```
   make dev-up
   ```
3. Run migrations:
   ```
   make migrate-up
   ```
4. (Optional) Seed demo data:
   ```
   make seed
   ```

## Run locally (Go)
Set env vars and run the server and worker in separate terminals:

```
$env:DB_URL = "postgres://amss:amss@localhost:5455/amss?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
$env:JWT_PRIVATE_KEY_PEM = "<your private key>"
$env:JWT_PUBLIC_KEY_PEM = "<your public key>"
$env:IMPORT_STORAGE_DIR = "./data/imports"
$env:APP_ENV = "development"
$env:PROMETHEUS_ENABLED = "true"

# Terminal 1

go run ./cmd/server

# Terminal 2

go run ./cmd/worker
```

## Environment variables
Required:
- `DB_URL` (Postgres connection string)
- `JWT_PRIVATE_KEY_PEM` (RSA private key PEM)
- `JWT_PUBLIC_KEY_PEM` (RSA public key PEM)

Optional:
- `APP_ENV` (default `development`)
- `HTTP_ADDR` (default `:8080`)
- `GRPC_ADDR` (default `:9090`)
- `REDIS_ADDR` (default `localhost:6379`)
- `IMPORT_STORAGE_DIR` (default `./data/imports`)
- `ACCESS_TOKEN_TTL` (default `15m`)
- `REFRESH_TOKEN_TTL` (default `168h`)
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `PROMETHEUS_ENABLED` (default `false`)
- `LOG_LEVEL` (default `info`)
- `CORS_ALLOWED_ORIGINS` (comma-separated)
- `WORKER_ID` (default `worker-1`)
- `RETENTION_CLEANUP_INTERVAL` (default `24h`)

Production templates:
- `deploy/.env.production.server.example`
- `deploy/.env.production.worker.example`

Secrets handling:
- Store JWT keys and DB credentials in a secret manager or CI/CD secret store.
- Do not commit real PEM values to git; inject them at deploy time.

## Seed data
Run `make seed` to apply the seed migration with a demo org and users.
Default password for all seeded users: `ChangeMe123!`.

## Using the API
- OpenAPI spec: `GET /openapi.yaml`
- Swagger UI: `GET /docs`

Example login:
```
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"org_id":"<org-uuid>","email":"admin@demo.local","password":"ChangeMe123!"}'
```

Use the returned access token:
```
curl -s http://localhost:8080/api/v1/organizations \
  -H "Authorization: Bearer <access_token>"
```

Notes:
- Most write endpoints support idempotency via the `Idempotency-Key` header.
- Rate limiting is enforced per org and on auth attempts.

## gRPC
- Proto: `api/proto/amss.proto`
- Default address: `:9090`

## Health and metrics
- `GET /health`
- `GET /ready`
- `GET /metrics` (when `PROMETHEUS_ENABLED=true`)

## Tests
Unit tests:
```
go test ./...
```
Includes a metrics smoke test that hits `/health` and `/metrics` and verifies key label sets.

Integration tests (use testcontainers by default, gated by `AMSS_INTEGRATION=1`):
```
AMSS_INTEGRATION=1 go test ./internal/infra/postgres -run TestPostgres
```
To use an existing Postgres instance instead of testcontainers, set `AMSS_TEST_DB_URL`.
To use an existing Redis instance for Redis integration tests, set `AMSS_TEST_REDIS_ADDR`.
Alternatively:
```
make test-integration
```

## 📚 Documentation

**New to AMSS?** Start here:

- **[Developer Onboarding Guide](docs/DEVELOPER_GUIDE.md)** - Complete system overview, architecture deep-dive, user journeys, and development workflows
- **[API & User Guide](docs/API_GUIDE.md)** - REST API reference, authentication, workflows, webhooks, and best practices
- **[Failure Modes & Recovery](docs/FAILURE_MODES.md)** - What happens when things fail, recovery procedures, backup/restore, disaster recovery
- **[Architecture Diagrams](docs/architecture-diagrams.md)** - System architecture and component interactions

## Repo layout (high level)
- `cmd/` entrypoints (`server`, `worker`)
- `internal/api/` REST + gRPC handlers and middleware
- `internal/app/` services and ports
- `internal/infra/` Postgres/Redis implementations
- `internal/jobs/` async workers
- `migrations/` SQL migrations
- `api/proto/` gRPC proto and generated code
- `docs/` comprehensive guides, diagrams, and design notes

## Deployment (Helm + GitHub Actions)
Helm chart and environment values:
- `deploy/helm/amss`
- `deploy/helm/values-uat.yaml`
- `deploy/helm/values-prod.yaml`

CI/CD:
- `.github/workflows/deploy.yml` builds/pushes Docker images and deploys to `amss-uat` on `main`.  
- Production deploys require a manual workflow dispatch to `prod` and environment approval.

Required GitHub Actions secrets:
- `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`
- `KUBE_CONFIG_DATA` (base64 kubeconfig)
- `UAT_DB_URL`, `UAT_JWT_PRIVATE_KEY_PEM`, `UAT_JWT_PUBLIC_KEY_PEM`
- `PROD_DB_URL`, `PROD_JWT_PRIVATE_KEY_PEM`, `PROD_JWT_PUBLIC_KEY_PEM`
