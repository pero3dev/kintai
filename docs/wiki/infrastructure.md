# Infrastructure

## Local Environment Topology

Primary local environment is Docker Compose in repository root.

```text
+----------------------+        +----------------------+
| frontend (nginx)     |        | backend (gin)        |
| localhost:3000       | -----> | localhost:8080       |
+----------------------+        +----------+-----------+
                                           |
                        +------------------+------------------+
                        |                                     |
                        v                                     v
                +---------------+                     +---------------+
                | postgres:16   |                     | redis:7       |
                | localhost:5433|                     | localhost:6379|
                +---------------+                     +---------------+
```

Optional monitoring profile adds Elasticsearch, Logstash, Kibana, Prometheus, Grafana, and OTel collector.

## Local Service Endpoints

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`
- Postgres (host): `localhost:5433`
- Redis (host): `localhost:6379`
- Kibana: `http://localhost:5601`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001`

## Compose and Build Notes

- Frontend Docker build context is repository root (`.`) so wiki markdown under `docs/wiki` can be included in image build.
- Frontend Dockerfile copies `frontend/` app sources and `docs/wiki` markdown assets.
- Backend container receives DB/Redis/OTel env vars from compose.

## Make Targets (Frequent)

```sh
make up            # start compose services
make down          # stop compose
make build         # compose build
make logs          # stream logs
make backend-test
make frontend-test
```

## Terraform Architecture (AWS)

Terraform root for dev: `infrastructure/environments/dev`.

```text
env/dev resources.tf
  -> module.vpc
  -> module.ecr
  -> module.alb
  -> module.ecs
  -> module.rds
  -> module.elasticache
  -> module.cdn
```

### AWS Module Roles

- `vpc`: VPC, public/private subnets, IGW, NAT, route tables
- `alb`: public ALB, backend/frontend target groups, listener rules
- `ecs`: cluster, task defs, services, CloudWatch logs, security groups
- `rds`: PostgreSQL instance + subnet group + SG
- `elasticache`: Redis network placement and SG
- `ecr`: backend/frontend image repositories
- `cdn`: CloudFront + S3 integration path to ALB frontend

## Runtime Data and Network Flow (Cloud)

```text
User
  -> CloudFront
  -> ALB
     -> ECS frontend task
     -> ECS backend task
        -> RDS PostgreSQL
        -> ElastiCache Redis
```

## Environment Variables and Secrets

Reference template: `.env.example`.

Critical values to treat as secrets in non-local env:

- `JWT_SECRET_KEY`
- `DATABASE_URL` (or db credentials)
- `AWS_*` credentials
- `SENTRY_DSN`

## Observability Pipeline

```text
Backend metrics/logs/traces
  -> OTel Collector (4317/4318)
     -> Prometheus exporter
     -> Elasticsearch exporter
     -> logging exporter
```

Prometheus scrapes backend metrics and collector metrics.
ELK stack ingests structured logs via Logstash pipeline.

## Operations Runbook (Local)

1. `make up`
2. verify `docker compose ps`
3. open `http://localhost:3000`
4. if schema mismatch, run migration/seed targets
5. monitor with `make logs`

## Failure Triage Checklist

- Frontend blank page: verify frontend container healthy and built assets present.
- API 502/timeout: inspect backend logs and DB connectivity.
- Login failing: verify JWT secret consistency and refresh flow.
- Slow responses: inspect DB, Redis, and query-heavy endpoints.
- Missing metrics/logs: verify Prometheus targets and OTel collector endpoint.
