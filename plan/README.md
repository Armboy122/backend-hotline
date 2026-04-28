# Backend Hotline Refactor Plan Index

> This folder is the execution map for `backend-hotline`. It is modeled after the provided `smart-cover-connect-backend/plan` folder, but scoped to this repo's actual Gin/GORM/Viper structure and Hotline domain.

## How to use this plan set

0. Read product/domain notes first:
   - `/Users/sakdithat/Downloads/hotline_prd.md`
   - `/Users/sakdithat/Downloads/hotline_domain_map.md`
   - `/Users/sakdithat/Downloads/backend_hotline_repo_evidence.md`
1. Read [`00-backend-architecture.md`](./00-backend-architecture.md) before changing code.
2. Check [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md) to avoid redoing completed work.
3. Work through [`02-execution-backlog.md`](./02-execution-backlog.md) in order.
4. Assign one task card at a time to an agent.
5. Every code task should follow RED -> GREEN -> REFACTOR where practical and end with `go test ./...`.

## Plan files

| Order | File | Purpose |
|---:|---|---|
| 0 | [`00-backend-architecture.md`](./00-backend-architecture.md) | Current/target architecture, module contracts, dependency rules |
| 1 | [`01-current-state-and-done-checklist.md`](./01-current-state-and-done-checklist.md) | Verified baseline, routes, partial refactor status, known gaps |
| 2 | [`02-execution-backlog.md`](./02-execution-backlog.md) | Global ordered backlog with dependencies |
| 3 | [`03-phase-a-test-and-architecture-hardening.md`](./03-phase-a-test-and-architecture-hardening.md) | Tests and architecture guardrails before moving more code |
| 4 | [`04-phase-b-taskdaily-vertical-slice.md`](./04-phase-b-taskdaily-vertical-slice.md) | Complete TaskDaily hexagonal slice |
| 5 | [`05-phase-c-monthly-plan-workflow.md`](./05-phase-c-monthly-plan-workflow.md) | Extract monthly plan/R2 workflow into module boundaries |
| 6 | [`06-phase-d-dashboard-and-masterdata.md`](./06-phase-d-dashboard-and-masterdata.md) | Dashboard query services and master data module pattern |
| 7 | [`07-phase-e-auth-user-deploy-hardening.md`](./07-phase-e-auth-user-deploy-hardening.md) | Auth/user safety, config, deploy, smoke checks |
| 8 | [`99-agent-task-template.md`](./99-agent-task-template.md) | Copy/paste template for assigning one task to Codex/Claude/GLM |
| 9 | [`release-checklist.md`](./release-checklist.md) | Release readiness checklist |
| 10 | [`runbook.md`](./runbook.md) | Runtime, migration, rollback, and troubleshooting notes |
| 11 | [`98-session-log-2026-04-28.md`](./98-session-log-2026-04-28.md) | This session's analysis and parallel-agent notes |
| 12 | [`08-phase-a-wave-1-task-board.md`](./08-phase-a-wave-1-task-board.md) | Wave 1 agent split and execution board |

## Immediate recommended next task

Start with **A1** in [`03-phase-a-test-and-architecture-hardening.md`](./03-phase-a-test-and-architecture-hardening.md): add architecture boundary tests.

Reason: this repo already has a partial TaskDaily hexagonal slice, while most handlers still use direct GORM. Boundary tests prevent future refactor work from making the transition more inconsistent.

## Quality gate for every phase

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

Do not proceed to the next phase if these fail.
