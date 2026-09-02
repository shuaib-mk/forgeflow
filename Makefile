.DEFAULT_GOAL := help

.PHONY: help setup dev build test test-unit test-integration test-e2e lint fmt verify compose-up compose-down migrate

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install frontend dependencies and prepare local configuration
	@test -f .env || cp .env.example .env
	cd web && npm ci

dev: ## Start PostgreSQL, Redis, API, worker, and dashboard
	docker compose up --build

build: ## Build all Go binaries and the dashboard
	go build -trimpath -o bin/forgeflow ./cmd/forgeflow
	go build -trimpath -o bin/forgeflow-api ./cmd/forgeflow-api
	go build -trimpath -o bin/forgeflow-worker ./cmd/forgeflow-worker
	cd web && npm run build

test: test-unit ## Run the default deterministic test suite
	cd web && npm test -- --run

test-unit: ## Run Go unit tests
	go test -race -count=1 ./internal/... ./pkg/... ./cli/...

test-integration: ## Run integration tests (requires Docker)
	go test -tags=integration -count=1 ./tests/integration/...

test-e2e: ## Run end-to-end smoke tests (requires the stack)
	go test -tags=e2e -count=1 ./tests/e2e/...

fmt: ## Format source files
	gofmt -w $$(find cmd internal pkg cli tests -name '*.go')
	cd web && npm run format

lint: ## Run static checks
	go vet ./cmd/... ./internal/... ./pkg/... ./cli/...
	cd web && npm run lint

verify: ## Run fresh-clone-equivalent unit, coverage, container, integration, and E2E checks
	./scripts/verify.sh

compose-up: ## Start dependencies
	docker compose up -d postgres redis

compose-down: ## Stop local services
	docker compose down

migrate: ## Apply database migrations
	go run ./cmd/forgeflow-migrate up
