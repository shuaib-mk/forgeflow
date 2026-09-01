#!/usr/bin/env sh
set -eu

test -f .env.example
test -f deployments/migrations/000001_initial.up.sql
test -f web/package-lock.json
test -z "$(gofmt -l cmd internal pkg cli tests)"
go vet ./cmd/... ./internal/... ./pkg/... ./cli/...
go test -race -count=1 ./internal/... ./pkg/... ./cli/...
(cd web && npm ci --ignore-scripts && npm run lint && npm test -- --run && npm run build && npm audit --audit-level=high)
