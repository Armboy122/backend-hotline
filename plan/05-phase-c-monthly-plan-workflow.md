# Phase C - Monthly Plan Workflow

> **Structure rule:** All code in this phase MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md). Each module lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. The pilot at `internal/modules/task/` is the reference. Do **not** create new files under `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` for monthly plan — put everything under `internal/modules/monthlyplan/`.

## Goal

Extract the large monthly plan handler into testable services and repository ports while preserving the existing `/v1/monthly-plans*` API and Cloudflare R2 upload behavior.

## C1 - Domain entities, errors, and ports

**Objective:** Define monthly plan domain structures, repository contract, and storage contract inside the module.

**Files:**

- Create: `internal/modules/monthlyplan/entity.go`
- Create: `internal/modules/monthlyplan/errors.go`
- Create: `internal/modules/monthlyplan/repository.go`
- Create: `internal/modules/monthlyplan/dto.go`

**Task cards:**

### C1.1 Domain entities

- [ ] `Period`
- [ ] `Settings`
- [ ] `PlanFile`
- [ ] `SubmissionStatus`
- [ ] `FileSizeLog`

### C1.2 Policy types (in entity.go or errors.go)

- [ ] Actor id
- [ ] Actor role
- [ ] Actor team id if available
- [ ] Admin override behavior

### C1.3 Repository interface (in repository.go)

- [ ] Find or create period
- [ ] Get/update settings
- [ ] Create/list/update/delete plan files
- [ ] Restore/hard delete file
- [ ] Create file size log
- [ ] Submission status query

### C1.4 Storage interface (in repository.go or separate port)

- [ ] Presign upload
- [ ] Presign download
- [ ] Delete object
- [ ] Object key generation policy stays outside handler

## C2 - Settings and period services

**Objective:** Move settings and period behavior out of `monthly_plan.go`.

**Files:**

- Create: `internal/modules/monthlyplan/service.go` (settings/period methods)
- Create: `internal/modules/monthlyplan/service_test.go`

**Task cards:**

- [ ] Validate year/month bounds
- [ ] Preserve default settings behavior
- [ ] Admin-only setting update policy is enforced by controller/middleware and tested at service boundary if actor is passed
- [ ] Response mapping matches existing DTO

## C3 - Upload services

**Objective:** Separate presign and confirm upload flows.

**Files:**

- Create: `internal/modules/monthlyplan/service.go` (upload methods)
- Create: `internal/modules/monthlyplan/repository_impl.go` (GORM + R2 adapter)
- Create: `internal/modules/monthlyplan/service_test.go`

**Task cards:**

### C3.1 Presign upload

- [ ] Validate period exists or is created per current behavior
- [ ] Validate file name, size, content type
- [ ] Validate allowed file types from settings
- [ ] Enforce lock state
- [ ] Enforce actor role/team policy
- [ ] Return presigned PUT URL and object key

### C3.2 Confirm upload

- [ ] Validate object key belongs to expected period/team
- [ ] Create `PlanFile` metadata row
- [ ] Write `FileSizeLog`
- [ ] Preserve master plan vs team-specific behavior
- [ ] Return existing file response shape

## C4 - File lifecycle services

**Objective:** Extract list/status/download/delete/restore/hard-delete behavior.

**Files:**

- Extend: `internal/modules/monthlyplan/service.go`
- Extend: `internal/modules/monthlyplan/service_test.go`

**Task cards:**

- [ ] List files filters by period/team/role like current handler
- [ ] Submission status aggregates by team correctly
- [ ] Download URL enforces access before signing
- [ ] Soft delete does not remove R2 object
- [ ] Restore only allowed for admin
- [ ] Hard delete removes metadata and R2 object per current behavior

## C5 - GORM repository implementation and R2 adapter

**Objective:** Move persistence/storage details behind module-local ports.

**Files:**

- Create: `internal/modules/monthlyplan/repository_impl.go`
- Modify: `pkg/s3/r2.go` only if a thin wrapper is needed

**Task cards:**

- [ ] Reuse existing models from `internal/models`
- [ ] Keep column-name helpers from `internal/models/columns.go`
- [ ] Use transactions for confirm upload and hard delete metadata where needed
- [ ] Keep R2 config and SDK details outside service layer

## C6 - Refactor controller and router

**Objective:** Make `MonthlyPlanController` parse/map/delegate only.

**Files:**

- Create: `internal/modules/monthlyplan/controller.go`
- Modify: `internal/router/router.go` only for constructor wiring if needed
- Retire or thin: `internal/handlers/v1/monthly_plan.go` once module is routed

**Task cards:**

- [ ] Constructor accepts repository or constructs adapters in one place consistently
- [ ] Controller keeps route-level auth context parsing
- [ ] Controller maps errors to existing envelope
- [ ] Existing route list remains unchanged

## Phase C acceptance

```bash
go test ./internal/modules/monthlyplan/... -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
