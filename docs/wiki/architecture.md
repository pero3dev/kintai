# Architecture

## Architecture Style

The platform is a modular-monolith style system:

- One backend deployment unit
- Explicit domain route packages under `backend/internal/apps/*`
- Shared cross-domain routes in `backend/internal/apps/shared`
- Frontend app switching via route prefixes and shared shell

## Context Diagram

```text
                        +----------------------+
                        |      Employees       |
                        | Managers / Admins    |
                        +----------+-----------+
                                   |
                                   v
+----------------------------------------------------------------+
|                 Kintai Web Platform (Single Product)           |
|  +------------------+  +------------------+  +--------------+  |
|  | Attendance App   |  | Expenses App     |  | HR App       |  |
|  | /attendance etc. |  | /expenses/*      |  | /hr/*        |  |
|  +------------------+  +------------------+  +--------------+  |
|                        +------------------+                    |
|                        | Internal Wiki     |                   |
|                        | /wiki/*           |                   |
|                        +------------------+                    |
+----------------------------------------------------------------+
```

## Runtime Container Diagram

```text
[Browser]
   |
   | HTTP :3000
   v
[Frontend]
   |   serves UI and calls /api/v1
   v
[Backend API :8080]
   |
   +--> [PostgreSQL :5432]  persistent domain data
   +--> [Redis :6379]       caching/session-like support
   +--> [OTel Collector] -> [Elasticsearch]
   +--> [Prometheus scrape endpoint]
```

## Backend Route Ownership

Route registration is centralized in `backend/internal/router/router.go` and delegated per domain package.

- Shared public/protected routes: `backend/internal/apps/shared/routes.go`
- Attendance routes: `backend/internal/apps/attendance_routes/routes.go`
- Expense routes: `backend/internal/apps/expense_routes/routes.go`
- HR routes: `backend/internal/apps/hr_routes/routes.go`

## Request Lifecycle (Protected API)

```text
Client request
  -> Gin Engine
  -> Recovery middleware
  -> RequestLogger middleware
  -> CORS middleware
  -> SecurityHeaders middleware
  -> RateLimit middleware
  -> /api/v1 group
  -> Auth middleware (JWT)
  -> optional RequireRole middleware
  -> Handler
  -> Service
  -> Repository
  -> Database
  -> Response
```

## Frontend Route and App Model

- Route tree is declared in `frontend/src/routes.tsx`.
- Shared shell is `Layout` (`frontend/src/components/layout/Layout.tsx`).
- Active app is derived from path by `getActiveApp()` in `frontend/src/config/apps.ts`.
- App switcher UI is `AppSwitcher` and navigates to each app base path.

## Domain Boundaries and Dependency Intent

```text
[Handlers]
  - HTTP parsing, validation, response shape
  - no heavy business branching

[Services]
  - business rules, orchestration, permissions inside use-cases

[Repositories]
  - DB IO, query details, persistence mapping

[Models/DTO]
  - transport and persistence structures
```

## Auth and Role Design

- JWT is required for all `/api/v1` protected routes.
- Role-based access is enforced by middleware `RequireRole(...)`.
- Frontend stores access/refresh tokens in Zustand persisted state.
- 401 handling in frontend triggers refresh-token flow and request retry.

## Data Domains (High-Level)

```text
users
  |- attendances
  |- leave_requests
  |- overtime_requests
  |- corrections
  |- notifications
  |- expense reports/comments/history
  |- hr entities (evaluation, goals, onboarding, etc.)
```

## Key Architectural Constraints

- Keep `internal/router` orchestration-only; domain logic belongs in service layers.
- Keep app route ownership explicit to avoid hidden coupling.
- Keep docs/wiki in sync with route and package ownership.
- Prefer additive migrations and reversible schema changes.

## Known Extension Points

- Add new app domain by adding:
  1. route package in `backend/internal/apps/*_routes`
  2. service/repository/handler modules
  3. frontend route group and app metadata
  4. wiki pages and tests
