# Frontend

## Runtime Shell

Frontend is a single React application with app-style route domains.
`Layout` provides shared navigation, app switcher, search UI, and top/bottom navigation behavior.

## Route Topology

Defined in `frontend/src/routes.tsx`.

```text
/                          home dashboard shell
/attendance, /leaves...    attendance domain pages
/expenses/*                expense domain pages
/hr/*                      HR domain pages
/wiki/*                    internal wiki pages
/login                     authentication page
```

## App Switching Model

`frontend/src/config/apps.ts` contains app metadata and route prefixes.
`getActiveApp(pathname)` resolves the current app by longest matching prefix.

```text
pathname -> getActiveApp -> Layout nav items -> AppSwitcher state
```

## Layout Responsibilities

`frontend/src/components/layout/Layout.tsx` handles:

- desktop + mobile navigation structure
- app-specific nav item rendering
- language toggle (`ja` <-> `en`)
- theme cycle (`system` -> `light` -> `dark`)
- logout and route transitions

## Internal Wiki Rendering Model

`frontend/src/pages/wiki/WikiPage.tsx`:

- loads markdown from `docs/wiki/*.md` (en)
- loads markdown from `docs/wiki/ja/*.md` (ja)
- picks locale from `i18n.language`
- parses headings, paragraphs, lists, code blocks
- renders in a card-based documentation UI

## API Client Architecture

`frontend/src/api/client.ts` provides two layers:

1. `openapi-fetch` client with request/response interceptors
2. explicit helper methods grouped by domain (`api.attendance`, `api.expenses`, `api.hr`, ...)

### Token Refresh Sequence

```text
Request with access token
  -> backend returns 401
  -> client executes /auth/refresh with refresh token
  -> on success: store new token pair
  -> retry original request once
  -> if refresh fails: logout + redirect /login
```

## State Management

- Auth state: `frontend/src/stores/authStore.ts` (Zustand + persist)
- Theme state: `frontend/src/stores/themeStore.ts`
- Server state: TanStack Query
- Route state: TanStack Router

## I18n Strategy

- i18n initialization: `frontend/src/i18n/index.ts`
- locale dictionaries: `frontend/src/i18n/locales/ja.json`, `en.json`
- default language: Japanese (`ja`)
- wiki docs are language-specific markdown sources

## Build and Chunking

Vite config (`frontend/vite.config.ts`) currently:

- dev server port: `3000`
- `/api` proxy to backend `http://localhost:8080`
- manual chunking for vendor/router/query/ui/i18n bundles

## Add-New-Page Playbook

1. Create page component under `frontend/src/pages/...`.
2. Add route entry in `frontend/src/routes.tsx`.
3. Add nav item in `Layout` if needed.
4. Add API client methods and query/mutation hooks.
5. Add tests (component + route behavior).
6. Update wiki docs where architecture changed.

## Add-New-App Playbook

1. Add app metadata in `frontend/src/config/apps.ts`.
2. Add route group in `routes.tsx`.
3. Add app nav logic in `Layout.tsx`.
4. Add pages and app-specific state/API calls.
5. Add tests for app switcher and app routes.

## Frontend Quality Checklist

- Route paths and app prefix ownership are explicit.
- All authenticated API calls include token flow coverage.
- Language toggle does not break labels or wiki content.
- Mobile drawer and desktop sidebar both remain navigable.
- Build succeeds without type errors.

## Useful Commands

```sh
cd frontend
npm run dev
npm run test
npm run build
npm run test:e2e
```
