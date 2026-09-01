# ADR 0003: Execute commands without a shell

- Status: accepted
- Date: 2026-09-01

## Decision

Workflow definitions separate an executable from its argument list. The worker invokes it directly with a fixed working directory, timeout, cancellation context, and bounded output buffer.

## Consequences

Shell interpolation, pipelines, redirection, and command substitution do not occur implicitly. Users who genuinely require shell behavior must invoke an explicit checked-in script. This materially reduces injection mistakes, but it does not turn the local runner into a hostile multi-tenant sandbox.

