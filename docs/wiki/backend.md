# Backend

## Boot Sequence

Backend starts from `backend/cmd/server/main.go`.

```text
Load config
  -> initialize logger
  -> connect PostgreSQL (GORM)
  -> development-only AutoMigrate checks
  -> construct repositories
  -> construct services (dependency injection)
  -> construct handlers
  -> construct middleware
  -> setup Gin router
  -> start HTTP server
  -> graceful shutdown on SIGINT/SIGTERM
```

## Directory Responsibilities

- `backend/cmd/server`: process entrypoint and bootstrap
- `backend/internal/router`: route orchestration only
- `backend/internal/middleware`: cross-cutting concerns (auth, rate limit, etc.)
- `backend/internal/handler`: HTTP request/response layer
- `backend/internal/service`: use-case and business logic layer
- `backend/internal/repository`: persistence and query layer
- `backend/internal/model`: entity and DTO definitions
- `backend/internal/apps/*`: domain boundary packages for modular split

## Dependency Injection Shape

Service container is `service.Services` (see `backend/internal/service/service.go`).
It aggregates domain services (attendance, expense, HR, export, etc.).

```text
Repositories --> Services --> Handlers --> Router
```

## Middleware Stack

Configured in `backend/internal/router/router.go` in this order:

1. `Recovery()`
2. `RequestLogger()`
3. `CORS()`
4. `SecurityHeaders()`
5. `RateLimit()`

Then protected routes add `Auth()`, and selected groups add `RequireRole(...)`.

## Auth Flow

Login and refresh behavior in `service.AuthService`:

- `POST /api/v1/auth/login`: verifies credentials, issues access/refresh JWT
- `POST /api/v1/auth/refresh`: validates refresh token, rotates tokens
- `POST /api/v1/auth/logout`: revokes refresh token records

Token claims include: `sub`, `email`, `role`, `exp`, `iat`.

## Core Route Domains

- Shared: auth, notifications, profile, projects, holidays, exports
- Attendance: clock in/out, attendance summary, leaves, overtime, corrections
- Expense: reports, approval flows, templates, policies, notifications
- HR: employee lifecycle, evaluations, goals, training, surveys, org chart

## Typical Feature Request Flow

```text
HTTP JSON request
  -> Handler bind/validate
  -> Service executes use-case
  -> Repository runs DB operations
  -> Service composes response model
  -> Handler returns HTTP status + payload
```

## Error and Status Conventions

- Validation errors: `400`
- Unauthenticated: `401`
- Unauthorized role: `403`
- Not found / business errors: domain-dependent `404/409/422`
- Unexpected server errors: `500`

Common error payload shape follows `model.ErrorResponse`.

## Configuration Model

From `backend/internal/config/config.go`:

- `APP_ENV`, `APP_PORT`
- `DATABASE_URL`, `REDIS_URL`
- `JWT_SECRET_KEY`, token expiry values
- `ALLOWED_ORIGINS`
- rate limit settings
- logging and observability endpoints

Production guard: default dev JWT secret is rejected.

## Data and Migration Guidelines

- Local development may run guarded AutoMigrate.
- Team-standard schema migration should use `golang-migrate` files.
- Keep migrations backward-compatible for rolling deployments.
- For destructive changes, use staged rollout and fallback path.

## Add-New-Endpoint Playbook

1. Add repository method and tests.
2. Add service method and business tests.
3. Add handler endpoint and request/response tests.
4. Register route in domain route package.
5. Add frontend API client method and UI/test updates.
6. Update wiki docs and route matrix.

## Operational Commands

```sh
# backend unit/integration style tests
cd backend
go test ./... -v

# lint
cd backend
golangci-lint run ./...

# generate swagger
cd backend
swag init -g cmd/server/main.go -o docs
```

## Troubleshooting Checklist

- 401 everywhere: verify JWT secret mismatch and token expiration.
- CORS issues: verify `ALLOWED_ORIGINS` and frontend origin.
- Slow endpoints: inspect query paths in repository layer first.
- Startup DB failure: verify `DATABASE_URL`, container health, network.
