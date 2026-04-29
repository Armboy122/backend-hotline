# Current State and Done Checklist

## Verified baseline

- Repo: `backend-hotline`
- Module: `backend-hotlines3`
- Current API prefix: `/v1`
- Stack: Go 1.24, Gin, GORM/PostgreSQL, Viper, JWT, R2/S3 client
- Snapshot date: 2026-04-28
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

### Partial architecture refactor

- [x] `TaskDaily` has a domain entity
- [x] `TaskDaily` has an outbound repository interface
- [x] `TaskDaily` has a GORM repository implementation for list
- [x] `TaskDaily` has a list usecase
- [x] `TaskHandler.List` uses the task list usecase
- [ ] `TaskHandler.GetByID` uses a usecase
- [ ] `TaskHandler.Create` uses a usecase
- [ ] `TaskHandler.Update` uses a usecase
- [ ] `TaskHandler.Delete` uses a usecase
- [ ] `TaskHandler.ListByTeam` uses a usecase/query service
- [x] `TaskHandler.List` delegates to `ListTasksUseCase` via port interface
- [x] `TaskHandler.GetByID` delegates to `GetTaskUseCase`
- [x] `TaskHandler.Create` delegates to `CreateTaskUseCase`
- [x] `TaskHandler.Update` delegates to `UpdateTaskUseCase`
- [x] `TaskHandler.Delete` delegates to `DeleteTaskUseCase`
- [x] `TaskHandler.ListByTeam` delegates to `ListTasksByTeamUseCase`
- [x] `TaskHandler.ListByFilter` uses a usecase/query service

### Handler-heavy areas

- [ ] Auth handler still directly uses GORM
- [ ] User handler still directly uses GORM
- [ ] Monthly plan handler still directly uses GORM and R2 client
- [ ] Dashboard handler still directly uses GORM aggregation queries
- [ ] Master data handlers still directly use GORM
- [ ] Upload handler directly owns R2 flow

## Known gaps / risks

| Gap | Impact | Preferred plan |
|---|---|---|
| Mixed architecture in TaskDaily | Future changes may duplicate rules across usecase and handler | Complete TaskDaily vertical slice first |
| Monthly plan handler is large and owns storage + policy + persistence | Hard to test role/team restrictions and file states | Extract usecases and storage port |
| Dashboard aggregates live in handler | Hard to test filters/performance regressions | Extract query service/repository |
| Master data CRUD duplication | Repeated validation/error behavior | Establish a small module pattern after core slices |
| Auth/User are security-sensitive and handler-heavy | Password/JWT regressions possible | Add tests before extracting |
| README documents `/api` while router uses `/v1` | Onboarding confusion | Update docs after route behavior is confirmed |
| No architecture boundary tests | Agents can add wrong dependencies unnoticed | Add in Phase A |

## Do not change yet

- Do not rename API prefix.
- Do not rename DB columns.
- Do not replace GORM globally.
- Do not split every master data handler before TaskDaily and MonthlyPlan are stable.
