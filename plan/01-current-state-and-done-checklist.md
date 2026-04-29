# Current State and Done Checklist

## Verified baseline

- Repo: `backend-hotline`
- Module: `backend-hotlines3`
- Current API prefix: `/v1`
- Stack: Go 1.24, Gin, GORM/PostgreSQL, Viper, JWT, R2/S3 client
- Snapshot date: 2026-04-29
- Current architecture target: Hinghoi-style feature-first packages under `internal/feature/<feature>/...`
- Deprecated transitional source: `internal/modules/*`, `internal/domain/*`, `internal/app/*`, `internal/port/*`, `internal/adapter/out/*`, `internal/handlers/v1/*`
- Plan source docs:
  - `/Users/sakdithat/Downloads/hotline_prd.md`
  - `/Users/sakdithat/Downloads/hotline_domain_map.md`
  - `/Users/sakdithat/Downloads/backend_hotline_repo_evidence.md`

## Implemented routes observed

```text
GET    /health

POST   /v1/auth/login
POST   /v1/auth/register
POST   /v1/auth/refresh
POST   /v1/auth/logout
GET    /v1/auth/me

GET    /v1/teams
GET    /v1/teams/:id
POST   /v1/teams
PUT    /v1/teams/:id
DELETE /v1/teams/:id

GET    /v1/job-types
GET    /v1/job-types/:id
POST   /v1/job-types
PUT    /v1/job-types/:id
DELETE /v1/job-types/:id

GET    /v1/job-details
GET    /v1/job-details/:id
POST   /v1/job-details
PUT    /v1/job-details/:id
DELETE /v1/job-details/:id
POST   /v1/job-details/:id/restore

GET    /v1/feeders
GET    /v1/feeders/:id
POST   /v1/feeders
PUT    /v1/feeders/:id
DELETE /v1/feeders/:id

GET    /v1/stations
GET    /v1/stations/:id
POST   /v1/stations
PUT    /v1/stations/:id
DELETE /v1/stations/:id

GET    /v1/peas
GET    /v1/peas/:id
POST   /v1/peas
POST   /v1/peas/bulk
PUT    /v1/peas/:id
DELETE /v1/peas/:id

GET    /v1/operation-centers
GET    /v1/operation-centers/:id
POST   /v1/operation-centers
PUT    /v1/operation-centers/:id
DELETE /v1/operation-centers/:id

GET    /v1/tasks
GET    /v1/tasks/by-team
GET    /v1/tasks/by-filter
GET    /v1/tasks/:id
POST   /v1/tasks
PUT    /v1/tasks/:id
DELETE /v1/tasks/:id

POST   /v1/upload/image
DELETE /v1/upload/*key

GET    /v1/monthly-plans/settings
PUT    /v1/monthly-plans/settings
DELETE /v1/monthly-plans/files/:id/permanent
POST   /v1/monthly-plans/files/:id/restore
GET    /v1/monthly-plans/:year/:month
GET    /v1/monthly-plans/:year/:month/files
GET    /v1/monthly-plans/:year/:month/status
POST   /v1/monthly-plans/:year/:month/files/presign
POST   /v1/monthly-plans/:year/:month/files
DELETE /v1/monthly-plans/files/:id
GET    /v1/monthly-plans/files/:id/download

GET    /v1/dashboard/summary
GET    /v1/dashboard/top-jobs
GET    /v1/dashboard/top-feeders
GET    /v1/dashboard/feeder-matrix
GET    /v1/dashboard/stats

GET    /v1/users
GET    /v1/users/:id
POST   /v1/users
PUT    /v1/users/:id
DELETE /v1/users/:id
PUT    /v1/users/:id/password
```

## Completed checklist

### Foundation

