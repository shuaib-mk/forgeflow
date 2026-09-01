# Architecture overview

ForgeFlow is a modular monolith with independently runnable API and worker processes. This keeps local operation simple while preserving boundaries that permit future extraction.

```text
Browser / CLI -> REST API -> services -> PostgreSQL
                              |             |
                              +-> events ---+-> audit
                              +-> Redis queue -> worker -> workflow executor
```

The API owns validation, authentication middleware, request IDs, and HTTP representation. Feature services own authorization and transactions. Repositories isolate persistence. Redis carries short-lived job delivery; PostgreSQL remains the source of truth for run state. In-process typed events decouple audit and notification reactions without becoming a second source of truth.

Workflow processes receive a fixed working directory, bounded output, timeouts, and cancellation. Arguments are executed directly rather than through a shell. This prevents shell expansion but does not make the runner safe for hostile tenants.

See the ADRs in this directory for key trade-offs.

