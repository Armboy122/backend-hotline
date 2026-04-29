# Backend Hotline Hinghoi-Style Structure Reset

## Goal

Reshape `backend-hotline` toward the structure used by `/Users/sakdithat/Desktop/Devpool/hinghoi-backend`, while preserving the current public `/v1` API behavior.

This replaces the previous `internal/modules/<module>` direction. Existing `internal/modules/*` work is now treated as transitional source code only.

## Reference Shape

The preferred backend shape is feature-first:

```text
cmd/<app>/main.go                         # later phase, optional for this repo
internal/server/hotline_server/           # later phase server/router composition
  server.go
  gin.go

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
  helper/                                # optional
  constant/                              # optional

pkg/db/
  db.go
  postgres.go
  models/
  migrations/
```

Current repo-specific bridge rules:

- Keep `main.go`, `internal/router`, `internal/models`, and `internal/database` until a later phase moves server/db ownership.
- Keep all public routes under `/v1`.
- Keep response envelopes unchanged.
- Migrate one feature at a time into `internal/feature/<feature>/...`.
- Do not add new feature code under `internal/modules`, `internal/domain`, `internal/app`, `internal/port`, `internal/adapter/out`, or `internal/handlers/v1`.
- Legacy packages may remain only as migration sources until the matching feature is fully moved.

## Phase 0: Hinghoi Baseline

### 0.1 Update The Plan Set

Files:

- `plan/README.md`
- `plan/00-backend-architecture.md`
- `plan/01-current-state-and-done-checklist.md`
- `plan/02-execution-backlog.md`
- Phase files `03` through `07`

Acceptance:

- The plan names `internal/feature/<feature>/...` as canonical.
- The plan no longer presents `internal/modules/<module>` as the desired end-state.
- Completed old module work is documented as transitional source.

### 0.2 Create A Task Feature Pilot

Files:

- `internal/feature/task/controller/initiator.go`
- `internal/feature/task/controller/v1.go`
- `internal/feature/task/service/initiator.go`
- `internal/feature/task/repository/initiator.go`
- `internal/feature/task/dto/dto.go`
- `internal/feature/task/entity/entity.go`
- `internal/feature/task/mapper/mapper.go`
- `internal/router/router.go`

Acceptance:

- `/v1/tasks*` routes are wired through `internal/feature/task/controller`.
- The pilot can still delegate to legacy handler/usecase/repository code during Phase 0.
- No route behavior changes.

### 0.3 Update Architecture Guardrails

Files:

- `internal/architecture/feature_layout_test.go`
- `internal/architecture/architecture_test.go`

Checks:

- `internal/feature/task` scaffold exists.
- Router uses the task feature entrypoint.
- `service` packages do not import Gin, GORM, legacy handlers, or persistence models.
- `controller` packages do not import GORM or persistence models.
- `dto`, `entity`, and `mapper` packages stay framework-free.

## Phase 0 Acceptance

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

## End-State Acceptance

The whole migration is done when:

- Active features are navigable from `internal/feature/<feature>/...`.
- HTTP parsing lives in `controller`.
- Business behavior lives in `service`.
- Database access lives in `repository`.
- Mapping lives in `mapper`.
- DTOs and feature entities live in `dto` and `entity`.
- The old split across `internal/modules`, `internal/domain`, `internal/app`, `internal/port`, `internal/adapter/out`, and `internal/handlers/v1` has been retired feature by feature.
