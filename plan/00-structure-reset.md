# Backend Hotline Structure Reset Implementation Plan

> **For Hermes:** Use `subagent-driven-development` to implement this plan task-by-task.

**Goal:** Reshape `backend-hotline` toward the SCC backend style: module-first, vertical-slice ownership, and module-local wiring, while preserving all current `/v1` behavior.

**Architecture:** Keep Go, Gin, GORM, PostgreSQL, JWT, and Docker. Introduce `internal/modules/<module>/` as the canonical home for each feature, with controller/service/repository/repository_impl/dto/errors grouped together. Preserve compatibility during the transition with thin wrappers, then remove the legacy layer-first packages once each module is fully migrated.

**Tech Stack:** Go 1.24, Gin, GORM, PostgreSQL, Docker, Cloudflare R2, JWT

---

## Why this phase exists

`backend-hotline` already has a solid hexagonal base, but the current layout is still layer-first:

- `internal/domain/<module>`
- `internal/app/<module>/usecase`
- `internal/port/outbound/repository`
- `internal/adapter/out/persistence/gorm`
- `internal/handlers/v1`

That is correct, but it is not the style that feels natural for this project. The SCC backend groups each feature inside one module folder, which makes the code easier to navigate and easier to hand off to another agent.

This phase is the bridge between the current hexagonal layout and the SCC-style module-first layout.

---

## Target module shape

Each migrated module should converge on:

```text
internal/modules/<module>/
  controller.go
  service.go
  repository.go
  repository_impl.go
  dto.go
  errors.go
  entity.go          # when the module needs its own domain shape
```

## Migration rule

- Keep public `/v1` routes unchanged.
- Keep response envelopes unchanged.
- Migrate one module at a time.
- Do not remove legacy packages until the module is green and routed through the new module entrypoint.
- Keep tests alongside the module being migrated.

---

## Phase 0: Document the new structure and freeze the baseline

### Task 0.1: Update plan index and backlog

**Objective:** Make this structure reset the first thing the repo does before later phases.

**Files:**
- Modify: `plan/README.md`
- Modify: `plan/02-execution-backlog.md`
- Modify: `plan/01-current-state-and-done-checklist.md` if needed for the new baseline

**Acceptance:**
- The new structure reset phase appears before the older phase sequence.
- Later phases are clearly marked as downstream from this reset.

### Task 0.2: Record the SCC-style preference in Obsidian

**Objective:** Persist the architecture preference in the Hotline vault so future sessions stay aligned.

**Files:**
- Create or update: `Hotline/Hotline - Architecture Preference.md`
- Update: `Hotline/README.md` if a hub link is helpful

**Acceptance:**
- The note says the preferred backend shape is module-first / vertical-slice, matching SCC.
- The note links back to the repo plan files.

### Task 0.3: Create the module-first task pilot

**Objective:** Add a module-local entrypoint for TaskDaily while preserving current behavior.

**Files:**
- Create: `internal/modules/task/controller.go`
- Create: `internal/modules/task/service.go`
- Create: `internal/modules/task/repository.go`
- Create: `internal/modules/task/repository_impl.go`
- Create: `internal/modules/task/dto.go`
- Create: `internal/modules/task/errors.go`
- Create or move compatibility glue in `internal/handlers/v1/task.go`
- Modify: `internal/router/router.go`

**Implementation rule:**
- The module-local entrypoint should own construction for the task flow.
- Existing usecase and repository code may remain temporarily behind wrappers.
- The route registration must point at the new module entrypoint, not at the old layer-first constructor.

**Verification:**
```bash
go test ./... 
go vet ./... 
go build -o /tmp/hotlines-api main.go
```

### Task 0.4: Add boundary tests for the new module shape

**Objective:** Prevent regressions while the task pilot is being moved.

**Files:**
- Create or modify: `internal/architecture/architecture_test.go`
- Create or modify: `internal/modules/task/*_test.go`

**Checks:**
- No Gin import in service/usecase logic.
- No GORM import in controller logic.
- Module repository interfaces return domain entities or local DTOs, not GORM models.
- `/v1/tasks*` response shapes remain unchanged, with existing v1 handler tests continuing to cover behavior during the transition.

