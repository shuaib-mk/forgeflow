#!/usr/bin/env sh
set -eu

printf '[1/6] Checking repository inputs and formatting\n'
test -f .env.example
test -f deployments/migrations/000001_initial.up.sql
test -f web/package-lock.json
test -z "$(gofmt -l cmd internal pkg cli tests)"

printf '[2/6] Running Go static analysis and unit tests\n'
go vet ./cmd/... ./internal/... ./pkg/... ./cli/...
go test -race -count=1 -coverprofile=coverage.out ./internal/... ./pkg/... ./cli/...
go test -count=1 -coverprofile=core-coverage.out ./internal/analytics ./internal/auth ./internal/authorization ./internal/events ./internal/projects ./internal/tasks ./internal/workflows
./scripts/check-coverage.sh core-coverage.out 70

printf '[3/6] Running dashboard formatting, lint, type checks, coverage, and build\n'
(cd web && npm ci --ignore-scripts && npm run format:check && npm run lint && npm run typecheck && npm run test:coverage && npm run build && npm audit --audit-level=high)

printf '[4/6] Building and starting the complete Docker stack\n'
docker compose up -d --build --wait

printf '[5/6] Running PostgreSQL/Redis integration tests\n'
FORGEFLOW_DATABASE_URL='postgres://forgeflow:forgeflow@localhost:5432/forgeflow?sslmode=disable' \
FORGEFLOW_REDIS_URL='redis://localhost:6379/0' \
go test -tags=integration -count=1 ./tests/integration/...

printf '[6/6] Running end-to-end tests against the live API\n'
FORGEFLOW_E2E_URL='http://localhost:8080' go test -tags=e2e -count=1 ./tests/e2e/...

printf 'VERIFICATION PASSED: unit, coverage, frontend, integration, E2E, and container checks completed successfully.\n'
