# Phase C - Monthly Plan Feature Migration

## Goal

Move monthly plan settings, period, upload, file lifecycle, and submission status behavior into `internal/feature/monthlyplan`.

Status: completed on 2026-04-29.

Previous `internal/modules/monthlyplan` and `internal/app/monthlyplan/usecase` code was used as migration source and retired.

## Target Files

```text
internal/feature/monthlyplan/controller/initiator.go
internal/feature/monthlyplan/controller/v1.go
internal/feature/monthlyplan/service/initiator.go
internal/feature/monthlyplan/service/v1.go
internal/feature/monthlyplan/repository/initiator.go
internal/feature/monthlyplan/repository/v1.go
internal/feature/monthlyplan/dto/dto.go
internal/feature/monthlyplan/entity/entity.go
internal/feature/monthlyplan/mapper/mapper.go
```

## Tasks

- [x] Define feature entity and policy types in `entity`
- [x] Define request/response DTOs in `dto`
- [x] Move HTTP parsing and error mapping into `controller`
- [x] Move settings and period behavior into `service`
- [x] Move presign and confirm upload behavior into `service`
- [x] Move list/status/download/delete/restore/hard-delete behavior into `service`
- [x] Move GORM and R2 persistence/storage into `repository`
- [x] Move conversion code into `mapper`
- [x] Retire monthly plan code from `internal/modules`, `internal/app`, `internal/port`, `internal/adapter/out`, and `internal/handlers/v1`

## Compatibility Requirements

- Existing `/v1/monthly-plans*` routes remain unchanged.
- R2 details are hidden behind repository/storage boundaries.
- Admin-only behavior remains enforced.
- Non-admin team restrictions remain tested.

## Acceptance

```bash
go test ./internal/feature/monthlyplan/... -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
