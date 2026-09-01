# Local development

Copy `.env.example` to `.env`. The Compose profile is the supported fresh-clone path and needs only Docker Desktop:

```sh
docker compose up --build
```

For host development, install Go 1.24, Node 24, PostgreSQL 17, and Redis 7. Run `make setup`, `make compose-up`, `make migrate`, then start `go run ./cmd/forgeflow-api`, `go run ./cmd/forgeflow-worker`, and `npm run dev --prefix web` in separate terminals.

The normal test suite is service-independent. Integration and E2E tests are explicitly tagged. Reproduce CI with `make fmt lint test build test-integration`.

## Troubleshooting

- `/ready` returns 503: verify the database URL and that migrations ran.
- Runs stay queued: start the worker and check Redis connectivity.
- A repository is rejected: it must resolve beneath `FORGEFLOW_WORKSPACE_ROOT`.
- The UI cannot reach the API: set `VITE_API_URL` at frontend build time or use the Compose proxy defaults.

