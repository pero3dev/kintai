# Testing

## Testing Strategy

Testing is layered by change impact and risk.

```text
          +-----------------------------+
          | E2E (critical user flows)   |
          +-----------------------------+
          | Integration/branch tests    |
          +-----------------------------+
          | Unit tests (services/utils) |
          +-----------------------------+
```

## Scope Selection Rules

- Small UI label/layout change: focused frontend component tests.
- API contract or business rule change: backend service/handler/repository tests.
- Route or auth behavior change: router/middleware branch tests.
- Cross-app user flow change: add or update Playwright E2E.

## Backend Test Layers

- Service tests: use-case and business branching
- Handler tests: request binding, status code, response payload shape
- Repository tests: query behavior and DB edge conditions
- Router/middleware tests: auth, role restrictions, route accessibility

Useful examples:

- `backend/internal/router/router_branch_test.go`
- `backend/internal/middleware/middleware_branch_test.go`
- `backend/internal/service/*_test.go`

## Frontend Test Layers

- Component tests: rendering, interaction, state transitions
- Layout/navigation tests: app switching, role-dependent menus, language toggle
- Page tests: route-specific behavior and content rendering
- Store tests: persisted/auth/theme logic

Useful examples:

- `frontend/src/components/layout/Layout.test.tsx`
- `frontend/src/pages/wiki/WikiPage.test.tsx`
- `frontend/src/config/apps.test.ts`

## E2E Setup

Playwright config: `frontend/playwright.config.ts`

- base URL: `http://localhost:3000`
- web server command: `pnpm dev`
- chromium project enabled
- traces on first retry

## Load Test Setup (k6)

Scenario management: `backend/loadtest/k6`

- `scenarios/load-profile.js`: normal -> peak -> spike
- `scenarios/high-concurrency.js`: many concurrent users/sessions
- `scenarios/soak-endurance.js`: long-running endurance

## CI Pipeline Coverage

Workflow: `.github/workflows/ci.yml`

- backend lint
- backend tests with Postgres + Redis service containers
- frontend lint + type check
- frontend tests with coverage
- frontend Playwright E2E
- docker image build on `main`

## Core Commands

```sh
# frontend
cd frontend
pnpm test
pnpm test:coverage
pnpm build
pnpm test:e2e

# backend
cd backend
go test ./... -v -race -coverprofile=coverage.out -covermode=atomic

# combined shortcuts
make backend-test
make frontend-test
make frontend-e2e
make loadtest-k6-load-profile
make loadtest-k6-high-concurrency
make loadtest-k6-soak
```

## Test Authoring Guidelines

- Name tests by behavior, not implementation detail.
- Keep one assertion theme per test case when possible.
- Include both success and failure paths for business rules.
- Prefer deterministic test data and explicit timestamps.
- Mock only the boundary you need; avoid over-mocking.

## Pull Request Test Checklist

1. Run tests for touched modules locally.
2. Run build for changed frontend/backend scope.
3. Add regression tests for fixed defects.
4. Document residual risk if any paths remain untested.

## Flaky Test Triage

- Check clock/timezone assumptions.
- Check shared mutable state across tests.
- Check async waits and retry timing.
- Check external dependency leakage (network, fs, env).
- Capture Playwright traces and screenshots when failing.

## Exit Criteria

- All required CI jobs green.
- No unexplained test skip.
- New feature has at least one behavior-focused test.
- Wiki docs updated for architectural behavior changes.

