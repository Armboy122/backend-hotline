# Execution Backlog

> **Structure rule:** All work from Phase A (M1) onward MUST follow the module-first vertical-slice layout defined in [`00-structure-reset.md`](./00-structure-reset.md). New code lives under `internal/modules/<module>/` (`controller.go`, `service.go`, `repository.go`, `repository_impl.go`, `dto.go`, `errors.go`, `entity.go`). The legacy `internal/domain/`, `internal/app/`, `internal/port/`, `internal/adapter/out/`, and `internal/handlers/v1/` paths are migration sources only — every milestone below ends with the affected feature owned by `internal/modules/<module>/`.

## Dependency graph

```text
0. Structure reset & module-first baseline             [DONE]
  -> A. Test + architecture hardening                 [TODO]
       -> B. TaskDaily vertical slice                 [TODO]
            -> C. Monthly plan workflow               [TODO]
                 -> D. Dashboard + master data pattern [TODO]
                      -> E. Auth/user + deploy hardening [TODO]
```

## Milestones

### S0 - Structure reset and module-first baseline [DONE]

- [x] S0.1 update plan index and backlog so the reset is the first prerequisite
- [x] S0.2 record the SCC-style preference in Obsidian
- [x] S0.3 create the module-first task pilot under `internal/modules/task/`
- [x] S0.4 wire `/v1/tasks` through the new module entrypoint
- [x] S0.5 add boundary tests for the module-first shape

Acceptance:

- The repo now has a module-first wiring shell for TaskDaily.
- Later phases can build on the new module shape instead of the old layer-first path.
- `go test ./...` still passes.

### M1 - Safe foundation [TODO]

- [ ] A1 architecture boundary tests
- [ ] A2 config and middleware tests
- [ ] A3 standard response/error mapping tests
- [ ] A4 TaskDaily list behavior tests around current usecase
- [ ] A5 smoke script skeleton for core `/v1` endpoints

Acceptance:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

### M2 - TaskDaily module complete [TODO]

> Target home: `internal/modules/task/`. Any remaining logic in `internal/app/task/usecase/`, `internal/port/outbound/repository/task_repository.go`, `internal/adapter/out/persistence/gorm/task_repository.go`, and `internal/handlers/v1/task.go` must be folded into the module by the end of M2.

- [ ] B1 `internal/modules/task/repository.go` interface covers get/create/update/delete/list-by-team/list-by-filter; `repository_impl.go` provides the GORM implementation
- [ ] B2 `internal/modules/task/service.go` exposes get/create/update/delete usecases
- [ ] B3 `internal/modules/task/controller.go` delegates all task endpoints to the service (router wires the module entrypoint only)
- [ ] B4 task validation and date/ID parsing locked by tests inside `internal/modules/task/`
- [ ] B5 response compatibility verified for nested team/job/feeder/station data via module-local DTOs

Acceptance:

- Existing `/v1/tasks*` routes still work.
- Pagination defaults remain `page=1`, `limit=50`, max `100`.
- Soft delete behavior is preserved.
- `go test ./...` passes.

### M3 - Monthly plan workflow modularized [TODO]

> Target home: `internal/modules/monthlyplan/`. Persistence and R2 storage are module-local adapters (`repository_impl.go` and a `storage_impl.go` or sibling file) so the monthly-plan flow lives in one folder.

- [ ] C1 define monthly plan entities and errors inside `internal/modules/monthlyplan/entity.go` and `errors.go`
- [ ] C2 define monthly plan repository/storage interfaces in `internal/modules/monthlyplan/repository.go`
- [ ] C3 add settings and period usecases on `internal/modules/monthlyplan/service.go`
- [ ] C4 add presign and confirm upload usecases on `internal/modules/monthlyplan/service.go` (R2 stays behind the module-local storage adapter)
- [ ] C5 add list/status/download/delete/restore/hard-delete usecases on `internal/modules/monthlyplan/service.go`
- [ ] C6 keep role/team policy tested inside `internal/modules/monthlyplan/`

Acceptance:

- Existing `/v1/monthly-plans*` routes remain compatible.
- R2 details are behind a storage port.
- Admin-only behavior remains enforced.
- Non-admin team restrictions are tested.

### M4 - Dashboard and master data cleaned [TODO]

> Target homes: `internal/modules/dashboard/` and one `internal/modules/<masterdata>/` per master data resource (`jobtype`, `jobdetail`, `team`, `feeder`, `station`, `pea`, `operationcenter`). No new code in `internal/handlers/v1/` or `internal/adapter/out/persistence/gorm/` for these features — migrate into modules.

- [ ] D1 dashboard module under `internal/modules/dashboard/` with `repository.go`/`repository_impl.go` for queries and `service.go` for aggregations
- [ ] D2 dashboard filter parsing tests inside `internal/modules/dashboard/`
- [ ] D3 extract master data pattern for one representative module under `internal/modules/jobtype/`
- [ ] D4 apply the same module-first pattern to remaining master data resources in small batches (`jobdetail`, `team`, `feeder`, `station`, `pea`, `operationcenter`)
- [ ] D5 update README route documentation from `/api` to `/v1`

Acceptance:

- Dashboard aggregation responses remain compatible.
- Master data CRUD behavior remains compatible.
- Handler direct GORM usage is reduced without broad generic abstractions.

### M5 - Auth/user/release hardening [TODO]

> Target homes: `internal/modules/auth/` and `internal/modules/user/`. Direct GORM/JWT calls in `internal/handlers/v1/auth.go` and `internal/handlers/v1/user.go` move into the module's `service.go` and `repository_impl.go`.

- [ ] E1 `internal/modules/auth/service.go` tests for login/refresh/me/logout
- [ ] E2 `internal/modules/user/service.go` tests for CRUD/change password
- [ ] E3 replace direct auth/user GORM logic with module services (delete old layer-first files once migrated)
- [ ] E4 production config validation and `.env.example` review
- [ ] E5 Docker/build/runbook/release checklist verified
- [ ] E6 smoke script runs a representative auth/task/monthly-plan/dashboard path

Acceptance:

- Password hashes are never exposed.
- Login updates `lastLogin` without `Save()` full-row side effects.
- Admin-only user management remains enforced.
- Release checklist is usable by another agent.

## Agent dispatch rule

Each unchecked checkbox above maps to one task assignment. Do not assign a whole milestone to one agent. Use [`99-agent-task-template.md`](./99-agent-task-template.md).
