# Backend Hotline Refactor Plan Index

> This folder is the execution map for `backend-hotline`. It is modeled after the provided `smart-cover-connect-backend/plan` folder, but scoped to this repo's actual Gin/GORM/Viper structure and Hotline domain.

## Structure rule (READ FIRST)

> **All code from Phase A onward MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md).**
>
> Each module lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. Do **not** create new code in the legacy layer-first locations (`internal/domain/`, `internal/app/`, `internal/port/`, `internal/adapter/`, `internal/handlers/v1/`); only modify them as part of a migration into a module folder. The example pilot lives at `internal/modules/task/`.

## How to use this plan set

0. Read product/domain notes first:
   - `/Users/sakdithat/Downloads/hotline_prd.md`
   - `/Users/sakdithat/Downloads/hotline_domain_map.md`
   - `/Users/sakdithat/Downloads/backend_hotline_repo_evidence.md`
1. Read [`00-structure-reset.md`](./00-structure-reset.md) before changing code.
2. Then read [`00-backend-architecture.md`](./00-backend-architecture.md) for the detailed rules.
3. Check [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md) to avoid redoing completed work.
4. Work through [`02-execution-backlog.md`](./02-execution-backlog.md) in order.
5. Assign one task card at a time to an agent.
6. Every code task should follow RED -> GREEN -> REFACTOR where practical and end with `go test ./...`.

## Plan files

| Order | File | Purpose |
|---:|---|---|
| 0 | [`00-structure-reset.md`](./00-structure-reset.md) | SCC-style module-first prerequisite phase and migration bridge |
| 1 | [`00-backend-architecture.md`](./00-backend-architecture.md) | Current/target architecture, module contracts, dependency rules |
| 2 | [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md) | Verified baseline, routes, partial refactor status, known gaps |
| 3 | [`02-execution-backlog.md`](./02-execution-backlog.md) | Global ordered backlog with dependencies |
| 4 | [`03-phase-a-test-and-architecture-hardening.md`](./03-phase-a-test-and-architecture-hardening.md) | Tests and architecture guardrails before moving more code |
| 5 | [`04-phase-b-taskdaily-vertical-slice.md`](./04-phase-b-taskdaily-vertical-slice.md) | Complete TaskDaily hexagonal slice |
| 6 | [`05-phase-c-monthly-plan-workflow.md`](./05-phase-c-monthly-plan-workflow.md) | Extract monthly plan/R2 workflow into module boundaries |
| 7 | [`06-phase-d-dashboard-and-masterdata.md`](./06-phase-d-dashboard-and-masterdata.md) | Dashboard query services and master data module pattern |
| 8 | [`07-phase-e-auth-user-deploy-hardening.md`](./07-phase-e-auth-user-deploy-hardening.md) | Auth/user safety, config, deploy, smoke checks |
| 9 | [`99-agent-task-template.md`](./99-agent-task-template.md) | Copy/paste template for assigning one task to Codex/Claude/GLM |
| 10 | [`release-checklist.md`](./release-checklist.md) | Release readiness checklist |
| 11 | [`runbook.md`](./runbook.md) | Runtime, migration, rollback, and troubleshooting notes |
| 12 | [`98-session-log-2026-04-28.md`](./98-session-log-2026-04-28.md) | This session's analysis and parallel-agent notes |
| 13 | [`08-phase-a-wave-1-task-board.md`](./08-phase-a-wave-1-task-board.md) | Wave 1 agent split and execution board |

## Phase progress

- ✅ **Phase 0** — Structure reset & module-first baseline
- ⬜ **M1** — Safe foundation
- ⬜ **M2** — TaskDaily module complete
- ⬜ **M3** — Monthly plan workflow modularized
- ⬜ **M4** — Dashboard and master data cleaned
- ⬜ **M5** — Auth/user/release hardening

## Structure rule (applies to ALL phases)

> Every phase MUST follow the **module-first vertical-slice** layout defined in [`00-structure-reset.md`](./00-structure-reset.md). Each feature lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. The pilot at `internal/modules/task/` is the reference. Do **not** create new files under the legacy `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` paths — migrate into `internal/modules/` instead.

## Immediate recommended next task

Start with **M1** in [`02-execution-backlog.md`](./02-execution-backlog.md): test + architecture hardening on top of the new module-first baseline.

## Quality gate for every phase

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

Do not proceed to the next phase if these fail.
