# Execution Backlog

> **Structure rule:** new feature work belongs under `internal/feature/<feature>/...`, following the Hinghoi backend shape. Existing `internal/modules/*` code is transitional source only.

## Dependency Graph

```text
0. Hinghoi-style structure reset                  [DONE]
  -> A. Test + architecture hardening             [DONE]
       -> B. TaskDaily feature migration          [DONE]
            -> C. Monthly plan feature migration  [DONE]
                 -> D. Dashboard + master data    [DONE]
                      -> E. Auth/user/release     [DONE]
```

## S0 - Hinghoi-Style Structure Reset [DONE]

- [x] S0.1 update plan index/backlog to make Hinghoi-style `internal/feature` canonical
- [x] S0.2 document that `internal/modules` is deprecated transitional source
- [x] S0.3 create TaskDaily pilot under `internal/feature/task`
- [x] S0.4 wire `/v1/tasks` through `taskrepository -> taskservice -> taskcontroller`
- [x] S0.5 update architecture tests for `internal/feature`

Acceptance:

- `/v1/tasks*` routes are still compatible.
- `internal/feature/task` has controller/service/repository/dto/entity/mapper packages.
- `go test ./...`, `go vet ./...`, and `go build -o /tmp/hotlines-api main.go` pass.

## M1 - Safe Foundation [DONE]

- [x] A1 architecture boundary tests
- [x] A2 config and middleware tests
- [x] A3 standard response/error mapping tests
- [x] A4 TaskDaily list behavior tests around current usecase
- [x] A5 smoke script skeleton for core `/v1` endpoints

## M2 - TaskDaily Feature Migration [DONE]

Target home: `internal/feature/task/`.

- [x] B1 move TaskDaily HTTP parsing from `internal/handlers/v1/task.go` into `internal/feature/task/controller/v1.go`
- [x] B2 move TaskDaily business behavior from `internal/app/task/usecase/*` and `internal/modules/task/*` into `internal/feature/task/service`
- [x] B3 move TaskDaily persistence from `internal/adapter/out/persistence/gorm/task_repository.go` into `internal/feature/task/repository`
- [x] B4 move TaskDaily DTO/entity aliases into owned `dto`, `entity`, and `mapper` code
- [x] B5 retire TaskDaily imports from `internal/modules/task`, `internal/app/task`, `internal/port/outbound/repository/task_repository.go`, and `internal/handlers/v1/task.go`

Completed: 2026-04-29

Acceptance:

- Existing `/v1/tasks*` routes still work.
- Pagination defaults remain `page=1`, `limit=50`, max `100`.
- Soft delete behavior is preserved.
- Task feature tests live under `internal/feature/task/...`.

## M3 - Monthly Plan Feature Migration [DONE]

Target home: `internal/feature/monthlyplan/`.

- [x] C1 create `controller`, `service`, `repository`, `dto`, `entity`, and `mapper` packages
- [x] C2 move settings and period usecases into `service`
- [x] C3 move presign and confirm upload usecases into `service`
- [x] C4 move list/status/download/delete/restore/hard-delete behavior into `service`
- [x] C5 move GORM persistence and R2 storage adapters behind `repository`
- [x] C6 retire monthly-plan logic from `internal/modules/monthlyplan`, `internal/app/monthlyplan`, `internal/port`, `internal/adapter/out`, and `internal/handlers/v1/monthly_plan.go`

Completed: 2026-04-29

Acceptance:

- Existing `/v1/monthly-plans*` routes remain compatible.
- R2 details are behind the feature repository/storage boundary.
- Admin-only and team-scoped behavior remain tested.

## M4 - Dashboard And Master Data [DONE]

Target homes:

- `internal/feature/dashboard/`
- `internal/feature/jobtype/`
- `internal/feature/jobdetail/`
- `internal/feature/team/`
- `internal/feature/feeder/`
- `internal/feature/station/`
- `internal/feature/pea/`
- `internal/feature/operationcenter/`

Tasks:

- [x] D0 remove `internal/modules/dashboard` and `internal/modules/masterdata` route entrypoints
- [x] D0.1 wire dashboard routes through `internal/feature/dashboard`
- [x] D0.2 wire master data routes through `internal/feature/masterdata`
- [x] D1 move dashboard aggregation queries into `internal/feature/dashboard/repository`
- [x] D2 move dashboard orchestration into `internal/feature/dashboard/service`
- [x] D3 move dashboard HTTP parsing/mapping into `internal/feature/dashboard/controller`
- [x] D4 migrate `jobtype` first as the master data reference pattern
- [x] D5 apply the pattern to `jobdetail`, `feeder`, `pea`, and `operationcenter`
- [x] D6 update README route docs from `/api` to `/v1`

Completed: 2026-04-29

Acceptance:

- Dashboard aggregation responses remain compatible.
- Master data CRUD behavior remains compatible.
- Direct business GORM usage is removed from legacy handlers for migrated features.

## M5 - Auth/User/Server/Deploy Hardening [DONE]

Target homes:

- `internal/feature/auth/`
- `internal/feature/user/`
- later: `internal/server/hotline_server/`
- later: `pkg/db/`

Tasks:

- [x] E0 wire auth and user routes through `internal/feature/auth` and `internal/feature/user` entrypoints
- [x] E1 move login/refresh/logout/me into `internal/feature/auth`
- [x] E2 move user CRUD/change password into `internal/feature/user`
- [x] E3 review production config validation and `.env.example`
- [x] E4 move server/router composition toward `internal/server/hotline_server`
- [x] E5 evaluate moving DB models/connections toward `pkg/db` after feature migration is stable
- [x] E6 verify Docker/build/runbook/release checklist and smoke script

Completed: 2026-04-29

Acceptance:

- Password hashes are never exposed.
- Login updates `lastLogin` without full-row side effects.
- Admin-only user management remains enforced.
- Release checklist is usable by another agent.
