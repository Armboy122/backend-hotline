# Backend Hotline Architecture

## Goal

Move `backend-hotline` to a Hinghoi-style feature-first backend without breaking current `/v1` API behavior. Preserve response shapes, filters, role behavior, and Cloudflare R2 upload flows.

## Current Stack

- Go 1.24
- Gin HTTP router
- GORM + PostgreSQL
- Viper + `config.yaml`
- JWT auth
- Cloudflare R2/S3-compatible presigned URLs
- Recovery, timeout, rate limit, gzip, cache header, and CORS middleware

## Current Repository Shape

The repo is still mixed:

```text
main.go
internal/router/
internal/handlers/v1/
internal/modules/                  # deprecated transitional source
internal/domain/                   # legacy source during migration
internal/app/                      # legacy source during migration
internal/port/
internal/adapter/out/
internal/models/
internal/database/
pkg/jwt/
pkg/password/
pkg/s3/
```

## Target Repository Shape

Use the Hinghoi pattern:

```text
internal/feature/<feature>/
  controller/
    initiator.go
    v1.go
  service/
    initiator.go
    v1.go
  repository/
    initiator.go
    v1.go
  dto/
    dto.go
  entity/
    entity.go
  mapper/
    mapper.go
  helper/                 # optional
  constant/               # optional

internal/server/hotline_server/    # later phase
  server.go
  gin.go

pkg/db/                            # later phase
  db.go
  postgres.go
  models/
  migrations/
```

## Dependency Rules

Allowed:

```text
router/server -> feature controllers
controller    -> service interface + HTTP DTO mapping
service       -> repository interface + feature entity/policy
repository    -> GORM models, SQL details, storage clients
mapper        -> entity/DTO/model conversion
```

Forbidden:

```text
new business rules in internal/router
new feature code in internal/modules
new feature code in internal/domain/internal/app/internal/port/internal/adapter/out
service imports github.com/gin-gonic/gin
service imports gorm.io/gorm
controller imports gorm.io/gorm or internal/models
dto/entity/mapper imports Gin, GORM, Viper, or AWS SDK
repository interfaces return raw persistence models as API DTOs
```

Phase 0 exception:

- `internal/feature/task/controller` may delegate to the legacy `v1.TaskHandler` until M2 retires that bridge.
- Existing `internal/modules/*` and legacy layer-first packages may remain as migration sources.

## Feature Ownership

| Data / Flow | Target Feature |
|---|---|
| auth/login/refresh/me/logout | `internal/feature/auth` |
| user management | `internal/feature/user` |
| operation centers | `internal/feature/operationcenter` |
| PEAs | `internal/feature/pea` |
| stations | `internal/feature/station` |
| feeders | `internal/feature/feeder` |
| teams | `internal/feature/team` |
| job types | `internal/feature/jobtype` |
| job details | `internal/feature/jobdetail` |
| task_dailies | `internal/feature/task` |
| monthly plan settings/files/status | `internal/feature/monthlyplan` |
| planning calendar projection | `internal/feature/planningcalendar` |
| dashboard aggregates | `internal/feature/dashboard` |
| upload/R2 generic flow | `internal/feature/upload` or shared pkg adapter after review |


## Request Flow

```text
HTTP request
  -> controller parses path/query/body/auth context
  -> service validates input and role/team policy
  -> repository reads/writes GORM models and storage clients
  -> mapper converts model/entity/DTO shapes
  -> controller returns existing standard response envelope
```

## Domain Rules To Preserve

- `TaskDaily` pagination defaults stay `page=1`, `limit=50`, max `100`.
- Task filters keep query names: `workDate`, `teamId`, `jobTypeId`, `feederId`, plus existing year/month endpoints.
- Monthly plan presign and confirm remain separate steps.
- Non-admin users only access monthly plan files permitted by role/team.
- Admin can manage settings, restore files, and hard delete files.
- Dashboard stays read-model heavy; prefer repository query methods plus service orchestration.
- Password hashes must never be exposed.

## Testing Strategy

- Service tests use fake repositories/storage.
- Controller tests use fake services and `httptest`.
- Repository tests use real DB setup only for risky SQL behavior.
- Architecture tests inspect imports/source for boundary violations.
- Smoke tests verify representative `/v1` flows after each major phase.

## Migration Rules

- Preserve public `/v1` routes unless a separate compatibility plan exists.
- Keep `dto.StandardResponse` response envelope compatible.
- Do not rename DB columns during architecture-only refactors.
- Do not drop tables/columns without a separate migration and rollback plan.
- Move one feature at a time.
- Retire the old source only after the matching feature is green and routed through `internal/feature`.
