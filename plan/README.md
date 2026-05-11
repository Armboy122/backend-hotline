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
| 14 | [`09-product-prd-and-userflow-draft.md`](./09-product-prd-and-userflow-draft.md) | Product PRD and user flow draft for verification |
| 15 | [`10-hotline-prd-latest-and-kanban-scope.md`](./10-hotline-prd-latest-and-kanban-scope.md) | Latest verified PRD, open questions, role model draft, and Kanban scope |
| 16 | [`11-k0-decision-matrix.md`](./11-k0-decision-matrix.md) | K0 decision matrix: Q1-Q10 analysis, recommendations, blocking dependencies, role matrix draft |
| 17 | [`12-performance-rbac-monthly-plan-replan.md`](./12-performance-rbac-monthly-plan-replan.md) | 2026-05-09 replan: performance-first, strict super_admin/admin split, and yearly monthly-plan flow |
| 18 | [`performance-baseline-2026-05-09.md`](./performance-baseline-2026-05-09.md) | HP0 measured frontend/backend performance baseline |
| 19 | [`performance-hp1-frontend-fixes-2026-05-09.md`](./performance-hp1-frontend-fixes-2026-05-09.md) | HP1 frontend performance fixes and measurements |
| 20 | [`performance-backend-hp4b-2026-05-09.md`](./performance-backend-hp4b-2026-05-09.md) | HP4B backend/API bottleneck fixes and before/after timings |
| 21 | [`13-work-planning-and-large-job-prd-discovery.md`](./13-work-planning-and-large-job-prd-discovery.md) | Discovery PRD for team plan, calendar, contact directory, and future `งานระดมทีม` |
| 22 | [`14-session-handoff-2026-05-09.md`](./14-session-handoff-2026-05-09.md) | Handoff summary for the completed performance/RBAC/monthly-plan stabilization round |
| 23 | [`15-team-plan-largework-implementation-plan.md`](./15-team-plan-largework-implementation-plan.md) | Implementation PRD for monthly-plan correction, team plan, calendar, contact directory, and `งานระดมทีม` |
| 24 | [`16-planning-domain-api-contract.md`](./16-planning-domain-api-contract.md) | HNP-01 implementation-ready API/RBAC/DB/frontend DTO contract for team plan, calendar, contacts, daily-report prefill, and `งานระดมทีม` |
| 25 | [`17-contact-directory-implementation.md`](./17-contact-directory-implementation.md) | HNX-01 implemented backend contact directory routes, contact fields, RBAC, and verification notes |
| 26 | [`17-planning-frontend-ux-contract.md`](./17-planning-frontend-ux-contract.md) | HNP-02 frontend UX/RBAC/API integration contract for planning calendar, team plan, contacts, daily-report prefill, and `งานระดมทีม` |
| 27 | [`18-hnq-final-release-readiness-and-handoff-2026-05-10.md`](./18-hnq-final-release-readiness-and-handoff-2026-05-10.md) | HNQ final release readiness and handoff report covering implemented flows, role matrix, gates, risks, and manual validation steps |
| 28 | [`28-hotline-large-work-execution-replan-2026-05-11.md`](./28-hotline-large-work-execution-replan-2026-05-11.md) | Product-approved replan for `งานระดมทีม`: teamlead creates/edits, assigns work to teams, worker todo execution with before/after photos |
| 29 | [`29-hotline-large-work-execution-qa-handoff-2026-05-11.md`](./29-hotline-large-work-execution-qa-handoff-2026-05-11.md) | Integration QA handoff for `งานระดมทีม` execution gates, blockers, risks, and manual QA script |

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
