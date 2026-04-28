# Backend Hotline Architecture

## Goal

Move `backend-hotline` from handler-heavy Gin/GORM code toward a modular, testable backend without breaking existing `/v1` API behavior. The refactor must preserve response shapes, filters, role behavior, and Cloudflare R2 upload flows.

## Current stack

- Go 1.24
- Gin HTTP router
- GORM + PostgreSQL
- Viper + `config.yaml`
- JWT auth
- Cloudflare R2/S3-compatible presigned URLs
- Middleware for recovery, timeout, rate limit, gzip, cache headers, and CORS

## Current repository shape

```text
main.go
cmd/
  migrate/
  fix-schema/
  fix-workdate/
  lowercase-columns/
  rename-column/
  test-read/
internal/
  adapter/out/persistence/gorm/task_repository.go
  app/task/usecase/list_tasks.go
  config/
  database/
  domain/task/entity.go
  dto/response.go
  handlers/v1/
  middleware/
  models/
  port/outbound/repository/task_repository.go
  router/router.go
pkg/
  jwt/
  password/
  s3/
```

## Target repository shape

Use module/vertical-slice structure incrementally. Do not big-bang move the whole repo.

```text
internal/
  domain/
    task/
    monthlyplan/
    user/
    masterdata/
  app/
    task/usecase/
    monthlyplan/usecase/
    dashboard/query/
    masterdata/usecase/
    user/usecase/
    auth/usecase/
  port/
    inbound/http/v1/          optional final home for handlers
    outbound/repository/
    outbound/storage/
  adapter/
    out/persistence/gorm/
    out/storage/r2/
  handlers/v1/                allowed during migration
  router/
  models/
```

## Layering rules

### Allowed

```text
router.SetupRouter -> handlers/controllers
handler/controller -> usecase/service interface
usecase/service    -> repository/storage interfaces + domain models
repository_impl    -> GORM models and SQL details
storage_impl       -> R2/S3 client details
```

### Forbidden

```text
new business rules in internal/router
new direct GORM queries in handlers once that endpoint has a usecase
usecase/service imports github.com/gin-gonic/gin
domain imports GORM, Gin, Viper, or AWS SDK
adapter/out imports handlers
repository interfaces return GORM models
```

## Module file contract

Each migrated domain should converge on:

```text
domain/<name>/entity.go
domain/<name>/errors.go
app/<name>/usecase/*.go
port/outbound/repository/<name>_repository.go
adapter/out/persistence/gorm/<name>_repository.go
handlers/v1/<name>.go
```

Optional:

```text
domain/<name>/policy.go
app/<name>/usecase/*_test.go
adapter/out/storage/r2/*.go
adapter/out/persistence/gorm/*_test.go
handlers/v1/*_test.go
```

## Data ownership

| Data | Owner |
|---|---|
| users/auth/session claims | `auth` and `user` |
| operation_centers, peas, stations, feeders | `masterdata` |
| teams | `team` or `masterdata/team` |
| job_types, job_details | `workcatalog` or `masterdata/job` |
| task_dailies | `task` |
| monthly_plans, monthly_plan_settings, plan_files, file_size_logs | `monthlyplan` |
| dashboard aggregates | `dashboard` query/read model |
| R2 object operations | `storage` adapter behind port |

## Domain rules

- `TaskDaily` is the main operational work record.
- Task list endpoints must keep pagination default `page=1`, `limit=50`, max `100`.
- Task filters must remain compatible with existing query names: `workDate`, `teamId`, `jobTypeId`, `feederId`, plus legacy/year-month endpoints.
- Monthly plan flow must keep presign and confirm as separate steps.
- Non-admin users must only access monthly plan files permitted by their team/role.
- Admin can manage settings, restore files, and hard delete files.
- Dashboard is read-model heavy; prefer query services over a thick domain model.
- Password hashes must never be exposed in responses.

## State transition architecture

```text
HTTP request
  -> handler parses path/query/body/auth context
  -> usecase validates input and role/team policy
  -> repository transaction reads/writes GORM models
  -> storage adapter signs or deletes R2 object when needed
  -> usecase returns domain/result DTO
  -> handler returns existing standard response envelope
```

## Testing strategy

- Usecase tests use fake repositories/storage.
- Handler tests use fake usecases and `httptest`.
- Repository tests use a real or sqlite-compatible DB strategy only where practical.
- Architecture tests inspect imports/source for boundary violations.
- Smoke tests verify core `/v1` flows after each major phase.

## Migration/refactor rules

- Preserve public `/v1` routes unless an explicit compatibility plan exists.
- Keep response envelopes compatible with `dto.StandardResponse`.
- Do not rename DB columns during architecture refactor tasks.
- Do not drop tables/columns without a separate migration and recovery plan.
- When replacing handler logic with a usecase, keep the old endpoint behavior covered by tests first.
