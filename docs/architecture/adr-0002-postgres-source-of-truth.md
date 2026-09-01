# ADR 0002: PostgreSQL is the source of truth

- Status: accepted
- Date: 2026-09-01

## Decision

Persist domain and job state in PostgreSQL. Redis is only the delivery mechanism for runnable job IDs.

## Rationale

Queue messages may be delivered more than once. Workers claim a run atomically in PostgreSQL, so replay is safe and lost Redis data can be reconstructed from queued runs.

