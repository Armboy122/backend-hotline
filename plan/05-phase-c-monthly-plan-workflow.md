# Phase C - Monthly Plan Workflow

## Goal

Extract the large monthly plan handler into testable usecases and ports while preserving the existing `/v1/monthly-plans*` API and Cloudflare R2 upload behavior.

## C1 - Domain and ports

**Objective:** Define monthly plan domain structures, repository contract, and storage contract.

**Files:**

- Create: `internal/domain/monthlyplan/entity.go`
- Create: `internal/domain/monthlyplan/errors.go`
- Create: `internal/domain/monthlyplan/policy.go`
- Create: `internal/port/outbound/repository/monthly_plan_repository.go`
- Create: `internal/port/outbound/storage/object_storage.go`

**Task cards:**

### C1.1 Domain entities

- [ ] `Period`
- [ ] `Settings`
- [ ] `PlanFile`
- [ ] `SubmissionStatus`
- [ ] `FileSizeLog`

### C1.2 Policy types

- [ ] Actor id
- [ ] Actor role
- [ ] Actor team id if available
- [ ] Admin override behavior

### C1.3 Repository interface

- [ ] Find or create period
- [ ] Get/update settings
- [ ] Create/list/update/delete plan files
- [ ] Restore/hard delete file
- [ ] Create file size log
- [ ] Submission status query

### C1.4 Storage interface

- [ ] Presign upload
- [ ] Presign download
- [ ] Delete object
- [ ] Object key generation policy stays outside handler

## C2 - Settings and period usecases

**Objective:** Move settings and period behavior out of `monthly_plan.go`.

**Files:**

- Create: `internal/app/monthlyplan/usecase/get_or_create_period.go`
- Create: `internal/app/monthlyplan/usecase/get_settings.go`
- Create: `internal/app/monthlyplan/usecase/update_settings.go`
- Create tests

**Task cards:**

- [ ] Validate year/month bounds
- [ ] Preserve default settings behavior
- [ ] Admin-only setting update policy is enforced by handler/middleware and tested at usecase boundary if actor is passed
- [ ] Response mapping matches existing DTO

## C3 - Upload usecases

**Objective:** Separate presign and confirm upload flows.

**Files:**

- Create: `internal/app/monthlyplan/usecase/presign_upload.go`
- Create: `internal/app/monthlyplan/usecase/confirm_upload.go`
- Create tests with fake repository and fake storage

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

## C4 - File lifecycle usecases

**Objective:** Extract list/status/download/delete/restore/hard-delete behavior.

**Files:**

- Create: `internal/app/monthlyplan/usecase/list_files.go`
- Create: `internal/app/monthlyplan/usecase/get_submission_status.go`
- Create: `internal/app/monthlyplan/usecase/get_download_url.go`
- Create: `internal/app/monthlyplan/usecase/soft_delete_file.go`
- Create: `internal/app/monthlyplan/usecase/restore_file.go`
- Create: `internal/app/monthlyplan/usecase/hard_delete_file.go`
- Create tests

**Task cards:**

- [ ] List files filters by period/team/role like current handler
- [ ] Submission status aggregates by team correctly
- [ ] Download URL enforces access before signing
- [ ] Soft delete does not remove R2 object
- [ ] Restore only allowed for admin
- [ ] Hard delete removes metadata and R2 object per current behavior

## C5 - GORM repository and R2 adapter

**Objective:** Move persistence/storage details behind ports.

**Files:**

- Create: `internal/adapter/out/persistence/gorm/monthly_plan_repository.go`
- Create: `internal/adapter/out/storage/r2/object_storage.go`
- Modify: `pkg/s3/r2.go` only if a thin wrapper is needed

**Task cards:**

- [ ] Reuse existing models from `internal/models`
- [ ] Keep column-name helpers from `internal/models/columns.go`
- [ ] Use transactions for confirm upload and hard delete metadata where needed
- [ ] Keep R2 config and SDK details outside usecases

## C6 - Refactor handler

**Objective:** Make `MonthlyPlanHandler` parse/map/delegate only.

**Files:**

- Modify: `internal/handlers/v1/monthly_plan.go`
- Modify: `internal/router/router.go` only for constructor wiring if needed

**Task cards:**

- [ ] Constructor accepts usecases or constructs adapters in one place consistently
- [ ] Handler keeps route-level auth context parsing
- [ ] Handler maps errors to existing envelope
- [ ] Existing route list remains unchanged

## Phase C acceptance

```bash
go test ./internal/app/monthlyplan/... -v
go test ./internal/adapter/out/persistence/gorm -run TestMonthlyPlan -v
go test ./internal/handlers/v1 -run TestMonthlyPlan -v
go test ./...
```
