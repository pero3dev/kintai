# k6 Load Test Scenarios

This directory manages API load-testing scenarios with `k6`.

## Scenarios

- `scenarios/load-profile.js`
  - Normal -> Peak -> Spike stages for `/api/v1/users/me`
- `scenarios/high-concurrency.js`
  - Many concurrent users/sessions
- `scenarios/soak-endurance.js`
  - Long-running endurance profile

## Requirements

- Backend API running (default: `http://localhost:8080`)
- PostgreSQL and Redis available for backend
- `k6` installed locally, or run via Docker image

## Run (local k6)

```bash
cd backend/loadtest/k6
K6_BASE_URL=http://localhost:8080 k6 run scenarios/load-profile.js
K6_BASE_URL=http://localhost:8080 k6 run scenarios/high-concurrency.js
K6_BASE_URL=http://localhost:8080 k6 run scenarios/soak-endurance.js
```

## Run (Docker k6)

```bash
docker run --rm -i \
  -e K6_BASE_URL=http://host.docker.internal:8080 \
  -v "$PWD/backend/loadtest/k6:/scripts" \
  grafana/k6:latest run /scripts/scenarios/load-profile.js
```

## Environment Variables

- `K6_BASE_URL`: API base URL (default: `http://localhost:8080`)
- `K6_PASSWORD`: test user password (default: `password123`)
- `K6_EMAIL_PREFIX`: test user email prefix
- `K6_VUS`: VU count (high-concurrency/soak)
- `K6_DURATION`: test duration (high-concurrency/soak)
- `K6_THINK_TIME_MS`: think time in milliseconds (load-profile/soak)

Additional stage variables for `load-profile.js`:

- `K6_STAGE_NORMAL_DURATION`, `K6_STAGE_NORMAL_VUS`
- `K6_STAGE_PEAK_DURATION`, `K6_STAGE_PEAK_VUS`
- `K6_STAGE_SPIKE_DURATION`, `K6_STAGE_SPIKE_VUS`
- `K6_STAGE_COOLDOWN_DURATION`
