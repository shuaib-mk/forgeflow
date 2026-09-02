# Changelog

All notable changes are documented here. ForgeFlow follows Semantic Versioning.

## [Unreleased]

## [0.2.0] - 2026-09-02

### Added

- Initial monorepo and reproducible development tooling.
- Typed domain, configuration, event, and plugin foundations.
- Validated workflow definitions with DAG dependencies, retries, cancellation, and timeouts.
- PostgreSQL persistence, Redis job delivery, and concurrency-safe background workers.
- Versioned REST API with bearer sessions, role-based authorization, pagination, request IDs, structured logs, health checks, and metrics.
- CLI commands for project, task, workflow, run, configuration, and diagnostics workflows.
- Responsive API-backed dashboard with accessible operational states.
- Multi-stage non-root containers, Compose stack, development container, immutable CI actions, CodeQL, vulnerability checks, and dependency automation.
- OpenAPI 3.1 contract, architecture decisions, deployment guidance, and fresh-clone verification script.
- Complete web product loop for account creation, projects, tasks, repositories, workflows, live runs, logs, analytics, audit history, plugins, and settings.
- Enforced backend and frontend coverage thresholds, frontend formatting and type checks, and one-command full-stack verification.
- Structured reporting for unexpected API failures with request method and route context.

### Fixed

- Preserve command failures as failed step states instead of misclassifying them after context cleanup.
- Break the workflow-service and worker import cycle discovered by full toolchain verification.
- Return the initial organization from registration so a new user can immediately configure the dashboard and CLI.
- Exclude frontend dependency trees and generated TypeScript metadata from repository and container build contexts.
- Scope task status updates to their owning project and support unfiltered project task lists.
