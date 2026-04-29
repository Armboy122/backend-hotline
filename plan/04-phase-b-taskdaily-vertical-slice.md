# Phase B - TaskDaily Feature Migration

## Goal

Make TaskDaily fully owned by `internal/feature/task`, using the Hinghoi-style package split.

Phase 0 already routes `/v1/tasks*` through:

```text
internal/feature/task/repository
  -> internal/feature/task/service
     -> internal/feature/task/controller
```

The current controller still delegates to legacy handler/usecase/repository code. Phase B retires that bridge.

## Target Files

```text
internal/feature/task/controller/initiator.go
internal/feature/task/controller/v1.go
internal/feature/task/service/initiator.go
internal/feature/task/service/v1.go
internal/feature/task/repository/initiator.go
internal/feature/task/repository/v1.go
internal/feature/task/dto/dto.go
internal/feature/task/entity/entity.go
internal/feature/task/mapper/mapper.go
```

## Tasks

- [ ] Move path/query/body parsing from `internal/handlers/v1/task.go` into `controller/v1.go`
- [ ] Move list/get/create/update/delete behavior from legacy usecases into `service/v1.go`
- [ ] Move GORM persistence from `internal/adapter/out/persistence/gorm/task_repository.go` into `repository/v1.go`
- [ ] Own request/response DTOs under `dto`
- [ ] Own feature entity under `entity`
- [ ] Move model/entity/DTO conversion into `mapper`
- [ ] Delete or thin TaskDaily legacy code only after tests pass

## Compatibility Requirements

- `/v1/tasks`
- `/v1/tasks/by-team`
- `/v1/tasks/by-filter`
- `/v1/tasks/:id`
- pagination defaults: `page=1`, `limit=50`, max `100`
- soft delete behavior
- nested team/job/feeder/station response shape

## Acceptance

```bash
go test ./internal/feature/task/... -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
