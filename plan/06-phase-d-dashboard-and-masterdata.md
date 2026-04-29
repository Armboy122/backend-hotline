# Phase D - Dashboard and Master Data

> **Structure rule:** All code in this phase MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md). Each module lives under `internal/modules/<module>/` with `controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, and `entity.go` as needed. The pilot at `internal/modules/task/` is the reference. Do **not** create new files under `internal/domain/`, `internal/app/`, `internal/port/`, or `internal/adapter/out/` — put everything under `internal/modules/<module>/`.

## Goal

Reduce handler-level GORM queries for dashboard and master data after the core TaskDaily and MonthlyPlan slices are stable.

## D1 - Dashboard query service

**Objective:** Move dashboard aggregation queries behind a module-local query repository/service.

**Files:**

- Create: `internal/modules/dashboard/service.go`
- Create: `internal/modules/dashboard/repository.go`
- Create: `internal/modules/dashboard/repository_impl.go`
- Create: `internal/modules/dashboard/controller.go`
- Create: `internal/modules/dashboard/dto.go`
- Retire or thin: `internal/handlers/v1/dashboard.go` once module is routed

**Task cards:**

### D1.1 Define filter contract (in repository.go)

- [ ] `year`
- [ ] `month`
- [ ] `teamId`
- [ ] `jobTypeId`
- [ ] Optional date range if current endpoint supports it
- [ ] Validate and normalize input outside SQL builder

### D1.2 Extract summary/top jobs/top feeders

- [ ] Preserve current response DTOs
- [ ] Preserve cache strategy in router
- [ ] Preserve ordering and limit defaults

### D1.3 Extract feeder matrix/stats

- [ ] Keep existing concurrent query behavior if present
- [ ] Context must propagate to every DB call
- [ ] Avoid shared mutable state without synchronization

### D1.4 Tests

- [ ] Service tests with fake repository
- [ ] Controller tests with fake service
- [ ] Repository tests only for high-risk aggregation SQL

## D2 - Master data module pattern

**Objective:** Establish a repeatable pattern using one representative CRUD module before touching all master data.

Recommended first module: `jobtype`, because it is small and core to tasks.

**Files:**

- Create: `internal/modules/jobtype/controller.go`
- Create: `internal/modules/jobtype/service.go`
- Create: `internal/modules/jobtype/repository.go`
- Create: `internal/modules/jobtype/repository_impl.go`
- Create: `internal/modules/jobtype/dto.go`
- Create: `internal/modules/jobtype/errors.go`
- Retire or thin: `internal/handlers/v1/job_type.go` once module is routed

**Task cards:**

- [ ] List with existing count behavior
- [ ] Get by id
- [ ] Create validation
- [ ] Update field-specific updates
- [ ] Delete behavior preserved
- [ ] Controller only parses/maps/delegates

## D3 - Apply pattern to remaining master data in batches

**Objective:** Migrate repetitive CRUD safely without over-abstracting too early.

**Batch 1: work catalog**

Each module gets its own folder under `internal/modules/`:

- [ ] `internal/modules/jobdetail/` — job detail CRUD
- [ ] Restore behavior for job detail preserved
- [ ] Relationship to job type preserved

**Batch 2: organization hierarchy**

- [ ] `internal/modules/operationcenter/` — operation center CRUD
- [ ] `internal/modules/pea/` — PEA CRUD
- [ ] `internal/modules/station/` — station CRUD
- [ ] `internal/modules/feeder/` — feeder CRUD
- [ ] Nested response/count behavior preserved

**Batch 3: team**

- [ ] `internal/modules/team/` — team CRUD
- [ ] Task count behavior preserved
- [ ] Team relation to users/monthly plans preserved

## D4 - README/API docs cleanup

**Objective:** Align docs with actual `/v1` router behavior.

**Files:**

- Modify: `README.md`

**Task cards:**

- [ ] Replace stale `/api` examples with `/v1`
- [ ] Add auth/monthly-plan routes
- [ ] Add config and environment notes
- [ ] Add test/build commands
- [ ] Document current module-first architecture briefly

## Phase D acceptance

```bash
go test ./internal/modules/dashboard/... -v
go test ./internal/modules/jobtype/... -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
