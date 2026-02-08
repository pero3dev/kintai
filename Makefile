.PHONY: help up down build logs backend-test frontend-test lint migrate

# ========================================
# 勤怠管理システム - Makefile
# ========================================

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---- Docker Compose ----
up: ## 開発環境を起動
	docker compose up -d

down: ## 開発環境を停止
	docker compose down

build: ## コンテナをビルド
	docker compose build

logs: ## ログを表示
	docker compose logs -f

# ---- Backend ----
backend-run: ## バックエンドを起動 (ホットリロード)
	cd backend && air

backend-test: ## バックエンドのテストを実行
	cd backend && go test ./... -v -cover

backend-lint: ## バックエンドのLintを実行
	cd backend && golangci-lint run ./...

backend-swagger: ## Swaggerドキュメントを生成
	cd backend && swag init -g cmd/server/main.go -o docs

# ---- Frontend ----
frontend-dev: ## フロントエンド開発サーバーを起動
	cd frontend && npm run dev

frontend-build: ## フロントエンドをビルド
	cd frontend && npm run build

frontend-test: ## フロントエンドのテストを実行
	cd frontend && npm run test

frontend-lint: ## フロントエンドのLintを実行
	cd frontend && npm run lint

frontend-storybook: ## Storybookを起動
	cd frontend && npm run storybook

frontend-e2e: ## E2Eテストを実行
	cd frontend && npm run test:e2e

# ---- Database ----
migrate-up: ## マイグレーションを適用
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down: ## マイグレーションをロールバック
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## 新しいマイグレーションを作成 (NAME=xxx)
	migrate create -ext sql -dir backend/migrations -seq $(NAME)

seed: ## 開発用シードデータを投入
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

seed-reset: ## データベースをリセットしてシードを投入
	docker compose exec -T postgres psql -U kintai -d kintai -c "TRUNCATE users, departments, attendances, leave_requests, shifts, refresh_tokens CASCADE;"
	docker compose exec -T postgres psql -U kintai -d kintai < backend/seeds/seed.sql

# ---- Infrastructure ----
tf-init: ## Terraform初期化 (dev)
	cd infrastructure/environments/dev && terraform init

tf-plan: ## Terraformプラン確認 (dev)
	cd infrastructure/environments/dev && terraform plan

tf-apply: ## Terraform適用 (dev)
	cd infrastructure/environments/dev && terraform apply

# ---- Codegen ----
generate-api-client: ## OpenAPIからフロントエンドAPIクライアントを生成
	cd frontend && npm run generate:api

generate-mocks: ## Goモックを生成
	cd backend && go generate ./...
