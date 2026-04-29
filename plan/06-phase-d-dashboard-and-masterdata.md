# Phase D - Dashboard And Master Data

## Goal

Move dashboard aggregation and master data CRUD behavior into Hinghoi-style feature folders.

## Dashboard Target

```text
internal/feature/dashboard/controller/
internal/feature/dashboard/service/
internal/feature/dashboard/repository/
internal/feature/dashboard/dto/
internal/feature/dashboard/entity/
internal/feature/dashboard/mapper/
```

Tasks:

- [x] Create `internal/feature/dashboard` entrypoint packages
- [x] Wire dashboard routes through `internal/feature/dashboard`
- [x] Retire `internal/modules/dashboard`
- [x] Define filter contract for `year`, `month`, `teamId`, `jobTypeId`, and endpoint-specific filters
- [x] Move summary/top jobs/top feeders queries into repository
- [x] Move feeder matrix/stats queries into repository
- [x] Move orchestration and defaults into service
- [x] Move HTTP parsing/mapping into controller
- [ ] Add service tests with fake repository

## Master Data Target

Migrate one feature at a time:

```text
internal/feature/jobtype/
internal/feature/jobdetail/
internal/feature/team/
internal/feature/feeder/
internal/feature/station/
internal/feature/pea/
internal/feature/operationcenter/
```

Tasks:

- [x] Create `internal/feature/masterdata` transitional entrypoint packages
- [x] Wire master data routes through `internal/feature/masterdata`
- [x] Retire `internal/modules/masterdata`
- [x] Migrate `jobtype` first as the reference pattern
- [ ] Preserve list/get/create/update/delete behavior
- [ ] Preserve count/nested response behavior
- [ ] Preserve job detail restore behavior
- [x] Apply the same package split to `team`
- [x] Apply the same package split to `station`
- [x] Apply the same package split to remaining master data features: `jobdetail`, `feeder`, `pea`, and `operationcenter`
- [x] Update README route docs from `/api` to `/v1`

## Acceptance

```bash
go test ./internal/feature/dashboard/... -v
go test ./internal/feature/jobtype/... -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
