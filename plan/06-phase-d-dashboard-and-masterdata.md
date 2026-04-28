# Phase D - Dashboard and Master Data

## Goal

Reduce handler-level GORM queries for dashboard and master data after the core TaskDaily and MonthlyPlan slices are stable.

## D1 - Dashboard query service

**Objective:** Move dashboard aggregation queries behind a query repository/service.

**Files:**

- Create: `internal/app/dashboard/query/service.go`
- Create: `internal/port/outbound/repository/dashboard_repository.go`
- Create: `internal/adapter/out/persistence/gorm/dashboard_repository.go`
- Modify: `internal/handlers/v1/dashboard.go`

**Task cards:**

### D1.1 Define filter contract

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

- [ ] Query service tests with fake repository
- [ ] Handler tests with fake service
- [ ] Repository tests only for high-risk aggregation SQL

## D2 - Master data module pattern

**Objective:** Establish a repeatable pattern using one representative CRUD module before touching all master data.

Recommended first module: `job_type`, because it is small and core to tasks.

**Files:**

- Create: `internal/domain/masterdata/jobtype.go`
- Create: `internal/app/masterdata/usecase/job_type_*.go`
- Create: `internal/port/outbound/repository/job_type_repository.go`
- Create: `internal/adapter/out/persistence/gorm/job_type_repository.go`
- Modify: `internal/handlers/v1/job_type.go`

**Task cards:**

- [ ] List with existing count behavior
- [ ] Get by id
- [ ] Create validation
- [ ] Update field-specific updates
- [ ] Delete behavior preserved
- [ ] Handler only parses/maps/delegates

## D3 - Apply pattern to remaining master data in batches

**Objective:** Migrate repetitive CRUD safely without over-abstracting too early.

**Batch 1: work catalog**

- [ ] `job_detail`
- [ ] Restore behavior for job detail preserved
- [ ] Relationship to job type preserved

**Batch 2: organization hierarchy**

- [ ] `operation_center`
- [ ] `pea`
- [ ] `station`
- [ ] `feeder`
- [ ] Nested response/count behavior preserved

**Batch 3: team**

- [ ] `team`
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
- [ ] Document current refactor architecture briefly

## Phase D acceptance

```bash
go test ./internal/app/dashboard/... -v
go test ./internal/app/masterdata/... -v
go test ./internal/handlers/v1 -run 'Test.*(Dashboard|JobType|JobDetail|Feeder|Station|PEA|OperationCenter|Team)' -v
go test ./...
```
