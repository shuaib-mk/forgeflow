# ADR 0001: Start with a modular monolith

- Status: accepted
- Date: 2026-09-01

## Decision

Use one Go module with separate API and worker processes and feature-oriented internal packages.

## Rationale

Small teams gain simple transactions, one deployment unit, and fast refactoring. Process separation keeps long-running commands away from HTTP traffic. Package boundaries and repository interfaces allow later extraction based on observed load rather than speculation.

