# ForgeFlow

ForgeFlow is a local-first workflow platform for small engineering teams. It combines project and task management with safe repository automation, background workflow execution, audit history, analytics, a REST API, a CLI, and a responsive web dashboard.

> ForgeFlow is under active development. The current release is suitable for local evaluation and contribution; review the security guide before exposing it beyond a trusted network.

## Quick start

Requirements: Docker Desktop with Compose, Git, and Make (or equivalent commands).

```sh
cp .env.example .env
docker compose up --build
```

Open <http://localhost:5173>. The API is available at <http://localhost:8080>, with health checks at `/health` and `/ready`.

Open the dashboard, choose **Create account**, and enter your name, organization, email, and password. ForgeFlow creates the organization, selects it automatically, and signs you in. The opaque session token is stored in browser session storage so closing the tab clears it.

## Dashboard workflow

1. Create a project from **Projects**.
2. Open the project and add tasks, initialize or connect a repository, and create a workflow. Workflow commands run in the project's first connected repository (or the workspace root when none is connected).
3. Run a workflow from the project page.
4. Watch its status and logs update automatically, then review all executions under **Runs**.
5. Use **Analytics** and **Audit log** to review workspace activity.

The dashboard covers the complete local workflow; the CLI and REST API are optional automation interfaces.

## Features

- Local project, task, and Git repository organization
- Validated workflow DAGs with direct command execution, timeouts, retries, cancellation, and bounded logs
- Durable PostgreSQL state with Redis-backed background delivery and atomic duplicate-run prevention
- Owner, admin, developer, and viewer permissions enforced in application services
- Typed events, durable audit history, analytics counters, notifications schema, and trusted plugin contracts
- REST API, Go CLI, responsive React dashboard, health/readiness probes, structured logs, and Prometheus-format metrics
- Unit, integration, frontend, and E2E test layers with locked dependencies and immutable CI actions

## Architecture

ForgeFlow is a modular monolith with separately runnable API and worker processes. HTTP adapters call feature services; services enforce authorization and own transactions; repository adapters isolate PostgreSQL. Redis delivers only run IDs and PostgreSQL remains authoritative. Typed in-process events handle cross-feature reactions such as audit records. See [the architecture overview](docs/architecture/overview.md) and [ADRs](docs/architecture/).

## Repository map

- `cmd/`: API, worker, migration, and CLI entry points
- `internal/`: application services and adapters
- `pkg/`: stable client and public models
- `cli/`: CLI command tree
- `web/`: React dashboard
- `deployments/`: containers and database migrations
- `docs/`: architecture, API, development, and deployment guides
- `tests/`: integration and end-to-end suites

## Common commands

```sh
make setup             # install local dependencies
make build             # build all binaries and dashboard
make test              # deterministic unit and frontend tests
make test-integration  # PostgreSQL/Redis integration tests
make lint              # Go and TypeScript static checks
make dev               # full stack with live services
```

Detailed setup, design, CLI, API, and troubleshooting guidance lives in [`docs/`](docs/README.md).

## Configuration

Every server variable is documented in [`.env.example`](.env.example). The required production values are `FORGEFLOW_DATABASE_URL`, `FORGEFLOW_REDIS_URL`, and a random `FORGEFLOW_SESSION_SECRET` of at least 32 characters. `FORGEFLOW_WORKSPACE_ROOT` is the only filesystem tree repositories and workflow commands may access. Comma-separated `FORGEFLOW_ALLOWED_ORIGINS` controls browser access.

## CLI

Build with `go build -o bin/forgeflow ./cmd/forgeflow`, then configure a session:

```sh
forgeflow config set --api-url http://localhost:8080 --token SESSION_TOKEN --organization ORGANIZATION_ID
forgeflow doctor
forgeflow project list
forgeflow task create --project PROJECT_ID --title "Add release checks"
forgeflow workflow validate .forgeflow/workflow.yaml
forgeflow workflow run --workflow WORKFLOW_ID
forgeflow run logs --run RUN_ID
```

Run `forgeflow --help` or any subcommand with `--help` for the complete reference. Tokens are stored in the operating system's user configuration directory with owner-only file permissions and are redacted by `forgeflow config show`.

## REST API

The API is rooted at `/api/v1`. Public registration and login endpoints issue opaque bearer sessions; all domain routes require `Authorization: Bearer …`. Responses use one error envelope with a stable code, human message, request ID, and optional validation fields. Lists accept `page`, `pageSize`, filtering, and stable sorting. The complete contract is [OpenAPI 3.1](docs/api/openapi.yaml).

Operational endpoints live outside the versioned domain API: `/health` confirms the process is alive, `/ready` checks PostgreSQL and Redis, and `/metrics` exposes non-secret counters.

## Docker and deployment

`docker compose up --build` starts PostgreSQL 17, Redis 7, the migration job, API, worker, and nginx-hosted dashboard. Images use multi-stage builds and non-root application accounts. PostgreSQL and workspace data use named volumes; Redis is deliberately ephemeral because it is not authoritative. See [Docker deployment](docs/deployment/docker.md) for production boundaries and backup guidance.

## Testing

`make test` runs deterministic Go unit tests and dashboard tests. `make test-integration` applies migrations twice against a disposable-compatible PostgreSQL database. `make test-e2e` exercises a running stack. CI additionally runs the Go race detector, static analysis, frontend lint/build, container builds, CodeQL, `govulncheck`, and `npm audit`.

For a CI-equivalent host check, run `scripts/verify.sh`. Integration and E2E suites are tagged so the default fresh-machine suite never silently depends on external services.

## Troubleshooting

- `/ready` is degraded: start PostgreSQL and Redis and verify the URLs in `.env`.
- Runs remain queued: ensure `forgeflow-worker` is running and can reach Redis and PostgreSQL.
- Repository rejected: resolve the checkout beneath `FORGEFLOW_WORKSPACE_ROOT`; symlink escapes are rejected.
- Workflow command missing: executables must be installed inside the worker environment and commands do not use implicit shell expansion.
- Dashboard is empty: use the organization selector in the sidebar or Settings. Memberships are discovered automatically after sign-in.
- Port already allocated: change the host-side port in `docker-compose.yml`; internal ports should remain unchanged.

## Security

Never reuse the sample database password and replace `FORGEFLOW_SESSION_SECRET` before any shared deployment. Workflow commands execute on the host with the configured service account; repository roots and command policies are validated, but the worker is not a hostile multi-tenant sandbox. See [SECURITY.md](SECURITY.md).

## Contributing

Issues and pull requests are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), which describes local checks, commit conventions, and architectural boundaries.

Licensed under the [MIT License](LICENSE).
