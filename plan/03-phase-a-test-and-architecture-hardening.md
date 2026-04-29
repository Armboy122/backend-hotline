# Phase A - Test And Architecture Hardening

## Status

Done as baseline guardrails, then updated for the Hinghoi-style direction in Phase 0.

## Structure Rule

New guardrails must protect `internal/feature/<feature>/...`, not the deprecated `internal/modules` path.

## Scope

- Keep existing config/middleware/response tests.
- Keep existing legacy behavior tests until each feature is migrated.
- Add or update architecture tests when a feature moves.

## Required Architecture Checks

- `internal/feature/<feature>/service` must not import Gin, GORM, legacy handlers, or persistence models.
- `internal/feature/<feature>/controller` must not import GORM or persistence models.
- `internal/feature/<feature>/dto`, `entity`, and `mapper` must stay framework-free.
- Router/server composition must not contain business queries.
- Once a feature is migrated, no new work for that feature should land in `internal/modules`, `internal/domain`, `internal/app`, `internal/port`, `internal/adapter/out`, or `internal/handlers/v1`.

## Existing Completed Work

- Config and middleware tests exist.
- Standard response/error mapping tests exist.
- TaskDaily list behavior tests exist around the current implementation.
- Smoke script skeleton exists.

## Acceptance

```bash
go test ./internal/architecture -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
