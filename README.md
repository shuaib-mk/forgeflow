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

## Security

Never reuse the sample database password and replace `FORGEFLOW_SESSION_SECRET` before any shared deployment. Workflow commands execute on the host with the configured service account; repository roots and command policies are validated, but the worker is not a hostile multi-tenant sandbox. See [SECURITY.md](SECURITY.md).

## Contributing

Issues and pull requests are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), which describes local checks, commit conventions, and architectural boundaries.

Licensed under the [MIT License](LICENSE).

