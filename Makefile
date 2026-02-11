.PHONY: help up down build logs backend-test frontend-test lint migrate

# ========================================
# Kintai - Makefile
# ========================================

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Docker Compose ----
up: ## Start local services
	docker compose up -d

down: ## Stop local services
	docker compose down

build: ## Build containers
	docker compose build

logs: ## Tail logs
	docker compose logs -f

# ---- Backend ----
backend-run: ## Run backend (hot reload)
	cd backend && air

backend-test: ## Run backend tests
	cd backend && go test ./... -v -cover

backend-lint: ## Run backend lint
	cd backend && golangci-lint run ./...

backend-swagger: ## Generate Swagger docs
	cd backend && swag init -g cmd/server/main.go -o docs

# ---- Frontend ----
frontend-dev: ## Start frontend dev server
	cd frontend && pnpm dev

frontend-build: ## Build frontend
	cd frontend && pnpm build

frontend-test: ## Run frontend tests
	cd frontend && pnpm test

frontend-lint: ## Run frontend lint
	cd frontend && pnpm lint

frontend-storybook: ## Start Storybook
	cd frontend && pnpm storybook

frontend-e2e: ## Run E2E tests
	cd frontend && pnpm test:e2e

# ---- Database ----
migrate-up: ## Apply migrations
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back one migration
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration (NAME=xxx)
	migrate create -ext sql -dir backend/migrations -seq $(NAME)

seed: ## Load development seed data
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

seed-reset: ## Reset database and reseed
	docker compose exec -T postgres psql -U kintai -d kintai -c "TRUNCATE users, departments, attendances, leave_requests, shifts, refresh_tokens CASCADE;"
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

# ---- Infrastructure ----
tf-init: ## Terraform init (dev)
	cd infrastructure/environments/dev && terraform init

tf-plan: ## Terraform plan (dev)
	cd infrastructure/environments/dev && terraform plan

tf-apply: ## Terraform apply (dev)
	cd infrastructure/environments/dev && terraform apply

# ---- Codegen ----
generate-api-client: ## Generate frontend API client from OpenAPI
	cd frontend && pnpm generate:api

generate-mocks: ## Generate Go mocks
	cd backend && go generate ./...
