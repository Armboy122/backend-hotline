# Phase A - Test and Architecture Hardening

> **Structure rule:** All code in this phase MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md). Each module lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. The pilot at `internal/modules/task/` is the reference. Architecture tests here MUST also fail when new code lands in the legacy `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` paths instead of `internal/modules/<module>/`.

## Goal

Lock current behavior before moving more handler logic into module services, and add guardrails that keep new code inside `internal/modules/<module>/`. This phase should create guardrails, not perform broad refactors.

## A1 - Architecture boundary tests

**Objective:** Add tests that stop future work from placing dependencies in the wrong layer.

**Files:**

- Create: `internal/architecture/architecture_test.go`

**Task cards:**

### A1.1 Test services/usecases do not import Gin

- [ ] Walk `internal/modules/**/service.go` and any remaining `internal/app/**/*.go`
- [ ] Fail if a non-test service/usecase file imports `github.com/gin-gonic/gin`
- [ ] Add clear failure message naming the file

Run:

```bash
go test ./internal/architecture -run TestUsecasesDoNotImportGin -v
```

### A1.2 Test entity/domain files do not import frameworks

- [ ] Walk `internal/modules/**/entity.go`, `internal/modules/**/errors.go`, and any remaining `internal/domain/**/*.go`
- [ ] Fail on imports from `gin`, `gorm`, `viper`, AWS SDK, or local handlers/controllers
- [ ] Allow only standard library and value-library imports when justified

### A1.3 Test router has no business queries

- [ ] Inspect `internal/router/*.go`
- [ ] Fail if files contain `.Where(`, `.Find(`, `.Create(`, `.Save(`, `.Updates(`, `.Delete(`
- [ ] This keeps router as composition/routing only

### A1.4 Test migrated repositories do not leak GORM models

- [ ] Inspect `internal/modules/**/repository.go` and any remaining `internal/port/outbound/repository/*.go`
- [ ] Fail if repository interfaces return `internal/models` types
- [ ] Allow module entities/DTOs (declared in the same module folder) and primitive query structs only

### A1.5 Test the module-first layout stays the canonical home

- [ ] Walk `internal/modules/**/*.go` and confirm each migrated module owns the seven-file shape (`controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, `entity.go` when present)
- [ ] Fail if a controller imports GORM directly
- [ ] Fail if a service imports Gin
- [ ] Fail if any new code is added under `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` for a feature that is already migrated to `internal/modules/<module>/`

Final verification:

```bash
go test ./internal/architecture -v
go test ./...
```

## A2 - Config and middleware tests

**Objective:** Lock infrastructure behavior that should remain stable during refactor.

**Files:**

- Create: `internal/config/config_test.go`
- Create or modify: `internal/middleware/*_test.go`
- Modify only if tests expose bugs: `internal/config/config.go`, `internal/middleware/*.go`

**Task cards:**

### A2.1 Config default and override tests

- [ ] Test config loads default server/database/CORS values
- [ ] Test env/config overrides where supported
- [ ] Test CORS origin parsing trims blanks if applicable

### A2.2 Auth middleware tests

- [ ] Missing bearer token returns 401
- [ ] Malformed bearer token returns 401
- [ ] Valid token stores claims/user context
- [ ] Wrong role returns 403
- [ ] Allowed role calls downstream handler

### A2.3 Timeout/recovery/cache tests

- [ ] Recovery middleware maps panic to standard error envelope
- [ ] Timeout middleware does not break fast handlers
- [ ] Cache middleware sets expected public/private headers

Final verification:

```bash
go test ./internal/config ./internal/middleware -v
go test ./...
```

## A3 - Standard response and error mapping tests

**Objective:** Prevent response envelope drift while handlers are extracted.

**Files:**

- Create: `internal/dto/response_test.go` (the standard envelope still lives in `internal/dto` and is shared infra; do not migrate it into a module)
- Create if needed: `internal/modules/task/error_mapping_test.go` for any task-specific assertions; legacy mapping coverage in `internal/handlers/v1/error_mapping_test.go` may stay only until the matching handler is removed

**Task cards:**

### A3.1 Standard response JSON shape

- [ ] Marshal success response
- [ ] Assert `success`, `data`, `meta`, `error` names remain stable
- [ ] Assert empty optional fields behave as current clients expect

### A3.2 Validation error convention

- [ ] Capture current validation response convention in tests
- [ ] Decide whether invalid path/query/body returns 400 or 422 per current handler behavior
- [ ] Document any inconsistent handler behavior as follow-up

## A4 - TaskDaily list behavior tests

**Objective:** Lock the partially migrated TaskDaily list flow before completing the slice in Phase B.

**Files:**

- Create: `internal/modules/task/service_test.go` (covers list pagination/normalization)
- Create: `internal/modules/task/repository_impl_test.go` if DB test setup is practical
- Create or modify: `internal/modules/task/controller_test.go`
- Existing legacy tests under `internal/app/task/usecase/` and `internal/handlers/v1/` may stay until Phase B retires them, but new assertions belong in the module folder

**Task cards:**

### A4.1 Usecase pagination normalization

- [ ] Page less than 1 becomes 1 or is rejected per current behavior
- [ ] Limit less than 1 or over 100 becomes 50 or is rejected per current behavior
- [ ] Repository receives normalized query

### A4.2 Handler query parsing compatibility

- [ ] `workDate=YYYY-MM-DD` maps to filter
- [ ] `teamId`, `jobTypeId`, `feederId` map to int64 filters
- [ ] Invalid numeric/date input behavior is captured before changing it

### A4.3 Response compatibility

- [ ] Fake usecase returns nested team/job/feeder/station names
- [ ] Handler returns same `dto.TaskResponse` JSON shape
- [ ] Meta includes page, limit, total

## A5 - Smoke script skeleton

**Objective:** Create a lightweight manual/CI script for representative API checks.

**Files:**

- Create: `scripts/smoke.sh`

**Task cards:**

- [ ] Check `GET /health`
- [ ] Login and capture access token if credentials are provided
- [ ] Hit `GET /v1/tasks?page=1&limit=1`
- [ ] Hit `GET /v1/dashboard/summary`
- [ ] Optional monthly plan status check when token exists
- [ ] Exit non-zero on HTTP >= 400

## Phase A acceptance

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
