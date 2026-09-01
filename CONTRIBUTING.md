# Contributing

Thank you for helping improve ForgeFlow.

1. Create a focused branch from `main`.
2. Add behavior-focused tests with each change.
3. Run `make fmt lint test` before opening a pull request.
4. Use Conventional Commit subjects such as `feat:`, `fix:`, `test:`, and `docs:`.

Keep HTTP concerns in `internal/api`, domain rules in their feature package, database code behind repository interfaces, and reusable external types in `pkg`. Avoid adding dependencies unless the standard library or existing stack cannot express the requirement clearly.

Integration tests are opt-in because they start disposable PostgreSQL and Redis services. CI runs them on every pull request.

