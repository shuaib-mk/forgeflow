# Event system

The in-process event bus handles reactions that belong to the same deployment but not the initiating feature. Today, project creation emits a typed event consumed by audit logging. Handlers execute synchronously so callers know whether required reactions completed.

Use events for cross-feature reactions with typed payloads. Do not use them to hide ordinary function calls, transport jobs to workers, or replace database state. A handler must be deterministic and idempotent where retries are possible.

