.PHONY: help up down build logs backend-test frontend-test lint migrate

# ========================================
# 蜍､諤邂｡逅・す繧ｹ繝・Β - Makefile
# ========================================

help: ## 繝倥Ν繝励ｒ陦ｨ遉ｺ
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Docker Compose ----
up: ## 髢狗匱迺ｰ蠅・ｒ襍ｷ蜍・
	docker compose up -d

down: ## 髢狗匱迺ｰ蠅・ｒ蛛懈ｭ｢
	docker compose down

build: ## 繧ｳ繝ｳ繝・リ繧偵ン繝ｫ繝・
	docker compose build

logs: ## 繝ｭ繧ｰ繧定｡ｨ遉ｺ
	docker compose logs -f

# ---- Backend ----
backend-run: ## 繝舌ャ繧ｯ繧ｨ繝ｳ繝峨ｒ襍ｷ蜍・(繝帙ャ繝医Μ繝ｭ繝ｼ繝・
	cd backend && air

backend-test: ## 繝舌ャ繧ｯ繧ｨ繝ｳ繝峨・繝・せ繝医ｒ螳溯｡・
	cd backend && go test ./... -v -cover

backend-lint: ## 繝舌ャ繧ｯ繧ｨ繝ｳ繝峨・Lint繧貞ｮ溯｡・
	cd backend && golangci-lint run ./...

backend-swagger: ## Swagger繝峨く繝･繝｡繝ｳ繝医ｒ逕滓・
	cd backend && swag init -g cmd/server/main.go -o docs

# ---- Frontend ----
frontend-dev: ## 繝輔Ο繝ｳ繝医お繝ｳ繝蛾幕逋ｺ繧ｵ繝ｼ繝舌・繧定ｵｷ蜍・
	cd frontend && pnpm dev

frontend-build: ## 繝輔Ο繝ｳ繝医お繝ｳ繝峨ｒ繝薙Ν繝・
	cd frontend && pnpm build

frontend-test: ## 繝輔Ο繝ｳ繝医お繝ｳ繝峨・繝・せ繝医ｒ螳溯｡・
	cd frontend && pnpm test

frontend-lint: ## 繝輔Ο繝ｳ繝医お繝ｳ繝峨・Lint繧貞ｮ溯｡・
	cd frontend && pnpm lint

frontend-storybook: ## Storybook繧定ｵｷ蜍・
	cd frontend && pnpm storybook

frontend-e2e: ## E2E繝・せ繝医ｒ螳溯｡・
	cd frontend && pnpm test:e2e

# ---- Database ----
migrate-up: ## 繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ繧帝←逕ｨ
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down: ## 繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ繧偵Ο繝ｼ繝ｫ繝舌ャ繧ｯ
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## 譁ｰ縺励＞繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ繧剃ｽ懈・ (NAME=xxx)
	migrate create -ext sql -dir backend/migrations -seq $(NAME)

seed: ## 髢狗匱逕ｨ繧ｷ繝ｼ繝峨ョ繝ｼ繧ｿ繧呈兜蜈･
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

seed-reset: ## 繝・・繧ｿ繝吶・繧ｹ繧偵Μ繧ｻ繝・ヨ縺励※繧ｷ繝ｼ繝峨ｒ謚募・
	docker compose exec -T postgres psql -U kintai -d kintai -c "TRUNCATE users, departments, attendances, leave_requests, shifts, refresh_tokens CASCADE;"
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

# ---- Infrastructure ----
tf-init: ## Terraform蛻晄悄蛹・(dev)
	cd infrastructure/environments/dev && terraform init

tf-plan: ## Terraform繝励Λ繝ｳ遒ｺ隱・(dev)
	cd infrastructure/environments/dev && terraform plan

tf-apply: ## Terraform驕ｩ逕ｨ (dev)
	cd infrastructure/environments/dev && terraform apply

# ---- Codegen ----
generate-api-client: ## OpenAPI縺九ｉ繝輔Ο繝ｳ繝医お繝ｳ繝陰PI繧ｯ繝ｩ繧､繧｢繝ｳ繝医ｒ逕滓・
	cd frontend && pnpm generate:api

generate-mocks: ## Go繝｢繝・け繧堤函謌・
	cd backend && go generate ./...