---

## Phase 1: Make TaskDaily the canonical SCC-style module

### Task 1.1: Collapse task wiring into the module folder

**Objective:** Make TaskDaily navigable from one folder instead of four layers.

**Files:**
- Move or rewrite: `internal/modules/task/*`
- Reduce legacy coupling in: `internal/app/task/usecase/*`
- Reduce legacy coupling in: `internal/adapter/out/persistence/gorm/task_repository.go`
- Reduce legacy coupling in: `internal/handlers/v1/task.go`

**Acceptance:**
- A task feature owner can understand the whole flow by opening one folder.
- `router.go` only wires the module, it does not contain task-specific business logic.

### Task 1.2: Keep tests and behavior stable during the move

**Objective:** Preserve the `/v1/tasks` behavior while the internal structure changes.

**Files:**
- Modify: task handler tests
- Modify: task usecase tests
- Modify: task repository tests if needed

**Checks:**
- Pagination defaults remain stable.
- Validation errors still return the same status codes.
- Nested team/job/feeder/station data remains compatible.

---

## Phase 2: Repeat the same structure for monthly plan

### Task 2.1: Create a module-local monthly plan folder

**Objective:** Move monthly plan workflow into the same module-first style.

**Files:**
- Create: `internal/modules/monthlyplan/controller.go`
- Create: `internal/modules/monthlyplan/service.go`
- Create: `internal/modules/monthlyplan/repository.go`
- Create: `internal/modules/monthlyplan/repository_impl.go`
- Create: `internal/modules/monthlyplan/dto.go`
- Create: `internal/modules/monthlyplan/errors.go`
- Create: `internal/modules/monthlyplan/entity.go`

**Acceptance:**
- Settings, period, upload, confirm, list, restore, delete, and hard-delete work as one module.
- R2 storage stays behind a module-local adapter boundary.

### Task 2.2: Move monthly plan routes to the module entrypoint

**Objective:** Keep the same HTTP API while changing the internal ownership.

**Files:**
- Modify: `internal/router/router.go`
- Modify or retire: `internal/handlers/v1/monthly_plan.go`
- Modify or retire: `internal/app/monthlyplan/usecase/*`

---

## Phase 3: Convert the remaining smaller modules

### Task 3.1: Extract master data modules

**Objective:** Apply the same layout to reference-data modules.

**Files:**
- `internal/modules/team/*`
- `internal/modules/jobtype/*`
- `internal/modules/jobdetail/*`
- `internal/modules/feeder/*`
- `internal/modules/station/*`
- `internal/modules/pea/*`
- `internal/modules/operationcenter/*`

### Task 3.2: Convert auth and upload boundaries where needed

**Objective:** Keep auth and storage simple but module-owned.

**Files:**
- `internal/modules/auth/*`
- `internal/modules/upload/*`
- `internal/store/*` or equivalent shared infra if it stays global

---

## Phase 4: Remove the old layer-first leftovers

### Task 4.1: Delete or retire obsolete packages

**Objective:** Finish the migration so the repo no longer feels split across two architectures.

**Files to retire when safe:**
- `internal/domain/*`
- `internal/app/*` when a module has moved
- `internal/port/*` when a module has moved
- `internal/adapter/*` when a module has moved
- `internal/handlers/v1/*` when a module has moved
- `internal/dto/*` when its DTOs have been relocated

### Task 4.2: Tighten architecture tests

**Objective:** Make the module-first structure stick.

**Checks:**
- Controllers do not talk to GORM directly.
- Service logic does not import Gin.
- Repository implementations stay inside the module or adapter boundary.
- `go test ./...` passes cleanly.

---

## End-state acceptance

The migration is done when:

- Each active feature is navigable from one module folder.
- The public `/v1` API stays compatible.
- TaskDaily and Monthly Plan have module-local ownership.
- The repo no longer feels split between two competing backend styles.
- Later phases can start without first re-learning the architecture.