- [x] Go module exists
- [x] Gin router exists
- [x] Config loader exists
- [x] PostgreSQL/GORM connection exists
- [x] AutoMigrate exists
- [x] Standard response envelope exists
- [x] JWT package exists
- [x] Password hashing package exists
- [x] R2/S3 package exists
- [x] Recovery middleware exists
- [x] Timeout middleware exists
- [x] Rate limit middleware exists
- [x] Gzip middleware exists
- [x] Cache header middleware exists
- [x] CORS middleware exists
- [x] Health endpoint exists

### Hinghoi-style Phase 0

- [x] Plan set now names `internal/feature/<feature>/...` as canonical
- [x] `internal/modules` direction is deprecated in the plan
- [x] `internal/feature/task/controller` exists
- [x] `internal/feature/task/service` exists
- [x] `internal/feature/task/repository` exists
- [x] `internal/feature/task/dto` exists
- [x] `internal/feature/task/entity` exists
- [x] `internal/feature/task/mapper` exists
- [x] `/v1/tasks*` routes are wired through the task feature entrypoint
- [x] Architecture tests guard `internal/feature`

### Transitional completed work

- [x] TaskDaily is fully routed and implemented under `internal/feature/task`
- [x] TaskDaily legacy handler/usecase/port/adapter/domain/module files were retired
- [x] Monthly plan is fully routed and implemented under `internal/feature/monthlyplan`
- [x] Monthly plan legacy handler/usecase/port/adapter/domain/module files were retired
- [x] R2 monthly plan storage is behind the feature repository/storage boundary
- [x] `internal/modules/dashboard` and `internal/modules/masterdata` entrypoint wrappers were retired
- [x] Dashboard routes are wired through `internal/feature/dashboard`
- [x] Master data routes are wired through `internal/feature/masterdata`
- [x] `/v1/job-types` is fully routed and implemented under `internal/feature/jobtype`
- [x] `/v1/teams` is fully routed and implemented under `internal/feature/team`
- [x] `/v1/stations` is fully routed and implemented under `internal/feature/station`
- [x] Auth routes are wired through `internal/feature/auth`
- [x] User routes are wired through `internal/feature/user`

### Retired transitional handlers

- [x] Auth handler logic was moved into `internal/feature/auth`
- [x] User handler logic was moved into `internal/feature/user`
- [x] Dashboard handler logic was moved into `internal/feature/dashboard`
- [x] Master data handler logic was moved into feature packages
- [x] Upload flow remains a separate legacy area for now, but the release/runbook notes cover it explicitly

## Known gaps / risks

| Gap | Impact | Preferred plan |
|---|---|---|
| Dashboard feature currently bridges to legacy handler for business logic | Feature route entrypoint is readable, but aggregation logic is still hard to unit test | Move queries into `internal/feature/dashboard/repository`, orchestration into `service`, mapping into `controller` |
| Master data feature currently bridges to legacy handlers | Feature route entrypoint is readable, but CRUD duplication remains | Split `jobtype` first, then repeat for the remaining reference data features |
| Auth/User features currently bridge to legacy handlers | Security behavior is preserved, but handler-heavy GORM code remains | Add focused tests, then extract auth/user services carefully |
| Dashboard aggregates live in handler | Hard to test filters/performance regressions | Extract query service/repository |
| Master data CRUD duplication | Repeated validation/error behavior | Establish a Hinghoi-style `jobtype` feature first |
| Auth/User are security-sensitive and handler-heavy | Password/JWT regressions possible | Add tests before extracting |
| DB connection ownership now lives in `pkg/db`; `internal/models` remains the deferred migration target | Connection wiring is centralized, but model relocation is still intentionally deferred | Keep `internal/models` stable for now and only move it when the next architecture phase is explicitly approved |

Resolved for Phase A: architecture boundary tests now exist, so dependency regressions are caught.

## Do not change yet

- Do not rename API prefix.
- Do not rename DB columns.
- Do not replace GORM globally.
- Do not move `internal/models` into `pkg/db` until feature migration is stable.
- Do not split every master data handler before Dashboard and the first master data reference feature are stable under `internal/feature`.
