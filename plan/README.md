# Backend Hotline Refactor Plan Index

This folder is the execution map for moving `backend-hotline` to the Hinghoi backend style used by `/Users/sakdithat/Desktop/Devpool/hinghoi-backend`.

## Structure Rule

All new feature work must use:

```text
internal/feature/<feature>/
  controller/
  service/
  repository/
  dto/
  entity/
  mapper/
```

The previous `internal/modules/<module>` direction is deprecated. Existing module folders are transitional source only.

## How To Use This Plan Set

1. Read [`00-structure-reset.md`](./00-structure-reset.md).
2. Read [`00-backend-architecture.md`](./00-backend-architecture.md).
3. Check [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md).
4. Work through [`02-execution-backlog.md`](./02-execution-backlog.md) in order.
5. Keep each task scoped to one feature folder.
6. End every code task with the quality gate below.

## Plan Files

| Order | File | Purpose |
|---:|---|---|
| 0 | [`00-structure-reset.md`](./00-structure-reset.md) | Hinghoi-style feature-first baseline and migration bridge |
| 1 | [`00-backend-architecture.md`](./00-backend-architecture.md) | Target architecture and dependency rules |
| 2 | [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md) | Verified status and known transitional code |
| 3 | [`02-execution-backlog.md`](./02-execution-backlog.md) | Ordered backlog with dependencies |
| 4 | [`03-phase-a-test-and-architecture-hardening.md`](./03-phase-a-test-and-architecture-hardening.md) | Guardrails and regression tests |
| 5 | [`04-phase-b-taskdaily-vertical-slice.md`](./04-phase-b-taskdaily-vertical-slice.md) | Move TaskDaily fully into `internal/feature/task` |
| 6 | [`05-phase-c-monthly-plan-workflow.md`](./05-phase-c-monthly-plan-workflow.md) | Move monthly plan workflow into `internal/feature/monthlyplan` |
| 7 | [`06-phase-d-dashboard-and-masterdata.md`](./06-phase-d-dashboard-and-masterdata.md) | Move dashboard and master data into feature folders |
| 8 | [`07-phase-e-auth-user-deploy-hardening.md`](./07-phase-e-auth-user-deploy-hardening.md) | Move auth/user and finish release hardening |
| 9 | [`99-agent-task-template.md`](./99-agent-task-template.md) | Task assignment template |
| 10 | [`release-checklist.md`](./release-checklist.md) | Release readiness checklist |
| 11 | [`runbook.md`](./runbook.md) | Runtime, migration, rollback, and troubleshooting notes |
| 12 | [`98-session-log-2026-04-28.md`](./98-session-log-2026-04-28.md) | Previous session notes |
| 13 | [`08-phase-a-wave-1-task-board.md`](./08-phase-a-wave-1-task-board.md) | Historical Wave 1 board |

## Phase Progress

- ✅ **Phase 0** — Hinghoi-style structure reset and TaskDaily pilot
- ✅ **M1** — Safe foundation and test guardrails
- ✅ **M2** — TaskDaily fully moved to `internal/feature/task`
- ✅ **M3** — Monthly plan fully moved to `internal/feature/monthlyplan`
- ✅ **M4** — Dashboard and master data route entrypoints plus business extraction are complete
- ✅ **M5** — Auth/user route entrypoints, deploy hardening, and DB/server bridge updates completed; only release operational checklist remains

## Current Recommendation

Feature migration backlog is complete. Use [`release-checklist.md`](./release-checklist.md) and [`runbook.md`](./runbook.md) for pre-release validation and operational checks.

## Quality Gate

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
