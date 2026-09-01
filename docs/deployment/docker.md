# Docker deployment

Compose creates PostgreSQL, Redis, the migration job, API, worker, and dashboard. Named volumes retain database and workspace data. Services run as non-root users and expose only the API and dashboard.

Before sharing a deployment, replace every sample credential, terminate TLS at a trusted reverse proxy, restrict the worker's filesystem and network permissions, and back up the PostgreSQL volume. Health checks are configured for dependency ordering, not as an availability guarantee.

