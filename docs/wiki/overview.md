# Overview

## Purpose

This internal wiki is the single source of truth for technical implementation and operations of the Kintai platform.
The goal is to let engineers move from "I do not know this area" to "I can safely modify it" in one reading session.

## Who Should Read This

- New engineers onboarding to the repository
- Engineers implementing features or fixes
- Reviewers validating design and risk
- Operators troubleshooting production-like issues

## Reading Paths

- New to the repository:
  1. `overview.md`
  2. `architecture.md`
  3. `backend.md`
  4. `frontend.md`
  5. `infrastructure.md`
  6. `testing.md`
- Adding backend API: `architecture.md` -> `backend.md` -> `testing.md`
- Adding frontend page/app: `architecture.md` -> `frontend.md` -> `testing.md`
- Debugging local environment: `infrastructure.md` -> `testing.md`

## Product and Domain Scope

Current app domains in one product shell:

- Attendance app (`/`, `/attendance`, `/leaves`, `/overtime`, `/corrections`, etc.)
- Expenses app (`/expenses/*`)
- HR app (`/hr/*`)
- Internal Wiki app (`/wiki/*`)

## System Snapshot

```text
[Browser]
   |
   | http://localhost:3000
   v
[Frontend (Vite/Nginx)]
   |
   | /api/v1/*
   v
[Backend (Gin)] -----> [PostgreSQL]
   |
   +---------------> [Redis]
   |
   +---------------> [Prometheus / OTel / ELK]
```

## Repository Map

```text
kintai/
|- backend/                 Go API, domain logic, repository layer
|- frontend/                React app, routing, UI, app switching
|- docs/wiki/               English technical docs loaded by /wiki
|- docs/wiki/ja/            Japanese technical docs loaded by /wiki
|- infrastructure/          Terraform for AWS resources
|- monitoring/              Prometheus, OTel collector, Logstash config
|- docker-compose.yml       Local multi-service environment
|- Makefile                 Common development commands
```

## Technology Baseline

- Backend: Go 1.23, Gin, GORM, PostgreSQL, Redis, JWT
- Frontend: React 19, TanStack Router, TanStack Query, Zustand, i18next
- Infra: Docker Compose (local), Terraform + AWS modules (cloud)
- Observability: Prometheus, OpenTelemetry Collector, ELK stack
- CI: GitHub Actions (`.github/workflows/ci.yml`)

## Typical End-to-End Change Flow

1. Identify affected app and route ownership.
2. Confirm API contract and authorization role.
3. Implement backend changes (handler/service/repository).
4. Implement frontend changes (page/component/api client).
5. Add or update tests by changed scope.
6. Validate with local build and tests.
7. Open PR with risk notes and rollback idea.

## Definition of Done (Engineering)

- Feature behavior implemented and manually verified
- Existing tests still pass
- New behavior covered by tests
- No obvious security regression (auth, role checks, CORS)
- Wiki docs updated in the same PR for architecture-impacting changes

## Documentation Update Rules

- Keep implementation facts aligned with code paths.
- Prefer concrete file references over abstract statements.
- Include at least one diagram for cross-component topics.
- If behavior depends on role or environment, state it explicitly.
