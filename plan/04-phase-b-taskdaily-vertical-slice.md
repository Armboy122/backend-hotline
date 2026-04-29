# Phase B - TaskDaily Vertical Slice

## Goal

Complete the TaskDaily module so all `/v1/tasks*` endpoints use application usecases and repository ports. This creates the pattern for later domains.

## ✅ Phase B COMPLETE (2026-04-29)

All B1–B5 tasks delivered. `go vet`, `go build`, `go test ./...` all pass.

### Summary of changes:
- **B1:** Expanded `TaskRepository` interface with `GetByID`, `Create`, `Update`, `SoftDelete`, `ListByTeam`, `ListByFilter` + query/input structs
- **B1:** Implemented all 6 methods in GORM adapter (`internal/adapter/out/persistence/gorm/task_repository.go`)
- **B2:** Created 7 usecases: `GetTask`, `CreateTask`, `UpdateTask`, `DeleteTask`, `ListTasks` (existing), `ListTasksByTeam`, `ListTasksByFilter` + shared `errors.go`
- **B3:** Refactored `TaskHandler` to use single-method port interfaces, delegate all 7 endpoints to usecases
- **B4:** Updated existing handler tests for new struct fields, pagination normalization moved to usecase layer
- **B5:** Response compatibility maintained — `convertDomainTaskToResponse` handles nested team/job/feeder/station
- **Router:** Updated to create `gormadapter.NewTaskRepository(db)` before passing to `NewTaskHandler`

### Files modified:
- `internal/port/outbound/repository/task_repository.go` — expanded interface
- `internal/adapter/out/persistence/gorm/task_repository.go` — full GORM impl
- `internal/app/task/usecase/` — 8 new files (7 usecases + errors.go)
- `internal/handlers/v1/task.go` — complete handler rewrite
- `internal/handlers/v1/error_mapping_test.go` — updated for new struct
- `internal/app/task/usecase/list_tasks_test.go` — updated fakeTaskRepo
- `internal/router/router.go` — wiring update

## B1 - Expand task repository contract

**Objective:** Make the repository interface express all task operations without leaking GORM models.

**Files:**

- Modify: `internal/port/outbound/repository/task_repository.go`
- Modify: `internal/adapter/out/persistence/gorm/task_repository.go`
- Modify: `internal/domain/task/entity.go`
- Add tests as needed under `internal/adapter/out/persistence/gorm/`

**Task cards:**

### B1.1 Add query/input structs

- [ ] Add `TaskGetQuery`
- [ ] Add `TaskCreateInput`
- [ ] Add `TaskUpdateInput`
- [ ] Add `TaskDeleteCommand`
- [ ] Add `TaskByTeamQuery`
- [ ] Add `TaskByFilterQuery`

### B1.2 Add repository methods

- [ ] `GetByID(ctx, id)`
- [ ] `Create(ctx, input)`
- [ ] `Update(ctx, input)`
- [ ] `SoftDelete(ctx, id)`
- [ ] `ListByTeam(ctx, query)`
- [ ] `ListByFilter(ctx, query)`

### B1.3 Preserve GORM preload behavior

- [ ] Keep Team preload
- [ ] Keep JobType preload
- [ ] Keep JobDetail preload
- [ ] Keep Feeder -> Station -> OperationCenter preload

Run:

```bash
go test ./internal/adapter/out/persistence/gorm -run TestTaskRepository -v
```

## B2 - Add task usecases

**Objective:** Move task business/input validation out of handler methods.

**Files:**

- Create: `internal/app/task/usecase/get_task.go`
- Create: `internal/app/task/usecase/create_task.go`
- Create: `internal/app/task/usecase/update_task.go`
- Create: `internal/app/task/usecase/delete_task.go`
- Create: `internal/app/task/usecase/list_tasks_by_team.go`
- Create: `internal/app/task/usecase/list_tasks_by_filter.go`
- Create tests in same package

**Task cards:**

### B2.1 Get task usecase

- [ ] Reject id <= 0
- [ ] Map repository not found to domain/usecase not found error
- [ ] Return domain entity

### B2.2 Create task usecase

- [ ] Require workDate
- [ ] Require teamId, jobTypeId, jobDetailId
- [ ] Validate optional feederId
- [ ] Preserve latitude/longitude precision behavior
- [ ] Preserve urlsBefore/urlsAfter behavior

### B2.3 Update task usecase

- [ ] Reject id <= 0
- [ ] Only update intended fields
- [ ] Preserve existing soft-delete semantics
- [ ] Do not use `Save()` if field-specific updates are safer

### B2.4 Delete task usecase

- [ ] Soft delete only
- [ ] Not found returns not found error
- [ ] Delete is idempotency policy documented and tested

### B2.5 List by team/filter usecases

- [ ] Preserve current year/month/team filter behavior
- [ ] Preserve current ordering
- [ ] Preserve response totals/count logic

## B3 - Refactor TaskHandler

**Objective:** Make `internal/handlers/v1/task.go` delegate all business behavior to usecases.

**Files:**

- Modify: `internal/handlers/v1/task.go`

**Task cards:**

### B3.1 Constructor wiring

- [ ] `NewTaskHandler` constructs one GORM repository
- [ ] Wire all task usecases
- [ ] Keep handler struct explicit, not a service locator

### B3.2 Parse and map only

- [ ] Handler parses query/path/body
- [ ] Handler maps auth context only if needed
- [ ] Handler calls usecase
- [ ] Handler maps usecase result to `dto.TaskResponse`

### B3.3 Error mapping

- [ ] Validation errors -> current 400/422 convention
- [ ] Not found -> 404
- [ ] Repository/internal -> 500
- [ ] Standard response envelope preserved

## B4 - Compatibility tests

**Objective:** Prove `/v1/tasks*` behavior did not drift.

**Files:**

- Create/modify: `internal/handlers/v1/task_test.go`

**Task cards:**

- [ ] `GET /v1/tasks` with pagination and filters
- [ ] `GET /v1/tasks/:id` success/not found/invalid id
- [ ] `POST /v1/tasks` success/invalid body
- [ ] `PUT /v1/tasks/:id` success/not found
- [ ] `DELETE /v1/tasks/:id` success/not found
- [ ] `GET /v1/tasks/by-team`
- [ ] `GET /v1/tasks/by-filter`

## Phase B acceptance

```bash
go test ./internal/app/task/... ./internal/adapter/out/persistence/gorm ./internal/handlers/v1 -run 'Test.*Task' -v
go test ./...
go vet ./...
```
